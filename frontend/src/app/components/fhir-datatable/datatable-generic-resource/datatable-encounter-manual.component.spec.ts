import {ComponentFixture, TestBed} from '@angular/core/testing';
import {HttpClient, provideHttpClient, withXhr} from '@angular/common/http';
import {provideHttpClientTesting} from '@angular/common/http/testing';
import {provideRouter} from '@angular/router';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {of} from 'rxjs';
import {HTTP_CLIENT_TOKEN} from '../../../dependency-injection';
import {FastenApiService} from '../../../services/fasten-api.service';
import {ResourceFhir} from '../../../models/fasten/resource_fhir';
import {MedicalRecordWizardAddEncounterComponent} from '../../medical-record-wizard-add-encounter/medical-record-wizard-add-encounter.component';
import {GenerateR4ResourceLookup} from '../../medical-record-wizard/medical-record-wizard.utilities';
import {FhirDatatableModule} from '../fhir-datatable.module';
import {DatatableEncounterComponent} from './datatable-encounter.component';
import {FORMATTERS} from './utils';

const sourceId = '11111111-1111-4111-8111-111111111111';

describe('manual Encounter table display', () => {
  let table: ComponentFixture<DatatableEncounterComponent>;
  let wizard: ComponentFixture<MedicalRecordWizardAddEncounterComponent>;
  let submitted: any;

  beforeEach(async () => {
    submitted = undefined;
    await TestBed.configureTestingModule({
      imports: [FhirDatatableModule, MedicalRecordWizardAddEncounterComponent],
      providers: [provideRouter([]), provideHttpClient(withXhr()), provideHttpClientTesting(),
        {provide: HTTP_CLIENT_TOKEN, useExisting: HttpClient},
        {provide: NgbActiveModal, useValue: {close: value => submitted = value}},
        {provide: FastenApiService, useValue: {
          queryResources: query => {
            if (query.from !== 'Encounter') throw new Error('Unexpected synthetic query');
            return of({success: true, data: [{value: 0}]});
          },
          getResources: type => {
            if (type !== 'Encounter') throw new Error('Unexpected synthetic resource');
            return of([]);
          }
        }}]
    }).compileComponents();
    wizard = TestBed.createComponent(MedicalRecordWizardAddEncounterComponent);
    wizard.detectChanges();
    table = TestBed.createComponent(DatatableEncounterComponent);
    table.componentInstance.sourceId = sourceId;
    table.componentInstance.resourceListType = 'Encounter';
    table.componentInstance.totalElements = 1;
    table.detectChanges();
  });

  afterEach(() => { wizard.destroy(); table.destroy(); });

  function createManualEncounter(withEnd = true) {
    const component = wizard.componentInstance;
    component.changeTab('create');
    component.newEncounterForm.patchValue({
      code: {text: 'Synthetic office visit'},
      period_start: {year: 2024, month: 2, day: 29},
      period_end: withEnd ? {year: 2024, month: 3, day: 1} : null
    });
    component.submit();
    expect(submitted?.data).toBeDefined();
    submitted.data.source_resource_id = 'synthetic-manual-encounter';
    const lookup = GenerateR4ResourceLookup({encounter: submitted, attachments: [],
      organizations: [], practitioners: [], medications: [], procedures: [], labresults: []} as any);
    // Match the serialized raw resource returned after persistence, not the display model.
    return JSON.parse(JSON.stringify(lookup['Encounter']['synthetic-manual-encounter']));
  }

  async function render(raw: any) {
    const original = JSON.stringify(raw);
    table.componentInstance.renderList([new ResourceFhir({source_id: sourceId,
      source_resource_type: 'Encounter', source_resource_id: raw.id,
      fhir_version: '4.0.1', resource_raw: raw})]);
    table.detectChanges();
    await table.whenStable();
    table.detectChanges();
    expect(JSON.stringify(raw)).toBe(original);
    return table.nativeElement.textContent;
  }

  it('shows the service description produced by the real manual wizard in the row', async () => {
    const raw = createManualEncounter();
    expect(raw.serviceType.text).toBe('Synthetic office visit');
    expect(raw.type).toBeUndefined();
    const text = await render(raw);
    expect(text).toContain('Synthetic office visit');
    expect(text).toContain('2024-02-29');
    expect(text).toContain('2024-03-01');
  });

  it('does not invent an end date when the manual wizard leaves End blank', async () => {
    const raw = createManualEncounter(false);
    expect(raw.period.end).toBeUndefined();
    await render(raw);
    expect(table.componentInstance.rows[0].Period).toContain('2024-02-29');
    expect(table.componentInstance.rows[0].Period).toContain('Not provided');
  });

  it('retains an existing provider type when both type and serviceType are present', async () => {
    const raw = createManualEncounter();
    raw.type = [{text: 'Synthetic provider type'}];
    const text = await render(raw);
    expect(text).toContain('Synthetic provider type');
    expect(table.componentInstance.rows[0].Encounter).toBe('Synthetic provider type');
  });

  it('renders a coded service description when there is no type', async () => {
    const raw = createManualEncounter();
    raw.serviceType = {coding: [{system: 'urn:chartrove:synthetic', code: 'example', display: 'Synthetic coded visit'}]};
    expect(await render(raw)).toContain('example: Synthetic coded visit');
  });

  it('preserves selection identity and the raw resource', async () => {
    const raw = createManualEncounter();
    await render(raw);
    let selected: any;
    table.componentInstance.selectionChanged.subscribe(value => selected = value);
    table.componentInstance.onSelect({selected: [table.componentInstance.rows[0]]});
    expect(selected.source_id).toBe(sourceId);
    expect(selected.source_resource_type).toBe('Encounter');
    expect(selected.source_resource_id).toBe('synthetic-manual-encounter');
    expect(selected.resource.resource_raw).toEqual(raw);
  });
});

describe('period formatter missing endpoints', () => {
  const start = new Date(2024, 1, 29).toISOString();
  const end = new Date(2024, 2, 1).toISOString();
  for (const value of [undefined, null, '']) {
    it(`does not invent a start for ${JSON.stringify(value)}`, () => {
      expect(FORMATTERS.period({start: value, end})).toContain('Not provided -> 2024-03-01');
    });
    it(`does not invent an end for ${JSON.stringify(value)}`, () => {
      expect(FORMATTERS.period({start, end: value})).toContain('2024-02-29');
      expect(FORMATTERS.period({start, end: value})).toContain(' -> Not provided');
    });
  }
  it('retains both supplied endpoints', () => {
    const value = FORMATTERS.period({start, end});
    expect(value).toContain('2024-02-29');
    expect(value).toContain('2024-03-01');
    expect(value).not.toContain('Not provided');
  });
  it('retains the empty result for an absent Period', () => {
    expect(FORMATTERS.period(undefined)).toBe('');
    expect(FORMATTERS.period(null)).toBe('');
  });
});
