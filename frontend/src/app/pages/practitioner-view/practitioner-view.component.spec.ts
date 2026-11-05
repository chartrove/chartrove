import { ComponentFixture, TestBed } from '@angular/core/testing';
import { PractitionerViewComponent } from './practitioner-view.component';
import { ActivatedRoute } from '@angular/router';
import { BehaviorSubject, of } from 'rxjs';
import { HTTP_CLIENT_TOKEN } from 'src/app/dependency-injection';
import { HttpClient, provideHttpClient, withInterceptorsFromDi, withXhr } from '@angular/common/http';
import {HttpTestingController, provideHttpClientTesting} from '@angular/common/http/testing';

describe('PractitionerViewComponent', () => {
  let component: PractitionerViewComponent;
  let queryParams: BehaviorSubject<Record<string, string>>;
  let fixture: ComponentFixture<PractitionerViewComponent>;

  beforeEach(async () => {
    queryParams = new BehaviorSubject<Record<string, string>>({});
    await TestBed.configureTestingModule({
    imports: [PractitionerViewComponent],
    providers: [
        {
            provide: ActivatedRoute,
            useValue: {
                snapshot: {
                    paramMap: { get: (key: string) => 'test-practitioner-id' },
                },
                params: of({ id: 'test-practitioner-id' }),
                queryParams: queryParams.asObservable(),
            },
        },
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
    fixture = TestBed.createComponent(PractitionerViewComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  afterEach(() => TestBed.inject(HttpTestingController).verify());

  it('reports a missing practitioner after a successful empty response', () => {
    const request = TestBed.inject(HttpTestingController).expectOne(req =>
      req.urlWithParams.endsWith('/api/secure/resource/fhir?sourceResourceType=Practitioner'));
    expect(request.request.method).toBe('GET');
    request.flush({success: true, data: []});
    fixture.detectChanges();
    expect(component.isLoading).toBeFalse();
    expect(component.loadError).toBe('Practitioner not found');
    expect(fixture.nativeElement.textContent).toContain('Practitioner not found');
  });
  const sourceA = '11111111-1111-4111-8111-111111111111';
  const sourceB = '22222222-2222-4222-8222-222222222222';
  const rows = [sourceA, sourceB].map((source, i) => ({
    source_id: source, source_resource_type: 'Practitioner', source_resource_id: 'test-practitioner-id',
    resource_raw: {resourceType: 'Practitioner', id: 'test-practitioner-id', name: [{text: `Synthetic clinician ${i}`}], qualification: []}
  }));

  it('keeps a source-specific route bound when an older response arrives last', () => {
    const http = TestBed.inject(HttpTestingController);
    const log = spyOn(console, 'log');
    queryParams.next({source_id: sourceB});
    const requests = http.match(req => req.urlWithParams.endsWith('/api/secure/resource/fhir?sourceResourceType=Practitioner'));
    expect(requests.length).toBe(2);
    requests[1].flush({success: true, data: rows});
    requests[0].flush({success: true, data: rows});
    fixture.detectChanges();
    expect(component.practitioner?.source_id).toBe(sourceB);
    expect(component.practitioner?.full_name).toBe('Synthetic clinician 1');
    expect(component.loadError).toBe('');
    expect(component.isLoading).toBeFalse();
    expect(log.calls.allArgs().flat().some(arg => Array.isArray(arg) && arg.some(p => p?.full_name?.startsWith('Synthetic clinician')))).toBeFalse();
  });

  it('refuses to select the first equal ID without a source', () => {
    TestBed.inject(HttpTestingController).expectOne(req => req.urlWithParams.endsWith('/api/secure/resource/fhir?sourceResourceType=Practitioner'))
      .flush({success: true, data: rows});
    fixture.detectChanges();
    expect(component.practitioner).toBeNull();
    expect(component.loadError).toContain('multiple sources');
    expect(component.isLoading).toBeFalse();
  });

});
