import { ComponentFixture, TestBed } from '@angular/core/testing';
import { PractitionerModel } from '../../../lib/models/resources/practitioner-model';

import { MedicalRecordWizardEditPractitionerComponent } from './medical-record-wizard-edit-practitioner.component';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import {HTTP_CLIENT_TOKEN} from '../../dependency-injection';
import { HttpClient, provideHttpClient, withInterceptorsFromDi, withXhr } from '@angular/common/http';

describe('MedicalRecordWizardEditPractitionerComponent', () => {
  let component: MedicalRecordWizardEditPractitionerComponent;
  let fixture: ComponentFixture<MedicalRecordWizardEditPractitionerComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
    imports: [MedicalRecordWizardEditPractitionerComponent],
    providers: [NgbActiveModal, NgbModal, {
            provide: HTTP_CLIENT_TOKEN,
            useClass: HttpClient,
        }, provideHttpClient(withXhr(), withInterceptorsFromDi()), provideHttpClientTesting()]
})
    .compileComponents();

    fixture = TestBed.createComponent(MedicalRecordWizardEditPractitionerComponent);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  for (const address of [undefined, [], [{}]]) {
    it(`opens a practitioner with address ${JSON.stringify(address)}`, () => {
      const raw = {
        resourceType: 'Practitioner',
        name: [{ text: 'Synthetic clinician' }],
        address,
        telecom: [],
      };
      const original = JSON.stringify(raw);
      component.practitioner = new PractitionerModel(raw);

      expect(() => fixture.detectChanges()).not.toThrow();
      expect(component.practitionerForm.get('address.line1').value || '').toBe('');
      expect(component.practitionerForm.get('address.city').value || '').toBe('');
      expect(JSON.stringify(raw)).toBe(original);
    });
  }

  it('allows adding contact details when optional contacts are absent', () => {
    component.practitioner = new PractitionerModel({
      resourceType: 'Practitioner',
      name: [{ text: 'Synthetic clinician' }],
      address: [{ line: ['1 Synthetic Road'], city: 'Example City' }],
    });
    expect(() => fixture.detectChanges()).not.toThrow();
    expect(component.practitionerForm.get('address.line1').value).toBe('1 Synthetic Road');
    component.practitionerForm.patchValue({
      phone: '5550100', fax: '5550101', email: 'synthetic@example.invalid',
    });
    const close = spyOn(component.activeModal, 'close');

    expect(() => component.submit()).not.toThrow();
    expect(component.practitioner.telecom).toEqual([
      { system: 'phone', value: '5550100', use: 'work' },
      { system: 'fax', value: '5550101', use: 'work' },
      { system: 'email', value: 'synthetic@example.invalid', use: 'work' },
    ]);
    expect(component.practitioner.address[0].line).toEqual(['1 Synthetic Road']);
    expect(close).toHaveBeenCalledWith(component.practitioner);
  });

  it('preserves populated address and unrelated contacts while editing a phone', () => {
    component.practitioner = new PractitionerModel({
      resourceType: 'Practitioner',
      name: [{ text: 'Synthetic clinician' }],
      address: [{ line: ['1 Synthetic Road', 'Suite 2'], city: 'Example City',
        state: 'EX', postalCode: '00000', country: 'US' }],
      telecom: [{ system: 'url', value: 'https://example.invalid' },
        { system: 'phone', value: '5550100', use: 'work' }],
    });
    fixture.detectChanges();
    expect(component.practitionerForm.get('address').value).toEqual({
      line1: '1 Synthetic Road', line2: 'Suite 2', city: 'Example City',
      state: 'EX', zip: '00000', country: 'US',
    });
    component.practitionerForm.get('phone').setValue('5550199');
    spyOn(component.activeModal, 'close');
    component.submit();
    expect(component.practitioner.telecom).toEqual([
      { system: 'url', value: 'https://example.invalid' },
      { system: 'phone', value: '5550199', use: 'work' },
    ]);
    expect(component.practitioner.address[0].line).toEqual(['1 Synthetic Road', 'Suite 2']);
  });

  it('opens a practitioner without optional name or qualifications', () => {
    component.practitioner = new PractitionerModel({
      resourceType: 'Practitioner', name: [], qualification: [],
      address: [{}], telecom: [],
    });
    expect(() => fixture.detectChanges()).not.toThrow();
    expect(component.practitionerForm.get('name').value || '').toBe('');
    expect(component.practitionerForm.get('profession').value || '').toBe('');
  });
});
