import { ComponentFixture, TestBed } from '@angular/core/testing';

import { SourceDetailComponent } from './source-detail.component';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import {RouterTestingModule} from '@angular/router/testing';
import {ActivatedRoute, convertToParamMap, RouterModule} from '@angular/router';
import {HTTP_CLIENT_TOKEN} from '../../dependency-injection';
import { HttpClient, provideHttpClient, withInterceptorsFromDi, withXhr } from '@angular/common/http';
import { FhirDatatableModule } from 'src/app/components/fhir-datatable/fhir-datatable.module';
import {ResourceFhir} from '../../models/fasten/resource_fhir';

describe('SourceDetailComponent', () => {
  let component: SourceDetailComponent;
  let fixture: ComponentFixture<SourceDetailComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
    declarations: [SourceDetailComponent],
    imports: [RouterTestingModule, RouterModule, FhirDatatableModule],
    providers: [
        {
            provide: ActivatedRoute,
            useValue: { snapshot: { paramMap: convertToParamMap({ 'source_id': 'b64.c291cmNlOmF0aGVuYTphLTgwMDAwLkUtMTQ1NDU' }) } }
        },
        {
            provide: HTTP_CLIENT_TOKEN,
            useClass: HttpClient,
        },
        provideHttpClient(withXhr(), withInterceptorsFromDi()),
        provideHttpClientTesting(),
    ]
})
    .compileComponents();

    fixture = TestBed.createComponent(SourceDetailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('displays a text-only SMART Patient name and leaves missing addresses empty', () => {
    component.selectedPatient = new ResourceFhir({resource_raw: {resourceType: 'Patient', id: 'synthetic-only',
      name: [{text: 'Chartrove Synthetic Patient'}]}});
    expect(component.getPatientName()).toBe('Chartrove Synthetic Patient');
    expect(component.getPatientAddress()).toEqual([]);
  });

  it('keeps structured names and partial addresses readable', () => {
    component.selectedPatient = new ResourceFhir({resource_raw: {name: [{given: ['Synthetic'], family: 'Patient'}],
      address: [{line: ['Synthetic Street'], city: 'Example City'}]}});
    expect(component.getPatientName()).toBe('Patient, Synthetic');
    expect(component.getPatientAddress()).toEqual(['Synthetic Street', 'Example City']);
  });

  it('uses a clear fallback when a Patient has no name', () => {
    component.selectedPatient = new ResourceFhir({resource_raw: {resourceType: 'Patient', id: 'synthetic-only'}});
    expect(component.getPatientName()).toBe('Name not provided');
  });

  for (const example of [
    {name: 'absent', raw: {}},
    {name: 'explicitly null', raw: {birthDate: null}},
    {name: 'unparseable', raw: {birthDate: 'synthetic-invalid-date'}},
  ]) {
    it(`shows a missing-value label for an ${example.name} birth date without inventing an age`, () => {
      component.selectedSource = {id: 'synthetic-age-source', brand_id: '00000000-0000-0000-0000-000000000000'} as any;
      component.selectedPatient = new ResourceFhir({source_resource_id: 'synthetic-age-patient',
        resource_raw: {resourceType: 'Patient', id: 'synthetic-age-patient', ...example.raw}});
      const before = JSON.stringify(component.selectedPatient.resource_raw);
      fixture.detectChanges();
      const view = fixture.nativeElement as HTMLElement;
      const ageLabel = Array.from(view.querySelectorAll('div')).find(element => element.textContent.trim() === 'Age:');
      expect(ageLabel?.nextElementSibling?.textContent.trim()).toBe('Not provided');
      expect(view.textContent).not.toContain('NaN');
      expect(view.textContent).not.toContain('0 Years Old');
      expect(JSON.stringify(component.selectedPatient.resource_raw)).toBe(before);
    });
  }

  for (const example of [
    {name: 'a zero-year age', date: '2026-06-14', display: '0 Years Old'},
    {name: 'an existing adult age', date: '2000-06-15', display: '26 Years Old'},
  ]) {
    it(`retains ${example.name} in the rendered profile`, () => {
      jasmine.clock().install();
      try {
        jasmine.clock().mockDate(new Date('2026-06-15T00:00:00Z'));
        component.selectedSource = {id: 'synthetic-age-source', brand_id: '00000000-0000-0000-0000-000000000000'} as any;
        component.selectedPatient = new ResourceFhir({source_resource_id: 'synthetic-age-patient',
          resource_raw: {resourceType: 'Patient', id: 'synthetic-age-patient', birthDate: example.date}});
        fixture.detectChanges();
        const view = fixture.nativeElement as HTMLElement;
        const ageLabel = Array.from(view.querySelectorAll('div')).find(element => element.textContent.trim() === 'Age:');
        expect(ageLabel?.nextElementSibling?.textContent.trim()).toBe(example.display);
      } finally {
        jasmine.clock().uninstall();
      }
    });
  }
});
