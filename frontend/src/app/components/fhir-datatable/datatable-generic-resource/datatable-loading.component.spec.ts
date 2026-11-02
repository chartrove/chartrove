import {Type} from '@angular/core';
import {ComponentFixture, TestBed, fakeAsync, tick} from '@angular/core/testing';
import {HttpClient, provideHttpClient, withXhr} from '@angular/common/http';
import {HttpTestingController, provideHttpClientTesting, TestRequest} from '@angular/common/http/testing';
import {provideRouter} from '@angular/router';
import {HTTP_CLIENT_TOKEN} from '../../../dependency-injection';
import {ResourceFhir} from '../../../models/fasten/resource_fhir';
import {FhirDatatableModule} from '../fhir-datatable.module';
import {DatatableGenericResourceComponent} from './datatable-generic-resource.component';
import {DatatableEncounterComponent} from './datatable-encounter.component';
import {DatatableFallbackComponent} from './datatable-fallback.component';

const sourceId = '11111111-1111-4111-8111-111111111111';
const cases: {
  name: string;
  component: Type<DatatableGenericResourceComponent>;
  resourceType: string;
  raw: (id: string) => any;
  loaded: string;
  recovered: string;
  latest: string;
  stale: string;
}[] = [
  {
    name: 'encounter', component: DatatableEncounterComponent, resourceType: 'Encounter',
    raw: id => ({resourceType: 'Encounter', id, status: 'finished',
      class: {system: 'http://terminology.hl7.org/CodeSystem/v3-ActCode', code: 'AMB'},
      type: [{text: `Synthetic ${id}`}], period: {start: '2024-01-01T00:00:00Z'}}),
    loaded: 'Synthetic loaded', recovered: 'Synthetic recovered',
    latest: 'Synthetic latest', stale: 'Synthetic stale',
  },
  {
    name: 'fallback', component: DatatableFallbackComponent, resourceType: 'RiskAssessment',
    raw: id => ({resourceType: 'RiskAssessment', id, status: 'final', subject: {reference: 'Patient/synthetic'}}),
    loaded: 'loaded', recovered: 'recovered', latest: 'latest', stale: 'stale',
  },
];

for (const example of cases) {
  describe(`${example.name} table request feedback`, () => {
    let fixture: ComponentFixture<DatatableGenericResourceComponent>;
    let http: HttpTestingController;
    let initial: TestRequest;

    function pageRequest(page: number): TestRequest {
      return http.expectOne(request => request.url.endsWith('/secure/resource/fhir') &&
        request.params.get('sourceResourceType') === example.resourceType &&
        request.params.get('sourceID') === sourceId && request.params.get('page') === `${page}`);
    }

    function resource(id: string): ResourceFhir {
      return new ResourceFhir({source_id: sourceId, source_resource_type: example.resourceType,
        source_resource_id: id, fhir_version: '4.0.1', resource_raw: example.raw(id)});
    }

    beforeEach(async () => {
      await TestBed.configureTestingModule({
        imports: [FhirDatatableModule],
        providers: [provideRouter([]), provideHttpClient(withXhr()), provideHttpClientTesting(),
          {provide: HTTP_CLIENT_TOKEN, useExisting: HttpClient}],
      }).compileComponents();
      fixture = TestBed.createComponent(example.component);
      fixture.componentInstance.sourceId = sourceId;
      fixture.componentInstance.resourceListType = example.resourceType;
      fixture.componentInstance.totalElements = 60;
      http = TestBed.inject(HttpTestingController);
      fixture.detectChanges();
      initial = pageRequest(0);
    });

    afterEach(() => {
      fixture.destroy();
      http.verify();
    });

    it('shows busy feedback instead of an empty result while the HTTP response is pending', () => {
      const view = fixture.nativeElement as HTMLElement;
      expect(view.querySelector('[role="status"]')?.textContent).toContain('Loading');
      expect(view.querySelector('ngx-datatable')?.getAttribute('aria-busy')).toBe('true');
      expect(view.textContent).not.toContain('No data to display');
    });

    it('shows the empty result only after a successful empty response', async () => {
      initial.flush({success: true, data: []});
      fixture.detectChanges();
      await fixture.whenStable();
      fixture.detectChanges();
      const view = fixture.nativeElement as HTMLElement;
      expect(view.querySelector('[role="status"]')).toBeNull();
      expect(view.querySelector('ngx-datatable')?.getAttribute('aria-busy')).toBe('false');
      expect(view.textContent).toContain('No data to display');
    });

    it('ends busy feedback and renders the returned synthetic record', async () => {
      initial.flush({success: true, data: [resource('loaded')]});
      fixture.detectChanges();
      await fixture.whenStable();
      fixture.detectChanges();
      const view = fixture.nativeElement as HTMLElement;
      expect(view.querySelector('[role="status"]')).toBeNull();
      expect(view.textContent).toContain(example.loaded);
      expect(view.querySelector('ngx-datatable')?.getAttribute('aria-busy')).toBe('false');
    });

    it('ends busy feedback on HTTP failure without presenting an empty result or the response payload', fakeAsync(() => {
      initial.flush({error: 'SYNTHETIC_PAYLOAD_SHOULD_NOT_BE_SHOWN'}, {status: 500, statusText: 'Server Error'});
      expect(() => tick()).not.toThrow();
      fixture.detectChanges();
      const view = fixture.nativeElement as HTMLElement;
      expect(view.querySelector('[role="alert"]')?.textContent).toContain('resources');
      expect(view.querySelector('[role="status"]')).toBeNull();
      expect(view.querySelector('ngx-datatable')?.getAttribute('aria-busy')).toBe('false');
      expect(view.textContent).not.toContain('No data to display');
      expect(view.textContent).not.toContain('SYNTHETIC_PAYLOAD_SHOULD_NOT_BE_SHOWN');
    }));

    it('can retry a failed page through the actual HTTP service and display the response', fakeAsync(() => {
      initial.flush({}, {status: 503, statusText: 'Unavailable'});
      expect(() => tick()).not.toThrow();
      fixture.detectChanges();
      const view = fixture.nativeElement as HTMLElement;
      const retry = Array.from(view.querySelectorAll('button')).find(button => /retry/i.test(button.textContent || ''));
      expect(retry).toBeDefined();
      if (!retry) return;
      retry.click();
      const request = pageRequest(0);
      request.flush({success: true, data: [resource('recovered')]});
      fixture.detectChanges();
      tick(16);
      fixture.detectChanges();
      expect(view.textContent).toContain(example.recovered);
      expect(view.querySelector('[role="alert"]')).toBeNull();
      expect(view.querySelector('[role="status"]')).toBeNull();
    }));

    it('cancels superseded paging requests so a late response cannot replace the current page', async () => {
      initial.flush({success: true, data: []});
      fixture.componentInstance.changePage({offset: 1});
      const old = pageRequest(1);
      fixture.componentInstance.changePage({offset: 2});
      const current = pageRequest(2);
      expect(old.cancelled).toBeTrue();
      current.flush({success: true, data: [resource('latest')]});
      if (!old.cancelled) old.flush({success: true, data: [resource('stale')]});
      fixture.detectChanges();
      await fixture.whenStable();
      fixture.detectChanges();
      const view = fixture.nativeElement as HTMLElement;
      expect(view.textContent).toContain(example.latest);
      expect(view.textContent).not.toContain(example.stale);
      expect(view.querySelector('[role="status"]')).toBeNull();
    });

    it('aborts the pending HTTP request when its table is destroyed', () => {
      fixture.destroy();
      expect(initial.cancelled).toBeTrue();
    });
  });
}
