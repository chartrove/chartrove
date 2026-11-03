import { HttpClient, provideHttpClient, withXhr } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { HTTP_CLIENT_TOKEN } from '../dependency-injection';
import { AuthService } from './auth.service';

describe('AuthService JWT decoding compatibility', () => {
  const now = 1893456000; // 2030-01-01T00:00:00Z
  const legacyClaims = {
    iss: 'docker-fastenhealth',
    sub: 'synthetic-claims-user',
    iat: now - 1,
    exp: now + 60,
    full_name: 'Synthetic Claims User',
    email: 'synthetic@example.invalid',
    role: 'user',
  };
  let service: AuthService;
  let previousToken: string | null;

  // These fixtures exercise decoding only. Backend authentication verifies the
  // signature; no signing key or real account is used by this browser consumer.
  function tokenFor(claims: Record<string, unknown>): string {
    const bytes = new TextEncoder().encode(JSON.stringify(claims));
    const payload = btoa(String.fromCharCode(...bytes))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=+$/, '');
    return `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.${payload}.c3ludGhldGlj`;
  }

  function store(claims: Record<string, unknown>): string {
    const token = tokenFor(claims);
    localStorage.setItem(service.FASTEN_JWT_LOCALSTORAGE_KEY, token);
    return token;
  }

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withXhr()),
        provideHttpClientTesting(),
        { provide: HTTP_CLIENT_TOKEN, useExisting: HttpClient },
      ],
    });
    service = TestBed.inject(AuthService);
    previousToken = localStorage.getItem(service.FASTEN_JWT_LOCALSTORAGE_KEY);
    localStorage.removeItem(service.FASTEN_JWT_LOCALSTORAGE_KEY);
    jasmine.clock().install();
    jasmine.clock().mockDate(new Date(now * 1000));
  });

  afterEach(() => {
    jasmine.clock().uninstall();
    if (previousToken === null) localStorage.removeItem(service.FASTEN_JWT_LOCALSTORAGE_KEY);
    else localStorage.setItem(service.FASTEN_JWT_LOCALSTORAGE_KEY, previousToken);
  });

  it('reads the legacy Go claim shape without rewriting its stored token', () => {
    const token = store(legacyClaims);
    expect<unknown>(service.GetCurrentUser()).toEqual(legacyClaims);
    expect(service.GetAuthToken()).toBe(token);
  });

  it('retains access-token metadata and optional registered claims', () => {
    const claims = {
      ...legacyClaims,
      aud: ['synthetic-client'],
      nbf: now - 1,
      jti: 'synthetic-token-id',
      token_id: 'synthetic-optional-id',
      token_type: 'access',
    };
    store(claims);
    expect<unknown>(service.GetCurrentUser()).toEqual(claims);
  });

  it('decodes UTF-8 profile text without changing it', () => {
    store({ ...legacyClaims, full_name: 'Synthetic مستخدم — テスト' });
    expect(service.GetCurrentUser().full_name).toBe('Synthetic مستخدم — テスト');
  });

  it('publishes authenticated state for an unexpired token', async () => {
    store(legacyClaims);
    expect(await service.IsAuthenticated()).toBeTrue();
    expect(service.IsAuthenticatedSubject.value).toBeTrue();
  });

  it('clears prior authenticated state for an expired token', async () => {
    store(legacyClaims);
    await service.IsAuthenticated();
    store({ ...legacyClaims, exp: now - 1 });
    expect(await service.IsAuthenticated()).toBeFalse();
    expect(service.IsAuthenticatedSubject.value).toBeFalse();
  });

  it('rejects a token exactly at its expiry boundary', async () => {
    store({ ...legacyClaims, exp: now });
    expect(await service.IsAuthenticated()).toBeFalse();
  });

  it('does not authenticate a token without expiry', async () => {
    store({ sub: legacyClaims.sub, role: 'user' });
    expect(await service.IsAuthenticated()).toBeFalse();
  });

  it('clears prior authenticated state when the token is absent', async () => {
    store(legacyClaims);
    await service.IsAuthenticated();
    localStorage.removeItem(service.FASTEN_JWT_LOCALSTORAGE_KEY);
    expect(await service.IsAuthenticated()).toBeFalse();
    expect(service.IsAuthenticatedSubject.value).toBeFalse();
    expect(() => service.GetCurrentUser()).toThrowError('no auth token found');
  });

  for (const role of ['admin', 'user', 'Admin']) {
    it(`preserves the existing admin-role decision for ${role}`, () => {
      store({ ...legacyClaims, role });
      expect(service.IsAdmin()).toBe(role === 'admin');
    });
  }

  it('removes the decoded session on logout', async () => {
    store(legacyClaims);
    await service.IsAuthenticated();
    await service.Logout();
    expect(service.GetAuthToken()).toBeNull();
    expect(service.IsAuthenticatedSubject.value).toBeFalse();
  });

  it('preserves rejection of malformed stored tokens', async () => {
    localStorage.setItem(service.FASTEN_JWT_LOCALSTORAGE_KEY, 'synthetic-not-a-jwt');
    expect(() => service.GetCurrentUser()).toThrow();
    await expectAsync(service.IsAuthenticated()).toBeRejected();
  });
});
