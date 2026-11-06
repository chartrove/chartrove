import {of, Subject, throwError} from 'rxjs';
import {ActivatedRoute} from '@angular/router';
import {ReportLabsComponent} from './report-labs.component';
import {FastenApiService} from '../../services/fasten-api.service';
import {ResourceFhir} from '../../models/fasten/resource_fhir';

// Independently constructed synthetic records: Bundle fullUrl aliases deliberately
// differ from business IDs. Resolving URNs by stripping a prefix is incorrect.
describe('Lab report normalized reference identity', () => {
  let api: jasmine.SpyObj<FastenApiService>;
  let route: Subject<any>;
  let component: ReportLabsComponent;
  const source = '11111111-1111-4111-8111-111111111111';
  const params = {source_id: source, resource_type: 'DiagnosticReport', resource_id: 'synthetic-report'};
  const observation = (id = 'synthetic-result', code = 'SYNTHETIC-0'): ResourceFhir => new ResourceFhir({
    source_id: source, source_resource_type: 'Observation', source_resource_id: id,
    resource_raw: {resourceType: 'Observation', id, status: 'final', code: {coding: [{system: 'http://loinc.org', code, display: 'Synthetic value'}]}, valueInteger: 0}
  });
  const report = (children = [observation()]): ResourceFhir => new ResourceFhir({
    source_id: source, source_resource_type: 'DiagnosticReport', source_resource_id: 'synthetic-report',
    resource_raw: {resourceType: 'DiagnosticReport', id: 'synthetic-report', status: 'final', code: {text: 'Synthetic report'}, result: [{reference: 'urn:uuid:99999999-9999-4999-8999-999999999999'}]},
    related_resources: children
  });
  beforeEach(() => {
    api = jasmine.createSpyObj('FastenApiService', ['getResources', 'getResourceGraph', 'queryResources']);
    api.getResources.and.returnValue(of([report()]));
    api.getResourceGraph.and.returnValue(of({results: {DiagnosticReport: [report()]}}));
    api.queryResources.and.returnValue(of({success: true, data: []}));
    route = new Subject();
    component = new ReportLabsComponent(api, {params: route} as unknown as ActivatedRoute);
  });
  afterEach(() => (component as any).ngOnDestroy?.());

  it('uses normalized direct relationships without guessing an alias business ID or changing raw FHIR', () => {
    const root = report();
    const before = JSON.stringify(root);
    api.getResourceGraph.and.returnValue(of({results: {DiagnosticReport: [root]}}));
    let groups: string[];
    component.findLabResultCodesFilteredToReport(source, 'DiagnosticReport', 'synthetic-report').subscribe(value => groups = value);
    expect(groups).toEqual(['http://loinc.org|SYNTHETIC-0']);
    expect(api.getResourceGraph).toHaveBeenCalledOnceWith('MedicalHistory', [{source_id: source, source_resource_type: 'DiagnosticReport', source_resource_id: 'synthetic-report'}]);
    expect(api.getResources).not.toHaveBeenCalled();
    expect(JSON.stringify(root)).toBe(before);
  });

  it('does not include nested derived-from results or unrelated patient observations', () => {
    const direct = observation();
    direct.related_resources = [observation('synthetic-derived-input', 'SYNTHETIC-DECOY')];
    api.getResourceGraph.and.returnValue(of({results: {DiagnosticReport: [report([direct])]}}));
    component.findLabResultCodesFilteredToReport(source, 'DiagnosticReport', 'synthetic-report').subscribe(groups => {
      expect(groups).toEqual(['http://loinc.org|SYNTHETIC-0']);
    });
  });

  for (const mismatch of ['source_id', 'source_resource_type', 'source_resource_id']) {
    it(`rejects a returned root with a different ${mismatch}`, () => {
      const root = report();
      root[mismatch] = 'synthetic-unrelated';
      api.getResourceGraph.and.returnValue(of({results: {DiagnosticReport: [root]}}));
      let failed = false;
      component.findLabResultCodesFilteredToReport(source, 'DiagnosticReport', 'synthetic-report').subscribe({error: () => failed = true});
      expect(failed).toBeTrue();
    });
  }

  for (const roots of [[], [report(), report()]]) {
    it(`rejects ${roots.length ? 'ambiguous' : 'missing'} selected roots`, () => {
      api.getResourceGraph.and.returnValue(of({results: {DiagnosticReport: roots}}));
      let failed = false;
      component.findLabResultCodesFilteredToReport(source, 'DiagnosticReport', 'synthetic-report').subscribe({error: () => failed = true});
      expect(failed).toBeTrue();
    });
  }

  it('rejects a related Observation from a different source', () => {
    const child = observation();
    child.source_id = '22222222-2222-4222-8222-222222222222';
    api.getResourceGraph.and.returnValue(of({results: {DiagnosticReport: [report([child])]}}));
    let failed = false;
    component.findLabResultCodesFilteredToReport(source, 'DiagnosticReport', 'synthetic-report').subscribe({error: () => failed = true});
    expect(failed).toBeTrue();
  });

  it('terminates an unresolved reference as a failure instead of claiming an empty report', () => {
    api.getResourceGraph.and.returnValue(of({results: {DiagnosticReport: [report([])]}}));
    component.ngOnInit();
    route.next(params);
    expect(component.loading).toBeFalse();
    expect((component as any).loadError).toBeTrue();
    expect(component.isEmptyReport).toBeFalse();
  });

  it('completes a genuinely empty selected report', () => {
    const root = report([]);
    (root.resource_raw as any).result = [];
    api.getResourceGraph.and.returnValue(of({results: {DiagnosticReport: [root]}}));
    component.ngOnInit();
    route.next(params);
    expect(component.loading).toBeFalse();
    expect(component.isEmptyReport).toBeTrue();
    expect((component as any).loadError).toBeFalse();
  });

  it('terminates a failed graph request without displaying stale clinical data', () => {
    api.getResourceGraph.and.returnValue(throwError(() => new Error('Synthetic denied request')));
    component.ngOnInit();
    route.next(params);
    expect(component.loading).toBeFalse();
    expect((component as any).loadError).toBeTrue();
    expect(component.reportDisplayModel).toBeNull();
    expect(component.observationGroups).toEqual({});
  });

  it('deduplicates repeated result references and codes, preserving zero values and Unicode', () => {
    const root = report();
    (root.resource_raw as any).result.push({...((root.resource_raw as any).result[0])});
    (root.related_resources[0].resource_raw as any).code.text = 'Synthetic Δοκιμή 血液';
    const before = JSON.stringify(root);
    api.getResourceGraph.and.returnValue(of({results: {DiagnosticReport: [root]}}));
    component.findLabResultCodesFilteredToReport(source, 'DiagnosticReport', 'synthetic-report').subscribe(groups => {
      expect(groups).toEqual(['http://loinc.org|SYNTHETIC-0']);
    });
    expect(JSON.stringify(root)).toBe(before);
  });

  it('rejects identifier-only references rather than declaring a report empty', () => {
    const root = report([]);
    (root.resource_raw as any).result = [{identifier: {value: 'synthetic-unresolved'}}];
    api.getResourceGraph.and.returnValue(of({results: {DiagnosticReport: [root]}}));
    let failed = false;
    component.findLabResultCodesFilteredToReport(source, 'DiagnosticReport', 'synthetic-report').subscribe({error: () => failed = true});
    expect(failed).toBeTrue();
  });

  it('cancels a superseded report request and ignores late results', () => {
    const stale = new Subject<any>();
    api.getResourceGraph.and.returnValue(stale);
    component.ngOnInit();
    route.next(params);
    expect(stale.observed).toBeTrue();
    api.getResourceGraph.and.returnValue(of({results: {DiagnosticReport: [report([])]}}));
    route.next({...params, resource_id: 'synthetic-other-report'});
    expect(stale.observed).toBeFalse();
    expect(component.reportResourceId).toBe('synthetic-other-report');
  });

  it('cancels graph and route subscriptions when the page is destroyed', () => {
    const pending = new Subject<any>();
    api.getResourceGraph.and.returnValue(pending);
    component.ngOnInit();
    route.next(params);
    (component as any).ngOnDestroy?.();
    expect(pending.observed).toBeFalse();
    expect(route.observed).toBeFalse();
  });
});
