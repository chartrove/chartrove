import { TestBed } from '@angular/core/testing';

import { FastenApiService } from './fasten-api.service';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import {HTTP_CLIENT_TOKEN} from '../dependency-injection';
import { HttpClient, provideHttpClient, withInterceptorsFromDi, withXhr } from '@angular/common/http';
import {DashboardWidgetQuery} from '../models/widget/dashboard-widget-query';

describe('FastenApiService', () => {
  let service: FastenApiService;

  beforeEach(() => {
    TestBed.configureTestingModule({
    imports: [],
    providers: [
        {
            provide: HTTP_CLIENT_TOKEN,
            useClass: HttpClient,
        },
        provideHttpClient(withXhr(), withInterceptorsFromDi()),
        provideHttpClientTesting(),
    ]
});
    service = TestBed.inject(FastenApiService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('sends exact lab JSON text and an explicit native source without rounding unrelated fields',()=>{
    const http=TestBed.inject(HttpTestingController);
    const raw='{"resourceType":"Observation","id":"same-id","valueQuantity":{"value":3.5},"extension":[{"url":"urn:synthetic:exact","valueDecimal":0.1234567890123456789}]}';
    const source='22222222-2222-4222-8222-222222222222';
    service.updateResourceText('Observation','same-id',source,raw).subscribe();
    const req=http.expectOne(r=>r.method==='PATCH');
    expect(req.request.url).toContain('/secure/resource/fhir/Observation/same-id');
    expect(req.request.headers.get('Content-Type')).toBe('application/json');
    expect(req.request.body).toBe(`{"source_id":"${source}","resource_raw":${raw}}`);
    expect(req.request.body).toContain('0.1234567890123456789');
    req.flush({success:true});http.verify();
  });

  it('carries explicit source selection while preserving legacy request shapes', () => {
    const http = TestBed.inject(HttpTestingController);
    const source = '22222222-2222-4222-8222-222222222222';
    service.deleteResourceFhir('Practitioner', 'shared-id', source).subscribe();
    const scopedDelete = http.expectOne(req => req.method === 'DELETE');
    expect(scopedDelete.request.params.get('source_id')).toBe(source);
    scopedDelete.flush({success: true});
    service.deleteResourceFhir('Practitioner', 'unique-id').subscribe();
    const legacyDelete = http.expectOne(req => req.method === 'DELETE');
    expect(legacyDelete.request.params.has('source_id')).toBeFalse();
    legacyDelete.flush({success: true});
    const resource = {resourceType: 'Practitioner', id: 'shared-id'};
    service.updatePractitioner('shared-id', resource, source).subscribe();
    const scopedPut = http.expectOne(req => req.method === 'PUT');
    expect(scopedPut.request.body).toEqual({resource, source_id: source});
    scopedPut.flush({success: true});
    service.updatePractitioner('shared-id', resource).subscribe();
    const legacyPut = http.expectOne(req => req.method === 'PUT');
    expect(legacyPut.request.body).toEqual({resource});
    legacyPut.flush({success: true});
    http.verify();
  });

  it('preserves related attachments in a serialized multipart request for the native transport', async () => {
    const http = TestBed.inject(HttpTestingController);
    const resourceList = {
      resourceType: 'List' as const,
      status: 'current' as const,
      mode: 'working' as const,
      encounter: {reference: 'Encounter/synthetic-encounter'},
      contained: [{
        resourceType: 'DocumentReference' as const,
        status: 'current' as const,
        description: 'Synthetic López attachment\nSecond line',
        content: [{attachment: {contentType: 'text/plain', data: 'AAECA/8=', title: 'Synthetic.txt'}}],
      }],
    };
    const result = service.createRelatedResourcesFastenSource(resourceList).toPromise();
    const request = http.expectOne(req => req.url.endsWith('/secure/resource/related'));
    const contentType = request.request.headers.get('Content-Type');
    expect(typeof request.request.body).toBe('string');
    expect(contentType).toMatch(/^multipart\/form-data; boundary=.+/);
    if (typeof request.request.body === 'string' && contentType) {
      const decoded = await new Response(request.request.body, {headers: {'Content-Type': contentType}}).formData();
      const file = decoded.get('file') as File;
      expect(file.name).toBe('related.json');
      expect(file.type).toBe('application/json');
      expect(JSON.parse(await file.text())).toEqual(resourceList);
      const fields: string[] = [];
      decoded.forEach((_value, key) => fields.push(key));
      expect(fields).toEqual(['file']);
    }
    request.flush({success: true, data: {id: 'synthetic-source'}});
    expect((await result).id).toBe('synthetic-source');
    http.verify();
  });

  it('imports a Patient through the local bundle API without provider credentials', async () => {
    const http = TestBed.inject(HttpTestingController);
    const result = service.importSandboxPatient({resourceType: 'Patient', id: 'synthetic-patient'}).toPromise();
    http.expectOne(req => req.url.endsWith('/secure/source')).flush({success: true, data: []});
    const request = http.expectOne(req => req.url.endsWith('/secure/source/manual'));
    const decoded = await new Response(request.request.body, {headers: {'Content-Type': request.request.headers.get('Content-Type')}}).formData();
    const file = decoded.get('file') as File;
    expect(JSON.parse(await file.text())).toEqual({resourceType: 'Bundle', type: 'collection', entry: [{resource: {resourceType: 'Patient', id: 'synthetic-patient'}}]});
    request.flush({success: true, source: {id: 'local-source'}, data: {totalResources: 1}});
    expect((await result).source.id).toBe('local-source');
    http.verify();
  });

  it('sends a serialized multipart Patient upload that the native transport can carry', async () => {
    const http = TestBed.inject(HttpTestingController);
    const patient = {resourceType: 'Patient' as const, id: 'synthetic-patient', name: [{family: 'López', given: ['Camila']}], text: {status: 'generated' as const, div: '<div>Line one\nLine two</div>'}};
    const result = service.importSandboxPatient(patient).toPromise();
    http.expectOne(req => req.url.endsWith('/secure/source')).flush({success: true, data: []});
    const request = http.expectOne(req => req.url.endsWith('/secure/source/manual'));
    const contentType = request.request.headers.get('Content-Type');
    expect(typeof request.request.body).toBe('string');
    expect(contentType).toMatch(/^multipart\/form-data; boundary=.+/);
    if (typeof request.request.body === 'string' && contentType) {
      const decoded = await new Response(request.request.body, {headers: {'Content-Type': contentType}}).formData();
      const file = decoded.get('file') as File;
      expect(file.name).toBe('epic-sandbox.json');
      expect(file.type).toBe('application/fhir+json');
      expect(JSON.parse(await file.text())).toEqual({resourceType: 'Bundle', type: 'collection', entry: [{resource: patient}]});
      const fields: string[] = [];
      decoded.forEach((_value, key) => fields.push(key));
      expect(fields).toEqual(['file']);
    }
    request.flush({success: true, source: {id: 'local-source'}});
    expect((await result).source.id).toBe('local-source');
    http.verify();
  });

  it('reuses an existing synthetic import instead of creating duplicate local records', async () => {
    const http = TestBed.inject(HttpTestingController);
    const result = service.importSandboxPatient({resourceType: 'Patient', id: 'synthetic-patient'}).toPromise();
    http.expectOne(req => req.url.endsWith('/secure/source')).flush({success: true, data: [{id: 'already-local', patient: 'synthetic-patient', platform_type: 'manual'}]});
    expect(await result).toEqual({source: jasmine.objectContaining({id: 'already-local'}), alreadyImported: true});
    http.expectNone(req => req.url.endsWith('/secure/source/manual'));
    http.verify();
  });

  it('recognizes the pinned importer legacy patient key for the Epic sandbox patient', async () => {
    const http = TestBed.inject(HttpTestingController);
    const result = service.importSandboxPatient({resourceType: 'Patient', id: 'erXuFYUfucBZaryVksYEcMg3'}).toPromise();
    http.expectOne(req => req.url.endsWith('/secure/source')).flush({success: true, data: [
      {id: 'already-local', patient: 'rXuFYUfucBZaryVksYEcMg3', platform_type: 'manual'}
    ]});
    http.expectNone(req => req.url.endsWith('/secure/source/manual'));
    expect(await result).toEqual({source: jasmine.objectContaining({id: 'already-local'}), alreadyImported: true});
    http.verify();
  });

  it('does not reuse a direct provider snapshot as a legacy sandbox source', async () => {
    const http = TestBed.inject(HttpTestingController);
    const result = service.importSandboxPatient({resourceType: 'Patient', id: 'synthetic-patient'}).toPromise();
    http.expectOne(req => req.url.endsWith('/secure/source')).flush({success: true, data: [
      {id: 'direct-source', patient: 'synthetic-patient', platform_type: 'manual', endpoint_id: 'provider-scoped-id'}
    ]});
    http.expectOne(req => req.url.endsWith('/secure/source/manual')).flush({success: true, source: {id: 'sandbox-source'}});
    expect((await result).source.id).toBe('sandbox-source');
    http.verify();
  });

  it('sends a direct Patient snapshot with provider provenance and reads backend deduplication', async () => {
    const http = TestBed.inject(HttpTestingController);
    const patient = {resourceType: 'Patient' as const, id: 'patient-one'};
    const origin = {fhir_base_url: 'https://example.invalid/fhir', display: 'Example hospital'};
    const result = service.importSmartPatient(patient, origin).toPromise();
    http.expectNone(req => req.url.endsWith('/secure/source'));
    const request = http.expectOne(req => req.url.endsWith('/secure/source/manual'));
    const form = await new Response(request.request.body, {headers: {'Content-Type': request.request.headers.get('Content-Type')}}).formData();
    expect(JSON.parse(String(form.get('smart_origin')))).toEqual(origin);
    expect(JSON.parse(await (form.get('file') as File).text())).toEqual({resourceType: 'Bundle', type: 'collection', entry: [{resource: patient}]});
    request.flush({success: true, already_imported: true, source: {id: 'existing-direct-source'}});
    expect(await result).toEqual({source: jasmine.objectContaining({id: 'existing-direct-source'}), alreadyImported: true});
    http.verify();
  });

});
