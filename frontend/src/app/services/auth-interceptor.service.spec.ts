import { HttpErrorResponse, HttpHandler, HttpHeaders, HttpRequest, HttpResponse } from '@angular/common/http';
import {Router} from '@angular/router';
import {of,throwError} from 'rxjs';
import {AuthInterceptorService} from './auth-interceptor.service';
import {AuthService} from './auth.service';
import {environment} from '../../environments/environment';
import {GetEndpointAbsolutePath} from '../../lib/utils/endpoint_absolute_path';

describe('AuthInterceptorService Patient privacy', () => {
  for(const path of ['/api/auth/signup','/api/secure/users']) {
    it(`keeps session invalidation limited to authentication failures on ${path}`,()=>{
      const auth=jasmine.createSpyObj<AuthService>('AuthService',['GetAuthToken','Logout']);auth.GetAuthToken.and.returnValue('synthetic-existing-session');
      const router=jasmine.createSpyObj<Router>('Router',['navigateByUrl']);const handler=jasmine.createSpyObj<HttpHandler>('HttpHandler',['handle']);
      handler.handle.and.returnValue(throwError(()=>new HttpErrorResponse({status:403,error:{code:'registration_closed'}})));
      let rejected=false;new AuthInterceptorService(auth,router).intercept(new HttpRequest('POST',path,{}),handler).subscribe({error:()=>rejected=true});
      if(path==='/api/auth/signup') {expect(rejected).toBeTrue();expect(auth.Logout).not.toHaveBeenCalled();expect(router.navigateByUrl).not.toHaveBeenCalled();}
      else {expect(auth.Logout).toHaveBeenCalledWith(false);expect(router.navigateByUrl).toHaveBeenCalledWith('/auth/signin');}
    });
  }
  it('does not log a serialized Patient import or its authorization header', () => {
    const log = spyOn(console, 'log');
    const auth = jasmine.createSpyObj<AuthService>('AuthService', ['GetAuthToken']);
    auth.GetAuthToken.and.returnValue('synthetic-test-token');
    const handler = jasmine.createSpyObj<HttpHandler>('HttpHandler', ['handle']);
    handler.handle.and.returnValue(of(new HttpResponse({status: 200})));
    const interceptor = new AuthInterceptorService(auth, jasmine.createSpyObj<Router>('Router', ['navigateByUrl']));
    const body = '--synthetic-boundary\r\n{"resourceType":"Patient","id":"synthetic-only","name":[{"text":"Synthetic Patient"}]}';
    interceptor.intercept(new HttpRequest('POST', '/api/secure/source/manual', body), handler).subscribe();
    expect(handler.handle).toHaveBeenCalledTimes(1);
    expect(log).not.toHaveBeenCalled();
  });
});

