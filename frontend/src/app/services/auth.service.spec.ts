import { TestBed } from '@angular/core/testing';

import { AuthService } from './auth.service';
import { of } from 'rxjs';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import {HTTP_CLIENT_TOKEN} from '../dependency-injection';
import { HttpClient, provideHttpClient, withInterceptorsFromDi, withXhr } from '@angular/common/http';

describe('AuthService', () => {
  let service: AuthService;

  beforeEach(() => {
    TestBed.configureTestingModule({
    imports: [],
    providers: [
        {
            provide: HTTP_CLIENT_TOKEN,
            useClass: HttpClient,
        },
        provideHttpClient(withXhr(), withInterceptorsFromDi()),
        provideHttpClientTesting(),
    ]
});
    service = TestBed.inject(AuthService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('forwards synthetic provider tokens without writing them to the console', async () => {
    const state = 'chartrove-synthetic-idp-log-privacy';
    const previousState = localStorage.getItem(state);
    const previousToken = localStorage.getItem(service.FASTEN_JWT_LOCALSTORAGE_KEY);
    const providerTokens = {
      access_token: 'synthetic-access-secret-do-not-log',
      id_token: 'synthetic-identity-secret-do-not-log',
      token_type: 'Bearer',
    };
    const providerResponse = new Response(JSON.stringify(providerTokens), {
      status: 200, headers: { 'Content-Type': 'application/json' },
    });
    // A constructed Response needs the URL that fetch normally supplies to the SDK.
    Object.defineProperty(providerResponse, 'url', { value: 'https://wallet.hello.coop/oauth/token' });
    const fetchToken = spyOn(globalThis, 'fetch').and.returnValue(Promise.resolve(providerResponse));
    const localBackend = spyOn(TestBed.inject(HTTP_CLIENT_TOKEN) as unknown as HttpClient, 'post').and.returnValue(of({
      success: true, data: 'synthetic-local-auth-token',
    }));
    const logs = spyOn(console, 'log');
    try {
      localStorage.setItem(state, JSON.stringify({
        state, code_verifier: 'synthetic-local-code-verifier-12345678901234567890',
        redirect_uri: 'http://localhost/synthetic-idp-callback',
      }));
      const token = await service.IdpCallback('hello', state, 'synthetic-code');

      expect(fetchToken.calls.count()).toBe(1);
      expect(String(fetchToken.calls.mostRecent().args[0])).toBe('https://wallet.hello.coop/oauth/token');
      expect(localBackend.calls.mostRecent().args[0]).toMatch(/\/auth\/callback\/hello$/);
      expect(localBackend.calls.mostRecent().args[1]).toEqual(providerTokens);
      expect(token).toBe('synthetic-local-auth-token');
      expect(service.GetAuthToken()).toBe('synthetic-local-auth-token');
      expect(localStorage.getItem(state)).toBeNull();
      const messages = JSON.stringify(logs.calls.allArgs());
      expect(messages).not.toContain('synthetic-access-secret-do-not-log');
      expect(messages).not.toContain('synthetic-identity-secret-do-not-log');
    } finally {
      if (previousState === null) localStorage.removeItem(state);
      else localStorage.setItem(state, previousState);
      if (previousToken === null) localStorage.removeItem(service.FASTEN_JWT_LOCALSTORAGE_KEY);
      else localStorage.setItem(service.FASTEN_JWT_LOCALSTORAGE_KEY, previousToken);
    }
  });
});
