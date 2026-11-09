import {exactObservationPreview} from './observation-preview';

describe('exactObservationPreview', () => {
  it('preserves numeric lexical precision, zero, signs, trailing decimals and exponent spellings', () => {
    for (const value of ['1.234567890123456789', '3.500', '0', '-0.000', '1e-8', '9007199254740993']) {
      const raw = `{"resourceType":"Observation","valueQuantity":{"value":${value},"unit":"mg","code":"mg","comparator":"<"},"effectiveDateTime":"2026-11"}`;
      expect(exactObservationPreview(raw)).toEqual({value, unit: 'mg', unitCode: 'mg', comparator: '<', effective: '2026-11'});
    }
  });
  it('keeps false and textual content as recorded without inventing units or dates', () => {
    expect(exactObservationPreview('{"resourceType":"Observation","valueBoolean":false}')?.value).toBe('false');
    expect(exactObservationPreview('{"resourceType":"Observation","valueInteger":0}')?.value).toBe('0');
    expect(exactObservationPreview('{"resourceType":"Observation","valueString":"<img src=x>"}')).toEqual({value: '<img src=x>', unit: null, unitCode: null, comparator: null, effective: null});
  });
  it('refuses ambiguous, complex, malformed and absent-value records rather than rounding or choosing', () => {
    for (const extra of ['"valueQuantity":{"value":1,"value":2}', '"valueString":"one","valueString":"two"',
      '"valueInteger":1.5', '"valueInteger":2147483648', '"valueQuantity":{"value":"1.23"}',
      '"valueQuantity":{"value":1},"dataAbsentReason":{"text":"unknown"}',
      '"valueString":"x","valueBoolean":false', '"component":[],"valueQuantity":{"value":1}',
      '"valueRange":{"low":{"value":1},"high":{"value":2}}', '"valueQuantity":{"value":1,"unit":{}}']) {
      expect(exactObservationPreview(`{"resourceType":"Observation",${extra}}`)).toBeNull();
    }
    expect(exactObservationPreview('{"resourceType":"Patient","valueInteger":1}')).toBeNull();
    expect(exactObservationPreview('{"resourceType":"Observation","valueInteger":1,}')).toBeNull();
    expect(exactObservationPreview('{/*comment*/"resourceType":"Observation","valueInteger":1}')).toBeNull();
    expect(exactObservationPreview('{"resourceType":"Observation","valueInteger":1} {}')).toBeNull();
    expect(exactObservationPreview('{')).toBeNull();
  });
  it('bounds preview complexity and leaves large/deep originals to the JSON fallback', () => {
    expect(exactObservationPreview(' '.repeat(2 * 1024 * 1024 + 1))).toBeNull();
    expect(exactObservationPreview(`{"resourceType":"Observation","valueInteger":1,"extension":${'['.repeat(140)}0${']'.repeat(140)}}`)).toBeNull();
  });
});
