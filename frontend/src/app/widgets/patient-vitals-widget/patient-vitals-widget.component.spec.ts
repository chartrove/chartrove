import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PatientVitalsWidgetComponent } from './patient-vitals-widget.component';
import {FastenApiService} from '../../services/fasten-api.service';
import {HTTP_CLIENT_TOKEN} from '../../dependency-injection';
import { HttpClient } from '@angular/common/http';
import {RouterTestingModule} from '@angular/router/testing';
import {of} from 'rxjs';
import patientVitalsObservationFixture from "../fixtures/patient_vitals_observation.json"
import patientVitalsPatientFixture from "../fixtures/patient_vitals_patient.json"

describe('PatientVitalsWidgetComponent', () => {
  let component: PatientVitalsWidgetComponent;
  let fixture: ComponentFixture<PatientVitalsWidgetComponent>;
  let mockedFastenApiService

  beforeEach(async () => {
    mockedFastenApiService = jasmine.createSpyObj('FastenApiService', ['queryResources'])
    await TestBed.configureTestingModule({
      imports: [ PatientVitalsWidgetComponent, RouterTestingModule ],
      providers: [
        {
          provide: FastenApiService,
          useValue: mockedFastenApiService
        },
        {
          provide: HTTP_CLIENT_TOKEN,
          useClass: HttpClient,
        },
      ]
    })
    .compileComponents();
    mockedFastenApiService.queryResources.and.callFake(query => of(
      query.from === 'Patient' ? {success: true, data: []} : patientVitalsObservationFixture
    ));


    fixture = TestBed.createComponent(PatientVitalsWidgetComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });


  describe('vitalsProcessQueryResults', () => {
    describe('PatientVitals - PatientVitalsWidget', () => {
      it('should parse data', () => {
        expect(component).toBeTruthy();

        //setup
        //mock moment() date - see: https://github.com/moment/moment/blob/2e2a5b35439665d4b0200143d808a7c26d6cd30f/src/test/moment/now.js#L15


        //test

        try {

          let processedVitalsQueryResponse = component.processQueryResourcesSelectClause(component.widgetConfig.queries[0].q, patientVitalsObservationFixture)
          let processedPatientQueryResponse = component.processQueryResourcesSelectClause(component.widgetConfig.queries[1].q, patientVitalsPatientFixture)
          component.vitalsProcessQueryResults([processedVitalsQueryResponse, processedPatientQueryResponse])
        } finally {

        }

        //assert
        // name: string = ''
        // age: string = ''
        // gender: string = ''
        // vitalSigns: {
        //   display: string,
        //     code: string,
        //     date: string,
        //     value: string,
        //     unit: string
        // }[] = []


        expect(component.name).toEqual('Abraham100 Heller342')
        // expect(component.age).toEqual('21 years') //TODO: mocking date.now or moment() not working.
        expect(component.gender).toEqual('male')
        expect(component.vitalSigns.length).toEqual(16)
      });
    })
  })
  describe('age from projected Patient birth dates', () => {
    beforeEach(() => {
      jasmine.clock().install();
      jasmine.clock().mockDate(new Date('2026-06-15T12:00:00Z'));
    });

    afterEach(() => jasmine.clock().uninstall());

    function projectPatient(birthDate: unknown, includeBirthDate = true) {
      const patient: any = {
        resourceType: 'Patient', id: 'synthetic-vitals-patient',
        name: [{use: 'official', family: 'Synthetic', given: ['Patient']}],
        gender: 'unknown'
      };
      if (includeBirthDate) patient.birthDate = birthDate;
      const original = JSON.stringify(patient);
      const projected = component.processQueryResourcesSelectClause(
        component.widgetConfig.queries[1].q,
        {success: true, data: [{resource_raw: patient}]} as any
      );
      return {patient, original, projected};
    }

    function renderAge(projected: any[]) {
      component.vitalsProcessQueryResults([[], projected]);
      component.loading = false;
      component.isEmpty = false;
      fixture.detectChanges();
      return fixture.nativeElement.querySelector('.card-text').textContent;
    }

    for (const entry of [
      {label: 'absent', value: undefined, include: false},
      {label: 'null', value: null, include: true},
      {label: 'empty', value: '', include: true},
      {label: 'unparseable', value: 'not-a-date', include: true}
    ]) {
      it(`shows missing age for a ${entry.label} birth date without changing the Patient`, () => {
        const sample = projectPatient(entry.value, entry.include);
        if (!entry.include) expect(sample.projected[0].birthDate).toEqual([]);
        const text = renderAge(sample.projected);
        expect(text).toContain('Age not provided');
        expect(text).not.toContain('0 years');
        expect(text).not.toContain('NaN');
        expect(JSON.stringify(sample.patient)).toBe(sample.original);
      });
    }

    for (const entry of [
      {birthDate: '2026-06-14', display: '0 years'},
      {birthDate: '2000-06-15', display: '26 years'}
    ]) {
      it(`keeps valid ${entry.display} visible`, () => {
        const sample = projectPatient(entry.birthDate);
        expect(renderAge(sample.projected)).toContain(entry.display);
        expect(component.age).toBe(entry.display);
        expect(JSON.stringify(sample.patient)).toBe(sample.original);
      });
    }

    it('clears the previous age when refreshed Patient data has no birth date', () => {
      expect(renderAge(projectPatient('2000-06-15').projected)).toContain('26 years');
      const sample = projectPatient(undefined, false);
      const text = renderAge(sample.projected);
      expect(text).toContain('Age not provided');
      expect(text).not.toContain('26 years');
      expect(JSON.stringify(sample.patient)).toBe(sample.original);
    });

    it('clears the previous age when a refreshed result contains no Patient', () => {
      expect(renderAge(projectPatient('2000-06-15').projected)).toContain('26 years');
      const text = renderAge([]);
      expect(text).toContain('Age not provided');
      expect(text).not.toContain('26 years');
    });
  });

});
