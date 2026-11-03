import {Type} from '@angular/core';
import {ComponentFixture, TestBed} from '@angular/core/testing';
import {HttpClient, provideHttpClient, withXhr} from '@angular/common/http';
import {HttpTestingController, provideHttpClientTesting, TestRequest} from '@angular/common/http/testing';
import {provideRouter} from '@angular/router';
import {HTTP_CLIENT_TOKEN} from '../../../dependency-injection';
import {ResourceFhir} from '../../../models/fasten/resource_fhir';
import {FhirDatatableModule} from '../fhir-datatable.module';
import {DatatableGenericResourceComponent} from './datatable-generic-resource.component';
import {DatatableEncounterComponent} from './datatable-encounter.component';
import {DatatableFallbackComponent} from './datatable-fallback.component';

const syntheticSource = '11111111-1111-4111-8111-111111111111';
const tables: Array<{name: string; component: Type<DatatableGenericResourceComponent>; type: string}> = [
  {name: 'clinical', component: DatatableEncounterComponent, type: 'Encounter'},
  {name: 'fallback', component: DatatableFallbackComponent, type: 'RiskAssessment'},
];

for (const example of tables) {
  describe(`${example.name} table browser interactions`, () => {
    let fixture: ComponentFixture<DatatableGenericResourceComponent>;
    let http: HttpTestingController;
    let initial: TestRequest;

    function request(page: number): TestRequest {
      return http.expectOne(value => value.url.endsWith('/secure/resource/fhir') &&
        value.params.get('sourceResourceType') === example.type &&
        value.params.get('sourceID') === syntheticSource && value.params.get('page') === `${page}`);
    }

    function record(id: string): ResourceFhir {
      return new ResourceFhir({source_id: syntheticSource, source_resource_type: example.type,
        source_resource_id: id, fhir_version: '4.0.1', resource_raw: {
          resourceType: example.type, id, status: example.type === 'Encounter' ? 'finished' : 'final',
          class: {system: 'http://terminology.hl7.org/CodeSystem/v3-ActCode', code: 'AMB'},
          type: [{text: `Synthetic ${id}`}], subject: {reference: 'Patient/synthetic'},
          period: {start: '2024-01-01T12:00:00+03:00'},
          syntheticUnknown: {zero: 0, flag: false, text: 'Épreuve'},
        }});
    }

    async function settle(): Promise<void> {
      fixture.detectChanges();
      await fixture.whenStable();
      fixture.detectChanges();
    }

    function clickRow(label: string): void {
      const rows = Array.from((fixture.nativeElement as HTMLElement).querySelectorAll<HTMLElement>('.datatable-body-row'));
      const row = rows.find(value => value.textContent.includes(label));
      expect(row).withContext(`Rendered ${label} row`).toBeDefined();
      if (!row) throw new Error(`Missing rendered ${label} row`);
      const cell = row.querySelector<HTMLElement>('.datatable-body-cell');
      expect(cell).not.toBeNull();
      if (!cell) throw new Error('Missing actual table cell');
      cell.click();
      fixture.detectChanges();
    }

    beforeEach(async () => {
      await TestBed.configureTestingModule({imports: [FhirDatatableModule],
        providers: [provideRouter([]), provideHttpClient(withXhr()), provideHttpClientTesting(),
          {provide: HTTP_CLIENT_TOKEN, useExisting: HttpClient}],
      }).compileComponents();
      fixture = TestBed.createComponent(example.component);
      fixture.nativeElement.style.width = '900px';
      fixture.componentInstance.sourceId = syntheticSource;
      fixture.componentInstance.resourceListType = example.type;
      fixture.componentInstance.totalElements = 60;
      fixture.componentInstance.disabledResourceIds = ['disabled'];
      http = TestBed.inject(HttpTestingController);
      fixture.detectChanges();
      initial = request(0);
    });

    afterEach(() => { fixture.destroy(); http.verify(); });

    it('selects the clicked record, rejects a disabled record and preserves raw fields', async () => {
      const records = [record('eligible'), record('disabled')];
      const raw = JSON.stringify(records.map(value => value.resource_raw));
      const selected = jasmine.createSpy('selectionChanged');
      fixture.componentInstance.selectionChanged.subscribe(selected);
      initial.flush({success: true, data: records});
      await settle();
      clickRow('disabled');
      expect(selected).not.toHaveBeenCalled();
      clickRow('eligible');
      expect(selected).toHaveBeenCalledTimes(1);
      const result = selected.calls.mostRecent().args[0];
      expect(result.source_resource_id).toBe('eligible');
      expect(result.source_id).toBe(syntheticSource);
      expect(JSON.stringify(records.map(value => value.resource_raw))).toBe(raw);
      expect(JSON.stringify(result.resource.resource_raw)).toBe(JSON.stringify(records[0].resource_raw));
    });

    it('loads the clicked next page without an undefined selection or stale rows', async () => {
      const selected = jasmine.createSpy('selectionChanged');
      fixture.componentInstance.selectionChanged.subscribe(selected);
      initial.flush({success: true, data: [record('first')]});
      await settle();
      clickRow('first');
      expect(selected).toHaveBeenCalledTimes(1);
      const next = (fixture.nativeElement as HTMLElement).querySelector<HTMLElement>('[aria-label="go to next page"]');
      expect(next).not.toBeNull();
      if (!next) throw new Error('Missing actual next-page control');
      next.click();
      fixture.detectChanges();
      const pending = request(1);
      expect(fixture.componentInstance.currentPage.offset).toBe(1);
      expect(fixture.componentInstance.loading).toBeTrue();
      expect(fixture.nativeElement.textContent).not.toContain('Synthetic first');
      pending.flush({success: true, data: [record('second')]});
      await settle();
      const rows = (fixture.nativeElement as HTMLElement).querySelector('.datatable-body');
      expect(rows?.textContent).toContain('second');
      expect(rows?.textContent).not.toContain('first');
      expect(selected).toHaveBeenCalledTimes(1);
      clickRow('second');
      expect(selected).toHaveBeenCalledTimes(2);
      expect(selected.calls.mostRecent().args[0].source_resource_id).toBe('second');
    });
  });
}