describe('AuthInterceptorService application authority boundary', () => {
  let auth: jasmine.SpyObj<AuthService>;
  let router: jasmine.SpyObj<Router>;
  let handler: jasmine.SpyObj<HttpHandler>;
  let interceptor: AuthInterceptorService;
  const lighthouse = environment.lighthouse_api_endpoint_base.replace(/\/+$/, '') + '/catalog';
  const application = new URL(GetEndpointAbsolutePath(globalThis.location, environment.fasten_api_endpoint_base));
  const applicationPath = application.pathname.replace(/\/+$/, '');

  beforeEach(() => {
    auth = jasmine.createSpyObj<AuthService>('AuthService', ['GetAuthToken', 'Logout']);
    auth.GetAuthToken.and.returnValue('synthetic-private-native-session');
    router = jasmine.createSpyObj<Router>('Router', ['navigateByUrl']);
    handler = jasmine.createSpyObj<HttpHandler>('HttpHandler', ['handle']);
    handler.handle.and.returnValue(of(new HttpResponse({status: 200})));
    interceptor = new AuthInterceptorService(auth, router);
  });

  for (const endpoint of [
    lighthouse,
    'https://api.platform.fastenhealth.com/v1/app/convert/ccda/to/fhir',
    'https://synthetic-provider.example.invalid/fhir/Patient',
  ]) {
    it(`does not export the native application credential to ${endpoint}`, () => {
      const request = new HttpRequest('GET', endpoint);
      interceptor.intercept(request, handler).subscribe();
      const sent = handler.handle.calls.mostRecent().args[0];
      expect(sent).toBe(request);
      expect(sent.headers.has('Authorization')).toBeFalse();
      expect(auth.GetAuthToken).not.toHaveBeenCalled();
    });
  }

  for (const status of [401, 403]) {
    it(`does not turn an external ${status} into local logout or a swallowed error`, () => {
      const failure = new HttpErrorResponse({status, error: {code: 'synthetic-external-denial'}});
      handler.handle.and.returnValue(throwError(() => failure));
      let observed: unknown;
      interceptor.intercept(new HttpRequest('GET', lighthouse), handler).subscribe({error: error => observed = error});
      expect(observed).toBe(failure);
      expect(auth.Logout).not.toHaveBeenCalled();
      expect(router.navigateByUrl).not.toHaveBeenCalled();
      expect(auth.GetAuthToken).not.toHaveBeenCalled();
    });
  }

  it('preserves explicitly supplied external provider authorization without substituting a native JWT', () => {
    const request = new HttpRequest('POST', lighthouse, {synthetic: true}, {
      headers: new HttpHeaders({Authorization: 'Bearer synthetic-external-provider-token'}),
    });
    interceptor.intercept(request, handler).subscribe();
    expect(handler.handle.calls.mostRecent().args[0]).toBe(request);
    expect(handler.handle.calls.mostRecent().args[0].headers.get('Authorization')).toBe('Bearer synthetic-external-provider-token');
    expect(auth.GetAuthToken).not.toHaveBeenCalled();
  });

  it('does not treat an API-prefixed sibling path as application authority', () => {
    const request = new HttpRequest('GET', application.origin + applicationPath + '-not-api/private');
    interceptor.intercept(request, handler).subscribe();
    expect(handler.handle.calls.mostRecent().args[0]).toBe(request);
    expect(auth.GetAuthToken).not.toHaveBeenCalled();
  });

  for (const path of [applicationPath, applicationPath + '/secure/account/me']) {
    it(`retains the native credential for the configured application API ${path}`, () => {
      interceptor.intercept(new HttpRequest('GET', application.origin + path), handler).subscribe();
      expect(handler.handle.calls.mostRecent().args[0].headers.get('Authorization')).toBe('Bearer synthetic-private-native-session');
      expect(auth.GetAuthToken).toHaveBeenCalledTimes(1);
    });
  }

  it('retains local session invalidation for an actual protected API 401', () => {
    handler.handle.and.returnValue(throwError(() => new HttpErrorResponse({status: 401})));
    interceptor.intercept(new HttpRequest('GET', application.origin + applicationPath + '/secure/account/me'), handler).subscribe();
    expect(auth.Logout).toHaveBeenCalledWith(false);
    expect(router.navigateByUrl).toHaveBeenCalledWith('/auth/signin');
  });

  const subpathLocation = {protocol: 'https:', host: 'synthetic-local.example.invalid:9443', pathname: '/family/web/auth/signin'};
  for (const {endpoint, authorized} of [
    {endpoint: '/family/api/secure/account/me', authorized: true},
    {endpoint: '/family/api', authorized: true},
    {endpoint: '/family/apiary/private', authorized: false},
    {endpoint: '/family/api/../not-api/private', authorized: false},
    {endpoint: '//foreign.example.invalid/family/api/secure/account/me', authorized: false},
    {endpoint: 'https://synthetic-user:synthetic-password@synthetic-local.example.invalid:9443/family/api/secure/account/me', authorized: false},
  ]) {
    it(`pins a reverse-proxy API authority and segment boundary: ${endpoint}`, () => {
      const scoped = new AuthInterceptorService(auth, router, subpathLocation);
      const request = new HttpRequest('GET', endpoint);
      scoped.intercept(request, handler).subscribe();
      const sent = handler.handle.calls.mostRecent().args[0];
      if (authorized) {
        expect(sent.headers.get('Authorization')).toBe('Bearer synthetic-private-native-session');
      } else {
        expect(sent).toBe(request);
        expect(auth.GetAuthToken).not.toHaveBeenCalled();
      }
    });
  }

  const nativeLocation = {origin: 'null', protocol: 'wails:', host: 'wails.localhost', pathname: '/index.html'};
  for (const endpoint of ['/api/secure/account/me', 'wails://wails.localhost/api/secure/account/me']) {
    it(`authenticates native application URLs with opaque serialized origin: ${endpoint}`, () => {
      const native = new AuthInterceptorService(auth, router, nativeLocation);
      expect(() => native.intercept(new HttpRequest('GET', endpoint), handler).subscribe()).not.toThrow();
      expect(handler.handle).toHaveBeenCalledTimes(1);
      if (handler.handle.calls.count() === 1) {
        expect(handler.handle.calls.mostRecent().args[0].headers.get('Authorization')).toBe('Bearer synthetic-private-native-session');
      }
    });
  }

  for (const endpoint of [
    'wails://foreign.localhost/api/secure/account/me',
    'custom://wails.localhost/api/secure/account/me',
    'https://wails.localhost/api/secure/account/me',
  ]) {
    it(`does not conflate native opaque origins or protocols: ${endpoint}`, () => {
      const native = new AuthInterceptorService(auth, router, nativeLocation);
      const request = new HttpRequest('GET', endpoint);
      expect(() => native.intercept(request, handler).subscribe()).not.toThrow();
      expect(handler.handle).toHaveBeenCalledTimes(1);
      if (handler.handle.calls.count() === 1) {
        expect(handler.handle.calls.mostRecent().args[0]).toBe(request);
      }
      expect(auth.GetAuthToken).not.toHaveBeenCalled();
    });
  }
});
