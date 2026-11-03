import { HttpClient, provideHttpClient, withXhr } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';

import { EncounterModel } from '../../../lib/models/resources/encounter-model';
import { HTTP_CLIENT_TOKEN } from '../../dependency-injection';
import type { ResourceCreateAttachment } from '../../models/fasten/resource_create';
import { MedicalRecordWizardComponent } from './medical-record-wizard.component';

describe('Medical record wizard attachment selections', () => {
  const sourceID = '11111111-1111-4111-8111-111111111111';
  const firstID = '22222222-2222-4222-8222-222222222222';
  const secondID = '33333333-3333-4333-8333-333333333333';
  let fixture: ComponentFixture<MedicalRecordWizardComponent>;
  let http: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MedicalRecordWizardComponent],
      providers: [
        provideHttpClient(withXhr()), provideHttpClientTesting(), provideRouter([]),
        { provide: HTTP_CLIENT_TOKEN, useExisting: HttpClient },
        NgbActiveModal,
      ],
    }).compileComponents();
    http = TestBed.inject(HttpTestingController);
    fixture = TestBed.createComponent(MedicalRecordWizardComponent);
    fixture.detectChanges();
    http.expectOne(request => request.url.endsWith('/secure/source')).flush({
      success: true, data: [{ id: sourceID, platform_type: 'fasten' }],
    });
    const encounter = new EncounterModel({ resourceType: 'Encounter', id: 'synthetic-selection', status: 'finished' });
    encounter.source_id = sourceID;
    encounter.source_resource_id = 'synthetic-selection';
    fixture.componentInstance.addEncounter({ data: encounter, action: 'find' });
    fixture.detectChanges();
  });

  afterEach(() => {
    const root: HTMLElement = fixture.nativeElement;
    fixture.destroy();
    http.verify();
    // Angular removes its host during TestBed teardown. An inline panel may
    // remain inside that destroyed host; appendTo=body panels must be gone.
    for (const panel of document.querySelectorAll('ng-dropdown-panel')) {
      expect(root.contains(panel)).toBeTrue();
    }
    root.remove();
    expect(document.querySelector('ng-dropdown-panel')).toBeNull();
  });

  function requiredElement(root: ParentNode, selector: string): HTMLElement {
    const element = root.querySelector<HTMLElement>(selector);
    if (!element) throw new Error(`Missing attachment selection element: ${selector}`);
    return element;
  }

  async function stable(): Promise<void> {
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();
  }

  async function chooseSection(section: 'medications' | 'procedures'): Promise<HTMLElement> {
    const component = fixture.componentInstance;
    if (section === 'medications') component.addMedication();
    else component.addProcedure();
    await stable();
    const root: HTMLElement = fixture.nativeElement;
    const label = section === 'medications' ? 'Medications' : 'Procedures';
    const tab = [...root.querySelectorAll<HTMLElement>('[role="tab"]')]
      .find(element => element.textContent.includes(label));
    if (!tab) throw new Error(`Missing wizard section: ${label}`);
    tab.click();
    await stable();
    return requiredElement(root, 'ng-select');
  }

  async function open(select: HTMLElement): Promise<void> {
    if (!document.querySelector('ng-dropdown-panel')) {
      requiredElement(select, '.ng-select-container')
        .dispatchEvent(new MouseEvent('mousedown', { bubbles: true, button: 0 }));
      await stable();
    }
  }

  function options(): HTMLElement[] {
    return [...document.querySelectorAll<HTMLElement>('ng-dropdown-panel .ng-option:not(.ng-option-disabled)')];
  }

  async function choose(select: HTMLElement, label: string): Promise<void> {
    await open(select);
    const option = options().find(element => element.textContent.includes(label));
    if (!option) throw new Error(`Missing synthetic attachment option: ${label}`);
    option.click();
    await stable();
  }

  function attachment(id: string, name: string): ResourceCreateAttachment {
    return {
      id, name, identifier: [], category: { id: 'synthetic-text', text: 'Synthetic text' },
      file_type: 'text/plain', file_name: `${name}.txt`,
      file_content: 'data:text/plain;base64,U3ludGhldGlj', file_size: 9,
    };
  }

  for (const section of ['medications', 'procedures'] as const) {
    it(`keeps the ${section} attachment picker read-only when no attachment exists`, async () => {
      const select = await chooseSection(section);
      const before = fixture.componentInstance.form.getRawValue();
      await open(select);
      expect(document.querySelector('ng-dropdown-panel')).toBeNull();
      expect(fixture.componentInstance.form.getRawValue()).toEqual(before);
    });

    it(`selects, hides and removes ${section} attachment IDs without changing attachment data`, async () => {
      const select = await chooseSection(section);
      const component = fixture.componentInstance;
      component.addAttachment(attachment(firstID, 'Synthetic one'));
      component.addAttachment(attachment(secondID, 'Synthetic two'));
      await stable();
      const originalAttachments = component.attachments.getRawValue();
      const control = component[section].at(0).get('attachments');

      await choose(select, 'Synthetic one');
      expect(control.value).toEqual([firstID]);
      await open(select);
      expect(options().some(element => element.textContent.includes('Synthetic one'))).toBeFalse();
      await choose(select, 'Synthetic two');
      expect(control.value).toEqual([firstID, secondID]);
      expect(select.querySelectorAll('.ng-value-label').length).toBe(2);

      const selected = [...select.querySelectorAll<HTMLElement>('.ng-value')]
        .find(element => element.textContent.includes('Synthetic one'));
      if (!selected) throw new Error('Missing selected synthetic attachment');
      requiredElement(selected, '.ng-value-icon').click();
      await stable();
      expect(control.value).toEqual([secondID]);
      await open(select);
      expect(options().some(element => element.textContent.includes('Synthetic one'))).toBeTrue();
      expect(options().some(element => element.textContent.includes('Synthetic two'))).toBeFalse();
      expect(component.attachments.getRawValue()).toEqual(originalAttachments);
      expect(JSON.parse(JSON.stringify(component.form.getRawValue()))[section][0].attachments)
        .toEqual([secondID]);
    });
  }
});
