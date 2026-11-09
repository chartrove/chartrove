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
