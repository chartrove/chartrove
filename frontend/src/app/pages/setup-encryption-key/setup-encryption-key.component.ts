import { Component, OnInit, OnDestroy, ChangeDetectorRef, ChangeDetectionStrategy } from '@angular/core';
import { FormBuilder, FormControl, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import {firstValueFrom, Subject, timer} from 'rxjs';
import {takeUntil} from 'rxjs/operators';
import { FastenApiService } from 'src/app/services/fasten-api.service';

@Component({
    selector: 'app-setup-encryption-key',
    templateUrl: './setup-encryption-key.component.html',
    styleUrls: ['./setup-encryption-key.component.scss'],
    changeDetection: ChangeDetectionStrategy.Eager,
    standalone: false
})
export class SetupEncryptionKeyComponent implements OnInit, OnDestroy {
  encryptionKeyForm: FormGroup<{encryptionKey: FormControl<string>}>;
  isProcessing = false;
  error = false;
  isValidated = false;
  statusMessage: string = '';
  countdown: number = 5;
  private countdownInterval: ReturnType<typeof setInterval> | null = null;
  private readonly destroyed = new Subject<void>();
  private active = true;
  private operationEpoch = 0;

  constructor(
    private fb: FormBuilder,
    private fastenService: FastenApiService,
    private router: Router,
    private cdRef: ChangeDetectorRef,
  ) {}

  private async sleep(ms: number): Promise<void> {
    await firstValueFrom(timer(ms).pipe(takeUntil(this.destroyed)));
  }

  private current(epoch: number): boolean {
    return this.active && epoch === this.operationEpoch;
  }

  ngOnInit() {
    this.encryptionKeyForm = this.fb.nonNullable.group({
      encryptionKey: ['', Validators.required],
    });

  }

  ngOnDestroy(): void {
    this.active = false;
    this.operationEpoch++;
    this.destroyed.next();
    this.destroyed.complete();
    if (this.countdownInterval !== null) clearInterval(this.countdownInterval);
    this.countdownInterval = null;
    this.encryptionKeyForm?.reset({encryptionKey: ''});
    this.isProcessing = false;
    this.isValidated = false;
    this.error = false;
    this.statusMessage = '';
  }

  onEncryptionKeyChange(): void {
    if (!this.active || this.isProcessing) return;
    if (this.encryptionKeyForm.controls.encryptionKey.valid) {
      this.error = false;
      this.statusMessage = '';
    }
  }

  handleButtonClick(): void {
    if (!this.active || this.isProcessing) return;
    if (this.isValidated) {
      this.redirectToDashboard();
    } else {
      void this.validateAndSetKey();
    }
  }

  private initializeValidationProcess(): void {
    this.isProcessing = true;
    this.error = false;
    this.isValidated = false;
    this.encryptionKeyForm.get('encryptionKey')?.disable();
  }

  private handleSuccessfulSetup(): void {
    this.isValidated = true;
    this.isProcessing = false;
    this.encryptionKeyForm.reset({encryptionKey: ''});
    this.statusMessage = 'Done! Redirecting...';
    this.startCountdown();
  }

  private handleError(errorMessage: string): void {
    this.isProcessing = false;
    this.error = true;
    this.statusMessage = errorMessage;
    this.encryptionKeyForm.controls.encryptionKey.enable();
    this.cdRef.detectChanges();
  }

  async validateAndSetKey(): Promise<void> {
    if (!this.active || this.isProcessing || this.isValidated || !this.encryptionKeyForm.valid) return;
    const encryptionKey = this.encryptionKeyForm.controls.encryptionKey.value;
    const epoch = ++this.operationEpoch;
    this.initializeValidationProcess();
    let validating = true;
    try {
      this.statusMessage = 'Validating key...';
      this.cdRef.detectChanges();
      await this.sleep(500);
      if (!this.current(epoch)) return;
      await firstValueFrom(this.fastenService.validateEncryptionKey(encryptionKey).pipe(takeUntil(this.destroyed)));
      if (!this.current(epoch)) return;
      validating = false;
      this.statusMessage = 'Setting up key...';
      this.cdRef.detectChanges();
      await this.sleep(500);
      if (!this.current(epoch)) return;
      await firstValueFrom(this.fastenService.setupEncryptionKey(encryptionKey).pipe(takeUntil(this.destroyed)));
      if (!this.current(epoch)) return;
      this.statusMessage = 'Confirming local database is unlocked...';
      this.cdRef.detectChanges();
      await firstValueFrom(this.fastenService.confirmEncryptionReady().pipe(takeUntil(this.destroyed)));
      if (this.current(epoch)) this.handleSuccessfulSetup();
    } catch {
      if (!this.current(epoch)) return;
      this.handleError(validating ? 'Encryption key validation could not be confirmed. Check the key and server before retrying.' :
        'Encryption setup was not confirmed. It may already have changed; keep your saved key and check the server before retrying.');
    }
  }

  startCountdown(): void {
    if (!this.active || !this.isValidated) return;
    this.countdown = 5;
    if (this.countdownInterval !== null) clearInterval(this.countdownInterval);
    this.countdownInterval = setInterval(() => {
      this.countdown--;
      if (this.countdown <= 0) {
        this.redirectToDashboard();
      }
    }, 1000);
  }

  redirectToDashboard(): void {
    if (!this.active || !this.isValidated) return;
    if (this.countdownInterval !== null) clearInterval(this.countdownInterval);
    this.countdownInterval = null;
    void this.router.navigate(['/']).catch(() => {
      if (this.active) this.statusMessage = 'Navigation could not be completed.';
    });
  }
}
