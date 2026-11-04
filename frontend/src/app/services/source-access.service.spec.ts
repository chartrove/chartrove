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

  it('retains only source selection metadata from the owner API', () => {
    service.owned().subscribe(rows => expect(rows).toEqual([{id: 'synthetic', display: 'Source', platform_type: 'fasten'}]));
    http.expectOne(req => req.url.endsWith('/secure/source')).flush({success: true, data: [{id: 'synthetic', display: 'Source', platform_type: 'fasten', access_token: 'synthetic-secret', user: {password: 'synthetic-hash'}}]});
  });
  it('never supplies an owner identity in grant or edit requests', () => {
    service.grant('synthetic-source', 'synthetic-delegate', 'VIEW', null).subscribe();
    const grant = http.expectOne(req => req.url.endsWith('/source-access/synthetic-source/grants'));
    expect(grant.request.body).toEqual({delegate_username: 'synthetic-delegate', access_level: 'VIEW', expires_at: null});
    grant.flush({success: true, data: {}});
    const raw = {resourceType: 'Observation', id: 'synthetic-id'};
    service.edit('synthetic-source', 'Observation', 'synthetic-id', raw).subscribe();
    const edit = http.expectOne(req => req.url.endsWith('/source-access/synthetic-source/resources/Observation/synthetic-id'));
    expect(edit.request.body).toEqual({resource_raw: raw}); edit.flush({success: true, data: null});
  });
});
