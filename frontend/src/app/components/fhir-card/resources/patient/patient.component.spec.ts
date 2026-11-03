import { ComponentFixture, TestBed } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { NgbCollapseModule } from '@ng-bootstrap/ng-bootstrap';

import { PatientModel } from 'src/lib/models/resources/patient-model';
import { PatientComponent } from './patient.component';

describe('PatientComponent', () => {
  let component: PatientComponent;
  let fixture: ComponentFixture<PatientComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ PatientComponent, NgbCollapseModule, RouterTestingModule ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(PatientComponent);
    component = fixture.componentInstance;
    component.displayModel = new PatientModel({});
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  function fieldText(label: string): string {
    const row = Array.from(fixture.nativeElement.querySelectorAll('li') as NodeListOf<HTMLElement>)
      .find(element => element.querySelector('strong')?.textContent.trim() === label);
    return row?.textContent.replace(/\s+/g, ' ').trim();
  }

  function renderSyntheticPatient(raw: object): string {
    const original = JSON.stringify(raw);
    jasmine.clock().install();
    jasmine.clock().mockDate(new Date('2026-11-03T12:00:00.000Z'));
    try {
      component.displayModel = new PatientModel(raw);
      fixture.detectChanges();
      expect(JSON.stringify(raw)).toBe(original);
      return fieldText('Age:');
    } finally {
      jasmine.clock().uninstall();
    }
  }

  for (const birthDate of [undefined, null, '', 'not-a-date']) {
    it(`shows a neutral age for synthetic birthDate ${String(birthDate)}`, () => {
      const raw = {resourceType: 'Patient', id: 'synthetic-detail-age',
        ...(birthDate === undefined ? {} : {birthDate})};
      expect(renderSyntheticPatient(raw)).toBe('Age: Not provided');
      expect(fieldText('Date of Birth:')).toBe('Date of Birth:');
    });
  }

  it('preserves a valid zero age in the patient detail card', () => {
    expect(renderSyntheticPatient({resourceType: 'Patient', birthDate: '2026-11-03'}))
      .toBe('Age: 0');
    expect(fieldText('Date of Birth:')).toBe('Date of Birth: Nov 3, 2026');
  });

  it('preserves a valid adult age in the patient detail card', () => {
    expect(renderSyntheticPatient({resourceType: 'Patient', birthDate: '1996-01-01'}))
      .toBe('Age: 30');
    expect(fieldText('Date of Birth:')).toBe('Date of Birth: Jan 1, 1996');
  });

  it('refreshes from a valid age to a synthetic patient without birthDate', () => {
    expect(renderSyntheticPatient({resourceType: 'Patient', birthDate: '1996-01-01'}))
      .toBe('Age: 30');
    expect(renderSyntheticPatient({resourceType: 'Patient', id: 'synthetic-no-birthdate'}))
      .toBe('Age: Not provided');
  });

});
