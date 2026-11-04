import {ChangeDetectionStrategy, Component, OnDestroy, OnInit} from '@angular/core';
import {CommonModule} from '@angular/common';
import {FormsModule} from '@angular/forms';
import {forkJoin, Subject, takeUntil} from 'rxjs';
import {AccessSource, SharedResource, ResourceListItem, SourceAccessLevel, SourceAccessService, SourceAudit, SourceGrant} from '../../services/source-access.service';

@Component({
  selector: 'app-source-access', standalone: true, imports: [CommonModule, FormsModule],
  changeDetection: ChangeDetectionStrategy.Eager,
  templateUrl: './source-access.component.html'
})
export class SourceAccessComponent implements OnInit, OnDestroy {
  owned: AccessSource[] = [];
  shared: AccessSource[] = [];
  ownerSource = '';
  grants: SourceGrant[] = [];
  audit: SourceAudit[] = [];
  username = '';
  level: SourceAccessLevel = 'VIEW';
  expiration = '';
  newRaw = '{"resourceType":"Observation","status":"final","code":{"text":""}}';
  viewingSource = '';
  counts: Record<string, number> = {};
  resourceType = '';
  resources: ResourceListItem[] = [];
  searchText = '';
  appliedSearch = '';
  page = 0;
  selected: SharedResource | null = null;
  editRaw = '';
  busy = false;
  error = '';
  notice = '';
  private readonly destroy = new Subject<void>();
  private readonly ownerChanged = new Subject<void>();
  private readonly viewChanged = new Subject<void>();
  private readonly detailChanged = new Subject<void>();

