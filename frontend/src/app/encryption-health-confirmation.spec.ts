import {HttpTestingController, provideHttpClientTesting} from '@angular/common/http/testing';
import type {HttpRequest} from '@angular/common/http';
import {TestBed, fakeAsync, tick} from '@angular/core/testing';
import {Router} from '@angular/router';
import type {Observable} from 'rxjs';
import {AppModule} from './app.module';
import {AuthService} from './services/auth.service';
import {FastenApiService} from './services/fasten-api.service';
import {environment} from '../environments/environment';
import {GetEndpointAbsolutePath} from '../lib/utils/endpoint_absolute_path';

describe('Bounded fresh encryption confirmation through registered HTTP', () => {
  const endpoint = GetEndpointAbsolutePath(globalThis.location, environment.fasten_api_endpoint_base);
  const health = (request: HttpRequest<unknown>): boolean => request.url === endpoint + '/health';
  let service: FastenApiService;
  let http: HttpTestingController;

  beforeEach(() => {
    const auth = jasmine.createSpyObj<AuthService>('AuthService', ['GetAuthToken', 'Logout']);
    auth.GetAuthToken.and.returnValue('synthetic-health-test-token');
    const router = jasmine.createSpyObj<Router>('Router', ['navigateByUrl']);
    router.navigateByUrl.and.returnValue(Promise.resolve(true));
    TestBed.configureTestingModule({imports: [AppModule], providers: [
      provideHttpClientTesting(), {provide: AuthService, useValue: auth}, {provide: Router, useValue: router},
    ]});
    service = TestBed.inject(FastenApiService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('does not overlap requests and stops after a fresh literal-success unlocked observation', fakeAsync(() => {
    let ready = false;
    service.confirmEncryptionReady().subscribe({next: () => ready = true, error: () => fail('Synthetic positive confirmation must succeed')});
    tick(0);
    const first = http.expectOne(health);
    expect(first.request.headers.has('Cache-Control')).toBeFalse();
    expect(first.request.params.get('_encryption_confirmation')).toBeTruthy();
    tick(1000);
    http.expectNone(health);
    first.flush({success: true, data: {standby_mode: true, first_run_wizard: false}});
    expect(ready).toBeFalse();
    tick(500);
    const second = http.expectOne(health);
    expect(second.request.urlWithParams).not.toBe(first.request.urlWithParams);
    second.flush({success: true, data: {standby_mode: false, first_run_wizard: false}});
    expect(ready).toBeTrue();
    tick(10000);
    http.expectNone(health);
  }));

  it('uses a fresh non-secret cache-busting scope for every subscription, not identity authority', fakeAsync(() => {
    const observation = service.confirmEncryptionReady();
    const firstSubscription = observation.subscribe();
    tick(0);
    const first = http.expectOne(health);
    firstSubscription.unsubscribe();
    const secondSubscription = observation.subscribe();
    tick(0);
    const second = http.expectOne(health);
    expect(second.request.urlWithParams).not.toBe(first.request.urlWithParams);
    expect(second.request.headers.has('Cache-Control')).toBeFalse();
    secondSubscription.unsubscribe();
    expect(first.cancelled && second.cancelled).toBeTrue();
  }));

  it('does not infer unlocked state from malformed or false health metadata', fakeAsync(() => {
    let ready = false;
    const subscription = service.confirmEncryptionReady().subscribe({next: () => ready = true});
    for (const [index, body] of [
      {success: false, data: {standby_mode: false, first_run_wizard: true}},
      {success: true, data: {standby_mode: 'false', first_run_wizard: true}},
      {success: true, data: {standby_mode: false}},
      {success: true, data: {standby_mode: false, first_run_wizard: 'true'}},
    ].entries()) {
      tick(index === 0 ? 0 : 500);
      http.expectOne(health).flush(body);
      expect(ready).toBeFalse();
    }
    subscription.unsubscribe();
  }));

  it('does not accept success-shaped HTTP failure and can observe the subsequent healthy restart', fakeAsync(() => {
    let ready = false;
    service.confirmEncryptionReady().subscribe({next: () => ready = true});
    tick(0);
    http.expectOne(health).flush({success: true, data: {standby_mode: false, first_run_wizard: true}}, {status: 500, statusText: 'Synthetic restart'});
    expect(ready).toBeFalse();
    tick(500);
    http.expectOne(health).flush({success: true, data: {standby_mode: false, first_run_wizard: true}});
    expect(ready).toBeTrue();
  }));

  it('releases polling/request subscriptions on unsubscription without subsequent requests', fakeAsync(() => {
    const subscription = service.confirmEncryptionReady().subscribe();
    tick(0);
    const request = http.expectOne(health);
    subscription.unsubscribe();
    expect(request.cancelled).toBeTrue();
    tick(10000);
    http.expectNone(health);
  }));

  it('bounds all hung requests and the whole observation to thirty seconds', fakeAsync(() => {
    let failure: unknown;
    let ready = false;
    service.confirmEncryptionReady().subscribe({next: () => ready = true, error: error => failure = error});
    tick(30000);
    const requests = http.match(health);
    expect(requests.length).toBeGreaterThan(0);
    expect(requests.length).toBeLessThanOrEqual(7);
    expect(requests.every(request => request.cancelled)).toBeTrue();
    expect(failure).toEqual(jasmine.any(Error));
    expect(ready).toBeFalse();
    tick(10000);
    http.expectNone(health);
  }));

  for (const operation of [
    {name: 'generation', path: '/encryption-key', call: () => service.getEncryptionKey()},
    {name: 'validation', path: '/encryption-key/validate', call: () => service.validateEncryptionKey('synthetic-only')},
    {name: 'setup', path: '/encryption-key', call: () => service.setupEncryptionKey('synthetic-only')},
  ]) {
    it(`bounds pending ${operation.name} without retrying its request`, fakeAsync(() => {
      let failure: unknown;
      const result: Observable<unknown> = operation.call();
      result.subscribe({error: error => failure = error});
      const request = http.expectOne(endpoint + operation.path);
      tick(30000);
      expect(request.cancelled).toBeTrue();
      expect(failure).toEqual(jasmine.any(Error));
      tick(10000);
      http.expectNone(endpoint + operation.path);
    }));
  }
});
