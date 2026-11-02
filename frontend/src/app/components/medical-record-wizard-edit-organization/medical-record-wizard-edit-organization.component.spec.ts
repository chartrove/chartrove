import { ComponentFixture, TestBed } from '@angular/core/testing';
import { OrganizationModel } from '../../../lib/models/resources/organization-model';

import { MedicalRecordWizardEditOrganizationComponent } from './medical-record-wizard-edit-organization.component';
import {NgbActiveModal, NgbModal, NgbModalModule} from '@ng-bootstrap/ng-bootstrap';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import {HTTP_CLIENT_TOKEN} from '../../dependency-injection';
import { HttpClient, provideHttpClient, withInterceptorsFromDi, withXhr } from '@angular/common/http';

describe('MedicalRecordWizardEditOrganizationComponent', () => {
  let component: MedicalRecordWizardEditOrganizationComponent;
  let fixture: ComponentFixture<MedicalRecordWizardEditOrganizationComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
    imports: [MedicalRecordWizardEditOrganizationComponent],
    providers: [NgbModal, NgbActiveModal, {
            provide: HTTP_CLIENT_TOKEN,
            useClass: HttpClient,
        }, provideHttpClient(withXhr(), withInterceptorsFromDi()), provideHttpClientTesting()]
})
    .compileComponents();

    fixture = TestBed.createComponent(MedicalRecordWizardEditOrganizationComponent);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  for (const address of [undefined, [], [{}]]) {
    it(`opens a organization with address ${JSON.stringify(address)}`, () => {
      const raw = {
        resourceType: 'Organization',
        name: 'Synthetic organization',
        address,
        telecom: [],
      };
      const original = JSON.stringify(raw);
      component.organization = new OrganizationModel(raw);

      expect(() => fixture.detectChanges()).not.toThrow();
      expect(component.organizationForm.get('address.line1').value || '').toBe('');
      expect(component.organizationForm.get('address.city').value || '').toBe('');
      expect(JSON.stringify(raw)).toBe(original);
    });
  }

  it('allows adding contact details when optional contacts are absent', () => {
    component.organization = new OrganizationModel({
      resourceType: 'Organization',
      name: 'Synthetic organization',
      address: [{ line: ['1 Synthetic Road'], city: 'Example City' }],
    });
    expect(() => fixture.detectChanges()).not.toThrow();
    expect(component.organizationForm.get('address.line1').value).toBe('1 Synthetic Road');
    component.organizationForm.patchValue({
      phone: '5550100', fax: '5550101', email: 'synthetic@example.invalid',
    });
    const close = spyOn(component.activeModal, 'close');

    expect(() => component.submit()).not.toThrow();
    expect(component.organization.telecom).toEqual([
      { system: 'phone', value: '5550100', use: 'work' },
      { system: 'fax', value: '5550101', use: 'work' },
      { system: 'email', value: 'synthetic@example.invalid', use: 'work' },
    ]);
    expect(component.organization.addresses[0].line).toEqual(['1 Synthetic Road']);
    expect(close).toHaveBeenCalledWith(component.organization);
  });

  it('preserves populated address and unrelated contacts while editing a phone', () => {
    component.organization = new OrganizationModel({
      resourceType: 'Organization',
      name: 'Synthetic organization',
      address: [{ line: ['1 Synthetic Road', 'Suite 2'], city: 'Example City',
        state: 'EX', postalCode: '00000', country: 'US' }],
      telecom: [{ system: 'url', value: 'https://example.invalid' },
        { system: 'phone', value: '5550100', use: 'work' }],
    });
    fixture.detectChanges();
    expect(component.organizationForm.get('address').value).toEqual({
      line1: '1 Synthetic Road', line2: 'Suite 2', city: 'Example City',
      state: 'EX', zip: '00000', country: 'US',
    });
    component.organizationForm.get('phone').setValue('5550199');
    spyOn(component.activeModal, 'close');
    component.submit();
    expect(component.organization.telecom).toEqual([
      { system: 'url', value: 'https://example.invalid' },
      { system: 'phone', value: '5550199', use: 'work' },
    ]);
    expect(component.organization.addresses[0].line).toEqual(['1 Synthetic Road', 'Suite 2']);
  });
});
