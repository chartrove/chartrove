import {HttpTestingController, provideHttpClientTesting} from '@angular/common/http/testing';
import {TestBed, fakeAsync, tick} from '@angular/core/testing';
import {Router} from '@angular/router';
import {AppModule} from './app.module';
import {AuthService} from './services/auth.service';
import {FastenApiService} from './services/fasten-api.service';
import {environment} from '../environments/environment';
import {GetEndpointAbsolutePath} from '../lib/utils/endpoint_absolute_path';

// Registered configured Angular HTTP; neither real networking nor Wails runtime.
describe('Encryption HTTP contracts at configured root, prefix and remote API', () => {
  let previousBase: string;
  let http: HttpTestingController;
  let service: FastenApiService;

  beforeEach(() => {
    previousBase = environment.fasten_api_endpoint_base;
    const auth = jasmine.createSpyObj<AuthService>('AuthService', ['GetAuthToken', 'Logout']);
    auth.GetAuthToken.and.returnValue('synthetic-native-encryption-session');
    TestBed.configureTestingModule({imports: [AppModule], providers: [
      provideHttpClientTesting(), {provide: AuthService, useValue: auth},
      {provide: Router, useValue: jasmine.createSpyObj<Router>('Router', ['navigateByUrl'])},
    ]});
    http = TestBed.inject(HttpTestingController);
    service = TestBed.inject(FastenApiService);
  });

  afterEach(() => {
    environment.fasten_api_endpoint_base = previousBase;
    http.verify();
  });

  for (const base of ['/api', '/synthetic-prefix/api', 'https://synthetic-api.example.invalid/proxy/api']) {
    it(`retains configured endpoint/exact payload/authentication and rejects false ACK at ${base}`, () => {
      environment.fasten_api_endpoint_base = base;
      const endpoint = GetEndpointAbsolutePath(globalThis.location, base);
      const exactKey = ' synthetic-native-exact-key ';
      let failed = false;
      let succeeded = false;
      service.setupEncryptionKey(exactKey).subscribe({next: () => succeeded = true, error: () => failed = true});
      const request = http.expectOne(endpoint + '/encryption-key');
      expect(request.request.body).toEqual({encryption_key: exactKey});
      expect(request.request.headers.get('Authorization')).toBe('Bearer synthetic-native-encryption-session');
      request.flush({success: false});
      expect(succeeded).toBeFalse();
      expect(failed).toBeTrue();
    });

    it(`confirms only fresh unlocked metadata through the configured provider at ${base}`, fakeAsync(() => {
      environment.fasten_api_endpoint_base = base;
      const endpoint = GetEndpointAbsolutePath(globalThis.location, base);
      let ready = false;
      service.confirmEncryptionReady().subscribe({next: () => ready = true});
      tick(0);
      const request = http.expectOne(candidate => candidate.url === endpoint + '/health');
      expect(request.request.params.get('_encryption_confirmation')).toBeTruthy();
      expect(request.request.headers.has('Cache-Control')).toBeFalse();
      expect(request.request.headers.get('Authorization')).toBe('Bearer synthetic-native-encryption-session');
      request.flush({success: true, data: {standby_mode: false, first_run_wizard: true}});
      expect(ready).toBeTrue();
    }));
  }
});
