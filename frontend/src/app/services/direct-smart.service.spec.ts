import {TestBed} from '@angular/core/testing';
import {DirectSmartService, SMART_DESKTOP_BRIDGE, SMART_LOCAL_SESSION, SmartLocalSession} from './direct-smart.service';
import {SmartBridge, SmartConnection} from '../models/direct-smart';

describe('DirectSmartService', () => {
  let service: DirectSmartService;
  let bridge: jasmine.SpyObj<SmartBridge>;
  const owner = {id: '00000000-0000-4000-8000-000000000001', username: 'synthetic-owner'};
  const connection: SmartConnection = {id: 'local-handle', provider_id: 'epic-one', patient_id: 'patient-one', expires_at: 9999999999, owner};
  let session: SmartLocalSession;

  beforeEach(() => {
    session = {token: 'synthetic-local-session', ownerID: owner.id};
    bridge = jasmine.createSpyObj<SmartBridge>('SmartBridge', ['providers', 'connect', 'readPatient', 'readClinicalSnapshot', 'cancel', 'disconnect']);
    bridge.providers.and.resolveTo([{id: 'epic-one', display: 'Example hospital', fhir_base_url: 'https://example.invalid/fhir'}]);
    bridge.connect.and.resolveTo(connection);
    bridge.readPatient.and.resolveTo('{"resourceType":"Patient","id":"patient-one"}');
    bridge.cancel.and.resolveTo();
    bridge.disconnect.and.resolveTo();
    TestBed.configureTestingModule({providers: [{provide: SMART_DESKTOP_BRIDGE, useValue: async () => ({smart: bridge})},
      {provide: SMART_LOCAL_SESSION, useValue: () => session}]});
    service = TestBed.inject(DirectSmartService);
  });

  it('retrieves the authorized Patient without exposing credential fields', async () => {
    const result = await service.connect('epic-one', new AbortController().signal);
    expect(result.patient.id).toBe('patient-one');
    expect(bridge.connect).toHaveBeenCalledWith('epic-one', 'synthetic-local-session');
    expect(bridge.readPatient).toHaveBeenCalledWith('local-handle', 'synthetic-local-session');
    expect(Object.keys(result.connection).sort()).toEqual(['expires_at', 'id', 'owner', 'patient_id', 'provider_id']);
  });

  it('rejects unauthenticated or foreign-owner native connections before data retrieval', async () => {
    session = {token: '', ownerID: ''};
    await expectAsync(service.connect('epic-one', new AbortController().signal)).toBeRejected();
    expect(bridge.connect).not.toHaveBeenCalled();
    session = {token: 'synthetic-local-session', ownerID: owner.id};
    bridge.connect.and.resolveTo({...connection, owner: {...owner, id: '00000000-0000-4000-8000-000000000002'}});
    await expectAsync(service.connect('epic-one', new AbortController().signal)).toBeRejected();
    expect(bridge.readPatient).not.toHaveBeenCalled();
    expect(bridge.disconnect).toHaveBeenCalledWith('local-handle');
  });

  it('discards data returned after the local receiving identity changes', async () => {
    let finish: (value: string) => void;
    let started: () => void;
    const reading = new Promise<void>(resolve => started = resolve);
    bridge.readPatient.and.callFake(() => {started(); return new Promise(resolve => finish = resolve);});
    const pending = service.connect('epic-one', new AbortController().signal);
    await reading;
    session = {token: 'synthetic-other-session', ownerID: '00000000-0000-4000-8000-000000000002'};
    finish('{"resourceType":"Patient","id":"patient-one"}');
    await expectAsync(pending).toBeRejected();
    expect(bridge.disconnect).toHaveBeenCalledWith('local-handle');
  });

  it('rejects mismatched Patient responses, releases the connection, and hides payload errors', async () => {
    bridge.readPatient.and.resolveTo('{"resourceType":"Patient","id":"someone-else"}');
    await expectAsync(service.connect('epic-one', new AbortController().signal)).toBeRejectedWithError('The direct connection failed. Reconnect to try again.');
    expect(bridge.disconnect).toHaveBeenCalledWith('local-handle');
  });

  it('cleans up a connection returned after cancellation', async () => {
    let complete: (connection: SmartConnection) => void;
    let opened: () => void;
    const started = new Promise<void>(resolve => opened = resolve);
    bridge.connect.and.callFake(() => {opened(); return new Promise(resolve => complete = resolve);});
    const controller = new AbortController();
    const pending = service.connect('epic-one', controller.signal);
    await started;
    controller.abort();
    complete(connection);
    await expectAsync(pending).toBeRejected();
    expect(bridge.cancel).toHaveBeenCalledWith('epic-one');
    expect(bridge.disconnect).toHaveBeenCalledWith('local-handle');
    expect(bridge.readPatient).not.toHaveBeenCalled();
  });

  it('does not pass native error bodies to the UI', async () => {
    bridge.connect.and.rejectWith(new Error('sensitive provider response'));
    await expectAsync(service.connect('epic-one', new AbortController().signal)).toBeRejectedWithError('The direct connection failed. Reconnect to try again.');
  });

  it('discards a Patient read that finishes after cancellation', async () => {
    let complete: (value: string) => void;
    let opened: () => void;
    const reading = new Promise<void>(resolve => opened = resolve);
    bridge.readPatient.and.callFake(() => {opened(); return new Promise(resolve => complete = resolve);});
    const controller = new AbortController();
    const pending = service.connect('epic-one', controller.signal);
    await reading;
    controller.abort();
    complete('{"resourceType":"Patient","id":"patient-one"}');
    await expectAsync(pending).toBeRejected();
    expect(bridge.disconnect).toHaveBeenCalledWith('local-handle');
  });
});
