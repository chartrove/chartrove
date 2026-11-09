import {HttpClient} from '@angular/common/http';
import {HttpTestingController, provideHttpClientTesting, TestRequest} from '@angular/common/http/testing';
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

  async function encodedRequest(url: string): Promise<TestRequest> {
    for (let attempt = 0; attempt < 250; attempt++) {
      const requests = http.match(url);
      if (requests.length) {
        expect(requests.length).toBe(1);
        return requests[0];
      }
      await new Promise(resolve => setTimeout(resolve, 10));
    }
    throw new Error('Native encoded request did not reach the mocked transport');
  }

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

    it(`materializes native multipart without changing bytes/receiver or reselecting a replacement JWT through ${name}`, async () => {
      const url = 'wails://wails.localhost/api/secure/source/manual';
      const raw = '{"resourceType":"Patient","id":"synthetic-native-exact","decimal":1.234567890123456789}';
      const form = new FormData();
      form.append('file', new Blob([raw], {type: 'application/fhir+json'}), 'synthetic.json');
      form.append('expected_owner_id', '11111111-1111-4111-8111-111111111111');
      getClient().post(url, form).subscribe();
      auth.GetAuthToken.and.returnValue('synthetic-replacement-session');
      const request = await encodedRequest(url);
      expect(request.request.headers.get('Authorization')).toBe('Bearer synthetic-native-provider-session');
      expect(request.request.body instanceof ArrayBuffer).toBeTrue();
      const contentType = request.request.headers.get('Content-Type');
      if (!contentType) throw new Error('Native multipart content type is missing');
      const parsed = await new Response(request.request.body, {headers: {'Content-Type': contentType}}).formData();
      expect(parsed.get('expected_owner_id')).toBe('11111111-1111-4111-8111-111111111111');
      const file = parsed.get('file');
      if (!(file instanceof File)) throw new Error('Native multipart file is missing');
      expect(await file.text()).toBe(raw);
      request.flush({success: true});
      expect(auth.GetAuthToken).toHaveBeenCalledTimes(1);
    });

    it(`does not start native transport after encoding subscription is cancelled through ${name}`, async () => {
      const url = 'wails://wails.localhost/api/secure/source/manual';
      const form = new FormData();
      form.append('file', new Blob(['synthetic-cancelled-encoding']), 'synthetic.json');
      const subscription = getClient().post(url, form).subscribe();
      subscription.unsubscribe();
      await new Promise(resolve => setTimeout(resolve, 100));
      http.expectNone(url);
      expect(auth.Logout).not.toHaveBeenCalled();
    });

    it(`does not encode or authorize external multipart through ${name}`, () => {
      const url = 'https://synthetic-provider.example.invalid/optional-file';
      const form = new FormData();
      form.append('synthetic', 'untouched');
      getClient().post(url, form).subscribe();
      const request = http.expectOne(url);
      expect(request.request.body).toBe(form);
      expect(request.request.headers.has('Authorization')).toBeFalse();
      request.flush({success: true});
      expect(auth.GetAuthToken).not.toHaveBeenCalled();
    });

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
