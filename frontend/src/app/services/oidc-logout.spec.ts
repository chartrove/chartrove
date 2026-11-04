import { TestBed } from '@angular/core/testing';
import { HttpClient, provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { HTTP_CLIENT_TOKEN } from '../dependency-injection';
import { AuthService } from './auth.service';

describe('OIDC session logout', () => {
  let auth: AuthService; let http: HttpTestingController; let previous: string | null;
  const token = (type: string) => `e30.${btoa(JSON.stringify({sub: 'synthetic', token_type: type, exp: 9999999999}))}.c3ludGhldGlj`;
  beforeEach(() => {
    TestBed.configureTestingModule({providers: [provideHttpClient(), provideHttpClientTesting(), {provide: HTTP_CLIENT_TOKEN, useExisting: HttpClient}]});
    auth = TestBed.inject(AuthService); http = TestBed.inject(HttpTestingController); previous = auth.GetAuthToken();
  });
  afterEach(() => { http.verify(); if (previous === null) localStorage.removeItem(auth.FASTEN_JWT_LOCALSTORAGE_KEY); else localStorage.setItem(auth.FASTEN_JWT_LOCALSTORAGE_KEY, previous) });
  it('revokes the server session before completing logout', async () => {
    auth.setAuthToken(token('oidc')); const pending = auth.Logout();
    const req = http.expectOne(r => r.url.endsWith('/secure/auth/oidc/logout')); expect(req.request.method).toBe('POST');
    req.flush({success: true}); await pending; expect(auth.GetAuthToken()).toBeNull();
  });
  it('clears this device but reports unconfirmed server revocation', async () => {
    auth.setAuthToken(token('oidc')); const pending = auth.Logout();
    http.expectOne(r => r.url.endsWith('/secure/auth/oidc/logout')).flush({}, {status: 503, statusText: 'Unavailable'});
    await expectAsync(pending).toBeRejected(); expect(auth.GetAuthToken()).toBeNull();
  });
  it('avoids recursive revocation when an interceptor invalidates authentication', async () => {
    auth.setAuthToken(token('oidc')); await auth.Logout(false); expect(auth.GetAuthToken()).toBeNull();
  });
  for (const value of [token(''), 'malformed']) {
    it(`preserves local logout for ${value === 'malformed' ? 'malformed storage' : 'password sessions'}`, async () => {
      localStorage.setItem(auth.FASTEN_JWT_LOCALSTORAGE_KEY, value); await auth.Logout(); expect(auth.GetAuthToken()).toBeNull();
    });
  }
});
