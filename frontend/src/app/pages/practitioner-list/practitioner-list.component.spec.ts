import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PractitionerListComponent } from './practitioner-list.component';
import { HTTP_CLIENT_TOKEN } from 'src/app/dependency-injection';
import { HttpClient, provideHttpClient, withInterceptorsFromDi, withXhr } from '@angular/common/http';
import {HttpTestingController, provideHttpClientTesting} from '@angular/common/http/testing';

describe('PractitionerListComponent', () => {
  let component: PractitionerListComponent;
  let fixture: ComponentFixture<PractitionerListComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
    imports: [PractitionerListComponent],
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
    fixture = TestBed.createComponent(PractitionerListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  afterEach(() => TestBed.inject(HttpTestingController).verify());

  it('finishes loading an empty practitioner list', () => {
    const request = TestBed.inject(HttpTestingController).expectOne(req =>
      req.urlWithParams.endsWith('/api/secure/resource/fhir?sourceResourceType=Practitioner'));
    expect(request.request.method).toBe('GET');
    request.flush({success: true, data: []});
    const favorites = TestBed.inject(HttpTestingController).expectOne(req =>
      req.urlWithParams.endsWith('/api/secure/user/favorites?resource_type=Practitioner'));
    expect(favorites.request.method).toBe('GET');
    favorites.flush({success: true, data: []});
    fixture.detectChanges();
    expect(component.loading).toBeFalse();
    expect(component.practitioners).toEqual([]);
  });

  it('selects and removes only the exact successful source in an out-of-order batch', () => {
    const http = TestBed.inject(HttpTestingController);
    const sourceA = '11111111-1111-4111-8111-111111111111';
    const sourceB = '22222222-2222-4222-8222-222222222222';
    const sourceC = '33333333-3333-4333-8333-333333333333';
    const rows = [sourceA, sourceB, sourceC].map((source, i) => ({
      source_id: source, source_resource_type: 'Practitioner', source_resource_id: i === 2 ? 'unique-id' : 'shared-id',
      resource_raw: {resourceType: 'Practitioner', id: i === 2 ? 'unique-id' : 'shared-id', name: [{text: `Synthetic clinician ${i}`}]}
    }));
    http.expectOne(req => req.urlWithParams.endsWith('/api/secure/resource/fhir?sourceResourceType=Practitioner'))
      .flush({success: true, data: rows});
    http.expectOne(req => req.urlWithParams.endsWith('/api/secure/user/favorites?resource_type=Practitioner'))
      .flush({success: true, data: []});
    const [a, b] = component.practitioners;
    component.toggleSelection(a);
    expect(component.isSelected(b)).toBeFalse();
    expect(component.trackByPractitioner(0, a)).not.toBe(component.trackByPractitioner(1, b));
    component.selectedPractitioners = [...component.practitioners];
    spyOn(window, 'confirm').and.returnValue(true);
    spyOn(window, 'alert');
    component.deleteSelectedPractitioners();
    const deletes = http.match(req => req.method === 'DELETE');
    expect(deletes.length).toBe(3);
    deletes.find(r => r.request.params.get('source_id') === sourceB)!.flush({success: true});
    deletes.find(r => r.request.params.get('source_id') === sourceC)!.flush({success: true});
    deletes.find(r => r.request.params.get('source_id') === sourceA)!.flush({success: false}, {status: 503, statusText: 'Synthetic unavailable'});
    expect(component.practitioners.map(p => p.source_id)).toEqual([sourceA]);
    expect(component.selectedPractitioners).toEqual([]);
    http.expectOne(req => req.urlWithParams.endsWith('/api/secure/resource/fhir?sourceResourceType=Practitioner'))
      .flush({success: true, data: [rows[0]]});
    // Reloading records reuses the already loaded favorites; it does not fetch them again.
    http.expectNone(req => req.urlWithParams.endsWith('/api/secure/user/favorites?resource_type=Practitioner'));
    expect(component.practitioners.map(p => p.source_id)).toEqual([sourceA]);
  });
});
