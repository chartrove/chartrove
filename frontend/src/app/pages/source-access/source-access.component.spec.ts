import {ComponentFixture, TestBed} from '@angular/core/testing';
import {of, Subject, throwError} from 'rxjs';
import {SourceAccessComponent} from './source-access.component';
import {SharedResource, SourceAccessService} from '../../services/source-access.service';

describe('SourceAccessComponent authorization UI', () => {
  let fixture: ComponentFixture<SourceAccessComponent>;
  let component: SourceAccessComponent;
  let api: jasmine.SpyObj<SourceAccessService>;
  const record: SharedResource = {source_resource_type: 'Observation', source_resource_id: 'synthetic-id',
    resource_raw: {resourceType: 'Observation', id: 'synthetic-id', status: 'final'},
    resource_raw_text: '{"resourceType":"Observation","id":"synthetic-id","status":"final"}', editable: false};

  beforeEach(async () => {
    api = jasmine.createSpyObj('SourceAccessService', ['owned', 'shared', 'grants', 'audit', 'grant', 'revoke', 'summary', 'list', 'search', 'detail', 'edit', 'createLocal']);
    api.owned.and.returnValue(of([{id: 'synthetic-owned', display: 'Local', platform_type: 'fasten'}]));
    api.shared.and.returnValue(of([{id: 'synthetic-shared', display: 'Shared', access_level: 'VIEW'}]));
    api.grants.and.returnValue(of([])); api.audit.and.returnValue(of([]));
    api.summary.and.returnValue(of({source: {id: 'synthetic-shared', display: 'Shared'}, counts: {Observation: 1}}));
    api.list.and.returnValue(of([record])); api.detail.and.returnValue(of(record));
    api.search.and.returnValue(of([{source_resource_type:'Observation',source_resource_id:'synthetic-hit',sort_title:'Synthetic search match'}]));
    api.edit.and.returnValue(of(null));
    await TestBed.configureTestingModule({imports: [SourceAccessComponent], providers: [{provide: SourceAccessService, useValue: api}]}).compileComponents();
    fixture = TestBed.createComponent(SourceAccessComponent); component = fixture.componentInstance; fixture.detectChanges();
    component.viewingSource = 'synthetic-shared'; component.viewSource(); component.resourceType = 'Observation';
  });

  it('applies a source-scoped search and rechecks access when opening metadata hits', () => {
    component.searchText = 'Synthetic'; component.searchResources(); fixture.detectChanges();
    expect(api.search).toHaveBeenCalledWith('synthetic-shared', 'Observation', 'Synthetic', 0);
    expect(component.resources.length).toBe(1);
    component.open(component.resources[0]);
    expect(api.detail).toHaveBeenCalledWith('synthetic-shared', 'Observation', 'synthetic-hit');
    component.searchText = 'unapplied'; component.loadResources(1);
    expect(api.search).toHaveBeenCalledWith('synthetic-shared', 'Observation', 'Synthetic', 1);
  });

  it('explains transcript search only for DocumentReference while keeping source-scoped metadata requests', () => {
    component.resourceType='DocumentReference';fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain('reviewed transcripts from new local document imports');
    component.searchText='Synthetic transcript';component.searchResources();
    expect(api.search).toHaveBeenCalledWith('synthetic-shared','DocumentReference','Synthetic transcript',0);
    component.resourceType='Observation';fixture.detectChanges();
    expect(fixture.nativeElement.textContent).not.toContain('reviewed transcripts from new local document imports');
  });

  it('clears search state and pending results when changing sources or losing access', () => {
    const pending = new Subject<any>(); api.search.and.returnValue(pending);
    component.searchText = 'Synthetic'; component.searchResources(); component.viewSource();
    pending.next([{source_resource_type:'Observation',source_resource_id:'stale'}]);
    expect(component.resources).toEqual([]); expect(component.appliedSearch).toBe(''); expect(component.searchText).toBe('');
    component.resourceType = 'Observation'; component.searchText = 'Synthetic';
    api.search.and.returnValue(throwError(() => new Error('revoked'))); component.searchResources();
    expect(component.resources).toEqual([]); expect(component.selected).toBeNull(); expect(component.editRaw).toBe('');
  });

  it('renders VIEW or imported records as text without an editor or save action', () => {
    const raw = {...record.resource_raw, note: '<em>Synthetic clinical note</em>'};
    api.detail.and.returnValue(of({...record, resource_raw: raw, resource_raw_text: JSON.stringify(raw)}));
    component.open(record); fixture.detectChanges();
    const view = fixture.nativeElement as HTMLElement;
    expect(view.querySelector('#edit-record')).toBeNull();
    expect(view.querySelector('em')).toBeNull();
    expect(view.textContent).toContain('<em>Synthetic clinical note</em>');
    component.saveResource(); expect(api.edit).not.toHaveBeenCalled();
  });

  it('renders asynchronously received source metadata without a user action', async () => {
    const pending = new Subject<any[]>();
    api.owned.and.returnValue(pending);
    component.owned = [];
    fixture.detectChanges();
    fixture.autoDetectChanges();
    component.ngOnInit();
    await new Promise<void>(resolve => fixture.ngZone!.run(() => setTimeout(() => {
      pending.next([{id: 'synthetic-async', display: 'Asynchronous synthetic source'}]);
      pending.complete(); resolve();
    }, 0)));
    await fixture.whenStable();
    const view = fixture.nativeElement as HTMLElement;
    expect(view.querySelector('#owner-source')?.textContent).toContain('Asynchronous synthetic source');
  });

  it('clears stale medical content and editing authority after denial', () => {
    component.selected = {...record, editable: true}; component.editRaw = JSON.stringify(record.resource_raw);
    component.resources = [record];
    api.edit.and.returnValue(throwError(() => ({status: 404})));
    component.saveResource(); fixture.detectChanges();
    expect(api.edit).toHaveBeenCalledOnceWith('synthetic-shared', 'Observation', 'synthetic-id', JSON.stringify(record.resource_raw));
    expect(component.selected).toBeNull(); expect(component.editRaw).toBe(''); expect(component.resources).toEqual([]);
    expect(fixture.nativeElement.querySelector('#edit-record')).toBeNull();
    expect(component.error).toContain('revoked');
  });

  it('cancels in-flight details when the selected source changes', () => {
    const pending = new Subject<SharedResource>(); api.detail.and.returnValue(pending);
    component.open(record); component.viewingSource = 'synthetic-owned'; component.viewSource();
    pending.next({...record, editable: true});
    expect(component.selected).toBeNull(); expect(component.editRaw).toBe('');
  });

  it('does not accept a late grant list for a previously selected owner source', () => {
    const pending = new Subject<any[]>(); api.grants.and.returnValue(pending);
    component.ownerSource = 'synthetic-owned'; component.selectOwner();
    component.ownerSource = ''; component.selectOwner();
    pending.next([{id: 'synthetic-old-grant', delegate_username: 'synthetic-old'}]); pending.complete();
    expect(component.grants).toEqual([]);
  });

  it('rejects changed FHIR identity and invalid JSON before attempting a write', () => {
    component.selected = {...record, editable: true}; component.editRaw = '{"resourceType":"Patient","id":"changed"}';
    component.saveResource(); expect(api.edit).not.toHaveBeenCalled(); expect(component.error).toContain('cannot change');
    component.editRaw = 'null'; component.saveResource(); expect(api.edit).not.toHaveBeenCalled();
  });

  it('does not offer creation in a shared or externally sourced source', () => {
    component.ownerSource = 'synthetic-shared'; component.createLocal(); expect(api.createLocal).not.toHaveBeenCalled();
    component.owned = [{id: 'synthetic-provider', display: 'External', platform_type: 'manual'}];
    component.ownerSource = 'synthetic-provider'; component.createLocal(); expect(api.createLocal).not.toHaveBeenCalled();
  });

  it('preserves clinical decimal literals through display and save', () => {
    const exact = '{"resourceType":"Observation","id":"synthetic-id","status":"final","valueQuantity":{"value":0.1234567890123456789},"component":[{"valueQuantity":{"value":2.00}}]}';
    api.detail.and.returnValue(of({...record, editable: true, resource_raw: JSON.parse(exact), resource_raw_text: exact}));
    component.open(record);
    expect(component.editRaw).toBe(exact);
    component.saveResource();
    expect(api.edit).toHaveBeenCalledOnceWith('synthetic-shared', 'Observation', 'synthetic-id', exact);
  });

  function document(attachments: readonly Record<string, unknown>[]): SharedResource {
    const raw = {resourceType: 'DocumentReference', id: 'synthetic-document', content: attachments.map(attachment => ({attachment}))};
    return {source_resource_type: 'DocumentReference', source_resource_id: raw.id, resource_raw: raw, resource_raw_text: JSON.stringify(raw), editable: false};
  }

  function encoded(text: string): string {
    return btoa(Array.from(new TextEncoder().encode(text), byte => String.fromCharCode(byte)).join(''));
  }

  it('previews authorized inline UTF-8 document text literally and retains the exact raw record', () => {
    const text = 'Synthetic résumé · μg\n<em>Literal attachment text</em>';
    const doc = document([{contentType: 'application/pdf', data: 'synthetic-original'}, {contentType: 'text/plain', title: 'Synthetic <em>title</em>', data: encoded(text)}]);
    api.detail.and.returnValue(of(doc)); component.resourceType = 'DocumentReference';
    component.open(doc); fixture.detectChanges();
    const view = fixture.nativeElement as HTMLElement;
    expect(view.querySelector('#document-text-body')?.textContent).toBe(text);
    expect(view.querySelector('#document-text')?.textContent).toContain('Synthetic <em>title</em>');
    expect(view.querySelector('em')).toBeNull();
    expect(view.querySelector('#edit-record')).toBeNull();
    expect(view.textContent).toContain(doc.resource_raw_text!);
    expect(component.editRaw).toBe(doc.resource_raw_text!);
    component.saveResource(); expect(api.edit).not.toHaveBeenCalled();
  });

  it('keeps remote, ambiguous, invalid, oversized and non-UTF-8 document data in the raw record without fetching', () => {
    const fetch = spyOn(globalThis, 'fetch').and.rejectWith(new Error('Synthetic remote fetch must not happen'));
    const valid = {contentType: 'text/plain', data: encoded('Synthetic valid text')};
    const cases: Record<string, unknown>[][] = [
      [{contentType: 'text/plain', url: 'https://example.invalid/synthetic.txt'}],
      [{contentType: 'text/html', data: encoded('<em>Synthetic HTML document</em>')}],
      [valid, valid],
      [{contentType: 'text/plain', data: 'invalid-base64'}],
      [{contentType: 'text/plain', data: btoa(String.fromCharCode(255))}],
      [{contentType: 'text/plain; charset=iso-8859-1', data: btoa('Synthetic text')}],
      [{contentType: 'text/plain', data: btoa('x'.repeat(128 * 1024 + 1))}],
      [{contentType: 'text/plain', data: encoded('é'.repeat(64 * 1024 + 1))}],
      [...Array.from({length: 8}, () => ({contentType: 'image/png', data: 'synthetic-image'})), valid],
    ];
    for (const attachments of cases) {
      const doc = document(attachments); api.detail.and.returnValue(of(doc));
      component.resourceType = 'DocumentReference'; component.open(doc); fixture.detectChanges();
      expect(fixture.nativeElement.querySelector('#document-text-body')).toBeNull();
      expect(fixture.nativeElement.textContent).toContain(doc.resource_raw_text!);
      expect(component.editRaw).toBe(doc.resource_raw_text!);
    }
    expect(fetch).not.toHaveBeenCalled(); expect(api.edit).not.toHaveBeenCalled();
  });

  it('accepts the document preview byte limit and clears text on selection changes, stale responses and access denial', () => {
    const doc = document([{contentType: 'text/plain; charset="UTF-8"', data: btoa('x'.repeat(128 * 1024))}]);
    api.detail.and.returnValue(of(doc)); component.resourceType = 'DocumentReference';
    component.open(doc); fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('#document-text-body')?.textContent.length).toBe(128 * 1024);
    component.loadResources(1); fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('#document-text-body')).toBeNull();
    const pending = new Subject<SharedResource>(); api.detail.and.returnValue(pending);
    component.open(doc); fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('#document-text-body')).toBeNull();
    component.viewingSource = 'synthetic-owned'; component.viewSource(); pending.next(doc); fixture.detectChanges();
    expect(component.selected).toBeNull(); expect(fixture.nativeElement.querySelector('#document-text-body')).toBeNull();
    api.detail.and.returnValue(of(doc)); component.resourceType = 'DocumentReference'; component.open(doc); fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('#document-text-body')).not.toBeNull();
    api.detail.and.returnValue(throwError(() => ({status: 404}))); component.open(doc); fixture.detectChanges();
    expect(component.selected).toBeNull(); expect(component.editRaw).toBe('');
    expect(fixture.nativeElement.querySelector('#document-text-body')).toBeNull();
    api.detail.and.returnValue(of(doc)); component.open(doc); fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('#document-text-body')).not.toBeNull();
    fixture.destroy(); expect(component.documentText).toBeNull(); expect(component.selected).toBeNull();
  });

  it('sends a source-specific grant with expiration and can explicitly revoke it', () => {
    component.ownerSource = 'synthetic-owned'; component.username = 'synthetic-delegate'; component.level = 'EDIT';
    component.expiration = '2099-01-01T12:00';
    const grant = {id: 'synthetic-grant', delegate_username: 'synthetic-delegate', delegate_user_id: 'synthetic-user', access_level: 'EDIT' as const, expires_at: null, revoked_at: null};
    api.grant.and.returnValue(of(grant)); api.revoke.and.returnValue(of(null));
    component.saveGrant();
    expect(api.grant).toHaveBeenCalledOnceWith('synthetic-owned', 'synthetic-delegate', 'EDIT', new Date(component.expiration).toISOString());
    component.revoke(grant); expect(api.revoke).toHaveBeenCalledOnceWith('synthetic-owned', 'synthetic-grant');
  });
});
