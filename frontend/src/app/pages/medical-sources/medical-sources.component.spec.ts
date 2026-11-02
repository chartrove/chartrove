import { ComponentFixture, TestBed } from '@angular/core/testing';

import { MedicalSourcesComponent } from './medical-sources.component';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import {RouterTestingModule} from '@angular/router/testing';
import {HTTP_CLIENT_TOKEN} from '../../dependency-injection';
import { HttpClient, provideHttpClient, withInterceptorsFromDi, withXhr } from '@angular/common/http';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import { LoadingSpinnerComponent } from 'src/app/components/loading-spinner/loading-spinner.component';
import { MedicalSourcesFilterComponent } from 'src/app/components/medical-sources-filter/medical-sources-filter.component';
import { MedicalSourcesConnectedComponent } from 'src/app/components/medical-sources-connected/medical-sources-connected.component';
import { NgxDropzoneModule } from 'ngx-dropzone';
import {EpicSandboxService} from '../../services/epic-sandbox.service';
import {FastenApiService} from '../../services/fasten-api.service';
import {of, throwError} from 'rxjs';
import {Source} from '../../models/fasten/source';
import {DirectSmartSourcesComponent} from '../../components/direct-smart-sources/direct-smart-sources.component';
import {DirectSmartService} from '../../services/direct-smart.service';

