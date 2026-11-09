import {HttpClient} from '@angular/common/http';
import {HttpTestingController, provideHttpClientTesting} from '@angular/common/http/testing';
import {TestBed} from '@angular/core/testing';
import {Router} from '@angular/router';
import {AppModule} from './app.module';
import {HTTP_CLIENT_TOKEN} from './dependency-injection';
import {APPLICATION_HTTP_LOCATION} from './services/auth-interceptor.service';
import {AuthService} from './services/auth.service';

describe('Application HTTP providers with a synthetic native URL authority', () => {
  let http: HttpTestingController;
  let auth: jasmine.SpyObj<AuthService>;

  beforeEach(() => {
    auth = jasmine.createSpyObj<AuthService>('AuthService', ['GetAuthToken', 'Logout']);
    auth.GetAuthToken.and.returnValue('synthetic-native-provider-session');
    TestBed.configureTestingModule({
      imports: [AppModule],
      providers: [
        provideHttpClientTesting(),
        {provide: AuthService, useValue: auth},
        {provide: Router, useValue: jasmine.createSpyObj<Router>('Router', ['navigateByUrl'])},
        {provide: APPLICATION_HTTP_LOCATION, useValue: {protocol: 'wails:', host: 'wails.localhost', pathname: '/index.html'}},
      ],
    });
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  for (const {name, getClient} of [
    {name: 'HttpClient', getClient: () => TestBed.inject(HttpClient)},
    {name: 'HTTP_CLIENT_TOKEN', getClient: () => TestBed.inject(HTTP_CLIENT_TOKEN) as unknown as HttpClient},
  ]) {
    for (const url of ['/api/secure/source/manual', 'wails://wails.localhost/api/secure/source/manual']) {
      it(`keeps native API authentication through ${name}: ${url}`, () => {
        const body = {resourceType: 'Patient', id: 'synthetic-native-only'};
        getClient().post(url, body).subscribe();
        const request = http.expectOne(url);
        expect(request.request.headers.get('Authorization')).toBe('Bearer synthetic-native-provider-session');
        expect(request.request.body).toBe(body);
        request.flush({success: true});
      });
    }

    it(`does not equate different custom-scheme hosts through ${name}`, () => {
      const url = 'wails://foreign.localhost/api/secure/source/manual';
      getClient().get(url).subscribe();
      const request = http.expectOne(url);
      expect(request.request.headers.has('Authorization')).toBeFalse();
      request.flush({synthetic: true});
      expect(auth.GetAuthToken).not.toHaveBeenCalled();
      expect(auth.Logout).not.toHaveBeenCalled();
    });
  }
});
