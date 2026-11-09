export interface SmartProvider {
  readonly id: string;
  readonly display: string;
  readonly fhir_base_url: string;
}

export interface SmartConnection {
  readonly id: string;
  readonly provider_id: string;
  readonly patient_id: string;
  readonly expires_at: number;
  readonly owner: {readonly id: string; readonly username: string};
}

export interface SmartBridge {
  providers(): Promise<SmartProvider[]>;
  connect(providerID: string, localToken: string): Promise<SmartConnection>;
  readPatient(connectionID: string, localToken: string): Promise<string>;
  readClinicalSnapshot(connectionID: string, localToken: string): Promise<string>;
  cancel(providerID: string): Promise<void>;
  disconnect(connectionID: string): Promise<void>;
}

export interface SmartImportOrigin {
  readonly fhir_base_url: string;
  readonly display: string;
}

export interface SmartSnapshotImportResult {
  readonly source_id: string;
  readonly snapshot_id: string;
  readonly sha256: string;
  readonly patient_id: string;
  readonly created: number;
  readonly updated: number;
  readonly unchanged: number;
  readonly already_imported: boolean;
}

export function isSmartSnapshotImportResult(value: unknown): value is SmartSnapshotImportResult {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return false;
  const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
  return 'source_id' in value && typeof value.source_id === 'string' && uuid.test(value.source_id) &&
    'snapshot_id' in value && typeof value.snapshot_id === 'string' && uuid.test(value.snapshot_id) &&
    'sha256' in value && typeof value.sha256 === 'string' && /^[0-9a-f]{64}$/.test(value.sha256) &&
    'patient_id' in value && typeof value.patient_id === 'string' && /^[A-Za-z0-9.-]{1,64}$/.test(value.patient_id) &&
    'created' in value && typeof value.created === 'number' && Number.isSafeInteger(value.created) && value.created >= 0 &&
    'updated' in value && typeof value.updated === 'number' && Number.isSafeInteger(value.updated) && value.updated >= 0 &&
    'unchanged' in value && typeof value.unchanged === 'number' && Number.isSafeInteger(value.unchanged) && value.unchanged >= 0 &&
    value.created + value.updated + value.unchanged <= 5001 &&
    'already_imported' in value && typeof value.already_imported === 'boolean';
}
