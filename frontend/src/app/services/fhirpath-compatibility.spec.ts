import {evaluate} from 'fhirpath';
import {AddressModel} from '../../lib/models/datatypes/address-model';
import {ExplanationOfBenefitModel} from '../../lib/models/resources/explanation-of-benefit-model';
import {fhirVersions} from '../../lib/models/constants';
import {ResourceFhir} from '../models/fasten/resource_fhir';
import {FhirPathPipe} from '../pipes/fhir-path.pipe';
import {DashboardWidgetComponent} from '../widgets/dashboard-widget/dashboard-widget.component';

describe('FHIRPath application compatibility', () => {
  it('preserves zero, false, timestamp offsets and reference order through the display pipe', () => {
    const raw = {resourceType: 'Observation', id: 'synthetic-path', status: 'final',
      code: {text: 'Synthetic measurement'}, valueQuantity: {value: 0, unit: 'mg/dL'},
      effectiveDateTime: '2024-02-29T12:34:56.123+03:00', syntheticFalse: false,
      hasMember: [{reference: 'Observation/first'}, {reference: 'Observation/second'}]};
    const before = JSON.stringify(raw);
    const record = new ResourceFhir({resource_raw: raw});
    const pipe = new FhirPathPipe();
    expect(pipe.transform(record, 'missing', 'Observation.valueQuantity.value')).toBe('0');
    expect(pipe.transform(record, 'Observation.syntheticFalse')).toBe('false');
    expect(pipe.transform(record, 'Observation.effectiveDateTime')).toBe(raw.effectiveDateTime);
    expect(pipe.transform(record, 'Observation.hasMember.reference')).toBe('Observation/first, Observation/second');
    expect(pipe.transform(record, 'missing')).toBeNull();
    expect(JSON.stringify(raw)).toBe(before);
  });

  it('retains Unicode and ordered address lines in the real address model', () => {
    const raw = {line: ['Épreuve 1', 'بيت 2'], city: 'Montréal', state: 'QC',
      country: 'CA', postalCode: 'H0H 0H0'};
    const before = JSON.stringify(raw);
    const address = new AddressModel(raw);
    expect(address.line).toEqual(raw.line);
    expect(address.city).toBe(raw.city);
    expect(address.state).toBe(raw.state);
    expect(address.country).toBe(raw.country);
    expect(address.postalCode).toBe(raw.postalCode);
    const absent = new AddressModel({});
    expect(absent.line).toEqual([]);
    expect(absent.city).toBe('');
    expect(JSON.stringify(raw)).toBe(before);
  });

  it('selects the principal claim diagnosis without losing zero amounts or focal coverage', () => {
    const principal = {coding: [{system: 'urn:synthetic', code: 'principal', display: 'Synthetic principal'}]};
    const raw = {resourceType: 'ExplanationOfBenefit', id: 'synthetic-claim', status: 'active',
      diagnosis: [
        {type: [{coding: [{code: 'secondary'}]}], diagnosisCodeableConcept: {text: 'Synthetic secondary'}},
        {type: [{coding: [{code: 'principal'}]}], diagnosisCodeableConcept: principal},
      ],
      total: [{category: {text: 'Synthetic total'}, amount: {value: 0, currency: 'USD'}}],
      insurance: [{focal: false, coverage: {reference: 'Coverage/secondary'}},
        {focal: true, coverage: {reference: 'Coverage/primary'}}],
      item: [{sequence: 1, quantity: {value: 0}, net: {value: 0, currency: 'USD'}}]};
    const before = JSON.stringify(raw);
    const claim = new ExplanationOfBenefitModel(raw, fhirVersions.R4);
    expect(claim.code).toEqual(principal);
    expect(claim.total[0].amount.value).toBe(0);
    expect(claim.services[0].quantity).toBe(0);
    expect(claim.services[0].itemCost.value).toBe(0);
    expect(JSON.stringify(claim.insurance)).toBe(JSON.stringify({reference: 'Coverage/primary'}));
    expect(JSON.stringify(raw)).toBe(before);
  });

  it('keeps dashboard aliases, raw identity and scalar collections through its actual query mapper', () => {
    const api = jasmine.createSpyObj('FastenApiService', ['queryResources']);
    const widget = new DashboardWidgetComponent(api);
    const raw = {resourceType: 'Observation', id: 'synthetic-mapping',
      valueQuantity: {value: 0}, syntheticFalse: false, effectiveDateTime: '2024-01-01T12:00:00+03:00'};
    const before = JSON.stringify(raw);
    const map = widget.fhirPathMapQueryFn({from: 'Observation', where: {}, select: [
      'valueQuantity.value as measured', 'syntheticFalse AS flag',
      'effectiveDateTime as when', '* as raw',
    ]});
    const result = map(raw);
    expect(result['measured']).toEqual([0]);
    expect(result['flag']).toEqual([false]);
    expect(result['when']).toEqual([raw.effectiveDateTime]);
    expect(result['raw']).toBe(raw);
    expect(result['id']).toBe(raw.id);
    expect(result['resourceType']).toBe(raw.resourceType);
    expect(api.queryResources).not.toHaveBeenCalled();
    expect(JSON.stringify(raw)).toBe(before);
  });

  it('compares compatible units and retains an explicit zero quantity', () => {
    expect(evaluate({}, "1 'g' = 1000 'mg'")).toEqual([true]);
    expect(evaluate({}, "0 'mg' = 0 'g'")).toEqual([true]);
    expect(evaluate({}, "(0 'mg').toString()")).toEqual(["0 'mg'"]);
  });

  it('does not perform remote resolution under the synchronous application defaults', () => {
    const fetch = spyOn(globalThis, 'fetch').and.throwError('Synthetic network disabled');
    const raw = {resourceType: 'Observation', subject: {reference: 'https://synthetic.invalid/Patient/one'}};
    const before = JSON.stringify(raw);
    expect(() => evaluate(raw, 'Observation.subject.resolve()')).toThrow();
    expect(fetch).not.toHaveBeenCalled();
    expect(JSON.stringify(raw)).toBe(before);
  });
});
