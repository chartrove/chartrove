import { Component, Input, OnInit, ChangeDetectionStrategy } from '@angular/core';
import {formatObservationDate} from '../../../../../lib/utils/observation_date';
import { ObservationModel } from '../../../../../lib/models/resources/observation-model';
import { ChartConfiguration } from 'chart.js';
import { ChartroveChartsModule } from '../../../chartrove-charts.module';

const defaultChartHeight = 65;
const defaultChartEntryHeight = 30;

@Component({
    selector: 'observation-bar-chart',
    imports: [ChartroveChartsModule],
    templateUrl: './observation-bar-chart.component.html',
    changeDetection: ChangeDetectionStrategy.Eager,
    styleUrls: ['./observation-bar-chart.component.scss']
})
export class ObservationBarChartComponent implements OnInit {
  @Input() observations: ObservationModel[]

  chartHeight = defaultChartEntryHeight;

  // based on https://stackoverflow.com/questions/38889716/chartjs-2-stacked-bar-with-marker-on-top
  // https://stackoverflow.com/questions/62711919/chart-js-horizontal-lines-per-bar
  barChartData =[
    {
      label: "Reference",
      data: [],
      dataLabels: [],
      backgroundColor: "rgba(32, 75, 120,0.6)",
      hoverBackgroundColor: "rgba(32, 75, 120,0.2)",
      parsing: {
        xAxisKey: 'range'
      },
      tooltip: {
        callbacks: {
          label: this.formatTooltip
        }
      }
    },
    {
      label: "Result",
      data: [],
      // @ts-ignore
      dataLabels: [],
      borderColor: "rgba(0,0,0,1)",
      backgroundColor: "rgba(0,0,0,1)",
      hoverBackgroundColor: "rgba(0,0,0,1)",
      minBarLength: 3,
      barPercentage: 1,
      parsing: {
        xAxisKey: 'value'
      },
      // @ts-ignore
      tooltip: {
        callbacks: {
          label: this.formatTooltip
        }
      }
    }
  ]  as ChartConfiguration<'bar'>['data']['datasets']

  barChartLabels = [] // ["2020", "2018"] //["1","2","3","4","5","6","7","8"]

  barChartOptions = {
    indexAxis: 'y',
    maintainAspectRatio: false,
    legend:{
      display: false,
    },
    autoPadding: true,
    //add padding to fix tooltip cutoff
    layout: {
      padding: {
        left: 0,
        right: 4,
        top: 0,
        bottom: 10
      }
    },
    scales: {
      y: {
        stacked: true,
        ticks: {
          beginAtZero: true,
          fontSize: 10,
          min: 0,
        },
      },
      x: {
        scaleLabel:{
          display: false,
          padding: 4,
        },
        ticks: {
          beginAtZero: true,
          fontSize: 10,
          min: 0,
        },
      },
    }
  } as ChartConfiguration<'bar'>['options']

  barChartColors = [
    {
      backgroundColor: 'white'
    }
  ];

  constructor() { }

  ngOnInit(): void {
    if(!this.observations || !this.observations[0]) {
      return;
    }

    let currentValues = []
    let referenceRanges = []

    for(let observation of this.observations) {
      referenceRanges.push(this.extractReferenceRange(observation));
      this.barChartData[0]['dataLabels'].push(observation.reference_range.display());

      currentValues.push(this.extractCurrentValue(observation));
      this.barChartData[1]['dataLabels'].push(observation.value_model.display());

      this.barChartLabels.push(this.formatDate(observation.effective_date))
    }

    let xAxisMax = Math.max(Math.max(...currentValues.flat()), Math.max(...referenceRanges.flat())) * 1.3;
    this.barChartOptions.scales['x']['max'] = xAxisMax

    // @ts-ignore
    this.barChartData[0].data = this.updateNullMax(referenceRanges, xAxisMax);
    // @ts-ignore
    this.barChartData[1].data = this.updateNullMax(currentValues, xAxisMax);

    this.chartHeight = defaultChartHeight + (defaultChartEntryHeight * currentValues.length)
  }

  private extractReferenceRange(observation: ObservationModel): [number, number] {
    let refRange = observation.reference_range;

    // Keep an unknown upper limit distinct from an explicit zero. With no
    // reference at all, retain the existing empty [0, 0] chart placeholder.
    return [refRange.low_value ?? 0,
      refRange.high_value ?? (refRange.low_value != null ? null : 0)]
  }

  private extractCurrentValue(observation: ObservationModel): [any, any] {
    let valueObject = observation.value_model.valueObject();

    if (valueObject.range) {
      return [valueObject.range.low, valueObject.range.high];
    } else {
      return [valueObject.value, valueObject.value]
    }
  }

  // Helper method to update any [number, null] set to [number, max].
  // Necessary in order to display greater-than ranges that have no upper bound.
  private updateNullMax(array: any[][], max: number): any[][] {
    return array.map(values => {
      if (values[0] != null && values[1] == null) {
        return [values[0], max]
      } else {
        return [values[0], values[1]]
      }
    });
  }

  private formatDate(date: string | number | Date): string {
    return formatObservationDate(date);
  }

  private formatTooltip(context) {
    return `${context.dataset.label}: ${context.dataset.dataLabels[context.dataIndex]}`;
  }
}
