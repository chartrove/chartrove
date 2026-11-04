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
  });
  afterEach(() => http.verify());
  it('uses the server HttpOnly cookie without persisting browser proof in JavaScript', async () => {
    const storage = spyOn(sessionStorage, 'setItem');
    const pending = service.start('synthetic-password');
    const req = http.expectOne(r => r.url.endsWith('/secure/auth/oidc/link'));
    expect(req.request.body).toEqual({password: 'synthetic-password'});
    req.flush({success: true, data: {authorization_url: 'https://idp.example/authorize?state=state', state: 'state'}});
    expect(await pending).toBe('https://idp.example/authorize?state=state');
    expect(storage).not.toHaveBeenCalled();
    const complete = service.complete(new URLSearchParams('state=state&code=synthetic-code&token=untrusted'));
    const callback = http.expectOne(r => r.url.endsWith('/auth/oidc/complete'));
    expect(callback.request.body).toEqual({state: 'state', code: 'synthetic-code'});
    callback.flush({success: true, data: 'native-session-token'}); await complete;
    expect(auth.setAuthToken).toHaveBeenCalledOnceWith('native-session-token');
    expect(storage).not.toHaveBeenCalled();
  });
  for (const query of ['state=state&code=a&code=b', 'state=state&error=access_denied', 'state=state&state=other&code=c', 'state=state', 'code=c']) {
    it(`rejects invalid callback ${query}`, async () => {
      await expectAsync(service.complete(new URLSearchParams(query))).toBeRejected();
      expect(auth.setAuthToken).not.toHaveBeenCalled();
    });
  }
  it('does not install a token after a rejected state, expired proof or replay', async () => {
    const pending = service.complete(new URLSearchParams('state=unknown&code=c'));
    http.expectOne(r => r.url.endsWith('/auth/oidc/complete')).flush({success: false, data: 'untrusted'});
    await expectAsync(pending).toBeRejected(); expect(auth.setAuthToken).not.toHaveBeenCalled();
  });
});
