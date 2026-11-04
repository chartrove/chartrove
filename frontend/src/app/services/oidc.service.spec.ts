import { TestBed } from '@angular/core/testing';
import { HttpClient, provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { HTTP_CLIENT_TOKEN } from '../dependency-injection';
import { OIDCService } from './oidc.service';
import { AuthService } from './auth.service';

describe('OIDC browser transaction', () => {
  let service: OIDCService;
  let http: HttpTestingController;
  const auth = jasmine.createSpyObj('AuthService', ['setAuthToken']);
  beforeEach(() => {
    TestBed.configureTestingModule({providers: [provideHttpClient(), provideHttpClientTesting(), {provide: HTTP_CLIENT_TOKEN, useExisting: HttpClient}, {provide: AuthService, useValue: auth}]});
    service = TestBed.inject(OIDCService); http = TestBed.inject(HttpTestingController); auth.setAuthToken.calls.reset();
    sessionStorage.removeItem(service.storageKey);
  });
  afterEach(() => { http.verify(); sessionStorage.removeItem(service.storageKey) });
  it('keeps browser proof in this tab and sends no native identity or password in the redirect', async () => {
    const pending = service.start('synthetic-password');
    const req = http.expectOne(r => r.url.endsWith('/secure/auth/oidc/link'));
    expect(req.request.body).toEqual({password: 'synthetic-password'});
    req.flush({success: true, data: {authorization_url: 'https://idp.example/authorize?state=state', state: 'state', browser_secret: 'secret'}});
    expect(await pending).toBe('https://idp.example/authorize?state=state');
    expect(sessionStorage.getItem(service.storageKey)).not.toContain('synthetic-password');
    const complete = service.complete(new URLSearchParams('state=state&code=synthetic-code'));
    const callback = http.expectOne(r => r.url.endsWith('/auth/oidc/complete'));
    expect(callback.request.body).toEqual({state: 'state', browser_secret: 'secret', code: 'synthetic-code'});
    expect(sessionStorage.getItem(service.storageKey)).toBeNull();
    callback.flush({success: true, data: 'native-session-token'}); await complete;
    expect(auth.setAuthToken).toHaveBeenCalledOnceWith('native-session-token');
    await expectAsync(service.complete(new URLSearchParams('state=state&code=synthetic-code'))).toBeRejected();
  });
  for (const query of ['state=wrong&code=c', 'state=state&code=a&code=b', 'state=state&error=access_denied', 'state=state&state=other&code=c', 'state=state']) {
    it(`rejects invalid callback ${query}`, async () => {
      sessionStorage.setItem(service.storageKey, JSON.stringify({state: 'state', browser_secret: 'secret', expires: Date.now() + 300000}));
      await expectAsync(service.complete(new URLSearchParams(query))).toBeRejected();
      expect(auth.setAuthToken).not.toHaveBeenCalled();
    });
  }
  it('rejects expired proof and ignores callback tokens', async () => {
    sessionStorage.setItem(service.storageKey, JSON.stringify({state: 'state', browser_secret: 'secret', expires: 0}));
    await expectAsync(service.complete(new URLSearchParams('state=state&code=c&token=untrusted'))).toBeRejected();
    expect(auth.setAuthToken).not.toHaveBeenCalled();
  });
  it('does not install a token after a rejected exchange', async () => {
    sessionStorage.setItem(service.storageKey, JSON.stringify({state: 'state', browser_secret: 'secret', expires: Date.now() + 300000}));
    const pending = service.complete(new URLSearchParams('state=state&code=c'));
    http.expectOne(r => r.url.endsWith('/auth/oidc/complete')).flush({success: false, data: 'untrusted'});
    await expectAsync(pending).toBeRejected(); expect(auth.setAuthToken).not.toHaveBeenCalled();
  });
});