describe('MedicalSourcesComponent', () => {
  let component: MedicalSourcesComponent;
  let fixture: ComponentFixture<MedicalSourcesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
    declarations: [MedicalSourcesComponent, MedicalSourcesFilterComponent, MedicalSourcesConnectedComponent, DirectSmartSourcesComponent],
    imports: [RouterTestingModule, FormsModule, ReactiveFormsModule, LoadingSpinnerComponent, NgxDropzoneModule],
    providers: [
        { provide: DirectSmartService, useValue: { providers: async () => [] } },
        {
            provide: HTTP_CLIENT_TOKEN,
            useClass: HttpClient,
        },
        provideHttpClient(withXhr(), withInterceptorsFromDi()),
        provideHttpClientTesting(),
    ]
})
    .compileComponents();

    fixture = TestBed.createComponent(MedicalSourcesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('offers the Epic sandbox connection separately from production providers', () => {
    component.environment_name = 'desktop_prod';
    fixture.detectChanges();
    const button = fixture.nativeElement.querySelector('[data-testid="connect-epic-sandbox"]');
    expect(button).not.toBeNull();
    expect(fixture.nativeElement.textContent).toContain('Synthetic test data only');
  });

  it('shows progress and enables import after retrieving a synthetic Patient', async () => {
    component.environment_name = 'desktop_prod';
    let complete: (patient: any) => void;
    spyOn(TestBed.inject(EpicSandboxService), 'connect').and.callFake((_signal, progress) => {
      progress('Waiting for sandbox sign-in');
      return new Promise(resolve => complete = resolve);
    });
    const connect = spyOn(component, 'connectEpicSandbox').and.callThrough();
    fixture.detectChanges();
    fixture.nativeElement.querySelector('[data-testid="connect-epic-sandbox"]').click();
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('[data-testid="connect-epic-sandbox"]').disabled).toBeTrue();
    expect(fixture.nativeElement.textContent).toContain('Waiting for sandbox sign-in');
    complete({resourceType: 'Patient', id: 'synthetic-patient'});
    await connect.calls.mostRecent().returnValue;
    await fixture.whenStable();
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain('Import synthetic Patient');
  });

  it('allows retrying a failed local import without another provider sign-in', async () => {
    component.environment_name = 'desktop_prod';
    component.sandboxPatient = {resourceType: 'Patient', id: 'synthetic-patient'};
    spyOn(TestBed.inject(FastenApiService), 'importSandboxPatient').and.returnValues(
      throwError(new Error('synthetic failure')), of({source: new Source({id: 'imported-source'})}));
    await component.importEpicSandboxPatient();
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain('retry without signing in again');
    expect(component.sandboxPatient).toBeDefined();
    await component.importEpicSandboxPatient();
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain('View imported Patient');
    expect(component.sandboxPatient).toBeUndefined();
  });

  it('cancels a pending connection when the page is destroyed', async () => {
    let signal: AbortSignal;
    spyOn(TestBed.inject(EpicSandboxService), 'connect').and.callFake(abort => {
      signal = abort;
      return new Promise((_resolve, reject) => abort.addEventListener('abort', () => reject(new Error('cancelled'))));
    });
    const pending = component.connectEpicSandbox();
    fixture.destroy();
    await pending;
    expect(signal.aborted).toBeTrue();
  });

  const syntheticBundle = () => new File(
    ['{"resourceType":"Bundle","type":"collection","entry":[]}'],
    'synthetic.json', { type: 'application/json' });
  const isManualImport = request => request.url.endsWith('/secure/source/manual');
  const isConversion = request => request.url.startsWith('https://api.platform.fastenhealth.com/v1/app/convert/ccda/to/fhir');
  const uploadStatus = () => fixture.nativeElement.querySelector('ngx-dropzone').parentElement.querySelector('[role="status"]');

  it('shows pending upload status until the synthetic import response completes', async () => {
    const file = syntheticBundle();
    const pending = component.uploadSourceBundleHandler({ addedFiles: [file] });
    fixture.detectChanges();
    expect(uploadStatus()).not.toBeNull();
    expect(fixture.nativeElement.querySelector('ngx-dropzone').getAttribute('aria-busy')).toBe('true');
    expect(fixture.nativeElement.querySelector('ngx-dropzone input[type="file"]').disabled).toBeTrue();
    const request = TestBed.inject(HttpTestingController).expectOne(isManualImport);
    expect(request.request.body.get('file')).toBe(file);
    request.flush({ success: true, data: { id: 'synthetic-source' } });
    await pending;
    fixture.detectChanges();
    expect(uploadStatus()).toBeNull();
    expect(component.uploadedFile).toEqual([]);
    expect(component.uploadErrorMsg).toBe('');
    expect(fixture.nativeElement.querySelector('ngx-dropzone input[type="file"]').disabled).toBeFalse();
  });

  it('clears failed-upload progress and allows a retry', async () => {
    const http = TestBed.inject(HttpTestingController);
    const pending = component.uploadSourceBundleHandler({ addedFiles: [syntheticBundle()] });
    http.expectOne(isManualImport).flush({ error: 'Synthetic rejection' }, { status: 500, statusText: 'Test failure' });
    await pending;
    fixture.detectChanges();
    expect(uploadStatus()).toBeNull();
    expect(component.uploadedFile).toEqual([]);
    expect(fixture.nativeElement.textContent).toContain('Synthetic rejection');
    const retry = component.uploadSourceBundleHandler({ addedFiles: [syntheticBundle()] });
    fixture.detectChanges();
    expect(uploadStatus()).not.toBeNull();
    http.expectOne(isManualImport).flush({ success: true, data: { id: 'synthetic-retry' } });
    await retry;
    fixture.detectChanges();
    expect(uploadStatus()).toBeNull();
    expect(component.uploadErrorMsg).toBe('');
  });

  it('submits only one import while an upload is pending', async () => {
    const file = syntheticBundle();
    const first = component.uploadSourceBundleHandler({ addedFiles: [file] });
    const duplicate = component.uploadSourceBundleHandler({ addedFiles: [file] });
    const requests = TestBed.inject(HttpTestingController).match(isManualImport);
    expect(requests.length).toBe(1);
    for (const request of requests) request.flush({ success: true, data: { id: 'synthetic-source' } });
    await Promise.all([first, duplicate]);
  });

  it('ignores an empty accepted-file selection', async () => {
    await expectAsync(component.uploadSourceBundleHandler({ addedFiles: [], rejectedFiles: [] })).toBeResolved();
    expect(TestBed.inject(HttpTestingController).match(isManualImport).length).toBe(0);
    expect(component.uploadedFile).toEqual([]);
  });

  it('clears pending status when the user declines external CCDA conversion', async () => {
    let decide: (value: boolean) => void;
    spyOn(component, 'showCcdaWarningModal').and.returnValue(new Promise(resolve => decide = resolve));
    const pending = component.uploadSourceBundleHandler({ addedFiles: [new File(['<synthetic/>'], 'synthetic.xml', { type: 'text/xml' })] });
    fixture.detectChanges();
    expect(uploadStatus()).not.toBeNull();
    decide(false);
    await pending;
    fixture.detectChanges();
    expect(uploadStatus()).toBeNull();
    expect(component.uploadedFile).toEqual([]);
    const http = TestBed.inject(HttpTestingController);
    expect(http.match(isConversion).length).toBe(0);
    expect(http.match(isManualImport).length).toBe(0);
  });

  it('clears conversion failures without sending a local import', async () => {
    spyOn(component, 'showCcdaWarningModal').and.returnValue(Promise.resolve(true));
    const pending = component.uploadSourceBundleHandler({ addedFiles: [new File(['<synthetic/>'], 'synthetic.xml', { type: 'text/xml' })] });
    await Promise.resolve();
    fixture.detectChanges();
    expect(uploadStatus()).not.toBeNull();
    const http = TestBed.inject(HttpTestingController);
    http.expectOne(isConversion).flush({ error: 'Synthetic conversion rejection' }, { status: 500, statusText: 'Test failure' });
    await pending;
    fixture.detectChanges();
    expect(uploadStatus()).toBeNull();
    expect(component.uploadedFile).toEqual([]);
    expect(component.uploadErrorMsg).toContain('Error converting file:');
    expect(http.match(isManualImport).length).toBe(0);
  });

  it('keeps progress visible through synthetic conversion and the local import', async () => {
    spyOn(component, 'showCcdaWarningModal').and.returnValue(Promise.resolve(true));
    const pending = component.uploadSourceBundleHandler({ addedFiles: [new File(['<synthetic/>'], 'synthetic.xml', { type: 'text/xml' })] });
    await Promise.resolve();
    const http = TestBed.inject(HttpTestingController);
    http.expectOne(isConversion).flush({ resourceType: 'Bundle', type: 'collection', entry: [] });
    await Promise.resolve();
    fixture.detectChanges();
    expect(uploadStatus()).not.toBeNull();
    const request = http.expectOne(isManualImport);
    expect(request.request.body.get('file').name).toBe('synthetic.xml.converted.json');
    request.flush({ success: true, data: { id: 'synthetic-converted-source' } });
    await pending;
    fixture.detectChanges();
    expect(uploadStatus()).toBeNull();
    expect(component.uploadedFile).toEqual([]);
  });
});
