import {TestBed} from '@angular/core/testing';
import {provideRouter} from '@angular/router';
import {ObservationModel} from './observation-model';
import {QuantityModel} from '../datatypes/quantity-model';
import {ReferenceRangeModel} from '../datatypes/reference-range-model';
import {RangeModel} from '../datatypes/range-model';
import {observationR4Factory} from 'src/lib/fixtures/factories/r4/resources/observation-r4-factory';
import {quantityR4Factory} from 'src/lib/fixtures/factories/r4/datatypes/quantity-r4-factory';
import {ObservationComponent} from 'src/app/components/fhir-card/resources/observation/observation.component';
import {ObservationTableComponent} from 'src/app/components/fhir-card/common/observation-table/observation-table.component';
import {ObservationBarChartComponent} from 'src/app/components/fhir-card/common/observation-bar-chart/observation-bar-chart.component';
import {obsValue} from 'src/app/components/fhir-datatable/datatable-generic-resource/utils';

function syntheticObservation(value: object = {}): any {
  return {resourceType: 'Observation', id: 'synthetic-result', status: 'final',
    code: {text: 'Synthetic result'}, effectiveDateTime: '2020-01-01T00:00:00Z', ...value};
}

describe('Observation value boundaries', () => {
  for (const example of [
    {name: 'integer zero', raw: {valueInteger: 0}, object: {value: 0}, display: '0'},
    {name: 'boolean false', raw: {valueBoolean: false}, object: {value: false}, display: 'false'},
    {name: 'explicit empty string without replacement', raw: {valueString: ''}, object: {value: ''}, display: ''},
  ]) {
    it(`preserves ${example.name} in the display model without mutating the source`, () => {
      const raw = syntheticObservation(example.raw);
      const before = JSON.stringify(raw);
      const model = new ObservationModel(raw);
      expect(model.value_model?.valueObject()).toEqual(example.object);
      expect(model.value_model?.display()).toBe(example.display);
      expect(JSON.stringify(raw)).toBe(before);
    });
  }

  it('does not invent a reading for absent or null primitive values', () => {
    for (const raw of [{}, {valueInteger: null}, {valueBoolean: null}, {valueString: null}]) {
      expect(new ObservationModel(syntheticObservation(raw)).value_model?.display()).toBeUndefined();
    }
  });

  it('recognizes numeric zero as a present quantity', () => {
    const quantity = new QuantityModel({value: 0, unit: '1'});
    expect(quantity.hasValue()).toBeTrue();
    expect(quantity.display()).toBe('0 1');
    expect(quantity.valueObject()).toEqual({value: 0});
  });

  it('retains an explicit strict Quantity comparator when its value is zero', () => {
    expect(new QuantityModel({value: 0, unit: '1', comparator: '>'}).display()).toBe('> 0 1');
  });

  for (const example of [
    {name: 'zero lower bound', raw: {low: {value: 0, unit: '1'}}, low: 0, high: undefined, display: '>= 0 1'},
    {name: 'zero upper bound', raw: {high: {value: 0, unit: '1'}}, low: undefined, high: 0, display: '<= 0 1'},
    {name: 'zero at the start of a range', raw: {low: {value: 0, unit: '1'}, high: {value: 5, unit: '1'}}, low: 0, high: 5, display: '0 1 – 5 1'},
    {name: 'zero at the end of a range', raw: {low: {value: -5, unit: '1'}, high: {value: 0, unit: '1'}}, low: -5, high: 0, display: '-5 1 – 0 1'},
  ]) {
    it(`keeps ${example.name} in chartable bounds and displayed reference limits`, () => {
      const range = new ReferenceRangeModel(example.raw);
      expect(range.hasValue()).toBeTrue();
      expect(range.chartableReferenceRange()).toEqual({low: example.low, high: example.high});
      expect(range.display()).toBe(example.display);
    });
  }

  it('uses inclusive signs for single known nonzero FHIR Range limits', () => {
    expect(new RangeModel({low: {value: 6, unit: '1'}}).display()).toBe('>= 6 1');
    expect(new RangeModel({high: {value: 6, unit: '1'}}).display()).toBe('<= 6 1');
  });

  it('keeps explicit zero upper bounds in plotted reference data instead of replacing them with the axis maximum', () => {
    const chart = new ObservationBarChartComponent();
    chart.observations = [new ObservationModel(syntheticObservation({
      valueQuantity: {value: 10, unit: '1'},
      referenceRange: [{low: {value: -5, unit: '1'}, high: {value: 0, unit: '1'}}],
    }))];
    chart.ngOnInit();
    expect(chart.barChartData[0].data).toEqual([[-5, 0]]);
    expect(chart.barChartData[1].data).toEqual([[10, 10]]);
  });

  for (const low of [0, 5]) {
    it(`retains an unknown upper reference limit when the known lower limit is ${low}`, () => {
      const chart = new ObservationBarChartComponent();
      chart.observations = [new ObservationModel(syntheticObservation({
        valueQuantity: {value: 10, unit: '1'},
        referenceRange: [{low: {value: low, unit: '1'}}],
      }))];
      chart.ngOnInit();
      expect(chart.barChartData[0].data).toEqual([[low, 13]]);
      expect(chart.barChartData[0]['dataLabels']).toEqual([`>= ${low} 1`]);
    });
  }

  for (const example of [
    {name: 'integer zero', value: {valueInteger: 0}, display: '0'},
    {name: 'boolean false', value: {valueBoolean: false}, display: 'false'},
    {name: 'missing global value with text-only coding', value: {}, display: ''},
  ]) {
    it(`renders ${example.name} safely in the resource-table value getter`, () => {
      const raw = syntheticObservation(example.value);
      let result: string;
      expect(() => { result = obsValue(raw); }).not.toThrow();
      expect(result).toBe(example.display);
    });
  }
});

