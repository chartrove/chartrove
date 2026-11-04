import {TestBed} from '@angular/core/testing';
import {HttpClient, provideHttpClient} from '@angular/common/http';
import {HttpTestingController, provideHttpClientTesting} from '@angular/common/http/testing';
import {HTTP_CLIENT_TOKEN} from '../dependency-injection';
import {DocumentImportService} from './document-import.service';

describe('DocumentImportService review boundary', () => {
  let service: DocumentImportService, http: HttpTestingController;
  beforeEach(() => {
    TestBed.configureTestingModule({providers:[provideHttpClient(),provideHttpClientTesting(),{provide:HTTP_CLIENT_TOKEN,useExisting:HttpClient}]});
    service=TestBed.inject(DocumentImportService); http=TestBed.inject(HttpTestingController);
  });
  afterEach(() => http.verify());
  const file = new File(['synthetic original'],'synthetic.png',{type:'image/png'});
  it('requires review and sends only the original and reviewed fields to the selected source', async () => {
    await expectAsync(service.save('synthetic',file,'Synthetic','Synthetic text',false)).toBeRejected();
    http.expectNone(req => req.method === 'POST');
    const result = service.save('synthetic',file,'Synthetic','Corrected synthetic text',true);
    const request=http.expectOne(req=>req.url.endsWith('/secure/document-import/synthetic'));
    const body=request.request.body as FormData;
    expect(Array.from(body.keys()).sort()).toEqual(['file','reviewed','text','title']);
    expect(await (body.get('file') as File).text()).toBe('synthetic original');
    expect(body.get('text')).toBe('Corrected synthetic text'); expect(body.get('reviewed')).toBe('true');
    expect(request.request.headers.has('Content-Type')).toBeFalse();
    request.flush({success:true,data:{source_resource_id:'synthetic-saved',created:true}});
    expect((await result).source_resource_id).toBe('synthetic-saved');
  });
  it('rejects unconfirmed responses rather than presenting an upload as saved', async () => {
    const result=service.save('synthetic',file,'Synthetic','Synthetic text',true);
    http.expectOne(req=>req.method==='POST').flush({success:false});
    await expectAsync(result).toBeRejected();
    const sources=service.sources(); http.expectOne(req=>req.url.endsWith('/document-import/sources')).flush({success:false});
    await expectAsync(sources).toBeRejected();
  });
});
