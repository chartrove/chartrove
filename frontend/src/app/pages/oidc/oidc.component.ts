import { Component, OnInit, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule, Location } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink, ActivatedRoute } from '@angular/router';
import { OIDCService, OIDCIdentity, OIDCAudit } from '../../services/oidc.service';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-oidc', standalone: true, imports: [CommonModule, FormsModule, RouterLink],
  changeDetection: ChangeDetectionStrategy.Eager,
  template: `
    <main class="container py-4">
      <h1>Identity provider sign-in</h1>
      <p>Local password sign-in remains available. Linking changes how you sign in; it does not share medical records.</p>
      @if (message) { <p role="status">{{message}}</p> }
      @if (enabled && !busy) {
        @if (authenticated) {
          <form (ngSubmit)="link()">
            <label for="oidc-password">Confirm your local password to link an identity</label>
            <input id="oidc-password" name="password" type="password" autocomplete="current-password" [(ngModel)]="password" required class="form-control">
            <button type="submit" [disabled]="!password" class="btn btn-primary my-2">Link identity provider</button>
          </form>
        } @else {
          <p>Use an identity previously linked from your local Chartrove account settings.</p>
          <button type="button" class="btn btn-primary" (click)="login()">Continue to identity provider</button>
        }
      }
      @for (identity of identities; track identity.id) {
        <section class="border p-3 my-2"><p>{{identity.issuer}}</p>
          @if (identity.revoked_at) { <p>Revoked</p> } @else {
            <button type="button" class="btn btn-outline-danger" [disabled]="busy || !password" (click)="revoke(identity.id)">Revoke identity and its sessions</button>
          }
        </section>
      }
      @if (audit.length) { <h2>Recent sign-in history</h2>
        <ul>@for (event of audit; track event.id) { <li>{{event.created_at | date:'medium'}} — {{event.action}} — {{event.identity_id}}</li> }</ul>
      }
      <a routerLink="/auth/signin">Local sign-in</a> · <a routerLink="/settings">Settings</a>
    </main>`
})
export class OIDCComponent implements OnInit {
  enabled = false; authenticated = false; busy = false; password = ''; message = '';
  identities: OIDCIdentity[] = []; audit: OIDCAudit[] = [];
  constructor(private oidc: OIDCService, private auth: AuthService, private router: Router, private route: ActivatedRoute, private location: Location) {}
  async ngOnInit() {
    if (this.route.snapshot.routeConfig?.path === 'auth/oidc/callback') {
      const params = new URLSearchParams(globalThis.location.search);
      this.location.replaceState(this.router.serializeUrl(this.router.createUrlTree(['/auth/oidc'])));
      this.busy = true;
      try { await this.oidc.complete(params); await this.router.navigateByUrl('/dashboard') }
      catch { this.message = 'Identity sign-in failed. Use local sign-in or start again.' }
      finally { this.busy = false }
      return;
    }
    this.enabled = await this.oidc.enabled();
    this.authenticated = await this.auth.IsAuthenticated();
    if (this.authenticated) await this.refresh();
    if (!this.enabled) this.message = 'Identity provider sign-in is unavailable. Local sign-in remains available.';
  }
  private async refresh() { try { [this.identities, this.audit] = await Promise.all([this.oidc.identities(), this.oidc.audit()]) } catch { this.message = 'Unable to load identity settings.' } }
  async login() { await this.begin() }
  async link() { const password = this.password; this.password = ''; await this.begin(password) }
  private async begin(password?: string) {
    this.busy = true;
    try { globalThis.location.assign(await this.oidc.start(password)) }
    catch { this.message = 'Unable to start identity sign-in. Check your password or use local sign-in.'; this.busy = false }
  }
  async revoke(id: string) {
    const password = this.password; this.password = ''; this.busy = true;
    try { await this.oidc.revoke(id, password); this.message = 'Identity and its sessions revoked. Use local sign-in to continue.'; this.identities = []; this.audit = [] }
    catch { this.message = 'Revocation was not confirmed. Check your local password and try again.' }
    finally { this.busy = false }
  }
}