  constructor(private api: SourceAccessService) {}
  ngOnInit(): void {
    forkJoin({owned: this.api.owned(), shared: this.api.shared()}).pipe(takeUntil(this.destroy)).subscribe({
      next: data => { this.owned = data.owned; this.shared = data.shared; },
      error: () => this.error = 'Unable to load source permissions.'
    });
  }
  ngOnDestroy(): void { this.destroy.next(); this.destroy.complete(); }
  get canCreateLocal(): boolean { return this.owned.some(s => s.id === this.ownerSource && s.platform_type === 'fasten'); }
  sourceLabel(source: AccessSource): string { return source.display || (source.platform_type === 'fasten' ? 'Locally managed records' : source.id); }
  grantLabel(id: string): string { return this.grants.find(grant => grant.id === id)?.delegate_username || id; }
  selectOwner(): void {
    this.ownerChanged.next(); this.grants = []; this.audit = []; this.error = ''; this.notice = '';
    if (!this.ownerSource) { return; }
    forkJoin({grants: this.api.grants(this.ownerSource), audit: this.api.audit(this.ownerSource)})
      .pipe(takeUntil(this.ownerChanged), takeUntil(this.destroy)).subscribe({
        next: data => { this.grants = data.grants; this.audit = data.audit; },
        error: () => this.error = 'Source permissions are unavailable.'
      });
  }
  saveGrant(): void {
    if (!this.ownerSource || !this.username || this.busy) { return; }
    let expires: string | null = null;
    if (this.expiration) {
      const date = new Date(this.expiration);
      if (!Number.isFinite(date.getTime()) || date.getTime() <= Date.now()) { this.error = 'Choose a future expiration.'; return; }
      expires = date.toISOString();
    }
    this.busy = true; this.error = ''; this.notice = '';
    this.api.grant(this.ownerSource, this.username, this.level, expires).pipe(takeUntil(this.destroy)).subscribe({
      next: () => { this.busy = false; this.selectOwner(); this.notice = 'Source permission saved.'; },
      error: () => { this.busy = false; this.error = 'Permission was not saved. Check the existing account name and expiration.'; }
    });
  }
  revoke(grant: SourceGrant): void {
    if (!this.ownerSource || this.busy) { return; }
    this.busy = true; this.error = ''; this.notice = '';
    this.api.revoke(this.ownerSource, grant.id).pipe(takeUntil(this.destroy)).subscribe({
      next: () => { this.busy = false; this.selectOwner(); this.notice = 'Permission revoked.'; },
      error: () => { this.busy = false; this.error = 'Permission could not be revoked. Refresh and try again.'; }
    });
  }
  viewSource(): void {
    this.searchText = ''; this.appliedSearch = '';
    this.viewChanged.next(); this.detailChanged.next(); this.resources = []; this.counts = {}; this.selected = null;
    this.editRaw = ''; this.resourceType = ''; this.page = 0; this.error = ''; this.notice = '';
    if (!this.viewingSource) { return; }
    this.api.summary(this.viewingSource).pipe(takeUntil(this.viewChanged), takeUntil(this.destroy)).subscribe({
      next: data => this.counts = data.counts,
      error: () => this.clearUnavailable()
    });
  }
  loadResources(page = 0): void {
    this.viewChanged.next(); this.detailChanged.next(); this.selected = null; this.editRaw = ''; this.resources = [];
    this.page = page; this.error = '';
    if (!this.viewingSource || !this.resourceType) { return; }
    const request = this.appliedSearch ? this.api.search(this.viewingSource, this.resourceType, this.appliedSearch, page) : this.api.list(this.viewingSource, this.resourceType, page);
    request.pipe(takeUntil(this.viewChanged), takeUntil(this.destroy)).subscribe({
      next: rows => this.resources = rows,
      error: () => this.clearUnavailable()
    });
  }
  searchResources(): void {
    if (this.searchText.length > 128) { this.error = 'Search text must be at most 128 characters.'; return; }
    this.appliedSearch = this.searchText.trim(); this.loadResources();
  }
  clearSearch(): void { this.searchText = ''; this.searchResources(); }
  open(resource: ResourceListItem): void {
    this.detailChanged.next(); this.selected = null; this.editRaw = ''; this.error = '';
    this.api.detail(this.viewingSource, resource.source_resource_type, resource.source_resource_id)
      .pipe(takeUntil(this.detailChanged), takeUntil(this.viewChanged), takeUntil(this.destroy)).subscribe({
        next: data => {
          if (typeof data.resource_raw_text !== 'string') { this.clearUnavailable(); return; }
          this.selected = data; this.editRaw = data.resource_raw_text;
        },
        error: () => this.clearUnavailable()
      });
  }
  saveResource(): void {
    const selected = this.selected;
    if (!selected?.editable || this.busy) { return; }
    let raw: Record<string, unknown>;
    try { raw = this.parseRaw(this.editRaw); } catch { this.error = 'Enter a valid FHIR JSON object.'; return; }
    if (raw['id'] !== selected.source_resource_id || raw['resourceType'] !== selected.source_resource_type) {
      this.error = 'The resource type and ID cannot change.'; return;
    }
    this.busy = true; this.error = ''; this.notice = '';
    // Parse only to validate identity; transmit the original text so clinical
    // decimals and their expressed precision never pass through JS numbers.
    this.api.edit(this.viewingSource, selected.source_resource_type, selected.source_resource_id, this.editRaw).pipe(takeUntil(this.destroy)).subscribe({
      next: () => { this.busy = false; this.notice = 'Record saved.'; this.open(selected); },
      error: () => { this.busy = false; this.clearUnavailable(); }
    });
  }
  createLocal(): void {
    if (!this.canCreateLocal || this.busy) { return; }
    try { this.parseRaw(this.newRaw); } catch { this.error = 'Enter a valid FHIR JSON object.'; return; }
    this.busy = true; this.error = ''; this.notice = '';
    this.api.createLocal(this.ownerSource, this.newRaw).pipe(takeUntil(this.destroy)).subscribe({
      next: () => { this.busy = false; this.selectOwner(); this.notice = 'New locally managed record created.'; },
      error: () => { this.busy = false; this.error = 'The local record could not be created.'; }
    });
  }
  private parseRaw(value: string): Record<string, unknown> {
    const raw: unknown = JSON.parse(value);
    if (!raw || typeof raw !== 'object' || Array.isArray(raw) || typeof (raw as Record<string, unknown>)['resourceType'] !== 'string') { throw new Error('Invalid FHIR object'); }
    return raw as Record<string, unknown>;
  }
  private clearUnavailable(): void {
    this.searchText = ''; this.appliedSearch = '';
    this.viewChanged.next(); this.detailChanged.next(); this.selected = null; this.editRaw = ''; this.resources = []; this.counts = {};
    this.error = 'Source or record unavailable. Access may have expired or been revoked.';
  }
}
