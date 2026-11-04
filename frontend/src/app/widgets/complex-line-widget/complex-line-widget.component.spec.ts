import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ComplexLineWidgetComponent } from './complex-line-widget.component';
import {provideHttpClient, HttpClient} from '@angular/common/http';
import {provideHttpClientTesting, HttpTestingController} from '@angular/common/http/testing';
import {provideRouter} from '@angular/router';
import {HTTP_CLIENT_TOKEN} from '../../dependency-injection';
import {AuthService} from '../../services/auth.service';
import {DashboardWidgetConfig} from '../../models/widget/dashboard-widget-config';
import {BaseChartDirective} from 'ng2-charts';
import {By} from '@angular/platform-browser';
import {GridStack} from 'gridstack';
import {GridstackComponent, NgGridStackWidget} from '../../components/gridstack/gridstack.component';

describe('ComplexLineWidgetComponent', () => {
  let component: ComplexLineWidgetComponent;
  let fixture: ComponentFixture<ComplexLineWidgetComponent>;
  let http: HttpTestingController;

  function config(): DashboardWidgetConfig {
    return {
      item_type: 'complex-line-widget', title_text: 'Synthetic quantity', description_text: 'Synthetic query results',
      width: 10, height: 10,
      queries: [{q: {select: ['issued as label', 'valueQuantity.value as data'], from: 'Observation', where: {code: 'synthetic-quantity'}}, dataset_options: {label: 'Synthetic series'}}],
      parsing: {xAxisKey: 'label', yAxisKey: 'data'},
    };
  }

  function response(values: number[]) {
    return {success: true, data: values.map((value, i) => ({
      source_id: 'synthetic-source', source_resource_id: `synthetic-${i}`,
      resource_raw: {resourceType: 'Observation', id: `synthetic-${i}`, issued: `2000-01-0${i + 1}T01:30:00-05:00`, valueQuantity: {value, unit: 'kg'}},
    }))};
  }

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ ComplexLineWidgetComponent, GridstackComponent ],
      providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([]),
        {provide: HTTP_CLIENT_TOKEN, useExisting: HttpClient}, {provide: AuthService, useValue: {}}],
    })
    .compileComponents();

    fixture = TestBed.createComponent(ComplexLineWidgetComponent);
    component = fixture.componentInstance;
    fixture.nativeElement.style.cssText = 'display:block;width:640px;height:360px';
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  async function settle() {
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();
  }

  it('queries configured records and renders their actual zero and nonzero values', async () => {
    const configuration = config();
    const originalConfig = JSON.stringify(configuration);
    const data = response([0, 2.5]);
    const originalRecords = JSON.stringify(data);
    fixture.componentRef.setInput('widgetConfig', configuration);
    fixture.detectChanges();
    const request = http.expectOne(req => req.url.endsWith('/secure/query'));
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual(configuration.queries[0].q);
    request.flush(data);
    await settle();
    expect(fixture.nativeElement.textContent).toContain('Synthetic quantity');
    expect(fixture.nativeElement.textContent).not.toContain('Diabetes Tracking');
    expect(fixture.nativeElement.textContent).not.toContain('140 mg/dl');
    const directive = fixture.debugElement.query(By.directive(BaseChartDirective)).injector.get(BaseChartDirective);
    directive.chart.update('none');
    const points = directive.chart.getDatasetMeta(0).data;
    const scale = directive.chart.scales['y'];
    expect(scale.getValueForPixel(points[0].getProps(['y'], true)['y'])).toBeCloseTo(0, 6);
    expect(scale.getValueForPixel(points[1].getProps(['y'], true)['y'])).toBeCloseTo(2.5, 6);
    expect(directive.chart.data.labels).toEqual(data.data.map(resource => resource.resource_raw.issued));
    expect(JSON.stringify(configuration)).toBe(originalConfig);
    expect(JSON.stringify(data)).toBe(originalRecords);
    const destroy = spyOn(directive.chart, 'destroy').and.callThrough();
    fixture.destroy();
    expect(destroy).toHaveBeenCalledTimes(1);
  });

  it('shows an empty result without substituting demonstration values', async () => {
    fixture.componentRef.setInput('widgetConfig', config());
    fixture.detectChanges();
    http.expectOne(req => req.url.endsWith('/secure/query')).flush({success: true, data: []});
    await settle();
    expect(fixture.nativeElement.textContent).toContain('No matching records');
    expect(fixture.nativeElement.querySelector('canvas')).toBeNull();
    expect(fixture.nativeElement.textContent).not.toContain('140');
  });

  it('reports a query failure without exposing its details or rendering fake data', async () => {
    fixture.componentRef.setInput('widgetConfig', config());
    fixture.detectChanges();
    http.expectOne(req => req.url.endsWith('/secure/query')).flush('SYNTHETIC PRIVATE DETAIL', {status: 500, statusText: 'Synthetic failure'});
    await settle();
    expect(fixture.nativeElement.querySelector('[role="alert"]').textContent).toContain('Unable to load chart');
    expect(fixture.nativeElement.textContent).not.toContain('SYNTHETIC PRIVATE DETAIL');
    expect(fixture.nativeElement.querySelector('canvas')).toBeNull();
  });

  it('requires configuration instead of showing invented clinical values', async () => {
    await settle();
    http.expectNone(req => req.url.endsWith('/secure/query'));
    expect(fixture.nativeElement.textContent).toContain('No queries configured');
    expect(fixture.nativeElement.querySelector('canvas')).toBeNull();
  });

  it('cancels its pending query when the widget is removed', () => {
    fixture.componentRef.setInput('widgetConfig', config());
    fixture.detectChanges();
    const request = http.expectOne(req => req.url.endsWith('/secure/query'));
    fixture.destroy();
    expect(request.cancelled).toBeTrue();
  });

  it('cancels a replaced query and clears the previous chart when configuration changes', async () => {
    fixture.componentRef.setInput('widgetConfig', config());
    fixture.detectChanges();
    const previous = http.expectOne(req => req.url.endsWith('/secure/query'));
    const replacement = config();
    replacement.title_text = 'Replacement synthetic query';
    replacement.queries[0].q.where = {code: 'replacement-synthetic'};
    fixture.componentRef.setInput('widgetConfig', replacement);
    fixture.detectChanges();
    expect(previous.cancelled).toBeTrue();
    expect(fixture.nativeElement.querySelector('canvas')).toBeNull();
    const next = http.expectOne(req => req.url.endsWith('/secure/query'));
    expect(next.request.body).toEqual(replacement.queries[0].q);
    next.flush(response([2.5]));
    await settle();
    expect(fixture.nativeElement.textContent).toContain(replacement.title_text);
    expect(fixture.nativeElement.querySelector('canvas')).not.toBeNull();
    const removed = {...replacement, queries: []};
    fixture.componentRef.setInput('widgetConfig', removed);
    await settle();
    expect(fixture.nativeElement.querySelector('canvas')).toBeNull();
    expect(fixture.nativeElement.textContent).toContain('No queries configured');
    expect(component.chartDatasets).toEqual([]);
    http.expectNone(req => req.url.endsWith('/secure/query'));
  });

  it('keeps missing numeric values as gaps rather than inventing zero', async () => {
    const data = response([0, 2.5, 5]);
    delete data.data[1].resource_raw.valueQuantity.value;
    fixture.componentRef.setInput('widgetConfig', config());
    fixture.detectChanges();
    http.expectOne(req => req.url.endsWith('/secure/query')).flush(data);
    await settle();
    const directive = fixture.debugElement.query(By.directive(BaseChartDirective)).injector.get(BaseChartDirective);
    directive.chart.update('none');
    const points = directive.chart.getDatasetMeta(0).data;
    expect(points[0]['skip']).toBeFalse();
    expect(points[1]['skip']).toBeTrue();
    expect(points[2]['skip']).toBeFalse();
    expect(component.chartDatasets[0].data[1]).toEqual(jasmine.objectContaining({y: null}));
  });

  it('rejects ambiguous multi-valued coordinates instead of selecting an arbitrary reading', async () => {
    const configuration = config();
    configuration.queries[0].q.select = ['issued as label', 'component.valueQuantity.value as data'];
    fixture.componentRef.setInput('widgetConfig', configuration);
    fixture.detectChanges();
    http.expectOne(req => req.url.endsWith('/secure/query')).flush({success: true, data: [{resource_raw: {
      resourceType: 'Observation', issued: '2000-01-01T00:00:00Z',
      component: [{valueQuantity: {value: 1}}, {valueQuantity: {value: 2}}],
    }}]});
    await settle();
    expect(fixture.nativeElement.querySelector('[role="alert"]').textContent).toContain('Unable to load chart');
    expect(fixture.nativeElement.querySelector('canvas')).toBeNull();
  });

  it('keeps independently queried series aligned to their own labels and styles', async () => {
    const configuration = config();
    configuration.queries.push({q: {...configuration.queries[0].q, where: {code: 'synthetic-second'}},
      dataset_options: {label: 'Second synthetic series', borderColor: '#123456'}});
    fixture.componentRef.setInput('widgetConfig', configuration);
    fixture.detectChanges();
    const requests = http.match(req => req.url.endsWith('/secure/query'));
    expect(requests.length).toBe(2);
    requests[0].flush(response([0]));
    const second = response([2.5]);
    second.data[0].resource_raw.issued = '2000-01-02T01:30:00-05:00';
    requests[1].flush(second);
    await settle();
    const directive = fixture.debugElement.query(By.directive(BaseChartDirective)).injector.get(BaseChartDirective);
    directive.chart.update('none');
    expect(directive.chart.data.labels).toEqual(['2000-01-01T01:30:00-05:00', '2000-01-02T01:30:00-05:00']);
    expect(directive.chart.data.datasets[1].label).toBe('Second synthetic series');
    expect(directive.chart.data.datasets[1].borderColor).toBe('#123456');
    expect(directive.chart.getDatasetMeta(0).data[0].getProps(['x'], true)['x']).toBeLessThan(
      directive.chart.getDatasetMeta(1).data[0].getProps(['x'], true)['x']);
  });

  it('uses the actual dashboard GridStack configuration path and preserves its saved query', async () => {
    const previousTypes = {...GridstackComponent.selectorToType};
    const previousAddRemove = GridStack.addRemoveCB;
    const previousSave = GridStack.saveCB;
    const gridFixture = TestBed.createComponent(GridstackComponent);
    try {
      gridFixture.nativeElement.style.width = '640px';
      gridFixture.componentInstance.options = {column: 12, cellHeight: 60};
      gridFixture.detectChanges();
      const configuration = config();
      const original = JSON.stringify(configuration);
      const grid = gridFixture.componentInstance.grid;
      grid.addWidget({type: 'complex-line-widget', w: 12, h: 6, widgetConfig: configuration} as NgGridStackWidget);
      gridFixture.detectChanges();
      const request = http.expectOne(req => req.url.endsWith('/secure/query'));
      expect(request.request.body).toEqual(configuration.queries[0].q);
      request.flush(response([0, 2.5]));
      gridFixture.detectChanges();
      await gridFixture.whenStable();
      gridFixture.detectChanges();
      expect(gridFixture.nativeElement.textContent).toContain(configuration.title_text);
      const chart = gridFixture.debugElement.query(By.directive(BaseChartDirective)).injector.get(BaseChartDirective).chart;
      chart.update('none');
      const point = chart.getDatasetMeta(0).data[1];
      expect(chart.scales['y'].getValueForPixel(point.getProps(['y'], true)['y'])).toBeCloseTo(2.5, 6);
      expect((grid.save(false) as NgGridStackWidget[])[0].widgetConfig).toEqual(configuration);
      expect(JSON.stringify(configuration)).toBe(original);
    } finally {
      gridFixture.destroy();
      GridstackComponent.selectorToType = previousTypes;
      GridStack.addRemoveCB = previousAddRemove;
      GridStack.saveCB = previousSave;
    }
  });

  it('should create', () => {
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });
});
