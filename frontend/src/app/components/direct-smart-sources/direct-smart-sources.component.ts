import {Component, EventEmitter, OnDestroy, OnInit, Output, ChangeDetectionStrategy} from '@angular/core';
import {Patient} from 'fhir/r4';
import {firstValueFrom} from 'rxjs';
import {DirectSmartCleanupError, DirectSmartService} from '../../services/direct-smart.service';
import {FastenApiService} from '../../services/fasten-api.service';
import {SmartConnection, SmartProvider} from '../../models/direct-smart';
import {Source} from '../../models/fasten/source';

function isRecord(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === 'object' && !Array.isArray(value);
}
function isPreviewPatient(value: unknown): value is Patient {
  return isRecord(value) && value['resourceType'] === 'Patient' && typeof value['id'] === 'string';
}

@Component({
    selector: 'app-direct-smart-sources', templateUrl: './direct-smart-sources.component.html',
    changeDetection: ChangeDetectionStrategy.Eager,
    standalone: false
})
export class DirectSmartSourcesComponent implements OnInit, OnDestroy {
  @Output() imported = new EventEmitter<void>();
  providers: SmartProvider[] = [];
  selectedProviderID = '';
  loading = true;
  error = '';
  status = '';
  controller?: AbortController;
  patient?: Patient;
  connection?: SmartConnection;
  activeProvider?: SmartProvider;
  importedSource?: Source;
  importing = false;
  disconnecting = false;
  retrieving = false;
  confirmed = false;
  observationCount?: number;
  private snapshotJSON?: string;
  private destroyed = false;

  constructor(private direct: DirectSmartService, private api: FastenApiService) {}

  ngOnInit(): void { void this.loadProviders(); }

  async loadProviders(): Promise<void> {
    if (this.controller || this.importing || this.disconnecting || this.patient) return;
    this.loading = true;
    this.error = '';
    try {
      const providers = await this.direct.providers();
      if (!this.destroyed) { this.providers = providers; this.selectedProviderID = providers[0]?.id || ''; }
    } catch {
      if (!this.destroyed) this.error = 'Could not load direct providers. Check this installation’s configuration and try again.';
    } finally { if (!this.destroyed) this.loading = false; }
  }

  async connect(): Promise<void> {
    const provider = this.providers.find(item => item.id === this.selectedProviderID);
    if (!provider || this.controller || this.importing || this.disconnecting || this.patient || this.destroyed) return;
    this.activeProvider = provider;
    this.importedSource = undefined;
    this.error = '';
    this.status = 'Your system browser will open. Sign in with the selected provider to retrieve a Patient preview.';
    const controller = this.controller = new AbortController();
    try {
      const result = await this.direct.connect(provider.id, controller.signal);
      if (this.destroyed || controller.signal.aborted || !this.direct.isCurrentIdentity(result.connection)) {
        await this.direct.disconnect(result.connection.id);
        if (!this.destroyed && !controller.signal.aborted) {
          this.status = '';
          this.error = 'The local identity changed. Reconnect before importing.';
        }
        return;
      }
      this.connection = result.connection;
      this.patient = result.patient;
      // Preserve the native raw Patient, including numeric extensions. Only the
      // public fullUrl string is encoded, never a parsed clinical resource.
      const uri = `${provider.fhir_base_url.replace(/\/+$/, '')}/Patient/${encodeURIComponent(result.connection.patient_id)}`;
      this.snapshotJSON = `{"resourceType":"Bundle","type":"collection","entry":[{"fullUrl":${JSON.stringify(uri)},"resource":${result.patientJSON}}]}`;
      this.confirmed = false;
      this.observationCount = undefined;
      this.status = 'Patient retrieved. Review the Patient and receiving local identity. Nothing has been saved yet.';
    } catch (error: unknown) {
      if (!this.destroyed) {
        this.status = controller.signal.aborted ? 'Connection cancelled.' : '';
        if (error instanceof DirectSmartCleanupError) this.error = error.message;
        else if (!controller.signal.aborted) this.error = 'The direct connection failed. Reconnect to try again.';
      }
    } finally { if (this.controller === controller) this.controller = undefined; }
  }

  async retrieveClinical(): Promise<void> {
    const connection = this.connection;
    if (!connection || !this.patient || this.controller || this.importing || this.disconnecting || this.destroyed) return;
    if (!this.direct.isCurrentIdentity(connection)) { await this.discardStaleIdentity(); return; }
    const controller = this.controller = new AbortController();
    this.retrieving = true;
    this.confirmed = false;
    this.error = '';
    this.status = 'Retrieving a bounded Patient/Observation collection. Retrieval does not save records.';
    try {
      const raw = await this.direct.readClinicalSnapshot(connection, controller.signal);
      if (this.destroyed) return;
      if (controller.signal.aborted) { this.clearPreview(); this.status = 'Clinical retrieval cancelled; preview discarded.'; return; }
      if (!this.direct.isCurrentIdentity(connection) || this.connection?.id !== connection.id) { await this.discardStaleIdentity(); return; }
      const bundle: unknown = JSON.parse(raw);
      if (!isRecord(bundle) || bundle['resourceType'] !== 'Bundle' || bundle['type'] !== 'collection' || !Array.isArray(bundle['entry'])) throw new Error();
      let patient: Patient | undefined;
      let observations = 0;
      for (const entry of bundle['entry']) {
        if (!isRecord(entry) || !isRecord(entry['resource'])) throw new Error();
        const resource = entry['resource'];
        if (isPreviewPatient(resource) && resource.id === connection.patient_id && !patient) patient = resource;
        else if (resource['resourceType'] === 'Observation') observations++;
        else throw new Error();
      }
      if (!patient || observations > 5000) throw new Error();
      this.patient = patient;
      this.observationCount = observations;
      this.snapshotJSON = raw;
      this.status = `Clinical preview ready: Patient and ${observations} Observations. Review and confirm before importing; retrieval does not save records.`;
    } catch (error: unknown) {
      if (!this.destroyed) {
        if (controller.signal.aborted) {
          this.clearPreview();
          this.status = 'Clinical retrieval cancelled; preview discarded.';
        } else {
          this.status = '';
          this.error = error instanceof DirectSmartCleanupError ? error.message : 'Clinical retrieval did not complete; no new or partial collection was accepted. The previous complete preview remains; review it before importing or reconnect.';
        }
      }
    } finally {
      this.retrieving = false;
      if (this.controller === controller) this.controller = undefined;
    }
  }

