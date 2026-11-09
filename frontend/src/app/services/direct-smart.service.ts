import {Inject, Injectable, InjectionToken, inject} from '@angular/core';
import {Patient} from 'fhir/r4';
import {DesktopBridge, GetDesktopBridge} from '../../lib/utils/desktop_bridge';
import {SmartBridge, SmartConnection, SmartProvider} from '../models/direct-smart';
import {AuthService} from './auth.service';

function isRecord(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === 'object' && !Array.isArray(value);
}

export class DirectSmartCleanupError extends Error {
  constructor() { super('Could not release the direct session. Close Chartrove to clear it.'); }
}

export const SMART_DESKTOP_BRIDGE = new InjectionToken<() => Promise<Pick<DesktopBridge, 'smart'>>>(
  'SMART_DESKTOP_BRIDGE', {providedIn: 'root', factory: () => GetDesktopBridge});

export interface SmartLocalSession {readonly token: string; readonly ownerID: string}
export const SMART_LOCAL_SESSION = new InjectionToken<() => SmartLocalSession>('SMART_LOCAL_SESSION', {
  providedIn: 'root', factory: () => {
    const auth = inject(AuthService);
    return () => ({token: auth.GetAuthToken(), ownerID: auth.GetCurrentUser().native_user_id || ''});
  }
});

@Injectable({providedIn: 'root'})
export class DirectSmartService {
  constructor(@Inject(SMART_DESKTOP_BRIDGE) private getDesktopBridge: () => Promise<Pick<DesktopBridge, 'smart'>>,
    @Inject(SMART_LOCAL_SESSION) private getLocalSession: () => SmartLocalSession) {}

  private localSession(): SmartLocalSession | undefined {
    try {
      const session = this.getLocalSession();
      return session.token && /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/.test(session.ownerID) ? session : undefined;
    } catch { return undefined; }
  }

  isCurrentIdentity(connection: SmartConnection): boolean {
    return !!connection.owner?.id && this.localSession()?.ownerID === connection.owner.id;
  }

  private async native(): Promise<SmartBridge> {
    try {
      const {smart} = await this.getDesktopBridge();
      if (!smart) throw new Error();
      return smart;
    } catch { throw new Error('Direct connections require the current Chartrove desktop host.'); }
  }

  async providers(): Promise<SmartProvider[]> {
    try {
      const providers = await (await this.native()).providers();
      if (!Array.isArray(providers)) throw new Error();
      return providers.map(provider => {
        if (!provider || typeof provider.id !== 'string' || !/^[a-z0-9][a-z0-9-]{0,63}$/.test(provider.id) ||
          typeof provider.display !== 'string' || !provider.display || provider.display.length > 100 ||
          typeof provider.fhir_base_url !== 'string' || new URL(provider.fhir_base_url).protocol !== 'https:') throw new Error();
        return {id: provider.id, display: provider.display, fhir_base_url: provider.fhir_base_url};
      });
    } catch {
      throw new Error('Could not load direct providers. Check this installation’s configuration and try again.');
    }
  }

  async connect(providerID: string, signal: AbortSignal): Promise<{connection: SmartConnection, patient: Patient, patientJSON: string}> {
    const session = this.localSession();
    if (!session) throw new Error('Sign in again before opening a direct provider connection.');
    const smart = await this.native();
    let connectionID: string | undefined;
    let keep = false;
    const cleanup: Promise<PromiseSettledResult<void>[]>[] = [];
    const abort = () => {
      // Teardown must handle both promises, even if the native app is quitting.
      cleanup.push(Promise.allSettled([smart.cancel(providerID), ...(connectionID ? [smart.disconnect(connectionID)] : [])]));
    };
    signal.addEventListener('abort', abort, {once: true});
    try {
      if (signal.aborted) throw new Error();
      const received = await smart.connect(providerID, session.token);
      connectionID = typeof received?.id === 'string' ? received.id : undefined;
      if (signal.aborted || !connectionID || received.provider_id !== providerID ||
        typeof received.patient_id !== 'string' || !/^[A-Za-z0-9.-]{1,64}$/.test(received.patient_id) ||
        !Number.isFinite(received.expires_at) || received.expires_at <= Date.now() / 1000 ||
        !received.owner || received.owner.id !== session.ownerID || typeof received.owner.username !== 'string' ||
        !received.owner.username || received.owner.username.length > 256) throw new Error();
      const connection: SmartConnection = {id: connectionID, provider_id: providerID,
        patient_id: received.patient_id, expires_at: received.expires_at,
        owner: {id: received.owner.id, username: received.owner.username}};
      if (!this.isCurrentIdentity(connection)) throw new Error();
      const patientJSON = await smart.readPatient(connectionID, session.token);
      const parsed: unknown = JSON.parse(patientJSON);
      if (signal.aborted || !this.isCurrentIdentity(connection) || !isRecord(parsed) || parsed['resourceType'] !== 'Patient' || parsed['id'] !== connection.patient_id) throw new Error();
      // The Go host verifies the same identity. Only the FHIR Patient crosses
      // this boundary; no provider credentials or arbitrary request URLs do.
      const patient = parsed as unknown as Patient;
      keep = true;
      return {connection, patient, patientJSON};
    } catch {
      throw new Error(signal.aborted ? 'Connection cancelled.' : 'The direct connection failed. Reconnect to try again.');
    } finally {
      signal.removeEventListener('abort', abort);
      if (!keep && connectionID) cleanup.push(Promise.allSettled([smart.disconnect(connectionID)]));
      const results = await Promise.all(cleanup);
      if (results.some(group => group.some(result => result.status === 'rejected'))) {
        throw new DirectSmartCleanupError();
      }
    }
  }

  async readClinicalSnapshot(connection: SmartConnection, signal: AbortSignal): Promise<string> {
    const session = this.localSession();
    if (!session || !this.isCurrentIdentity(connection)) throw new Error('The local identity changed. Reconnect before retrieving records.');
    const smart = await this.native();
    const cleanup: Promise<PromiseSettledResult<void>[]>[] = [];
    const abort = () => { cleanup.push(Promise.allSettled([smart.disconnect(connection.id)])); };
    signal.addEventListener('abort', abort, {once: true});
    try {
      if (signal.aborted) throw new Error();
      const raw = await smart.readClinicalSnapshot(connection.id, session.token);
      if (signal.aborted || !this.isCurrentIdentity(connection) || typeof raw !== 'string' ||
        new TextEncoder().encode(raw).byteLength > 16 * 1024 * 1024) throw new Error();
      return raw;
    } catch {
      throw new Error(signal.aborted ? 'Clinical retrieval cancelled.' : 'Clinical retrieval did not complete. No partial collection is available.');
    } finally {
      signal.removeEventListener('abort', abort);
      const outcomes = await Promise.all(cleanup);
      if (outcomes.some(group => group.some(outcome => outcome.status === 'rejected'))) throw new DirectSmartCleanupError();
    }
  }

  async disconnect(connectionID: string): Promise<void> {
    try { await (await this.native()).disconnect(connectionID); }
    catch { throw new DirectSmartCleanupError(); }
  }
}
