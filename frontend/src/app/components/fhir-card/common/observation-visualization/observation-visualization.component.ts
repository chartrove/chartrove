import { Component, Input, OnInit, ChangeDetectionStrategy } from '@angular/core';
import { ObservationModel } from '../../../../../lib/models/resources/observation-model';
import { QuantityModel } from '../../../../../lib/models/datatypes/quantity-model';

import { ObservationBarChartComponent } from '../observation-bar-chart/observation-bar-chart.component';
import { ObservationTableComponent } from '../observation-table/observation-table.component';

@Component({
    selector: 'observation-visualization',
    imports: [ObservationBarChartComponent, ObservationTableComponent],
    templateUrl: './observation-visualization.component.html',
    changeDetection: ChangeDetectionStrategy.Eager,
    styleUrls: ['./observation-visualization.component.scss']
})
export class ObservationVisualizationComponent implements OnInit {
  @Input() observations: ObservationModel[]
  @Input() preferredVisualizationType?: string = 'bar'

  visualizationType: string = ''

  constructor() { }

  ngOnInit(): void {
    if(!this.observations || !this.observations.length) {
      return;
    }

    this.visualizationType = this.pickVisualizationType(this.preferredVisualizationType, this.observations)
  }

  // A categorical or missing reading must never become a numeric endpoint just
  // because an earlier reading permits a chart. The table preserves all rows.
  private pickVisualizationType(preferredType: string, observations: ObservationModel[]): string {
    const supported = observations[0]?.value_model?.visualizationTypes() ?? [];
    // A common numeric axis must not imply conversion between reported units.
    // Compare supplied metadata exactly; the table retains each literal value.
    const units = observations.map(({value_model}) => value_model instanceof QuantityModel
      ? JSON.stringify([value_model.system ?? null, value_model.code ?? null, value_model.unit ?? null]) : null);
    const sameUnits = units.every(unit => unit === units[0]);
    const shared = supported.filter(type => (type !== 'bar' || sameUnits)
      && observations.every(observation => observation.value_model?.visualizationTypes().includes(type)));
    return shared.includes(preferredType) ? preferredType : shared[0] ?? 'table';
  }
}
