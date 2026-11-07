import { Subject } from 'rxjs';
import { MedicalRecordWizardComponent } from './medical-record-wizard.component';

describe('Confirmed encounter unlink display', () => {
  const sourceA = '11111111-1111-4111-8111-111111111111';
  const sourceB = '22222222-2222-4222-8222-222222222222';

  function deferred() {
    let resolve!: (value: boolean) => void;
    const result = new Promise<boolean>(done => resolve = done);
    return { componentInstance: {}, result, resolve };
  }

  function resource(source = sourceA, id = 'synthetic-procedure') {
    return {
      source_id: source, source_resource_type: 'Procedure', source_resource_id: id,
      resource_raw: {
        resourceType: 'Procedure', id, status: 'completed',
        code: { text: 'Synthetic café صفر' }, performedDateTime: '2020-01-02',
      },
    };
  }

  function encounter(source = sourceA, id = 'synthetic-encounter') {
    return {
      source_id: source, source_resource_type: 'Encounter', source_resource_id: id,
      related_resources: {
        Procedure: [resource(source), resource(sourceB, 'unrelated-procedure')],
        Practitioner: [resource(sourceB, 'unrelated-group')],
      },
    };
  }

  function harness() {
    const dialogs: ReturnType<typeof deferred>[] = [];
    const response = new Subject<any>();
    const remove = jasmine.createSpy('removeEncounterRelatedResource').and.returnValue(response);
    const open = jasmine.createSpy('open').and.callFake(() => {
      const dialog = deferred();
      dialogs.push(dialog);
      return dialog;
    });
    const component = new MedicalRecordWizardComponent({} as any, { open } as any, {
      removeEncounterRelatedResource: remove,
    } as any);
    const record = encounter();
    component.existingEncounter = record as any;
    return { component, record, response, remove, open, dialogs };
  }

  beforeEach(() => spyOn(window, 'alert'));

  it('waits for confirmation from the API and removes only the matching displayed tuple', async () => {
    const h = harness();
    const selected = h.record.related_resources.Procedure[0];
    const unrelated = h.record.related_resources.Procedure[1];
    const raw = JSON.stringify(selected.resource_raw);
    h.component.handleUnlinkRequested({ ...selected } as any);
    h.dialogs[0].resolve(true);
    await Promise.resolve();
    expect(h.remove).toHaveBeenCalledOnceWith('synthetic-encounter', 'synthetic-procedure', 'Procedure');
    expect(h.record.related_resources.Procedure).toEqual([selected, unrelated]);
    expect(h.component.submitWizardLoading).toBeTrue();
    h.response.next({});
    h.response.complete();
    expect(h.record.related_resources.Procedure).toEqual([unrelated]);
    expect(JSON.stringify(selected.resource_raw)).toBe(raw);
    expect(h.record.related_resources.Practitioner.length).toBe(1);
    expect(h.component.submitWizardLoading).toBeFalse();
  });

  it('preserves every source and raw field when the server rejects ambiguous IDs', async () => {
    const h = harness();
    h.record.related_resources.Procedure[1] = resource(sourceB);
    const before = JSON.stringify(h.record);
    h.component.handleUnlinkRequested(h.record.related_resources.Procedure[0] as any);
    h.dialogs[0].resolve(true);
    await Promise.resolve();
    h.response.error({ status: 409 });
    expect(JSON.stringify(h.record)).toBe(before);
    expect(h.component.submitWizardLoading).toBeFalse();
    expect(window.alert).toHaveBeenCalled();
  });

  it('does not send an absent source tuple or remove the last unrelated row', async () => {
    const h = harness();
    const before = JSON.stringify(h.record);
    h.component.handleUnlinkRequested(resource(sourceB) as any);
    if (h.dialogs.length) h.dialogs[0].resolve(true);
    await Promise.resolve();
    expect(h.remove).not.toHaveBeenCalled();
    expect(JSON.stringify(h.record)).toBe(before);
    h.response.complete();
  });

  it('rejects a confirmation after the selected encounter changes', async () => {
    const h = harness();
    const before = JSON.stringify(h.record);
    h.component.handleUnlinkRequested(h.record.related_resources.Procedure[0] as any);
    const replacement = encounter(sourceB, 'replacement-encounter');
    const replacementBefore = JSON.stringify(replacement);
    h.component.existingEncounter = replacement as any;
    h.dialogs[0].resolve(true);
    await Promise.resolve();
    expect(h.remove).not.toHaveBeenCalled();
    expect(JSON.stringify(h.record)).toBe(before);
    expect(JSON.stringify(replacement)).toBe(replacementBefore);
    h.response.complete();
  });

  it('does not apply an earlier response to a replacement encounter', async () => {
    const h = harness();
    h.component.handleUnlinkRequested(h.record.related_resources.Procedure[0] as any);
    h.dialogs[0].resolve(true);
    await Promise.resolve();
    const replacement = encounter(sourceB, 'replacement-encounter');
    const before = JSON.stringify(replacement);
    h.component.existingEncounter = replacement as any;
    h.response.next({});
    h.response.complete();
    expect(JSON.stringify(replacement)).toBe(before);
    expect(h.component.submitWizardLoading).toBeFalse();
  });

  it('allows only one in-flight unlink when two dialogs are confirmed', async () => {
    const h = harness();
    h.component.handleUnlinkRequested(h.record.related_resources.Procedure[0] as any);
    h.component.handleUnlinkRequested(h.record.related_resources.Procedure[1] as any);
    h.dialogs[0].resolve(true);
    h.dialogs[1].resolve(true);
    await Promise.resolve();
    expect(h.remove.calls.count()).toBe(1);
    h.response.complete();
  });

  it('preserves the record and sends no request when confirmation is declined', async () => {
    const h = harness();
    const before = JSON.stringify(h.record);
    h.component.handleUnlinkRequested(h.record.related_resources.Procedure[0] as any);
    h.dialogs[0].resolve(false);
    await Promise.resolve();
    expect(h.remove).not.toHaveBeenCalled();
    expect(JSON.stringify(h.record)).toBe(before);
    expect(h.component.submitWizardLoading).toBeFalse();
    h.response.complete();
  });
});
