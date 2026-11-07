import {ChangeDetectorRef, Component, Input, OnInit, ChangeDetectionStrategy} from '@angular/core';
import { NgbCollapseModule } from '@ng-bootstrap/ng-bootstrap';

import {BadgeComponent} from '../../common/badge/badge.component';
import {TableComponent} from '../../common/table/table.component';
import {Router, RouterModule} from '@angular/router';
import {TableRowItem, TableRowItemDataType} from '../../common/table/table-row-item';
import {ObservationModel} from '../../../../../lib/models/resources/observation-model';
import { ObservationVisualizationComponent } from '../../common/observation-visualization/observation-visualization.component';

@Component({
    imports: [BadgeComponent, TableComponent, RouterModule, NgbCollapseModule, ObservationVisualizationComponent],
    providers: [],
    selector: 'fhir-observation',
    templateUrl: './observation.component.html',
    changeDetection: ChangeDetectionStrategy.Eager,
    styleUrls: ['./observation.component.scss']
})
export class ObservationComponent implements OnInit {
  @Input() displayModel: ObservationModel
  @Input() showDetails: boolean = true
  @Input() isCollapsed: boolean = false

  tableData: TableRowItem[] = []
  componentTableData: TableRowItem[][] = []
  displayVisualization: boolean = true

  constructor(public changeRef: ChangeDetectorRef, public router: Router) { }

  ngOnInit(): void {
    if(!this.displayModel){
      return
    }

    const visualizationTypes = this.displayModel?.value_model?.visualizationTypes() ?? [];

    // The card already displays tabular values; missing values have no chart.
    this.displayVisualization = visualizationTypes.some(type => type !== 'table');

    this.tableData.push(
      {
        label: 'Observed on',
        data: this.displayModel?.effective_date,
        enabled: !!this.displayModel?.effective_date,
      },
      {
        label: 'Issued on',
        data: this.displayModel?.issued,
        enabled: !!this.displayModel?.issued,
      },
      {
        label: 'Subject',
        data: this.displayModel?.subject,
        data_type: TableRowItemDataType.Reference,
        enabled: !!this.displayModel?.subject,
      },
      {
        label: 'Coding',
        data: this.displayModel?.code,
        data_type: TableRowItemDataType.CodableConcept,
        enabled: !!this.displayModel?.code,
      },
      {
        label: 'Value',
        data: this.displayModel?.value_model?.display(),
        enabled: !!this.displayModel?.value_model,
      },
      {
        label: 'Reference',
        data: this.displayModel?.reference_range.display(),
        enabled: !!this.displayModel?.reference_range.hasValue(),
      }
    )

    const components = this.displayModel.fhirResource?.component;
    this.componentTableData = (Array.isArray(components) ? components : []).map(raw => {
      // Reuse scalar formatting without changing the parent or original resource.
      // Components keep their own code, value and reference range in source order.
      const component = new ObservationModel(raw);
      return [
        {label: 'Component', data: component.code, data_type: TableRowItemDataType.CodableConcept, enabled: true},
        {label: 'Value', data: component.value_model?.display() ?? 'Value unavailable', enabled: true},
        {label: 'Reference', data: component.reference_range.display(), enabled: component.reference_range.hasValue()},
      ];
    });
  }

  markForCheck(){
    this.changeRef.markForCheck()
  }
}
