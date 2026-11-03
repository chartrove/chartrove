import {Component, Input, OnInit, OnDestroy, ChangeDetectionStrategy} from '@angular/core';
import {ResourceFhir} from '../../models/fasten/resource_fhir';
import {EncounterModel} from '../../../lib/models/resources/encounter-model';
import {RecResourceRelatedDisplayModel} from '../../../lib/utils/resource_related_display_model';
import {DiagnosticReportModel} from '../../../lib/models/resources/diagnostic-report-model';
import {FastenDisplayModel} from '../../../lib/models/fasten/fasten-display-model';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';

@Component({
    selector: 'app-report-medical-history-timeline-panel',
    templateUrl: './report-medical-history-timeline-panel.component.html',
    styleUrls: ['./report-medical-history-timeline-panel.component.scss'],
    changeDetection: ChangeDetectionStrategy.Eager,
    standalone: false
})
export class ReportMedicalHistoryTimelinePanelComponent implements OnInit, OnDestroy {
  @Input() resourceFhir: ResourceFhir
  displayModel: EncounterModel
  private destroyed = false;

  constructor(private modalService: NgbModal) { }

  ngOnInit(): void {
    if (!this.resourceFhir) {
      return;
    }

    let parsed = RecResourceRelatedDisplayModel(this.resourceFhir)
    this.displayModel = parsed.displayModel as EncounterModel
  }

  diagnosticReportLink(diagnosticReportRaw: FastenDisplayModel): string {
    let diagnosticReport = diagnosticReportRaw as DiagnosticReportModel
    return diagnosticReport?.is_category_lab_report ?
      '/labs/report/'+ diagnosticReport?.source_id + '/' + diagnosticReport?.source_resource_type + '/' + diagnosticReport?.source_resource_id :
      '/explore/'+ diagnosticReport?.source_id + '/resource/' + diagnosticReport?.source_resource_id + '/'
  }


  async openMedicalRecordWizard(): Promise<void> {
    const encounter = this.displayModel;
    const {MedicalRecordWizardComponent} = await import('../medical-record-wizard/medical-record-wizard.component');
    if (this.destroyed) return;
    const modalRef = this.modalService.open(MedicalRecordWizardComponent, {
      size: 'xl',
    });
    modalRef.componentInstance.existingEncounter = encounter;
  }

  ngOnDestroy(): void {
    this.destroyed = true;
  }
}
