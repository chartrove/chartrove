import { ComponentFixture, TestBed } from '@angular/core/testing';
import { HttpClient, provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import { HTTP_CLIENT_TOKEN } from '../../dependency-injection';
import { EncounterModel } from '../../../lib/models/resources/encounter-model';
import { MedicalRecordWizardComponent } from './medical-record-wizard.component';

const sourceID = '11111111-1111-4111-8111-111111111111';

describe('Medical record save lifecycle', () => {
  let fixture: ComponentFixture<MedicalRecordWizardComponent>;
  let http: HttpTestingController;
  let close: jasmine.Spy;

  beforeEach(async () => {
    close = jasmine.createSpy('close');
    await TestBed.configureTestingModule({
      imports: [MedicalRecordWizardComponent],
      providers: [
        provideHttpClient(), provideHttpClientTesting(),
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

  function chooseSyntheticEncounter(): void {
    const encounter = new EncounterModel({ resourceType: 'Encounter', id: 'synthetic-encounter', status: 'finished' });
    encounter.source_id = sourceID;
    encounter.source_resource_id = 'synthetic-encounter';
    fixture.componentInstance.form.get('encounter').setValue({ data: encounter, action: 'find' });
    fixture.detectChanges();
  }

  function submit(): void {
    const form: HTMLFormElement = fixture.nativeElement.querySelector('form');
    form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    fixture.detectChanges();
  }

  it('blocks saving before an encounter is selected without throwing', () => {
    const button: HTMLButtonElement = fixture.nativeElement.querySelector('button[type="submit"]');
    expect(button.disabled).toBeTrue();
    expect(() => submit()).not.toThrow();
    http.expectNone(request => request.url.endsWith('/secure/resource/related'));
    expect(fixture.componentInstance.submitWizardLoading).toBeFalse();
  });

  it('sends one request for repeated form submissions and shows saving status', () => {
    chooseSyntheticEncounter();
    submit();
    submit();
    const requests = http.match(request => request.url.endsWith('/secure/resource/related'));
    expect(requests.length).toBe(1);
    const root: HTMLElement = fixture.nativeElement;
    expect(root.querySelector<HTMLButtonElement>('button[type="submit"]').disabled).toBeTrue();
    expect(root.querySelector('[role="status"]')?.textContent).toContain('Saving');
    for (const request of requests) request.flush({ success: true, data: {} });
    expect(close).toHaveBeenCalledTimes(1);
  });

  it('allows retry after failure while preserving the unsaved form', () => {
    chooseSyntheticEncounter();
    const before = JSON.stringify(fixture.componentInstance.form.getRawValue());
    submit();
    http.expectOne(request => request.url.endsWith('/secure/resource/related')).flush({}, { status: 500, statusText: 'Synthetic failure' });
    fixture.detectChanges();
    expect(close).not.toHaveBeenCalled();
    expect(fixture.componentInstance.submitWizardLoading).toBeFalse();
    expect(JSON.stringify(fixture.componentInstance.form.getRawValue())).toBe(before);
    expect(fixture.nativeElement.querySelector('button[type="submit"]').disabled).toBeFalse();
    submit();
    http.expectOne(request => request.url.endsWith('/secure/resource/related')).flush({ success: true, data: {} });
    expect(close).toHaveBeenCalledTimes(1);
  });

  it('does not automatically log the synthetic clinical resource during serialization', () => {
    chooseSyntheticEncounter();
    const warning = spyOn(console, 'warn');
    submit();
    http.expectOne(request => request.url.endsWith('/secure/resource/related')).flush({ success: true, data: {} });
    expect(warning).not.toHaveBeenCalled();
  });
});
