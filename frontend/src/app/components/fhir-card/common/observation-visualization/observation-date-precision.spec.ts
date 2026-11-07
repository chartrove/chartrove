import {TestBed} from '@angular/core/testing';
import {ObservationTableComponent} from '../observation-table/observation-table.component';
import {ObservationBarChartComponent} from '../observation-bar-chart/observation-bar-chart.component';
import {ObservationModel} from '../../../../../lib/models/resources/observation-model';

describe('Observation table and chart date precision', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({imports: [ObservationTableComponent, ObservationBarChartComponent]}).compileComponents();
  });

  for (const example of [
    {name: 'year-only precision without an inferred month or day', effective: '2020', issued: undefined, expected: '2020'},
    {name: 'month-only precision without an inferred day', effective: '2020-02', issued: undefined, expected: '2020-02'},
    {name: 'an explicitly supplied calendar day', effective: '2020-02-03', issued: undefined, expected: 'Feb 3, 2020'},
    {name: 'missing observation time despite a supplied issuance time', effective: undefined, issued: '2020-02-04T12:00:00Z', expected: 'Unknown date'},
  ]) {
    function model() {
      const raw = {resourceType: 'Observation', id: 'synthetic-date-precision', status: 'final',
        code: {text: 'Synthetic date precision'}, valueQuantity: {value: 12, unit: 'mg'},
        ...(example.effective ? {effectiveDateTime: example.effective} : {}),
        ...(example.issued ? {issued: example.issued} : {})};
      return {raw, before: JSON.stringify(raw), value: new ObservationModel(raw)};
    }

    it(`keeps ${example.name} in the actual table`, () => {
      const original = model();
      const fixture = TestBed.createComponent(ObservationTableComponent);
      fixture.componentInstance.observations = [original.value];
      fixture.detectChanges();
      const date = fixture.nativeElement.querySelector('tbody tr td') as HTMLTableCellElement;
      expect(date.textContent.trim()).toBe(example.expected);
      expect(JSON.stringify(original.raw)).toBe(original.before);
      expect(original.value.effective_date).toBe(example.effective);
    });

    it(`keeps ${example.name} in the actual chart label`, () => {
      const original = model();
      const fixture = TestBed.createComponent(ObservationBarChartComponent);
      fixture.componentInstance.observations = [original.value];
      fixture.detectChanges();
      expect(fixture.componentInstance.barChartLabels).toEqual([example.expected]);
      expect(JSON.stringify(original.raw)).toBe(original.before);
      expect(original.value.effective_date).toBe(example.effective);
    });
  }
});
