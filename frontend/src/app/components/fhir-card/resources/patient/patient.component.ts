import { CommonModule, DatePipe } from '@angular/common';
import { ChangeDetectorRef, Component, Input, OnInit, ChangeDetectionStrategy } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { NgbCollapseModule } from '@ng-bootstrap/ng-bootstrap';
import { PatientModel } from '../../../../../lib/models/resources/patient-model';
import { BadgeComponent } from '../../common/badge/badge.component';
import { TableComponent } from '../../common/table/table.component';
import { FhirCardComponentInterface } from '../../fhir-card/fhir-card-component-interface';

@Component({
    imports: [NgbCollapseModule, CommonModule, BadgeComponent, TableComponent, RouterModule],
    providers: [DatePipe],
    selector: 'fhir-patient',
    templateUrl: './patient.component.html',
    changeDetection: ChangeDetectionStrategy.Eager,
    styleUrls: ['./patient.component.scss']
})
export class PatientComponent implements OnInit, FhirCardComponentInterface {
  @Input() displayModel: PatientModel;
  @Input() showDetails: boolean = true;
  @Input() isCollapsed: boolean = false;

  constructor(public changeRef: ChangeDetectorRef, public router: Router, private datePipe: DatePipe) { }

  get patientAgeDisplay(): number | string {
    const age = this.displayModel?.patient_age;
    return typeof age === 'number' && Number.isFinite(age) ? age : 'Not provided';
  }

  get patientBirthDateDisplay(): string {
    try {
      return this.datePipe.transform(this.displayModel?.patient_birthdate, 'mediumDate') ?? '';
    } catch {
      return '';
    }
  }

  ngOnInit(): void {
  }

  markForCheck(){
    this.changeRef.markForCheck()
  }

}
