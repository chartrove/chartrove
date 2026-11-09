import {HttpTestingController, provideHttpClientTesting} from '@angular/common/http/testing';
import {ComponentFixture, fakeAsync, flushMicrotasks, TestBed} from '@angular/core/testing';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import QRCode from 'qrcode';
import {AppModule} from '../../app.module';
import {AuthService} from '../../services/auth.service';
import {SettingsComponent} from './settings.component';

interface OwnedModal {
  readonly result: Promise<unknown>;
  readonly dismiss: jasmine.Spy;
  readonly close: jasmine.Spy;
  finish(): void;
}

class SettingsModalProbe {
  readonly refs: OwnedModal[] = [];
  open(): OwnedModal {
    let resolve: (value: unknown) => void = () => undefined;
    const result = new Promise<unknown>(done => { resolve = done; });
    const finish = () => resolve(undefined);
    const ref = {result, finish, dismiss: jasmine.createSpy('owned modal dismiss').and.callFake(finish), close: jasmine.createSpy('owned modal close').and.callFake(finish)};
    this.refs.push(ref);
    return ref;
  }
  current(): OwnedModal {
    const ref = this.refs[this.refs.length - 1];
    if (!ref) throw new Error('Expected an owned pairing modal');
    return ref;
  }
}

// Actual component/template/registered application HTTP providers, mocked only at
// HTTP, modal and QR encoder boundaries. Browser UUID claims are stale-context
// aids here, not successful server authentication or real mobile/visual QA.
describe('Settings companion credential privacy', () => {
  const owner = '11111111-1111-4111-8111-111111111111';
  const replacement = '22222222-2222-4222-8222-222222222222';
  const credential = 'synthetic-private-device-credential-do-not-log';
  let fixture: ComponentFixture<SettingsComponent>;
  let component: SettingsComponent;
  let http: HttpTestingController;
  let auth: AuthService;
  let modals: SettingsModalProbe;
  let qr: jasmine.Spy;
  let resolveQR: (value: string) => void;
  let rejectQR: (reason: unknown) => void;
  let consoleSpies: jasmine.Spy[];

  function session(id: string): string {
    return btoa(JSON.stringify({alg: 'HS256'})) + '.' + btoa(JSON.stringify({sub: 'synthetic-profile', native_user_id: id, exp: 4102444800})) + '.synthetic';
  }

  beforeEach(async () => {
    modals = new SettingsModalProbe();
    await TestBed.configureTestingModule({imports: [AppModule], providers: [provideHttpClientTesting(), {provide: NgbModal, useValue: modals}]}).compileComponents();
    auth = TestBed.inject(AuthService);
    auth.setAuthToken(session(owner));
    consoleSpies = [spyOn(console, 'log'), spyOn(console, 'error'), spyOn(console, 'warn'), spyOn(console, 'debug'), spyOn(console, 'info')];
    const pending = new Promise<string>((resolve, reject) => { resolveQR = resolve; rejectQR = reject; });
    qr = spyOn(QRCode, 'toDataURL').and.callFake(() => pending);
    http = TestBed.inject(HttpTestingController);
    fixture = TestBed.createComponent(SettingsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
    http.expectOne(request => request.url.endsWith('/secure/account/me')).flush({success: true, data: {id: owner, username: 'synthetic-owner', role: 'user'}});
    http.expectOne(request => request.url.endsWith('/secure/access/token')).flush({success: true, data: []});
    fixture.detectChanges();
  });

  afterEach(() => {
    fixture.destroy();
    http.verify({ignoreCancelled: true});
    localStorage.removeItem('token');
  });

  function open(): OwnedModal {
    const button: HTMLButtonElement = fixture.nativeElement.querySelector('.card-header button');
    button.click();
    fixture.detectChanges();
    return modals.current();
  }

  function issued(): void {
    http.expectOne(request => request.method === 'POST' && request.url.endsWith('/secure/access/token')).flush({success: true, data: credential});
    http.expectOne(request => request.method === 'GET' && request.url.endsWith('/secure/access/token')).flush({success: true, data: []});
    http.expectOne(request => request.url.endsWith('/secure/sync/discovery')).flush({success: true, data: {server_base_urls: ['https://synthetic.example.invalid'], sync_endpoint: 'api/secure/resource/fhir'}});
  }

  function expectCleared(): void {
    expect(component.accessToken).toBe('');
    expect(component.qrCodeData).toBe('');
    expect(component.qrCodeUrl).toBeNull();
    expect(component.step).toBe('askDetails');
  }

  it('keeps explicit successful QR handoff but never logs its credential response', fakeAsync(() => {
    open();
    component.connectDevice();
    issued();
    resolveQR('data:image/png;base64,synthetic');
    flushMicrotasks();
    expect(component.step).toBe('showQR');
    expect(component.qrCodeData).toContain(credential);
    expect(qr).toHaveBeenCalled();
    expect(JSON.stringify(consoleSpies.flatMap(spy => spy.calls.allArgs()))).not.toContain(credential);
  }));

  it('clears prior secrets at a new attempt and handles actual intercepted500 without false success', fakeAsync(() => {
    open();
    component.accessToken = 'synthetic-prior-private-credential';
    component.qrCodeData = 'synthetic-prior-private-qr';
    component.step = 'showQR';
    component.connectDevice();
    expectCleared();
    http.expectOne(request => request.method === 'POST').flush({success: false, error: 'synthetic-private-error-body'}, {status: 500, statusText: 'Synthetic storage unavailable'});
    flushMicrotasks();
    expectCleared();
    expect(component.isLoading).toBeFalse();
    expect(component.errorMessage).toContain('Review saved devices');
    expect(component.errorMessage).not.toContain('synthetic-private-error-body');
    expect(qr).not.toHaveBeenCalled();
    http.expectNone(request => request.url.endsWith('/secure/sync/discovery'));
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain('Review saved devices');
  }));

  it('refuses malformed successful credential data without discovery or QR emission', () => {
    open();
    component.connectDevice();
    http.expectOne(request => request.method === 'POST').flush({success: true, data: {private: credential}});
    expectCleared();
    expect(component.hasError).toBeTrue();
    expect(qr).not.toHaveBeenCalled();
    http.expectNone(request => request.method === 'GET');
  });

  it('closing its modal cancels pending client emission and clears secrets, not server persistence', fakeAsync(() => {
    const modal = open();
    component.connectDevice();
    const request = http.expectOne(request => request.method === 'POST');
    modal.finish();
    flushMicrotasks();
    expect(request.cancelled).toBeTrue();
    expectCleared();
    expect(component.isLoading).toBeFalse();
  }));

  it('cannot repopulate a closed modal from a late QR encoder promise', fakeAsync(() => {
    const modal = open();
    component.connectDevice();
    issued();
    modal.finish();
    flushMicrotasks();
    resolveQR('data:image/png;base64,synthetic-private-late-qr');
    flushMicrotasks();
    expectCleared();
  }));

  it('rejects a replacement profile before accepting an old in-flight token response', fakeAsync(() => {
    open();
    component.connectDevice();
    const request = http.expectOne(request => request.method === 'POST');
    auth.setAuthToken(session(replacement));
    fixture.detectChanges();
    expect(request.cancelled).toBeTrue();
    expectCleared();
    expect(modals.current().dismiss).toHaveBeenCalled();
    expect(component.errorMessage).toContain('previous profile');
    expect(qr).not.toHaveBeenCalled();
    http.expectNone(request => request.method === 'GET');
    flushMicrotasks();
  }));

  it('clears an already prepared credential on profile replacement without inferring revocation', fakeAsync(() => {
    open();
    component.connectDevice();
    issued();
    resolveQR('data:image/png;base64,synthetic');
    flushMicrotasks();
    expect(component.qrCodeData).toContain(credential);
    auth.setAuthToken(session(replacement));
    fixture.detectChanges();
    expectCleared();
    expect(modals.current().dismiss).toHaveBeenCalled();
    expect(component.errorMessage).toContain('previous profile');
    flushMicrotasks();
  }));

  it('destroys only owned client work and cannot retain a credential afterward', fakeAsync(() => {
    open();
    component.connectDevice();
    const request = http.expectOne(request => request.method === 'POST');
    fixture.destroy();
    expect(request.cancelled).toBeTrue();
    expectCleared();
    flushMicrotasks();
  }));

  it('does not issue a second credential for competing connect clicks', () => {
    open();
    component.connectDevice();
    component.connectDevice();
    const request = http.expectOne(request => request.method === 'POST');
    request.flush({success: false});
    expectCleared();
    expect(component.isLoading).toBeFalse();
  });

  it('does not let an older modal completion clear the new owned pairing', fakeAsync(() => {
    const older = open();
    older.finish();
    const newer = open();
    component.connectDevice();
    issued();
    flushMicrotasks();
    expect(component.accessToken).toBe(credential);
    expect(newer.dismiss).not.toHaveBeenCalled();
    resolveQR('data:image/png;base64,synthetic');
    flushMicrotasks();
    expect(component.qrCodeData).toContain(credential);
  }));

  it('checks a profile replacement at a late QR response even before change detection', fakeAsync(() => {
    open();
    component.connectDevice();
    issued();
    auth.setAuthToken(session(replacement));
    resolveQR('data:image/png;base64,synthetic-private-late-qr');
    flushMicrotasks();
    expectCleared();
    expect(component.errorMessage).toContain('previous profile');
  }));

  it('refuses missing immutable identity before creating another credential', fakeAsync(() => {
    open();
    auth.setAuthToken(session(''));
    component.connectDevice();
    http.expectNone(request => request.method === 'POST');
    expectCleared();
    expect(component.errorMessage).toContain('previous profile');
    flushMicrotasks();
  }));

  it('keeps QR rejection payloads private and clears the now uncertain handoff', fakeAsync(() => {
    open();
    component.connectDevice();
    issued();
    rejectQR(new Error(credential));
    flushMicrotasks();
    expectCleared();
    expect(component.errorMessage).toContain('Review saved devices');
    expect(JSON.stringify(consoleSpies.flatMap(spy => spy.calls.allArgs()))).not.toContain(credential);
  }));

  it('does not claim successful device removal from an intercepted500 response', () => {
    component.tokens = [{tokenId: 'synthetic-device', name: 'synthetic', issuedAt: '2026-11-01', expiresAt: '2026-12-01'}];
    spyOn(window, 'confirm').and.returnValue(true);
    component.deleteToken('synthetic-device');
    http.expectOne(request => request.method === 'DELETE').flush({success: false}, {status: 500, statusText: 'Synthetic storage unavailable'});
    expect(component.tokens.length).toBe(1);
    expect(component.errorMessage).toContain('removal was not confirmed');
    http.expectNone(request => request.method === 'GET');
  });

  it('does not silently turn malformed device metadata into a zero-device assertion', () => {
    component.loadTokens();
    http.expectOne(request => request.method === 'GET').flush({success: true, data: [{token_id: 'synthetic', issued_at: '2026-11-01', expires_at: '2026-12-01'}, {}]});
    fixture.detectChanges();
    expect(component.hasError).toBeTrue();
    expect(component.tokens).toEqual([]);
    expect(fixture.nativeElement.textContent).not.toContain('No devices connected yet');
  });

  it('clears credentials when discovery fails after issuance without claiming server rollback', () => {
    open();
    component.connectDevice();
    http.expectOne(request => request.method === 'POST').flush({success: true, data: credential});
    http.expectOne(request => request.url.endsWith('/secure/access/token')).flush({success: true, data: []});
    http.expectOne(request => request.url.endsWith('/secure/sync/discovery')).flush({success: false}, {status: 500, statusText: 'Synthetic discovery unavailable'});
    expectCleared();
    expect(component.errorMessage).toContain('Review saved devices');
    expect(qr).not.toHaveBeenCalled();
  });

  it('observes cross-window storage identity changes without waiting for a template tick', fakeAsync(() => {
    open();
    component.connectDevice();
    const request = http.expectOne(request => request.method === 'POST');
    auth.setAuthToken(session(replacement));
    window.dispatchEvent(new StorageEvent('storage', {key: 'token'}));
    expect(request.cancelled).toBeTrue();
    expectCleared();
    expect(component.errorMessage).toContain('previous profile');
    flushMicrotasks();
  }));

  it('preserves explicit device name and zero-expiration semantics through registered authentication', () => {
    open();
    component.newDeviceName = 'Synthetic personal device';
    component.newDeviceExpiration = 0;
    component.connectDevice();
    const request = http.expectOne(request => request.method === 'POST');
    expect(request.request.body).toEqual({name: 'Synthetic personal device', expiration: 0});
    expect(request.request.headers.get('Authorization')).toBe('Bearer ' + auth.GetAuthToken());
    request.flush({success: false});
    expectCleared();
  });

  it('dismisses only its owned modal, not another active application dialog', fakeAsync(() => {
    const unrelated = modals.open();
    const owned = open();
    fixture.destroy();
    flushMicrotasks();
    expect(owned.dismiss).toHaveBeenCalled();
    expect(unrelated.dismiss).not.toHaveBeenCalled();
  }));

  it('does not connect a device before the active server profile has been confirmed', () => {
    component.currentUser = null;
    fixture.detectChanges();
    const button: HTMLButtonElement = fixture.nativeElement.querySelector('.card-header button');
    expect(button.disabled).toBeTrue();
    component.connectDevice();
    http.expectNone(request => request.method === 'POST');
  });
});
