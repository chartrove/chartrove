import {HttpClient} from '@angular/common/http';
import {ChangeDetectionStrategy, Component, DoCheck, Inject, OnDestroy, OnInit, TemplateRef} from '@angular/core';
import {DomSanitizer, SafeUrl} from '@angular/platform-browser';
import {NgbModal, NgbModalRef} from '@ng-bootstrap/ng-bootstrap';
import QRCode from 'qrcode';
import {fromEvent, Subject} from 'rxjs';
import {finalize, takeUntil, timeout} from 'rxjs/operators';
import {environment} from '../../../environments/environment';
import {GetEndpointAbsolutePath} from '../../../lib/utils/endpoint_absolute_path';
import {HTTP_CLIENT_TOKEN} from '../../dependency-injection';
import {AuthService} from '../../services/auth.service';

interface DeviceTokenView {
  tokenId: string;
  name: string;
  issuedAt: string;
  expiresAt: string;
}

interface SettingsUser {
  username: string;
  full_name?: string;
  email?: string;
  role?: string;
}

interface ServerDiscovery {
  server_base_urls: string[];
  sync_endpoint: string;
}

function record(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function successfulData(response: unknown): unknown {
  return record(response) && response['success'] === true ? response['data'] : undefined;
}

const issuanceUncertain = 'Could not confirm a new device token. Review saved devices before retrying.';
const pairingUncertain = 'Could not prepare the device connection. Review saved devices before retrying.';

@Component({
  selector: 'app-settings',
  templateUrl: './settings.component.html',
  styleUrls: ['./settings.component.scss'],
  changeDetection: ChangeDetectionStrategy.Eager,
  standalone: false
})
export class SettingsComponent implements OnInit, OnDestroy, DoCheck {
  readonly localDocumentImport = !environment.environment_desktop && !environment.environment_cloud;
  tokens: DeviceTokenView[] = [];
  qrCodeUrl: SafeUrl | null = null;
  qrCodeData = '';
  isLoading = false;
  hasError = false;
  isRawQrCodeCollapsed = true;
  errorMessage = '';
  accessToken = '';
  serverInfo: ServerDiscovery | null = null;
  currentUser: SettingsUser | null = null;
  newDeviceName = '';
  newDeviceExpiration = 0;
  readonly expirationOptions = [
    {value: 0, label: 'No Expiration'},
    {value: 7, label: '7 Days'},
    {value: 30, label: '30 Days'},
    {value: 60, label: '60 Days'},
    {value: 90, label: '90 Days'},
  ];
  step: 'askDetails' | 'showQR' = 'askDetails';

  private readonly stopped = new Subject<void>();
  private readonly pairingStopped = new Subject<void>();
  private destroyed = false;
  private profileChanged = false;
  private ownerID: string | null = null;
  private generation = 0;
  private modal: NgbModalRef | null = null;

  constructor(
    @Inject(HTTP_CLIENT_TOKEN) private http: HttpClient,
    private sanitizer: DomSanitizer,
    private modalService: NgbModal,
    private auth: AuthService,
  ) {}

  ngOnInit(): void {
    this.ownerID = this.readOwner();
    if (!this.ownerID) {
      this.setError('An active local profile is required. Sign in again before connecting a device.');
      return;
    }
    fromEvent<StorageEvent>(window, 'storage').pipe(takeUntil(this.stopped)).subscribe(event => {
      if ((event.key === this.auth.FASTEN_JWT_LOCALSTORAGE_KEY || event.key === null) && this.readOwner() !== this.ownerID) this.discardProfile();
    });
    this.loadCurrentUser();
    this.loadTokens();
  }

  ngDoCheck(): void {
    if (!this.destroyed && !this.profileChanged && this.ownerID && this.readOwner() !== this.ownerID) {
      this.discardProfile();
    }
  }

  ngOnDestroy(): void {
    this.destroyed = true;
    this.cancelPairing();
    this.stopped.next();
    this.stopped.complete();
    this.pairingStopped.complete();
    const modal = this.modal;
    this.modal = null;
    modal?.dismiss('settings closed');
    this.tokens = [];
    this.currentUser = null;
  }

  private endpoint(path: string): string {
    return `${GetEndpointAbsolutePath(globalThis.location, environment.fasten_api_endpoint_base)}/secure/${path}`;
  }

  // Browser UUID is only a stale-context aid. Signed server requests and the
  // UUID-bound token row remain authority; this does not authenticate claims.
  private readOwner(): string | null {
    try {
      const id = this.auth.GetCurrentUser().native_user_id;
      if (typeof id !== 'string' || !/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(id) || /^0{8}-0{4}-0{4}-0{4}-0{12}$/.test(id)) return null;
      return id.toLowerCase();
    } catch {
      return null;
    }
  }

  private active(): boolean {
    if (this.destroyed || this.profileChanged || !this.ownerID) return false;
    if (this.readOwner() !== this.ownerID) {
      this.discardProfile();
      return false;
    }
    return true;
  }

  private currentPairing(generation: number): boolean {
    return generation === this.generation && this.modal !== null && this.active();
  }

  private clearCredential(): void {
    this.accessToken = '';
    this.qrCodeData = '';
    this.qrCodeUrl = null;
    this.serverInfo = null;
    this.step = 'askDetails';
    this.isRawQrCodeCollapsed = true;
  }

  private cancelPairing(): void {
    this.generation++;
    this.pairingStopped.next();
    this.clearCredential();
    this.isLoading = false;
  }

  private discardProfile(): void {
    this.profileChanged = true;
    this.cancelPairing();
    this.stopped.next();
    const modal = this.modal;
    this.modal = null;
    modal?.dismiss('profile changed');
    this.currentUser = null;
    this.tokens = [];
    this.setError('Profile changed. Any pending token may already exist for the previous profile. Review its saved devices. Leave Settings and reopen it for the active profile.');
  }

  loadCurrentUser(): void {
    if (!this.active()) return;
    this.http.get<unknown>(this.endpoint('account/me')).pipe(timeout(30000), takeUntil(this.stopped)).subscribe({
      next: response => {
        if (!this.active()) return;
        const user = successfulData(response);
        if (!record(user) || typeof user['id'] !== 'string' || user['id'].toLowerCase() !== this.ownerID || typeof user['username'] !== 'string') {
          this.setError('The active profile could not be confirmed.');
          return;
        }
        this.currentUser = {
          username: user['username'],
          full_name: typeof user['full_name'] === 'string' ? user['full_name'] : undefined,
          email: typeof user['email'] === 'string' ? user['email'] : undefined,
          role: typeof user['role'] === 'string' ? user['role'] : undefined,
        };
      },
      error: () => { if (this.active()) this.setError('The active profile could not be loaded.'); },
    });
  }

  generateAccessToken(): void {
    if (this.isLoading || !this.modal || !this.active() || !this.currentUser) return;
    this.cancelPairing();
    const generation = this.generation;
    this.hasError = false;
    this.errorMessage = '';
    this.isLoading = true;
    const body = {name: this.newDeviceName, expiration: this.newDeviceExpiration};
    this.http.post<unknown>(this.endpoint('access/token'), body).pipe(
      timeout(30000), takeUntil(this.stopped), takeUntil(this.pairingStopped),
      finalize(() => { if (generation === this.generation) this.isLoading = false; }),
    ).subscribe({
      next: response => {
        if (!this.currentPairing(generation)) return;
        const token = successfulData(response);
        if (typeof token !== 'string' || !token || token.length > 16384) {
          this.clearCredential();
          this.setError(issuanceUncertain);
          return;
        }
        this.accessToken = token;
        this.newDeviceName = '';
        this.step = 'showQR';
        this.loadTokens();
        this.getServerDiscovery();
      },
      error: () => {
        if (!this.currentPairing(generation)) return;
        this.clearCredential();
        this.setError(issuanceUncertain);
      },
    });
  }

  generateAndShowQR(content: TemplateRef<unknown>): void {
    if (!this.active() || !this.currentUser) return;
    const previous = this.modal;
    this.modal = null;
    previous?.dismiss('new pairing');
    this.cancelPairing();
    this.hasError = false;
    this.errorMessage = '';
    this.newDeviceName = '';
    this.newDeviceExpiration = 0;
    const modal = this.modalService.open(content, {size: 'lg', centered: true, backdrop: 'static'});
    this.modal = modal;
    const finished = () => {
      if (this.modal !== modal) return;
      this.modal = null;
      this.cancelPairing();
    };
    void modal.result.then(finished, finished);
  }

  connectDevice(): void {
    this.generateAccessToken();
  }

  getServerDiscovery(): void {
    const generation = this.generation;
    if (!this.currentPairing(generation) || !this.accessToken) return;
    this.http.get<unknown>(this.endpoint('sync/discovery')).pipe(timeout(30000), takeUntil(this.stopped), takeUntil(this.pairingStopped)).subscribe({
      next: response => {
        if (!this.currentPairing(generation)) return;
        const info = successfulData(response);
        const rawURLs: unknown = record(info) ? info['server_base_urls'] : undefined;
        const urls: unknown[] = Array.isArray(rawURLs) ? rawURLs : [];
        const stringURLs = urls.filter((url): url is string => typeof url === 'string');
        if (!record(info) || !urls.length || stringURLs.length !== urls.length || !stringURLs.every(url => url.length > 0 && url.length <= 2048) || typeof info['sync_endpoint'] !== 'string' || !info['sync_endpoint'] || info['sync_endpoint'].length > 2048) {
          this.clearCredential();
          this.setError(pairingUncertain);
          return;
        }
        this.serverInfo = {server_base_urls: stringURLs, sync_endpoint: info['sync_endpoint']};
        this.generateQRCode();
      },
      error: () => {
        if (!this.currentPairing(generation)) return;
        this.clearCredential();
        this.setError(pairingUncertain);
      },
    });
  }

  generateQRCode(): void {
    const generation = this.generation;
    if (!this.currentPairing(generation) || !this.accessToken || !this.serverInfo) return;
    const token = this.accessToken;
    this.qrCodeData = JSON.stringify({token, server_base_urls: this.serverInfo.server_base_urls, sync_endpoint: this.serverInfo.sync_endpoint}, null, 2);
    try {
      void QRCode.toDataURL(this.qrCodeData, {errorCorrectionLevel: 'M', type: 'image/png', margin: 1, color: {dark: '#000000', light: '#FFFFFF'}}).then(url => {
        if (!this.currentPairing(generation) || this.accessToken !== token) return;
        if (!url.startsWith('data:image/png;base64,') || url.length > 2*1024*1024) {
          this.clearCredential();
          this.setError(pairingUncertain);
          return;
        }
        this.qrCodeUrl = this.sanitizer.bypassSecurityTrustUrl(url);
      }).catch(() => {
        if (!this.currentPairing(generation)) return;
        this.clearCredential();
        this.setError(pairingUncertain);
      });
    } catch {
      if (!this.currentPairing(generation)) return;
      this.clearCredential();
      this.setError(pairingUncertain);
    }
  }

  loadTokens(): void {
    if (!this.active()) return;
    this.http.get<unknown>(this.endpoint('access/token')).pipe(timeout(30000), takeUntil(this.stopped)).subscribe({
      next: response => {
        if (!this.active()) return;
        const rows = successfulData(response);
        if (!Array.isArray(rows)) {
          this.tokens = [];
          this.setError('Saved devices could not be loaded.');
          return;
        }
        const input: unknown[] = rows;
        const tokens: DeviceTokenView[] = [];
        for (const row of input) {
          if (!record(row) || typeof row['token_id'] !== 'string' || typeof row['issued_at'] !== 'string' || typeof row['expires_at'] !== 'string') {
            this.tokens = [];
            this.setError('Saved devices could not be confirmed.');
            return;
          }
          tokens.push({tokenId: row['token_id'], name: typeof row['name'] === 'string' ? row['name'] : '', issuedAt: row['issued_at'], expiresAt: row['expires_at']});
        }
        this.tokens = tokens;
      },
      error: () => {
        if (!this.active()) return;
        this.tokens = [];
        this.setError('Saved devices could not be loaded.');
      },
    });
  }

  deleteToken(tokenId: string): void {
    if (!this.active() || !confirm('Are you sure you want to delete this access token?') || !this.active()) return;
    this.http.delete<unknown>(this.endpoint('access/token'), {body: {token_id: tokenId}}).pipe(timeout(30000), takeUntil(this.stopped)).subscribe({
      next: response => {
        if (!this.active()) return;
        if (!record(response) || response['success'] !== true) {
          this.setError('Device removal was not confirmed. Review saved devices before retrying.');
          return;
        }
        this.cancelPairing();
        this.loadTokens();
      },
      error: () => { if (this.active()) this.setError('Device removal was not confirmed. Review saved devices before retrying.'); },
    });
  }

  formatDate(dateString: string | Date): string {
    return dateString ? new Date(dateString).toLocaleDateString() : 'N/A';
  }

  formatExpiryDate(dateString: string | Date): string {
    if (!dateString) return 'N/A';
    return `on ${new Date(dateString).toLocaleDateString('en-US', {weekday: 'short', month: 'short', day: '2-digit', year: 'numeric'})}`;
  }

  copyRawQR(): void {
    const generation = this.generation;
    if (!this.currentPairing(generation) || !this.qrCodeData) return;
    const failed = () => { if (this.currentPairing(generation)) this.setError('Connection details could not be copied.'); };
    try {
      if (!navigator.clipboard?.writeText) {
        failed();
        return;
      }
      void navigator.clipboard.writeText(this.qrCodeData).catch(failed);
    } catch {
      failed();
    }
  }

  openQRModal(content: TemplateRef<unknown>): void {
    this.generateAndShowQR(content);
  }

  private setError(message: string): void {
    this.hasError = true;
    this.errorMessage = message;
  }
}
