import { ComponentFixture, TestBed } from '@angular/core/testing';
import { HttpClient, provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideRouter } from '@angular/router';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import { HTTP_CLIENT_TOKEN } from '../../dependency-injection';
import { EncounterModel } from '../../../lib/models/resources/encounter-model';
import { MedicalRecordWizardComponent } from './medical-record-wizard.component';

const sourceID = '11111111-1111-4111-8111-111111111111';

describe('Medical record wizard section validation', () => {
  let fixture: ComponentFixture<MedicalRecordWizardComponent>;
  let http: HttpTestingController;
  let close: jasmine.Spy;

  beforeEach(async () => {
    close = jasmine.createSpy('close');
    await TestBed.configureTestingModule({
      imports: [MedicalRecordWizardComponent],
      providers: [
        provideHttpClient(), provideHttpClientTesting(), provideRouter([]),
        { provide: HTTP_CLIENT_TOKEN, useExisting: HttpClient },
        { provide: NgbActiveModal, useValue: { close, dismiss: jasmine.createSpy() } },
      ],
    }).compileComponents();
    http = TestBed.inject(HttpTestingController);
    fixture = TestBed.createComponent(MedicalRecordWizardComponent);
    fixture.detectChanges();
    http.expectOne(request => request.url.endsWith('/secure/source')).flush({
      success: true, data: [{ id: sourceID, platform_type: 'fasten' }],
    });
  });

  afterEach(() => http.verify());

  function tab(label: string): HTMLElement {
    const tabs = [...fixture.nativeElement.querySelectorAll('[role="tab"]')] as HTMLElement[];
    const match = tabs.find(element => element.textContent.includes(label));
    expect(match).withContext(label).toBeDefined();
    return match;
  }

  function chooseSyntheticEncounter(): void {
    const encounter = new EncounterModel({ resourceType: 'Encounter', id: 'synthetic-validation', status: 'finished' });
    encounter.source_id = sourceID;
    encounter.source_resource_id = 'synthetic-validation';
    fixture.componentInstance.addEncounter({ data: encounter, action: 'find' });
    fixture.detectChanges();
  }

  it('explains an initially disabled save button through the invalid encounter section', () => {
    expect(fixture.nativeElement.querySelector('button[type="submit"]').disabled).toBeTrue();
    expect(tab('Encounter').getAttribute('aria-invalid')).toBe('true');
    expect(fixture.nativeElement.querySelector('[role="alert"][data-validation-feedback]')?.textContent)
      .toContain('highlighted');
    http.expectNone(request => request.url.endsWith('/secure/resource/related'));
    expect(close).not.toHaveBeenCalled();
  });

  it('identifies an incomplete medication in an inactive section and keeps saving blocked', () => {
    chooseSyntheticEncounter();
    fixture.componentInstance.addMedication();
    fixture.detectChanges();
    expect(fixture.componentInstance.active).toBe('encounter');
    expect(tab('Medications').getAttribute('aria-invalid')).toBe('true');
    expect(tab('Encounter').getAttribute('aria-invalid')).not.toBe('true');
    expect(fixture.nativeElement.querySelector('button[type="submit"]').disabled).toBeTrue();
    fixture.componentInstance.onSubmit();
    http.expectNone(request => request.url.endsWith('/secure/resource/related'));
    expect(close).not.toHaveBeenCalled();
  });

  it('removes the invalid-section guidance when the incomplete row is removed', () => {
    chooseSyntheticEncounter();
    fixture.componentInstance.addMedication();
    fixture.detectChanges();
    expect(tab('Medications').getAttribute('aria-invalid')).toBe('true');
    fixture.componentInstance.deleteMedication(0);
    fixture.detectChanges();
    expect(tab('Medications').getAttribute('aria-invalid')).not.toBe('true');
    expect(fixture.nativeElement.querySelector('[data-validation-feedback]')).toBeNull();
    expect(fixture.nativeElement.querySelector('button[type="submit"]').disabled).toBeFalse();
    expect(close).not.toHaveBeenCalled();
  });
});