  cancel(): void { this.controller?.abort(); this.status = 'Cancelling direct operation…'; }

  private clearPreview(): void {
    this.patient = undefined;
    this.snapshotJSON = undefined;
    this.connection = undefined;
    this.confirmed = false;
    this.observationCount = undefined;
  }

  private async discardStaleIdentity(): Promise<void> {
    const connection = this.connection;
    this.clearPreview();
    this.importedSource = undefined;
    this.status = '';
    this.error = 'The local identity changed. Reconnect and confirm the receiving profile before importing.';
    if (connection) {
      try { await this.direct.disconnect(connection.id); }
      catch { this.error = 'Could not release the direct session. Close Chartrove to clear it.'; }
    }
  }

  async disconnect(): Promise<void> {
    if (this.importing || this.disconnecting) return;
    this.disconnecting = true;
    this.status = 'Discarding preview and releasing the direct session…';
    this.controller?.abort();
    const connection = this.connection;
    this.clearPreview();
    try {
      if (connection) await this.direct.disconnect(connection.id);
      if (!this.destroyed) this.status = 'Preview discarded and direct session released. Existing local records were not removed.';
    } catch {
      if (!this.destroyed) {
        this.status = '';
        this.error = 'Could not release the direct session. Close Chartrove to clear it.';
      }
    }
    finally { this.disconnecting = false; }
  }

  async importPatient(): Promise<void> {
    const provider = this.activeProvider;
    const connection = this.connection;
    const raw = this.snapshotJSON;
    if (!this.patient || !provider || !connection || !raw || this.controller || this.importing || this.disconnecting || this.destroyed) return;
    if (!this.direct.isCurrentIdentity(connection)) { await this.discardStaleIdentity(); return; }
    if (!this.confirmed) { this.error = 'Confirm your authority and the receiving local identity before importing.'; return; }
    this.importing = true;
    this.error = '';
    try {
      const result = await firstValueFrom(this.api.importClinicalSnapshot(raw,
        {fhir_base_url: provider.fhir_base_url, display: provider.display}, connection.owner.id, true));
      if (!this.destroyed && this.direct.isCurrentIdentity(connection) && this.connection?.id === connection.id) {
        this.importedSource = new Source({id: result.source_id});
        this.clearPreview();
        this.status = result.already_imported ? 'Previously imported; no changes applied. This does not mean older records are current.' :
          `Snapshot saved: ${result.created} created, ${result.updated} updated, ${result.unchanged} unchanged. Automatic synchronization is not enabled.`;
        this.imported.emit();
      } else if (!this.destroyed) {
        this.clearPreview();
        this.importedSource = undefined;
        this.status = 'The local identity changed. No previous-profile preview is shown; verify records after signing back in.';
      }
      try { await this.direct.disconnect(connection.id); }
      catch { if (!this.destroyed) this.error = 'Import completed, but the direct session could not be released. Close Chartrove to clear it.'; }
    } catch {
      if (!this.destroyed) {
        if (!this.direct.isCurrentIdentity(connection)) await this.discardStaleIdentity();
        else this.error = 'Import completion could not be confirmed. You may retry this exact snapshot; previously recorded originals are acknowledged without changing current records.';
      }
    } finally { this.importing = false; }
  }

  get patientName(): string {
    const names = this.patient?.name;
    if (!Array.isArray(names) || !names[0] || typeof names[0] !== 'object') return 'Name not provided';
    const name = names[0];
    if (typeof name.text === 'string' && name.text) return name.text;
    const given = Array.isArray(name.given) ? name.given.filter(item => typeof item === 'string') : [];
    return [...given, ...(typeof name.family === 'string' ? [name.family] : [])].join(' ') || 'Name not provided';
  }

  get birthDate(): string { return typeof this.patient?.birthDate === 'string' ? this.patient.birthDate : 'Not provided'; }
  get gender(): string { return typeof this.patient?.gender === 'string' ? this.patient.gender : 'Not provided'; }

  ngOnDestroy(): void {
    this.destroyed = true;
    this.controller?.abort();
    const connection = this.connection;
    this.clearPreview();
    if (connection) void this.direct.disconnect(connection.id).catch(() => { this.error = 'Close Chartrove to clear the direct session.'; });
  }
}
