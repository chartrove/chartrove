import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ObservationBarChartComponent } from './observation-bar-chart.component';
import { ObservationModel } from 'src/lib/models/resources/observation-model';
import { observationR4Factory } from 'src/lib/fixtures/factories/r4/resources/observation-r4-factory';
import { fhirVersions } from 'src/lib/models/constants';
import { BarElement, Chart } from 'chart.js';

describe('ObservationBarChartComponent', () => {
  let component: ObservationBarChartComponent;
  let fixture: ComponentFixture<ObservationBarChartComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ ObservationBarChartComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ObservationBarChartComponent);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it('renders zero measurements and inclusive reference bounds without changing raw observations', async () => {
    component.observations = [0, 6].map((value, index) => new ObservationModel({
      resourceType: 'Observation',
      id: `synthetic-chart-${index}`,
      status: 'final',
      code: { text: 'Synthetic measurement' },
      effectiveDateTime: `2024-01-0${index + 1}T12:00:00Z`,
      valueQuantity: { value, unit: 'mg' },
      referenceRange: [{ low: { value: 0, unit: 'mg' }, high: { value: 10, unit: 'mg' } }],
    }, fhirVersions.R4));
    const rawBefore = JSON.stringify(component.observations.map(value => value.fhirResource));
    fixture.nativeElement.style.width = '600px';
    fixture.detectChanges();
    await fixture.whenStable();
    const canvas = fixture.nativeElement.querySelector('canvas') as HTMLCanvasElement;
    const chart = Chart.getChart(canvas);
    expect(chart).toBeDefined();
    if (!chart) throw new Error('Synthetic observation chart was not registered');
    chart.update('none');
    expect(chart.width).toBeGreaterThan(0);
    expect(chart.data.datasets[0].data).toEqual([[0, 10], [0, 10]]);
    expect(chart.data.datasets[1].data).toEqual([[0, 0], [6, 6]]);
    expect(chart.options.scales?.['x'].max).toBe(13);
    const reference = chart.getDatasetMeta(0).data[0];
    expect(reference instanceof BarElement).toBeTrue();
    if (reference instanceof BarElement) expect(reference.getProps(['width'], true).width).toBeGreaterThan(0);
    expect(Number.isFinite(chart.getDatasetMeta(1).data[0].x)).toBeTrue();
    expect(JSON.stringify(component.observations.map(value => value.fhirResource))).toBe(rawBefore);
    fixture.destroy();
    expect(Chart.getChart(canvas)).toBeUndefined();
  });

  describe('updateNullMax', () => {
    it('fills only absent upper bounds and preserves explicit zero limits', () => {
      let test = [
        [5, null],
        [5, 0],
        [5, undefined],
        [0, null],
        [0, undefined],
        [0, 0],
        [4, 6]
      ]
      let expected = [
        [5, 8],
        [5, 0],
        [5, 8],
        [0, 8],
        [0, 8],
        [0, 0],
        [4, 6]
      ]

      expect(component['updateNullMax'](test, 8)).toEqual(expected);
    });
  });

  describe('extractReferenceRange', () => {
    it('returns the correct value when there is no reference range', () => {
      let observation = new ObservationModel(observationR4Factory.build(), fhirVersions.R4);

      expect(component['extractReferenceRange'](observation)).toEqual([0, 0])
    });

    it('returns the correct value when there is a reference range', () => {
      let observation = new ObservationModel(observationR4Factory.referenceRange(5, 10).build(), fhirVersions.R4);
      let observation2 = new ObservationModel(observationR4Factory.referenceRangeOnlyHigh(10).build(), fhirVersions.R4);
      let observation3 = new ObservationModel(observationR4Factory.referenceRangeOnlyLow(5).build(), fhirVersions.R4);

      expect(component['extractReferenceRange'](observation)).toEqual([5, 10])
      expect(component['extractReferenceRange'](observation2)).toEqual([0, 10])
      expect(component['extractReferenceRange'](observation3)).toEqual([5, null])
    });
  });

  describe('extractCurrentValue', () => {
    it('returns the correct value when the value is a range', () => {
      let observation = new ObservationModel(observationR4Factory.valueString('< 10').build(), fhirVersions.R4);
      let observation2 = new ObservationModel(observationR4Factory.valueString('> 10').build(), fhirVersions.R4);

      expect(component['extractCurrentValue'](observation)).toEqual([null, 10])
      expect(component['extractCurrentValue'](observation2)).toEqual([10, null])
    });

    it('returns the correct value when the value is a single value', () => {
      let observation = new ObservationModel(observationR4Factory.valueQuantity({ value: 5 }).build(), fhirVersions.R4);

      expect(component['extractCurrentValue'](observation)).toEqual([5, 5])
    });
  });
});
