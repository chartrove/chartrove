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
