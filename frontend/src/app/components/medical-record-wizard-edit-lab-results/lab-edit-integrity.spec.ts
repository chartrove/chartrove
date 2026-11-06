import {TestBed} from '@angular/core/testing';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {of} from 'rxjs';
import {MedicalRecordWizardEditLabResultsComponent} from './medical-record-wizard-edit-lab-results.component';
import {NlmClinicalTableSearchService} from '../../services/nlm-clinical-table-search.service';
import {FastenApiService} from '../../services/fasten-api.service';
import {SourceAccessService} from '../../services/source-access.service';
import {ObservationModel} from '../../../lib/models/resources/observation-model';

describe('Lab editor record identity and field integrity', () => {
  const A = '11111111-1111-4111-8111-111111111111';
  const B = '22222222-2222-4222-8222-222222222222';
  let modal: jasmine.SpyObj<NgbActiveModal>;
  let util: {checkValidity: jasmine.Spy; getFormFHIRData: jasmine.Spy};
  let previous: PropertyDescriptor | undefined;

  function raw(id: string, code: string, value: number): string {
    return `{"resourceType":"Observation","id":"${id}","status":"final","code":{"text":"Same synthetic label","coding":[{"system":"http://loinc.org","code":"${code}"}]},"subject":{"reference":"Patient/synthetic"},"effectiveDateTime":"2020-01-02T03:04:05+05:30","note":[{"text":"Retain café صفر"}],"extension":[{"url":"urn:synthetic:exact","valueDecimal":0.1234567890123456789}],"valueQuantity":{"value":${value},"unit":"mg/dL","system":"http://unitsofmeasure.org","code":"mg/dL","extension":[{"url":"urn:synthetic:value-meta","valueString":"keep"}]}}`;
  }
  function observation(source: string, id: string, code: string, value: number): ObservationModel {
    return Object.assign(new ObservationModel(JSON.parse(raw(id,code,value))), {source_id:source,source_resource_id:id});
  }
  function bundle(code: string, value: number) {
    return {resourceType:'Bundle',entry:[{resource:{resourceType:'Observation',id:'generated-widget-id',code:{text:'Same synthetic label',coding:[{system:'http://loinc.org',code}]},valueQuantity:{value,unit:'mg/dL',system:'http://unitsofmeasure.org',code:'mg/dL'}}}]};
  }
  async function editor(records: ObservationModel[], code = 'synthetic-code-2') {
    const questionnaire = {type:'LOINC',code:'synthetic-panel',codeSystem:'http://loinc.org',name:'Synthetic panel',lformsVersion:'34.0.0',items:[{linkId:'/question',question:'Same synthetic label',questionCode:code,questionCodeSystem:'http://loinc.org',dataType:'REAL',units:[{name:'mg/dL',code:'mg/dL',system:'http://unitsofmeasure.org'}]}]};
    await TestBed.configureTestingModule({imports:[MedicalRecordWizardEditLabResultsComponent],providers:[
      {provide:NgbActiveModal,useValue:modal},
      {provide:FastenApiService,useValue:{}},
      {provide:SourceAccessService,useValue:{detail:(source:string,_type:string,id:string)=>{
        const record = records.find(r=>r.source_id===source&&r.source_resource_id===id)!;
        const text = raw(id,record.code.coding[0].code,(record.fhirResource.valueQuantity as any).value);
        return of({source_id:source,source_resource_type:'Observation',source_resource_id:id,resource_raw:JSON.parse(text),resource_raw_text:text,editable:true});
      }}},
      {provide:NlmClinicalTableSearchService,useValue:{searchLabPanels:()=>of([]),searchLabPanelQuestionnaire:()=>of(questionnaire)}}
    ]}).compileComponents();
    const fixture=TestBed.createComponent(MedicalRecordWizardEditLabResultsComponent);
    const c=fixture.componentInstance;
    c.diagnosticReport={source_id:B,source_resource_type:'DiagnosticReport',source_resource_id:'synthetic-report',code:{coding:[{system:'http://loinc.org',code:'synthetic-panel'}]}} as any;
    c.observations=records;
    fixture.detectChanges();
    return c;
  }
  beforeEach(()=>{
    previous=Object.getOwnPropertyDescriptor(globalThis,'LForms');
    modal=jasmine.createSpyObj<NgbActiveModal>('modal',['close','dismiss']);
    util={checkValidity:jasmine.createSpy('validity').and.returnValue(null),getFormFHIRData:jasmine.createSpy('export').and.returnValue(bundle('synthetic-code-2',2))};
    Object.defineProperty(globalThis,'LForms',{configurable:true,value:{Util:util}});
  });
  afterEach(()=>{if(previous)Object.defineProperty(globalThis,'LForms',previous);else delete (globalThis as any).LForms;});

  it('updates only the selected source when another source has the same ID and code',async()=>{
    const c=await editor([observation(A,'same-id','synthetic-code-2',1),observation(B,'same-id','synthetic-code-2',2)]);
    c.onFormReady(); util.getFormFHIRData.and.returnValue(bundle('synthetic-code-2',3)); c.submit();
    expect(modal.close).toHaveBeenCalledTimes(1);
    const plan=modal.close.calls.mostRecent().args[0]; expect(plan.length).toBe(1);
    expect(plan[0].observation.source_id).toBe(B);
  });
  it('preserves original identity, date, references, notes and exact unrelated decimals',async()=>{
    const c=await editor([observation(B,'original-id','synthetic-code-2',2)]);
    c.onFormReady();util.getFormFHIRData.and.returnValue(bundle('synthetic-code-2',3));c.submit();
    expect(modal.close).toHaveBeenCalledTimes(1);
    const text=modal.close.calls.mostRecent().args[0][0]?.resourceRawText;
    expect(typeof text).toBe('string'); if(typeof text!=='string')return;
    expect(text).toContain('0.1234567890123456789');
    const result=JSON.parse(text);
    expect(result.id).toBe('original-id');expect(result.subject.reference).toBe('Patient/synthetic');
    expect(result.effectiveDateTime).toBe('2020-01-02T03:04:05+05:30');expect(result.note[0].text).toBe('Retain café صفر');
    expect(result.valueQuantity.value).toBe(3);expect(result.valueQuantity.extension[0].valueString).toBe('keep');
  });
  it('prefills by coding rather than a duplicate display label',async()=>{
    const c=await editor([observation(B,'first','synthetic-code-1',1),observation(B,'second','synthetic-code-2',2)]);
    expect(c.questionnaire.items[0].defaultAnswer).toBe(2);
  });
  it('refuses ambiguous repeated observations instead of writing both',async()=>{
    const c=await editor([observation(B,'first','synthetic-code-2',1),observation(B,'second','synthetic-code-2',2)]);
    c.onFormReady();util.getFormFHIRData.and.returnValue(bundle('synthetic-code-2',3));c.submit();
    expect(modal.close).not.toHaveBeenCalled();expect(c.formError.length).toBeGreaterThan(0);
  });
});
