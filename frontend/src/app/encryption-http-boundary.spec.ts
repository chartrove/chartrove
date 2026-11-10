import {HttpTestingController, provideHttpClientTesting} from '@angular/common/http/testing';
import {TestBed} from '@angular/core/testing';
import {Router} from '@angular/router';
import {AppModule} from './app.module';
import {AuthService} from './services/auth.service';
import {FastenApiService} from './services/fasten-api.service';
import {environment} from '../environments/environment';
import {GetEndpointAbsolutePath} from '../lib/utils/endpoint_absolute_path';

describe('Encryption acknowledgments through registered application HTTP providers', () => {
  let http: HttpTestingController;
  let service: FastenApiService;
  const privateMarker = 'synthetic-private-encryption-response-canary';
  const exactKey = ' synthetic-key-exact ';
  const endpoint = GetEndpointAbsolutePath(globalThis.location, environment.fasten_api_endpoint_base);

  beforeEach(() => {
    const auth = jasmine.createSpyObj<AuthService>('AuthService', ['GetAuthToken', 'Logout']);
    auth.GetAuthToken.and.returnValue('synthetic-encryption-http-token');
    const router = jasmine.createSpyObj<Router>('Router', ['navigateByUrl']);
    router.navigateByUrl.and.returnValue(Promise.resolve(true));
    TestBed.configureTestingModule({
      imports: [AppModule],
      providers: [provideHttpClientTesting(), {provide: AuthService, useValue: auth}, {provide: Router, useValue: router}],
    });
    http = TestBed.inject(HttpTestingController);
    service = TestBed.inject(FastenApiService);
  });

  afterEach(() => http.verify());

  for (const operation of [
    {name: 'setup', path: '/encryption-key', call: () => service.setupEncryptionKey(exactKey)},
    {name: 'validate', path: '/encryption-key/validate', call: () => service.validateEncryptionKey(exactKey)},
  ]) {
    for (const status of [200, 400, 401, 403, 500]) {
      it(`rejects private false ${operation.name} acknowledgment HTTP${status}`,  () => {
        let succeeded = false;
        let failure: unknown;
        operation.call().subscribe({next: () => succeeded = true, error: error => failure = error});
        const request = http.expectOne(endpoint + operation.path);
        expect(request.request.body).toEqual({encryption_key: exactKey});
        request.flush({success: false, error: privateMarker}, {status, statusText: status === 200 ? 'OK' : 'Denied'});
        expect(succeeded).toBeFalse();
        expect(failure).toEqual(jasmine.any(Error));
        expect(String(failure)).not.toContain(privateMarker);
      });
    }

    for (const body of [null, [], {}, {success: 'true'}, {success: 1}]) {
      it(`rejects malformed ${operation.name} acknowledgment ${JSON.stringify(body)}`, () => {
        let succeeded = false;
        let failed = false;
        operation.call().subscribe({next: () => succeeded = true, error: () => failed = true});
        http.expectOne(endpoint + operation.path).flush(body);
        expect(succeeded).toBeFalse();
        expect(failed).toBeTrue();
      });
    }

    for (const status of [400, 500]) {
      it(`rejects success-shaped ${operation.name} bodies on HTTP${status}`, () => {
        let succeeded = false;
        let failure: unknown;
        operation.call().subscribe({next: () => succeeded = true, error: error => failure = error});
        http.expectOne(endpoint + operation.path).flush({success: true, error: privateMarker}, {status, statusText: 'Denied'});
        expect(succeeded).toBeFalse();
        expect(failure).toEqual(jasmine.any(Error));
        expect(String(failure)).not.toContain(privateMarker);
      });
    }

    it(`preserves literal-success no-data ${operation.name} acknowledgment and exact outgoing key`, () => {
      let succeeded = false;
      operation.call().subscribe({next: () => succeeded = true, error: () => fail('Synthetic positive acknowledgment must succeed')});
      const request = http.expectOne(endpoint + operation.path);
      expect(request.request.body).toEqual({encryption_key: exactKey});
      request.flush({success: true});
      expect(succeeded).toBeTrue();
    });
  }

  for (const body of [null, {}, {success: false, data: privateMarker}, {success: true}, {success: true, data: 42}, {success: true, data: ''}]) {
    it(`rejects unavailable generated key contract ${JSON.stringify(body)}`, () => {
      let succeeded = false;
      let failure: unknown;
      service.getEncryptionKey().subscribe({next: () => succeeded = true, error: error => failure = error});
      http.expectOne(endpoint + '/encryption-key').flush(body);
      expect(succeeded).toBeFalse();
      expect(failure).toEqual(jasmine.any(Error));
      expect(String(failure)).not.toContain(privateMarker);
    });
  }

  for (const status of [400, 401, 403, 500]) {
    it(`rejects success-shaped generation body on HTTP${status}, including empty completed denial`, () => {
      let succeeded = false;
      let failure: unknown;
      service.getEncryptionKey().subscribe({next: () => succeeded = true, error: error => failure = error});
      http.expectOne(endpoint + '/encryption-key').flush({success: true, data: privateMarker}, {status, statusText: 'Denied'});
      expect(succeeded).toBeFalse();
      expect(failure).toEqual(jasmine.any(Error));
      expect(String(failure)).not.toContain(privateMarker);
    });
  }

  it('preserves exact generated key bytes after literal-success receipt', () => {
    let result: string | undefined;
    service.getEncryptionKey().subscribe({next: value => result = value, error: () => fail('Synthetic positive generation must succeed')});
    http.expectOne(endpoint + '/encryption-key').flush({success: true, data: exactKey});
    expect(result).toBe(exactKey);
  });
});
