import {ChangeDetectionStrategy, Component, OnChanges, OnDestroy, OnInit, SimpleChanges} from '@angular/core';
import {ChartOptions} from 'chart.js';
import {resolveObjectKey} from 'chart.js/helpers';
import {forkJoin, Subscription} from 'rxjs';
import {map} from 'rxjs/operators';
import {ChartroveChartsModule} from '../../components/chartrove-charts.module';
import {FastenApiService} from '../../services/fasten-api.service';
import {DashboardWidgetComponent} from '../dashboard-widget/dashboard-widget.component';
import {LoadingWidgetComponent} from '../loading-widget/loading-widget.component';

@Component({
  imports: [ChartroveChartsModule, LoadingWidgetComponent],
  selector: 'complex-line-widget',
  templateUrl: './complex-line-widget.component.html',
  changeDetection: ChangeDetectionStrategy.Eager,
  styleUrls: ['./complex-line-widget.component.scss']
})
export class ComplexLineWidgetComponent extends DashboardWidgetComponent<'line'> implements OnInit, OnChanges, OnDestroy {
  loadError = false;
  private request: Subscription;
  chartOptions: ChartOptions<'line'> = {
    responsive: true,
    maintainAspectRatio: false,
    parsing: {xAxisKey: 'x', yAxisKey: 'y'},
    elements: {line: {tension: 0}},
  };

  constructor(public fastenApi: FastenApiService) { super(fastenApi); }

  ngOnInit(): void {
    this.refresh();
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['widgetConfig'] && !changes['widgetConfig'].firstChange) this.refresh();
  }

  private refresh(): void {
    // A changed query must never reuse an earlier request or chart under its new title.
    this.request?.unsubscribe();
    this.chartDatasets = [];
    this.chartLabels = [];
    this.isEmpty = true;
    this.loadError = false;
    this.loading = true;
    const config = this.widgetConfig;
    if (!config?.queries?.length) {
      this.loading = false;
      return;
    }
    const xKey = config.parsing?.xAxisKey || config.parsing?.label || 'label';
    const yKey = config.parsing?.yAxisKey || 'data';
    this.request = forkJoin(config.queries.map(query => this.fastenApi.queryResources(query.q).pipe(
      map(response => {
        if (!response?.success) throw new Error('Query failed');
        return this.processQueryResourcesSelectClause(query.q, response);
      })
    ))).pipe(map(results => results.map((rows, index) => ({
      ...config.queries[index].dataset_options,
      label: config.queries[index].dataset_options?.label || `Series ${index + 1}`,
      parsing: {xAxisKey: 'x', yAxisKey: 'y'},
      data: rows.map(row => {
        const xValue = this.singleValue(resolveObjectKey(row || {}, xKey));
        const yValue = this.singleValue(resolveObjectKey(row || {}, yKey));
        if (xValue !== null && typeof xValue !== 'string' && (typeof xValue !== 'number' || !Number.isFinite(xValue))) {
          throw new Error('Unsupported chart label');
        }
        if (yValue !== null && (typeof yValue !== 'number' || !Number.isFinite(yValue))) {
          throw new Error('Unsupported chart value');
        }
        // Keep API rows and raw resources unchanged. Category labels are text;
        // missing values stay null gaps and numeric zero remains zero.
        return {...row, x: xValue === null ? null : String(xValue), y: yValue};
      }),
    })))).subscribe({
      next: datasets => {
        this.chartDatasets = datasets;
        this.chartLabels = [...new Set(datasets.flatMap(dataset => dataset.data.map(point => point.x)).filter(label => label !== null))];
        this.isEmpty = !datasets.some(dataset => dataset.data.some(point => point.x !== null && point.y !== null));
        this.loading = false;
      },
      error: () => {
        this.loadError = true;
        this.loading = false;
        this.isEmpty = true;
        this.chartDatasets = [];
      },
    });
  }

  private singleValue(value: unknown): unknown {
    if (Array.isArray(value)) {
      if (value.length > 1) throw new Error('Multiple values for one chart coordinate');
      value = value[0];
    }
    return value === undefined ? null : value;
  }

  ngOnDestroy(): void {
    this.request?.unsubscribe();
    this.chartDatasetsSubject.complete();
  }
}
