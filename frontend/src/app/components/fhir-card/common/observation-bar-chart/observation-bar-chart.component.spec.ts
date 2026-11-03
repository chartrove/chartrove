import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ObservationBarChartComponent } from './observation-bar-chart.component';
import { ObservationModel } from 'src/lib/models/resources/observation-model';
import { observationR4Factory } from 'src/lib/fixtures/factories/r4/resources/observation-r4-factory';
import { fhirVersions } from 'src/lib/models/constants';

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
