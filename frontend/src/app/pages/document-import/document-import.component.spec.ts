import {ComponentFixture, TestBed} from '@angular/core/testing';
import {provideRouter} from '@angular/router';
import {DocumentImportComponent} from './document-import.component';
import {DocumentImportService} from '../../services/document-import.service';
import {LocalDocumentOCRService, LocalExtraction} from '../../services/local-document-ocr.service';
import {AuthService} from '../../services/auth.service';

describe('DocumentImportComponent local review', () => {
  let fixture: ComponentFixture<DocumentImportComponent>, component: DocumentImportComponent;
  let api: jasmine.SpyObj<DocumentImportService>, ocr: jasmine.SpyObj<LocalDocumentOCRService>;
  let subject: string;
  beforeEach(async () => {
    subject='synthetic-owner';
    api=jasmine.createSpyObj('DocumentImportService',['sources','save']);
    api.sources.and.resolveTo([{id:'synthetic-source',display:'Synthetic local source'}]);
    api.save.and.resolveTo({source_id:'synthetic-source',source_resource_type:'DocumentReference',source_resource_id:'synthetic-id',original_sha256:'synthetic-fingerprint',created:true});
    ocr=jasmine.createSpyObj('LocalDocumentOCRService',['extract']);
    ocr.extract.and.resolveTo({text:'Synthetic extracted text',pages:1,mediaType:'image/png'});
    await TestBed.configureTestingModule({imports:[DocumentImportComponent],providers:[provideRouter([]),
      {provide:DocumentImportService,useValue:api},{provide:LocalDocumentOCRService,useValue:ocr},
      {provide:AuthService,useValue:{GetCurrentUser:()=>({sub:subject})}}]}).compileComponents();
    fixture=TestBed.createComponent(DocumentImportComponent);component=fixture.componentInstance;
    component.supported=true; await component.ngOnInit();
  });
  afterEach(()=>fixture.destroy());
  function selection(name='synthetic-scan.png'): Event {
    return {target:{files:[new File(['synthetic original'],name)],value:name}} as unknown as Event;
  }
  it('does no upload until review and retains the original while saving corrected text',async()=>{
    await component.selected(selection()); expect(api.save).not.toHaveBeenCalled();
    expect(component.originalURL.startsWith('blob:')).toBeTrue();
    await component.save();expect(api.save).not.toHaveBeenCalled();
    component.text='Corrected synthetic transcript';component.reviewed=true;await component.save();
    expect(api.save).toHaveBeenCalledWith('synthetic-source',component.file,'synthetic-scan.png','Corrected synthetic transcript',true);
    expect(await component.file.text()).toBe('synthetic original');expect(component.receipt.created).toBeTrue();
    const revoke=spyOn(URL,'revokeObjectURL').and.callThrough();component.clear();
    expect(revoke).toHaveBeenCalled();expect(component.text).toBe('');expect(component.file).toBeUndefined();
  });
  it('cancels pending extraction and ignores its late result',async()=>{
    let resolve:(value:LocalExtraction)=>void;
    ocr.extract.and.returnValue(new Promise(done=>resolve=done));
    const pending=component.selected(selection());const signal=ocr.extract.calls.mostRecent().args[1];
    component.cancel();expect(signal.aborted).toBeTrue();
    resolve({text:'Late private text',pages:1,mediaType:'image/png'});await pending;
    expect(component.text).toBe('');expect(component.originalURL).toBe('');expect(api.save).not.toHaveBeenCalled();
  });
  it('clears private data and refuses submission after the native account changes',async()=>{
    await component.selected(selection());component.reviewed=true;subject='synthetic-other';
    await component.save();expect(api.save).not.toHaveBeenCalled();expect(component.text).toBe('');expect(component.sources).toEqual([]);
  });
  it('discards source results if the account changes during loading',async()=>{
    let resolve:(value:any)=>void;api.sources.and.returnValue(new Promise(done=>resolve=done));
    const loading=component.ngOnInit();subject='synthetic-other';resolve([{id:'stale',display:'Stale source'}]);await loading;
    expect(component.sources).toEqual([]);expect(component.source).toBe('');
  });
  it('resets confirmation when the reviewed transcript changes and renders text without HTML',async()=>{
    await component.selected(selection());component.text='<img src=x onerror=invalid> Synthetic';component.reviewed=true;
    fixture.detectChanges();await fixture.whenStable();
    const area=fixture.nativeElement.querySelector('#document-transcript') as HTMLTextAreaElement;
    area.value='Corrected synthetic text';area.dispatchEvent(new Event('input'));fixture.detectChanges();
    expect(component.reviewed).toBeFalse();expect(fixture.nativeElement.querySelector('img')).toBeNull();
  });
  it('does not fetch sources or offer import in the existing desktop/cloud modes',async()=>{
    api.sources.calls.reset();component.supported=false;await component.ngOnInit();fixture.detectChanges();
    expect(api.sources).not.toHaveBeenCalled();expect(fixture.nativeElement.querySelector('#document-file')).toBeNull();
  });
});
