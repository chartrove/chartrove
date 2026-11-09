import {ComponentFixture, TestBed} from '@angular/core/testing';
import {CommonModule} from '@angular/common';
import {RouterTestingModule} from '@angular/router/testing';
import {of, Subject, throwError} from 'rxjs';
import {DirectSmartSourcesComponent} from './direct-smart-sources.component';
import {DirectSmartService} from '../../services/direct-smart.service';
import {FastenApiService} from '../../services/fasten-api.service';
import {SmartSnapshotImportResult} from '../../models/direct-smart';

describe('DirectSmartSourcesComponent', () => {
  let fixture: ComponentFixture<DirectSmartSourcesComponent>;
  let component: DirectSmartSourcesComponent;
  let direct: jasmine.SpyObj<DirectSmartService>;
  let api: jasmine.SpyObj<FastenApiService>;
  const provider = {id: 'epic-one', display: 'Example hospital', fhir_base_url: 'https://example.invalid/fhir'};
  const owner = {id: '00000000-0000-4000-8000-000000000001', username: 'synthetic-owner'};
  const connection = {id: 'handle-one', provider_id: 'epic-one', patient_id: 'patient-one', expires_at: 9999999999, owner};
  const patient = {resourceType: 'Patient' as const, id: 'patient-one', name: [{text: '<img src=x onerror=alert(1)>'}], birthDate: '2000-01-02'};
  const patientJSON = '{"resourceType":"Patient","id":"patient-one","name":[{"text":"<img src=x onerror=alert(1)>"}],"extension":[{"url":"https://example.invalid/unknown","valueDecimal":1.234567890123456789}]}';
  const clinicalJSON = `{"resourceType":"Bundle","type":"collection","entry":[{"resource":${patientJSON}},{"resource":{"resourceType":"Observation","id":"observation-one","status":"final","code":{"text":"synthetic-value"},"subject":{"reference":"Patient/patient-one"},"valueQuantity":{"value":2.345678901234567891,"unit":"mg"}}}]}`;
  const saved: SmartSnapshotImportResult = {source_id: '00000000-0000-4000-8000-000000000002', snapshot_id: '00000000-0000-4000-8000-000000000003', sha256: 'a'.repeat(64), patient_id: 'patient-one', created: 1, updated: 0, unchanged: 0, already_imported: false};

  beforeEach(async () => {
    direct = jasmine.createSpyObj<DirectSmartService>('DirectSmartService', ['providers', 'connect', 'disconnect', 'readClinicalSnapshot', 'isCurrentIdentity']);
    direct.providers.and.resolveTo([provider]);
    direct.connect.and.resolveTo({connection, patient, patientJSON});
    direct.disconnect.and.resolveTo();
    direct.readClinicalSnapshot.and.resolveTo(clinicalJSON);
    direct.isCurrentIdentity.and.returnValue(true);
    api = jasmine.createSpyObj<FastenApiService>('FastenApiService', ['importClinicalSnapshot']);
    api.importClinicalSnapshot.and.returnValue(of(saved));
    await TestBed.configureTestingModule({declarations: [DirectSmartSourcesComponent], imports: [CommonModule, RouterTestingModule],
      providers: [{provide: DirectSmartService, useValue: direct}, {provide: FastenApiService, useValue: api}]}).compileComponents();
    fixture = TestBed.createComponent(DirectSmartSourcesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();
  });

  it('previews escaped Patient and receiving identity, preserving raw numeric fields until explicit confirmation', async () => {
    fixture.nativeElement.querySelector('[data-testid="direct-smart-connect"]').click();
    await fixture.whenStable(); fixture.detectChanges();
    const preview = fixture.nativeElement.querySelector('[data-testid="direct-patient-preview"]');
    expect(preview.textContent).toContain('patient-one');
    expect(preview.textContent).toContain(owner.username);
    expect(preview.textContent).toContain(owner.id);
    expect(preview.textContent).toContain('<img src=x onerror=alert(1)>');
    expect(preview.querySelector('img')).toBeNull();
    expect(api.importClinicalSnapshot).not.toHaveBeenCalled();
    expect(fixture.nativeElement.querySelector('[data-testid="direct-smart-import"]').disabled).toBeTrue();
    await component.importPatient();
    expect(api.importClinicalSnapshot).not.toHaveBeenCalled();
    fixture.nativeElement.querySelector('#direct-smart-authority').click();
    fixture.detectChanges();
    fixture.nativeElement.querySelector('[data-testid="direct-smart-import"]').click();
    await fixture.whenStable(); fixture.detectChanges();
    const args = api.importClinicalSnapshot.calls.mostRecent().args;
    expect(args[0]).toContain(patientJSON);
    expect(args[0]).toContain('1.234567890123456789');
    expect(args.slice(1)).toEqual([{fhir_base_url: provider.fhir_base_url, display: provider.display}, owner.id, true]);
    expect(direct.disconnect).toHaveBeenCalledWith('handle-one');
    expect(fixture.nativeElement.textContent).toContain('View local records');
    expect(fixture.nativeElement.querySelector('[data-testid="direct-patient-preview"]')).toBeNull();
  });

  it('retrieves a complete clinical preview, resets confirmation and sends its exact raw text', async () => {
    await component.connect(); component.confirmed = true;
    await component.retrieveClinical(); fixture.detectChanges();
    expect(component.confirmed).toBeFalse();
    expect(component.observationCount).toBe(1);
    expect(fixture.nativeElement.querySelector('[data-testid="direct-clinical-preview"]').textContent).toContain('1 Observations');
    expect(api.importClinicalSnapshot).not.toHaveBeenCalled();
    component.confirmed = true;
    await component.importPatient();
    expect(api.importClinicalSnapshot).toHaveBeenCalledWith(clinicalJSON, {fhir_base_url: provider.fhir_base_url, display: provider.display}, owner.id, true);
    expect(api.importClinicalSnapshot.calls.mostRecent().args[0]).toContain('2.345678901234567891');
  });

  it('retains the exact preview on transport uncertainty and acknowledges historical retry without claiming current data', async () => {
    await component.connect(); component.confirmed = true;
    api.importClinicalSnapshot.and.returnValues(throwError(() => new Error('private body')), of({...saved, already_imported: true, created: 0}));
    await component.importPatient(); fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain('completion could not be confirmed');
    expect(fixture.nativeElement.textContent).not.toContain('private body');
    expect(component.patient).toBeDefined();
    const original = api.importClinicalSnapshot.calls.mostRecent().args[0];
    await component.importPatient(); fixture.detectChanges();
    expect(direct.connect).toHaveBeenCalledTimes(1);
    expect(api.importClinicalSnapshot.calls.mostRecent().args[0]).toBe(original);
    expect(fixture.nativeElement.textContent).toContain('Previously imported; no changes applied');
  });

  it('refuses a stale receiving profile before writing and removes its private preview', async () => {
    await component.connect(); component.confirmed = true;
    direct.isCurrentIdentity.and.returnValue(false);
    await component.importPatient();
    expect(api.importClinicalSnapshot).not.toHaveBeenCalled();
    expect(component.patient).toBeUndefined();
    expect(component.error).toContain('local identity changed');
    expect(direct.disconnect).toHaveBeenCalledWith(connection.id);
  });

  it('does not apply a previous-profile import response after identity changes', async () => {
    await component.connect(); component.confirmed = true;
    const responses = new Subject<SmartSnapshotImportResult>();
    api.importClinicalSnapshot.and.returnValue(responses);
    const imported = spyOn(component.imported, 'emit');
    const pending = component.importPatient();
    direct.isCurrentIdentity.and.returnValue(false);
    responses.next(saved); responses.complete();
    await pending;
    expect(component.patient).toBeUndefined();
    expect(component.importedSource).toBeUndefined();
    expect(imported).not.toHaveBeenCalled();
    expect(component.status).toContain('local identity changed');
  });

  it('cancels and releases a late connection when the view is destroyed', async () => {
    let finish: (result: {connection: typeof connection, patient: typeof patient, patientJSON: string}) => void;
    direct.connect.and.callFake(() => new Promise(resolve => finish = resolve));
    const pending = component.connect();
    const signal = direct.connect.calls.mostRecent().args[1];
    fixture.destroy();
    expect(signal.aborted).toBeTrue();
    finish({connection, patient, patientJSON});
    await pending;
    expect(direct.disconnect).toHaveBeenCalledWith('handle-one');
    expect(api.importClinicalSnapshot).not.toHaveBeenCalled();
    expect(component.patient).toBeUndefined();
  });

  it('keeps a prior complete preview, not a partial replacement, after failed clinical refresh', async () => {
    await component.connect(); await component.retrieveClinical();
    direct.readClinicalSnapshot.and.rejectWith(new Error('private provider reply'));
    await component.retrieveClinical();
    expect(component.observationCount).toBe(1);
    expect(component.error).toContain('previous complete preview remains');
    expect(component.error).not.toContain('private provider reply');
    component.confirmed = true;
    await component.importPatient();
    expect(api.importClinicalSnapshot.calls.mostRecent().args[0]).toBe(clinicalJSON);
  });

  it('does not import or confirm while clinical retrieval is in flight and discards a cancelled late result', async () => {
    await component.connect(); component.confirmed = true;
    let finish: (raw: string) => void;
    direct.readClinicalSnapshot.and.callFake(() => new Promise(resolve => finish = resolve));
    const pending = component.retrieveClinical();
    expect(component.confirmed).toBeFalse();
    const signal = direct.readClinicalSnapshot.calls.mostRecent().args[1];
    component.confirmed = true;
    await component.importPatient();
    expect(api.importClinicalSnapshot).not.toHaveBeenCalled();
    component.cancel();
    expect(signal.aborted).toBeTrue();
    finish(clinicalJSON);
    await pending;
    expect(component.patient).toBeUndefined();
    expect(component.retrieving).toBeFalse();
    expect(component.controller).toBeUndefined();
  });

  it('shows an actionable empty state without opening authorization', async () => {
    direct.providers.and.resolveTo([]);
    await component.loadProviders(); fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain('No direct providers are configured');
    expect(fixture.nativeElement.querySelector('[data-testid="direct-smart-connect"]')).toBeNull();
    expect(direct.connect).not.toHaveBeenCalled();
  });

  it('discards a preview without importing it', async () => {
    await component.connect();
    await component.disconnect(); fixture.detectChanges();
    expect(component.patient).toBeUndefined();
    expect(direct.disconnect).toHaveBeenCalledWith('handle-one');
    expect(api.importClinicalSnapshot).not.toHaveBeenCalled();
  });

  it('waits for discard cleanup before allowing another connection', async () => {
    await component.connect();
    let release: () => void;
    direct.disconnect.and.callFake(() => new Promise<void>(resolve => release = resolve));
    const discard = component.disconnect();
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('[data-testid="direct-smart-connect"]').disabled).toBeTrue();
    expect(fixture.nativeElement.querySelector('#direct-smart-provider').disabled).toBeTrue();
    await component.connect();
    expect(direct.connect).toHaveBeenCalledTimes(1);
    release(); await discard; fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('[data-testid="direct-smart-connect"]').disabled).toBeFalse();
    await component.connect();
    expect(direct.connect).toHaveBeenCalledTimes(2);
  });

  it('clears discard progress and releases the UI guard when cleanup fails', async () => {
    await component.connect();
    direct.disconnect.and.rejectWith(new Error('private provider payload'));
    await component.disconnect(); fixture.detectChanges();
    expect(component.status).toBe('');
    expect(component.error).toContain('Could not release the direct session');
    expect(fixture.nativeElement.textContent).not.toContain('private provider payload');
    expect(fixture.nativeElement.querySelector('[data-testid="direct-smart-connect"]').disabled).toBeFalse();
  });

  it('does not send another import while a confirmed write is in flight', async () => {
    await component.connect(); component.confirmed = true;
    const response = new Subject<SmartSnapshotImportResult>();
    api.importClinicalSnapshot.and.returnValue(response);
    const first = component.importPatient();
    const second = component.importPatient();
    expect(api.importClinicalSnapshot).toHaveBeenCalledTimes(1);
    response.next(saved); response.complete();
    await Promise.all([first, second]);
    expect(component.importing).toBeFalse();
  });
});
