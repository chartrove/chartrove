import { ComponentFixture, TestBed } from '@angular/core/testing';

import { MedicalRecordWizardComponent } from './medical-record-wizard.component';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import {HTTP_CLIENT_TOKEN} from '../../dependency-injection';
import { HttpClient, provideHttpClient, withInterceptorsFromDi, withXhr } from '@angular/common/http';
import {of} from 'rxjs';

describe('MedicalRecordWizardComponent', () => {
  let component: MedicalRecordWizardComponent;
  let fixture: ComponentFixture<MedicalRecordWizardComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
    imports: [MedicalRecordWizardComponent],
    providers: [NgbActiveModal, NgbModal, {
            provide: HTTP_CLIENT_TOKEN,
            useClass: HttpClient,
        }, provideHttpClient(withXhr(), withInterceptorsFromDi()), provideHttpClientTesting()]
})
    .compileComponents();

    fixture = TestBed.createComponent(MedicalRecordWizardComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});

describe('Source-bound medical record editing', () => {
  it('selects the source-specific lab report and sends only allowed native tuples',async()=>{
    const sourceA='11111111-1111-4111-8111-111111111111',sourceB='22222222-2222-4222-8222-222222222222';
    const report=(source:string)=>({source_id:source,source_resource_type:'DiagnosticReport',source_resource_id:'same-report',resource_raw:{resourceType:'DiagnosticReport',id:'same-report',status:'final',code:{text:'Synthetic'}},related_resources:[{source_id:source,source_resource_type:'Observation',source_resource_id:'same-observation',resource_raw:{resourceType:'Observation',id:'same-observation',status:'final',code:{text:'Synthetic'},valueQuantity:{value:2}}}]});
    const instance:any={}, selected=report(sourceB);
    const raw='{"resourceType":"Observation","id":"same-observation","valueQuantity":{"value":3.5}}';
    const modal={open:()=>({componentInstance:instance,result:Promise.resolve([{observation:{source_id:sourceB,source_resource_type:'Observation',source_resource_id:'same-observation'},resourceRawText:raw}])})};
    const update=jasmine.createSpy('updateResourceText').and.returnValue(of({success:true}));
    const component=new MedicalRecordWizardComponent({} as any,modal as any,{getResourceGraph:()=>of({results:{DiagnosticReport:[report(sourceA),selected]}}),updateResourceText:update} as any);
    component.openEditLabResultsModal({...selected} as any);await Promise.resolve();
    expect(instance.observations.length).toBe(1);expect(instance.observations[0].source_id).toBe(sourceB);
    expect(update).toHaveBeenCalledOnceWith('Observation','same-observation',sourceB,raw);
  });

  it('refuses a changed-source lab plan before dispatching any write',async()=>{
    const instance:any={}, update=jasmine.createSpy('updateResourceText');spyOn(window,'alert');
    const report={source_id:'selected-source',source_resource_type:'DiagnosticReport',source_resource_id:'report',resource_raw:{resourceType:'DiagnosticReport',id:'report',code:{text:'Synthetic'},status:'final'},related_resources:[{source_id:'selected-source',source_resource_type:'Observation',source_resource_id:'same-id',resource_raw:{resourceType:'Observation',id:'same-id',status:'final',code:{text:'Synthetic'},valueQuantity:{value:2}}}]};
    const modal={open:()=>({componentInstance:instance,result:Promise.resolve([{observation:{source_id:'another-source',source_resource_type:'Observation',source_resource_id:'same-id'},resourceRawText:'{}'}])})};
    const component=new MedicalRecordWizardComponent({} as any,modal as any,{getResourceGraph:()=>of({results:{DiagnosticReport:[report]}}),updateResourceText:update} as any);
    component.openEditLabResultsModal(report as any);await Promise.resolve();
    expect(update).not.toHaveBeenCalled();expect(window.alert).toHaveBeenCalled();
  });
  it('keeps the original source when the edited model changes its metadata', async () => {
    const selected = {source_resource_type: 'Practitioner', source_resource_id: 'shared-id', source_id: 'synthetic-source-B', name: []};
    const edited = {...selected, source_id: 'synthetic-source-A'};
    const update = jasmine.createSpy('updateResource').and.returnValue(of({success: true}));
    const modal = {open: () => ({componentInstance: {}, result: Promise.resolve(edited)})};
    const component = new MedicalRecordWizardComponent({} as any, modal as any, {updateResource: update} as any);
    component.openEditPractitionerModal(selected as any);
    await Promise.resolve();
    expect(update.calls.mostRecent().args[2].source_id).toBe('synthetic-source-B');
  });

  it('matches relative MedicationRequest references inside the selected source', () => {
    const first = {source_id: 'synthetic-source-A', medication_reference: {reference: 'Medication/shared-id'}};
    const second = {source_id: 'synthetic-source-B', medication_reference: {reference: 'Medication/shared-id'}};
    const instance: any = {};
    const modal = {open: () => ({componentInstance: instance, result: new Promise(() => {})})};
    const component = new MedicalRecordWizardComponent({} as any, modal as any, {} as any);
    component.existingEncounter = {related_resources: {MedicationRequest: [first, second]}} as any;
    component.openEditMedicationModal({source_id: 'synthetic-source-B', source_resource_id: 'shared-id'} as any);
    expect(instance.medicationRequest).toBe(second);
  });
});
