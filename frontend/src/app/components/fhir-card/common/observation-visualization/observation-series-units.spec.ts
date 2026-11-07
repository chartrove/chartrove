import {TestBed} from '@angular/core/testing';
import {ObservationVisualizationComponent} from './observation-visualization.component';
import {ObservationModel} from '../../../../../lib/models/resources/observation-model';

describe('Observation series reported-unit compatibility', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({imports: [ObservationVisualizationComponent]}).compileComponents();
  });

  const ucum = 'http://unitsofmeasure.org';
  const mg = {unit: 'mg/dL', system: ucum, code: 'mg/dL'};

  function render(values: object[]) {
    const raw = values.map((value, index) => ({resourceType: 'Observation', id: `synthetic-unit-${index}`,
      status: 'final', code: {coding: [{system: 'http://loinc.org', code: 'synthetic-same-code'}]},
      effectiveDateTime: `2020-01-0${index + 1}`, ...value}));
    const before = JSON.stringify(raw);
    const fixture = TestBed.createComponent(ObservationVisualizationComponent);
    fixture.componentInstance.observations = raw.map(value => new ObservationModel(value));
    expect(() => fixture.detectChanges()).not.toThrow();
    expect(JSON.stringify(raw)).toBe(before);
    return fixture.nativeElement as HTMLElement;
  }

  function table(view: HTMLElement, expected: string[]) {
    expect(view.querySelector('observation-bar-chart')).toBeNull();
    expect(view.querySelector('observation-table')).not.toBeNull();
    const rows = Array.from(view.querySelectorAll('observation-table tbody tr')) as HTMLTableRowElement[];
    expect(rows.map(row => row.cells[1].textContent.trim())).toEqual(expected);
  }

  it('retains differing reported units as literal table values without a common numeric axis', () => {
    table(render([{valueQuantity: {value: 100, ...mg}},
      {valueQuantity: {value: 5.5, unit: 'mmol/L', system: ucum, code: 'mmol/L'}}]), ['100 mg/dL', '5.5 mmol/L']);
  });

  it('does not equate identical labels with different supplied unit codes', () => {
    table(render([{valueQuantity: {value: 100, ...mg}},
      {valueQuantity: {value: 5.5, ...mg, code: 'different-code'}}]), ['100 mg/dL', '5.5 mg/dL']);
  });

  it('does not equate identical labels and codes from different unit systems', () => {
    table(render([{valueQuantity: {value: 100, ...mg}},
      {valueQuantity: {value: 5.5, ...mg, system: 'https://example.invalid/units'}}]), ['100 mg/dL', '5.5 mg/dL']);
  });

  it('does not assign a preceding quantity unit to a unitless integer', () => {
    table(render([{valueQuantity: {value: 100, ...mg}}, {valueInteger: 5}]), ['100 mg/dL', '5']);
  });

  it('retains the chart for quantities with identical supplied unit metadata', () => {
    const view = render([{valueQuantity: {value: 100, ...mg}}, {valueQuantity: {value: 110, ...mg}}]);
    expect(view.querySelector('observation-bar-chart')).not.toBeNull();
    expect(view.querySelector('observation-table')).toBeNull();
  });

  it('retains the chart for an individual quantity', () => {
    expect(render([{valueQuantity: {value: 100, ...mg}}]).querySelector('observation-bar-chart')).not.toBeNull();
  });

  it('retains the existing chart for a uniformly unitless integer series', () => {
    expect(render([{valueInteger: 5}, {valueInteger: 6}]).querySelector('observation-bar-chart')).not.toBeNull();
  });
});
