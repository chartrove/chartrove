import {ObservationValueCodeableConceptModel} from './observation-value-codeable-concept-model';
import {DataAbsentReasonModel} from './data-absent-reason-model';

describe('Scalar Observation coded-value boundaries', () => {
  it('retains a supplied code when optional display and text are absent', () => {
    const raw = {coding: [{system: 'https://example.invalid/codes', code: 'synthetic-value'}]};
    const before = JSON.stringify(raw);
    const model = new ObservationValueCodeableConceptModel(raw);
    expect(model.valueObject()).toEqual({value: 'synthetic-value'});
    expect(model.display()).toBe('synthetic-value');
    expect(JSON.stringify(raw)).toBe(before);
  });

  it('retains a supplied absence-reason code without inventing a translation', () => {
    const model = new DataAbsentReasonModel({coding: [{system: 'http://terminology.hl7.org/CodeSystem/data-absent-reason', code: 'not-performed'}]});
    expect(model.valueObject()).toEqual({value: 'not-performed'});
    expect(model.display()).toBe('not-performed (data absent)');
  });

  it('does not crash or invent a value for an empty coding array', () => {
    const model = new ObservationValueCodeableConceptModel({coding: []});
    expect(() => model.display()).not.toThrow();
    expect(model.valueObject()).toEqual({value: null});
    expect(model.display()).toBe('');
  });

  it('keeps supplied narrative text ahead of coding display and raw code', () => {
    const model = new ObservationValueCodeableConceptModel({text: 'Café 零 صفر', coding: [{display: 'Other display', code: 'other-code'}]});
    expect(model.valueObject()).toEqual({value: 'Café 零 صفر'});
    expect(model.display()).toBe('Café 零 صفر');
  });

  it('keeps the first supplied display ahead of its code without merging codings', () => {
    const model = new ObservationValueCodeableConceptModel({coding: [{display: 'First assertion', code: 'first-code'}, {display: 'Second assertion', code: 'second-code'}]});
    expect(model.display()).toBe('First assertion');
    expect(model.valueObject()).toEqual({value: 'First assertion'});
  });

  it('retains the existing no-value and unspecified-absence behavior', () => {
    expect(new ObservationValueCodeableConceptModel({}).valueObject()).toEqual({value: null});
    expect(new ObservationValueCodeableConceptModel({}).display()).toBe('');
    expect(new DataAbsentReasonModel({}).display()).toBe('(data absent)');
  });
});
