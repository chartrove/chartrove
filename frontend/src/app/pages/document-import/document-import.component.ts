import { ChangeDetectionStrategy, Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { environment } from '../../../environments/environment';
import { AuthService } from '../../services/auth.service';
import { DocumentImportService, DocumentSource, DocumentReceipt } from '../../services/document-import.service';
import { LocalDocumentOCRService, OCR_MAX_TEXT_BYTES } from '../../services/local-document-ocr.service';

@Component({
  selector:'app-document-import', standalone:true, imports:[CommonModule,FormsModule,RouterLink],
  changeDetection:ChangeDetectionStrategy.Eager, templateUrl:'./document-import.component.html'
})
export class DocumentImportComponent implements OnInit, OnDestroy {
  supported = !environment.environment_desktop && !environment.environment_cloud;
  sources: DocumentSource[] = []; source = ''; file: File | undefined;
  title = ''; text = ''; reviewed = false; processing = false; saving = false;
  page = 0; pages = 0; error = ''; notice = ''; originalURL = ''; receipt: DocumentReceipt | null = null;
  private controller: AbortController | undefined;
  private subject = ''; private destroyed = false;
  constructor(private ocr: LocalDocumentOCRService, private api: DocumentImportService, private auth: AuthService) {}
  async ngOnInit(): Promise<void> {
    if (!this.supported) return;
    try {
      this.subject = this.auth.GetCurrentUser().sub;
      const sources = await this.api.sources();
      if (!this.destroyed && this.sameAccount()) { this.sources = sources; if (sources.length === 1) this.source = sources[0].id; }
    } catch { if (!this.destroyed) this.error = 'Import sources are unavailable. Use an unlocked local Chartrove account.'; }
  }
  ngOnDestroy(): void { this.destroyed = true; this.clear(); }
  private sameAccount(): boolean {
    try { if (this.auth.GetCurrentUser().sub === this.subject) return true; } catch {}
    this.clear(); this.sources = []; this.source = ''; this.error = 'Your sign-in changed. Reopen document import before continuing.'; return false;
  }
  clear(): void {
    this.controller?.abort(); this.controller = undefined;
    if (this.originalURL) URL.revokeObjectURL(this.originalURL);
    this.originalURL = ''; this.file = undefined; this.title = ''; this.text = ''; this.reviewed = false;
    this.receipt = null; this.notice = ''; this.processing = false; this.page = 0; this.pages = 0;
  }
  cancel(): void { this.clear(); this.error = 'Local processing cancelled.'; }
  async selected(event: Event): Promise<void> {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0]; input.value = '';
    if (!file || this.saving || !this.sameAccount()) return;
    this.clear(); this.error = '';
    if (file.name.length > 200) { this.error = 'Use a filename no longer than 200 characters.'; return; }
    const controller = new AbortController(); this.controller = controller;
    this.file = file; this.title = file.name; this.processing = true;
    try {
      const result = await this.ocr.extract(file,controller.signal,(page,pages) => { if (this.controller === controller) { this.page = page; this.pages = pages; } });
      if (this.destroyed || controller.signal.aborted || this.controller !== controller || !this.sameAccount()) return;
      this.text = result.text; this.pages = result.pages;
      this.originalURL = URL.createObjectURL(new Blob([file],{type:result.mediaType}));
    } catch {
      if (this.controller === controller && !this.destroyed) { this.clear(); this.error = 'The document could not be processed. Use a valid PNG, JPEG or PDF within the limits, and try again.'; }
    } finally { if (this.controller === controller) this.processing = false; }
  }
  async save(): Promise<void> {
    if (this.processing || this.saving || !this.file || !this.source || !this.reviewed || !this.text.trim() || !this.title.trim() || !this.sameAccount()) return;
    if (new TextEncoder().encode(this.text).length > OCR_MAX_TEXT_BYTES) { this.error = 'The reviewed transcript must be no larger than 128 KiB.'; return; }
    this.saving = true; this.error = ''; this.notice = '';
    try {
      const receipt = await this.api.save(this.source,this.file,this.title,this.text,this.reviewed);
      if (!this.destroyed && this.sameAccount()) { this.receipt = receipt; this.notice = receipt.created ? 'Original document and reviewed transcript saved.' : 'This document and transcript were already imported.'; }
    } catch { if (!this.destroyed) this.error = 'Import was not confirmed. Check your account and source access before retrying.'; }
    finally { this.saving = false; }
  }
}
