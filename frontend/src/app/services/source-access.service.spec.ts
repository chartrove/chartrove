import {TestBed} from '@angular/core/testing';
import {HttpClient, provideHttpClient} from '@angular/common/http';
import {HttpTestingController, provideHttpClientTesting} from '@angular/common/http/testing';
import {HTTP_CLIENT_TOKEN} from '../dependency-injection';
import {SourceAccessService} from './source-access.service';

describe('SourceAccessService boundaries', () => {
  let service: SourceAccessService;
  let http: HttpTestingController;
  beforeEach(() => {
    TestBed.configureTestingModule({providers: [provideHttpClient(), provideHttpClientTesting(), {provide: HTTP_CLIENT_TOKEN, useExisting: HttpClient}]});
    service = TestBed.inject(SourceAccessService); http = TestBed.inject(HttpTestingController);
  });
  afterEach(() => http.verify());

  it('keeps search terms in the JSON body and supplies no owner identity', () => {
    service.search('synthetic-source', 'Observation', '100% follow_up!', 1).subscribe(rows => expect(rows.length).toBe(1));
    const request = http.expectOne(req => req.url.endsWith('/source-access/synthetic-source/search/Observation'));
    expect(request.request.method).toBe('POST');
    expect(request.request.params.keys()).toEqual([]);
    expect(request.request.body).toEqual({text: '100% follow_up!', page: 1});
    request.flush({success: true, data: [{source_resource_type: 'Observation', source_resource_id: 'synthetic-hit'}]});
  });

  it('retains only source selection metadata from the owner API', () => {
    service.owned().subscribe(rows => expect(rows).toEqual([{id: 'synthetic', display: 'Source', platform_type: 'fasten'}]));
    http.expectOne(req => req.url.endsWith('/secure/source')).flush({success: true, data: [{id: 'synthetic', display: 'Source', platform_type: 'fasten', access_token: 'synthetic-secret', user: {password: 'synthetic-hash'}}]});
  });
  it('never supplies an owner identity in grant or edit requests', () => {
    service.grant('synthetic-source', 'synthetic-delegate', 'VIEW', null).subscribe();
    const grant = http.expectOne(req => req.url.endsWith('/source-access/synthetic-source/grants'));
    expect(grant.request.body).toEqual({delegate_username: 'synthetic-delegate', access_level: 'VIEW', expires_at: null});
    grant.flush({success: true, data: {}});
    const raw = '{"resourceType":"Observation","id":"synthetic-id","valueQuantity":{"value":2.00}}';
    service.edit('synthetic-source', 'Observation', 'synthetic-id', raw).subscribe();
    const edit = http.expectOne(req => req.url.endsWith('/source-access/synthetic-source/resources/Observation/synthetic-id'));
    expect(edit.request.body).toBe(`{"resource_raw":${raw}}`);
    expect(edit.request.headers.get('Content-Type')).toBe('application/json');
    edit.flush({success: true, data: null});
  });

  it('creates local records without converting clinical numbers to JavaScript numbers', () => {
    const raw = '{"resourceType":"Observation","valueQuantity":{"value":0.1234567890123456789}}';
    service.createLocal('synthetic-source', raw).subscribe();
    const create = http.expectOne(req => req.url.endsWith('/source-access/synthetic-source/resources'));
    expect(create.request.body).toBe(`{"resource_raw":${raw}}`);
    expect(create.request.headers.get('Content-Type')).toBe('application/json');
    create.flush({success: true, data: {}});
  });
});
