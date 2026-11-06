import {TestBed} from '@angular/core/testing';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {of} from 'rxjs';
import {MedicalRecordWizardAddLabResultsComponent} from './medical-record-wizard-add-lab-results.component';
import {NlmClinicalTableSearchService} from '../../services/nlm-clinical-table-search.service';
import {GenerateR4ResourceLookup} from '../medical-record-wizard/medical-record-wizard.utilities';
import {ObservationModel} from '../../../lib/models/resources/observation-model';
import {ObservationTableComponent} from '../fhir-card/common/observation-table/observation-table.component';
import {ObservationBarChartComponent} from '../fhir-card/common/observation-bar-chart/observation-bar-chart.component';

describe('Explicit dates for newly entered lab results (Fasten #617 / Chartrove #186)',()=>{
  let previous: PropertyDescriptor | undefined;
  let close: jasmine.Spy;
  const observation={resourceType:'Observation',id:'synthetic-observation',status:'final',code:{text:'Synthetic lab'},valueQuantity:{value:0,unit:'mg/dL'},note:[{text:'Retain café صفر'}]};
  const bundle=()=>({resourceType:'Bundle',type:'transaction',entry:[{resource:{resourceType:'DiagnosticReport',id:'synthetic-report',status:'final',code:{text:'Synthetic panel'},result:[{reference:'Observation/synthetic-observation'}]}},{resource:{...observation}}]});
  async function ready(data=bundle()) {
    close=jasmine.createSpy('close');
    Object.defineProperty(globalThis,'LForms',{configurable:true,value:{Util:{checkValidity:()=>null,getFormFHIRData:()=>data}}});
    await TestBed.configureTestingModule({imports:[MedicalRecordWizardAddLabResultsComponent],providers:[
      {provide:NgbActiveModal,useValue:{close,dismiss:()=>{}}},
      {provide:NlmClinicalTableSearchService,useValue:{searchLabPanels:()=>of([]),searchLabPanelQuestionnaire:()=>of({type:'LOINC',code:'synthetic-panel',codeSystem:'http://loinc.org',items:[]})}},
    ]}).compileComponents();
    const fixture=TestBed.createComponent(MedicalRecordWizardAddLabResultsComponent);
    fixture.detectChanges();fixture.componentInstance.newLabPanelTypeaheadForm.get('data').setValue({id:'synthetic-panel'});fixture.detectChanges();
    fixture.nativeElement.querySelector('wc-lhc-form').dispatchEvent(new CustomEvent('onFormReady'));fixture.detectChanges();
    return fixture;
  }
  beforeEach(()=>previous=Object.getOwnPropertyDescriptor(globalThis,'LForms'));
  afterEach(()=>{if(previous)Object.defineProperty(globalThis,'LForms',previous);else delete (globalThis as any).LForms;});

  it('carries an explicitly entered leap-day through the creation pipeline into the Labs table and chart',async()=>{
    const original=bundle(), before=JSON.stringify(original),fixture=await ready(original);
    const input=fixture.nativeElement.querySelector('input[type="date"]') as HTMLInputElement;
    expect(input).withContext('The add-lab dialog must offer an explicit optional collection date').not.toBeNull();if(!input)return;
    expect(input.value).toBe('');input.value='2020-02-29';input.dispatchEvent(new Event('input'));fixture.detectChanges();
    fixture.nativeElement.querySelector('.modal-footer button').click();expect(close).toHaveBeenCalledTimes(1);
    const submitted=close.calls.mostRecent().args[0];
    const lookup=GenerateR4ResourceLookup({encounter:{action:'create',data:{source_resource_id:'synthetic-encounter',period_start:'2024-12-31'}},attachments:[],organizations:[],practitioners:[],medications:[],procedures:[],labresults:[submitted]} as any);
    const raw=lookup['Observation']['synthetic-observation'] as any;
    expect(raw.effectiveDateTime).toBe('2020-02-29');expect((lookup['DiagnosticReport']['synthetic-report'] as any).effectiveDateTime).toBe('2020-02-29');
    expect(raw.valueQuantity).toEqual(observation.valueQuantity);expect(raw.note).toEqual(observation.note);
    expect(JSON.stringify(original)).toBe(before);
    const model=new ObservationModel(raw),table=new ObservationTableComponent(),chart=new ObservationBarChartComponent();
    table.observations=[model];table.ngOnInit();chart.observations=[model];chart.ngOnInit();
    expect(table.rows[0][0]).toBe('Feb 29, 2020');expect(chart.barChartLabels).toEqual(['Feb 29, 2020']);
  });

  it('keeps unknown dates unknown when the optional date is blank',async()=>{
    const original=bundle(),fixture=await ready(original);fixture.nativeElement.querySelector('.modal-footer button').click();
    expect(close.calls.mostRecent().args[0].data).toEqual(original);
    const table=new ObservationTableComponent();table.observations=[new ObservationModel(original.entry[1].resource)];table.ngOnInit();expect(table.rows[0][0]).toBe('Unknown date');
  });

  it('refuses impossible or ambiguous dates before export or save',async()=>{
    const fixture=await ready(),component=fixture.componentInstance;
    for(const date of ['2023-02-29','1900-02-29','2020-04-31','0000-01-01','2020-13-01','2020-01-00','2020-2-01','02/01/2020','2020-02','2020-02-29T00:00:00Z',' 2020-02-29','10000-01-01']) {
      component.collectionDate.setValue(date);fixture.detectChanges();
      expect(component.collectionDate.invalid).withContext(date).toBeTrue();
      expect(fixture.nativeElement.querySelector('.modal-footer button').disabled).toBeTrue();
      component.submit();expect(close).not.toHaveBeenCalled();
    }
    for(const date of ['0001-01-01','1900-02-28','2000-02-29','2024-02-29','9999-12-31','']) {
      component.collectionDate.setValue(date);expect(component.collectionDate.valid).withContext(date).toBeTrue();
    }
  });

  it('preserves existing effective choices, primitive metadata, issuance and unrelated resources',async()=>{
    const original:any=bundle();
    original.entry[0].resource.effectivePeriod={start:'2019-01-01',end:'2019-01-02'};
    original.entry[0].resource.issued='2019-01-03T01:02:03+05:30';
    const choices=[{effectiveDateTime:'2018-02-03T04:05:06-07:00'},{effectivePeriod:{start:'2018-02-03'}},{effectiveTiming:{code:{text:'Source timing'}}},{effectiveInstant:'2018-02-03T04:05:06Z'},{_effectiveDateTime:{extension:[{url:'urn:synthetic:unknown',valueString:'retain'}]}}];
    original.entry=original.entry.slice(0,1).concat(choices.map((choice,i)=>({resource:{...observation,id:'synthetic-'+i,...choice}})),[{resource:{resourceType:'Patient',id:'synthetic-patient'}}]);
    const before=JSON.stringify(original),fixture=await ready(original);
    fixture.componentInstance.collectionDate.setValue('2020-02-29');fixture.detectChanges();fixture.componentInstance.submit();
    expect(close.calls.mostRecent().args[0].data).toEqual(original);expect(JSON.stringify(original)).toBe(before);
  });

  it('does not reload the panel or discard answers when its optional date changes',async()=>{
    const fixture=await ready(),widget=fixture.nativeElement.querySelector('wc-lhc-form'),q=fixture.componentInstance.questionnaire;
    const input=fixture.nativeElement.querySelector('input[type="date"]') as HTMLInputElement;
    for(const value of ['2020-02-29','2021-01-02','']) {
      input.value=value;input.dispatchEvent(new Event('input'));fixture.detectChanges();
      expect(fixture.nativeElement.querySelector('wc-lhc-form')).toBe(widget);expect(fixture.componentInstance.questionnaire).toBe(q);
      expect(fixture.componentInstance.formReady).toBeTrue();
    }
  });
});
