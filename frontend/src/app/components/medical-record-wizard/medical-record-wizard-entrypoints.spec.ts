import { HttpClient, provideHttpClient, withXhr } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ApplicationRef } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter, Router } from '@angular/router';
import { NgbModal, NgbModalConfig } from '@ng-bootstrap/ng-bootstrap';

import { EncounterModel } from '../../../lib/models/resources/encounter-model';
import { HTTP_CLIENT_TOKEN } from '../../dependency-injection';
import { ResourceCreatorComponent } from '../../pages/resource-creator/resource-creator.component';
import { ReportMedicalHistoryTimelinePanelComponent } from '../report-medical-history-timeline-panel/report-medical-history-timeline-panel.component';
import { MedicalRecordWizardComponent } from './medical-record-wizard.component';

describe('Medical record wizard entry points', () => {
  let modal: NgbModal;
  let http: HttpTestingController;
  let open: jasmine.Spy<NgbModal['open']>;
  let navigate: jasmine.Spy<Router['navigate']>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MedicalRecordWizardComponent],
      providers: [
        provideHttpClient(withXhr()), provideHttpClientTesting(), provideRouter([]),
        { provide: HTTP_CLIENT_TOKEN, useExisting: HttpClient },
      ],
    }).compileComponents();
    // Exercise real modal/content creation; transition duration is irrelevant
    // to the entry-point and dismissal/navigation contract tested here.
    TestBed.inject(NgbModalConfig).animation = false;
    modal = TestBed.inject(NgbModal);
    http = TestBed.inject(HttpTestingController);
    open = spyOn(modal, 'open').and.callThrough();
    navigate = spyOn(TestBed.inject(Router), 'navigate').and.resolveTo(true);
  });

  afterEach(() => {
    for (const call of open.calls.all()) call.returnValue.close();
    http.verify();
  });

  async function verifyWizard(): Promise<void> {
    TestBed.inject(ApplicationRef).tick();
    http.expectOne(request => request.url.endsWith('/secure/source')).flush({
      success: true,
      data: [{ id: '11111111-1111-4111-8111-111111111111', platform_type: 'fasten' }],
    });
    await TestBed.inject(ApplicationRef).whenStable();
    expect(open.calls.count()).toBe(1);
    expect(open.calls.mostRecent().args).toEqual([MedicalRecordWizardComponent, { size: 'xl' }]);
    expect(open.calls.mostRecent().returnValue.componentInstance)
      .toBeInstanceOf(MedicalRecordWizardComponent);
    expect(document.querySelector('ngb-modal-window')?.textContent).toContain('Medical Record Wizard');
  }

  for (const outcome of ['saved', 'dismissed'] as const) {
    it(`opens the create route and returns to medical history when ${outcome}`, async () => {
      const creator = new ResourceCreatorComponent(modal, TestBed.inject(Router));
      await creator.ngOnInit();
      await verifyWizard();
      const ref = open.calls.mostRecent().returnValue;
      if (outcome === 'saved') ref.close('synthetic-saved');
      else ref.dismiss('synthetic-dismiss');
      await TestBed.inject(ApplicationRef).whenStable();
      expect(navigate.calls.allArgs()).toEqual([[['/medical-history']]]);
    });
  }

  it('opens the timeline wizard with the same synthetic Encounter', async () => {
    const timeline = new ReportMedicalHistoryTimelinePanelComponent(modal);
    const encounter = new EncounterModel({ resourceType: 'Encounter', id: 'synthetic-entry', status: 'finished' });
    encounter.source_id = '11111111-1111-4111-8111-111111111111';
    encounter.source_resource_id = 'synthetic-entry';
    timeline.displayModel = encounter;
    await timeline.openMedicalRecordWizard();
    await verifyWizard();
    expect(open.calls.mostRecent().returnValue.componentInstance.existingEncounter).toBe(encounter);
    expect(navigate).not.toHaveBeenCalled();
  });

  it('retains the clicked Encounter if the timeline changes during loading', async () => {
    const timeline = new ReportMedicalHistoryTimelinePanelComponent(modal);
    const clicked = new EncounterModel({ resourceType: 'Encounter', id: 'synthetic-clicked', status: 'finished' });
    timeline.displayModel = clicked;
    const opening = timeline.openMedicalRecordWizard();
    timeline.displayModel = new EncounterModel({ resourceType: 'Encounter', id: 'synthetic-other', status: 'finished' });
    await opening;
    await verifyWizard();
    expect(open.calls.mostRecent().returnValue.componentInstance.existingEncounter).toBe(clicked);
  });

  it('does not open a late wizard after the create route is destroyed', async () => {
    const creator = new ResourceCreatorComponent(modal, TestBed.inject(Router));
    const opening = creator.ngOnInit();
    creator.ngOnDestroy();
    await opening;
    expect(open).not.toHaveBeenCalled();
    expect(navigate).not.toHaveBeenCalled();
  });

  it('does not open a late wizard after the timeline is destroyed', async () => {
    const timeline = new ReportMedicalHistoryTimelinePanelComponent(modal);
    const opening = timeline.openMedicalRecordWizard();
    timeline.ngOnDestroy();
    await opening;
    expect(open).not.toHaveBeenCalled();
  });
});
