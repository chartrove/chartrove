import {TestBed} from '@angular/core/testing';
import {ObservationVisualizationComponent} from './observation-visualization.component';
import {ObservationModel} from '../../../../../lib/models/resources/observation-model';

// A later categorical/absent reading must not become a numeric chart endpoint
// merely because the first same-code observation supports a bar chart.
describe('Observation series visualization compatibility', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({imports: [ObservationVisualizationComponent]}).compileComponents();
  });

  function render(values: object[]) {
    const raw = values.map((value, index) => ({resourceType: 'Observation', id: `synthetic-series-${index}`,
      status: 'final', code: {coding: [{system: 'https://example.invalid/codes', code: 'same-series'}]},
      effectiveDateTime: `2020-01-0${index + 1}T00:00:00Z`, ...value}));
    const before = JSON.stringify(raw);
    const models = raw.map(value => new ObservationModel(value));
    const fixture = TestBed.createComponent(ObservationVisualizationComponent);
    fixture.componentInstance.observations = models;
    expect(() => fixture.detectChanges()).not.toThrow();
    expect(JSON.stringify(raw)).toBe(before);
    return {fixture, view: fixture.nativeElement as HTMLElement, models};
  }

  function expectTable(view: HTMLElement, results: string[]) {
    expect(view.querySelector('observation-bar-chart')).toBeNull();
    expect(view.querySelector('observation-table')).not.toBeNull();
    const rows = Array.from(view.querySelectorAll('observation-table tbody tr')) as HTMLTableRowElement[];
    expect(rows.map(row => row.cells[1].textContent.trim())).toEqual(results);
  }

  it('keeps a later numeric-looking code out of the quantitative chart', () => {
    const {view, models} = render([{valueQuantity: {value: 12, unit: 'mg'}},
      {valueCodeableConcept: {coding: [{code: '200'}]}}]);
    expect(typeof models[1].value_model.valueObject().value).toBe('string');
    expectTable(view, ['12 mg', '200']);
  });

  it('retains the same mixed assertions when the categorical reading comes first', () => {
    const {view} = render([{valueCodeableConcept: {coding: [{code: '200'}]}},
      {valueQuantity: {value: 12, unit: 'mg'}}]);
    expectTable(view, ['200', '12 mg']);
  });

  it('keeps a coded absence reason alongside the measurement in the table', () => {
    const {view} = render([{valueQuantity: {value: 12, unit: 'mg'}},
      {dataAbsentReason: {coding: [{code: 'not-performed'}]}}]);
    expectTable(view, ['12 mg', 'not-performed (data absent)']);
  });

  it('preserves a later pending reading without a chart exception or invented zero', () => {
    const {view} = render([{valueQuantity: {value: 12, unit: 'mg'}}, {status: 'registered'}]);
    expectTable(view, ['12 mg', '']);
  });

  it('does not drop the series when its first reading is pending', () => {
    const {view} = render([{status: 'registered'}, {valueQuantity: {value: 12, unit: 'mg'}}]);
    expectTable(view, ['', '12 mg']);
  });

  it('retains the quantitative chart when every reading supports it', () => {
    const {view, fixture} = render([{valueQuantity: {value: 12, unit: 'mg'}}, {valueQuantity: {value: 13, unit: 'mg'}}]);
    expect(fixture.componentInstance.visualizationType).toBe('bar');
    expect(view.querySelector('observation-bar-chart')).not.toBeNull();
    expect(view.querySelector('observation-table')).toBeNull();
  });
});