describe('Observation synthetic fixture boundaries', () => {
  const examples: {name: string; build: () => any; pick: (raw: any) => any; expected: any}[] = [
    {name: 'observation quantity zero', build: () => observationR4Factory.valueQuantity({value: 0}).build(), pick: raw => raw.valueQuantity.value, expected: 0},
    {name: 'integer zero', build: () => observationR4Factory.valueInteger(0).build(), pick: raw => raw.valueInteger, expected: 0},
    {name: 'boolean false', build: () => observationR4Factory.valueBoolean(false).build(), pick: raw => raw.valueBoolean, expected: false},
    {name: 'explicit empty string', build: () => observationR4Factory.valueString('').build(), pick: raw => raw.valueString, expected: ''},
    {name: 'zero lower reference bound', build: () => observationR4Factory.referenceRange(0, 5).build(), pick: raw => raw.referenceRange[0].low.value, expected: 0},
    {name: 'zero upper reference bound', build: () => observationR4Factory.referenceRange(-5, 0).build(), pick: raw => raw.referenceRange[0].high.value, expected: 0},
    {name: 'only-low zero reference', build: () => observationR4Factory.referenceRangeOnlyLow(0).build(), pick: raw => raw.referenceRange[0].low.value, expected: 0},
    {name: 'only-high zero reference', build: () => observationR4Factory.referenceRangeOnlyHigh(0).build(), pick: raw => raw.referenceRange[0].high.value, expected: 0},
    {name: 'standalone quantity zero', build: () => quantityR4Factory.value(0).build(), pick: raw => raw.value, expected: 0},
  ];
  for (const example of examples) {
    it(`does not replace the supplied ${example.name} with a sample default`, () => {
      expect(example.pick(example.build())).toEqual(example.expected);
    });
  }

  it('uses an integral default for the FHIR integer fixture choice', () => {
    expect(Number.isInteger((observationR4Factory.valueInteger().build() as any).valueInteger)).toBeTrue();
  });
});

describe('Observation card and table boundary rendering', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ObservationComponent, ObservationTableComponent],
      providers: [provideRouter([])],
    }).compileComponents();
  });

  for (const example of [
    {name: 'integer zero', value: {valueInteger: 0}, expected: '0'},
    {name: 'boolean false', value: {valueBoolean: false}, expected: 'false'},
  ]) {
    it(`shows ${example.name} in the actual observation card`, () => {
      const fixture = TestBed.createComponent(ObservationComponent);
      fixture.componentInstance.showDetails = false;
      fixture.componentInstance.displayModel = new ObservationModel(syntheticObservation(example.value));
      expect(() => fixture.detectChanges()).not.toThrow();
      const rows = Array.from((fixture.nativeElement as HTMLElement).querySelectorAll('tr'));
      const valueRow = rows.find(row => row.cells[0]?.textContent.trim() === 'Value');
      expect(valueRow?.cells[1]?.textContent.trim()).toBe(example.expected);
    });
  }

  it('renders a pending observation without a global value or an invented result/chart', () => {
    const fixture = TestBed.createComponent(ObservationComponent);
    fixture.componentInstance.showDetails = false;
    fixture.componentInstance.displayModel = new ObservationModel(syntheticObservation({status: 'registered'}));
    expect(() => fixture.detectChanges()).not.toThrow();
    const view = fixture.nativeElement as HTMLElement;
    const valueRow = Array.from(view.querySelectorAll('tr')).find(row => row.cells[0]?.textContent.trim() === 'Value');
    expect(valueRow).toBeUndefined();
    expect(view.textContent).toContain('Synthetic result');
    expect(view.querySelector('observation-visualization')).toBeNull();
  });

  it('retains a zero-only reference bound in the rendered observation table', () => {
    const fixture = TestBed.createComponent(ObservationTableComponent);
    fixture.componentInstance.observations = [new ObservationModel(syntheticObservation({
      valueInteger: 1, referenceRange: [{low: {value: 0, unit: '1'}}],
    }))];
    fixture.detectChanges();
    const view = fixture.nativeElement as HTMLElement;
    expect(view.textContent).toContain('Reference Range');
    expect(view.querySelector('tbody tr')?.textContent).toContain('>= 0 1');
  });
});
