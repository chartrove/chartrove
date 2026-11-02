import { ComponentFixture, TestBed } from '@angular/core/testing';
import { PractitionerCreateComponent } from './practitioner-create.component';
import { HTTP_CLIENT_TOKEN } from 'src/app/dependency-injection';
import { HttpClient, provideHttpClient, withInterceptorsFromDi, withXhr } from '@angular/common/http';
import {HttpTestingController, provideHttpClientTesting} from '@angular/common/http/testing';
import {Router} from '@angular/router';

describe('PractitionerCreateComponent', () => {
  let component: PractitionerCreateComponent;
  let fixture: ComponentFixture<PractitionerCreateComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
    imports: [PractitionerCreateComponent],
    providers: [
        {
            provide: HTTP_CLIENT_TOKEN,
            useClass: HttpClient,
        },
        provideHttpClient(withXhr(), withInterceptorsFromDi()),
        provideHttpClientTesting(),
    ]
}).compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(PractitionerCreateComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  afterEach(() => TestBed.inject(HttpTestingController).verify());

  it('initializes a new practitioner form with an empty count response', () => {
    const request = TestBed.inject(HttpTestingController).expectOne(req => req.url.endsWith('/api/secure/query'));
    expect(request.request.method).toBe('POST');
    expect(request.request.body.from).toBe('Practitioner');
    request.flush({success: true, data: []});
    fixture.detectChanges();
    expect(component.totalPractitioners).toBe(0);
  });

  function preparePrivateCreation() {
    const http = TestBed.inject(HttpTestingController);
    http.expectOne(req => req.url.endsWith('/api/secure/query'))
      .flush({success: true, data: []});
    const logs = [spyOn(console, 'log'), spyOn(console, 'error'), spyOn(console, 'warn')];
    spyOn(window, 'alert');
    spyOn(TestBed.inject(Router), 'navigate').and.resolveTo(true);
    component.newPractitionerForm.patchValue({
      name: 'Canary PRIVATE_CANARY, Avery',
      profession: {text: 'Synthetic specialty', identifier: []},
    });
    return {http, logs};
  }

  it('creates a practitioner without logging its clinical payload or server response', () => {
    const {http, logs} = preparePrivateCreation();
    component.submit();
    expect(component.isSubmitting).toBeTrue();
    const request = http.expectOne(req => req.url.endsWith('/secure/practitioners'));
    expect(request.request.body.resource.resourceType).toBe('Practitioner');
    expect(request.request.body.resource.name[0].text).toBe('Canary PRIVATE_CANARY, Avery');
    request.flush({success: true, data: {private_details: 'response_PRIVATE_CANARY'}});
    expect(component.isSubmitting).toBeFalse();
    expect(TestBed.inject(Router).navigate).toHaveBeenCalledWith(['/practitioners']);
    expect(JSON.stringify(logs.flatMap(log => log.calls.allArgs()))).not.toContain('PRIVATE_CANARY');
  });

  it('preserves a failed practitioner form for retry without logging the raw error', () => {
    const {http, logs} = preparePrivateCreation();
    component.submit();
    http.expectOne(req => req.url.endsWith('/secure/practitioners'))
      .flush({error: 'error_PRIVATE_CANARY'}, {status: 503, statusText: 'Unavailable'});
    expect(component.isSubmitting).toBeFalse();
    expect(component.newPractitionerForm.get('name').value).toBe('Canary PRIVATE_CANARY, Avery');
    expect(TestBed.inject(Router).navigate).not.toHaveBeenCalled();
    component.submit();
    const retry = http.expectOne(req => req.url.endsWith('/secure/practitioners'));
    expect(retry.request.body.resource.name[0].text).toBe('Canary PRIVATE_CANARY, Avery');
    retry.flush({success: true});
    expect(component.isSubmitting).toBeFalse();
    expect(TestBed.inject(Router).navigate).toHaveBeenCalledTimes(1);
    expect(JSON.stringify(logs.flatMap(log => log.calls.allArgs()))).not.toContain('PRIVATE_CANARY');
  });
});
