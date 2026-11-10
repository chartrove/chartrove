import {ComponentFixture, TestBed, fakeAsync, flushMicrotasks, tick} from '@angular/core/testing';
import {Router} from '@angular/router';
import {By} from '@angular/platform-browser';
import {Subject, of, throwError} from 'rxjs';
import {AppModule} from './app.module';
import {GetEncryptionKeyWizardComponent} from './pages/get-encryption-key-wizard/get-encryption-key-wizard.component';
import {SetupEncryptionKeyComponent} from './pages/setup-encryption-key/setup-encryption-key.component';
import {FastenApiService} from './services/fasten-api.service';
import {ToastService} from './services/toast.service';

describe('Encryption setup acknowledgment, confirmation and private lifecycle', () => {
  const key = 'synthetic-exact-key-for-owned-ui';
  const privateMarker = 'synthetic-private-encryption-ui-canary';
  let confirmation: Subject<void>;
  let api: {
    getEncryptionKey: jasmine.Spy;
    validateEncryptionKey: jasmine.Spy;
    setupEncryptionKey: jasmine.Spy;
    confirmEncryptionReady: jasmine.Spy;
  };
  let router: jasmine.SpyObj<Router>;
  let toast: jasmine.SpyObj<ToastService>;

  beforeEach(async () => {
    confirmation = new Subject<void>();
    api = {
      getEncryptionKey: jasmine.createSpy('getEncryptionKey').and.returnValue(of(key)),
      validateEncryptionKey: jasmine.createSpy('validateEncryptionKey').and.returnValue(of(undefined)),
      setupEncryptionKey: jasmine.createSpy('setupEncryptionKey').and.returnValue(of(undefined)),
      confirmEncryptionReady: jasmine.createSpy('confirmEncryptionReady').and.returnValue(confirmation.asObservable()),
    };
    router = jasmine.createSpyObj<Router>('Router', ['navigate', 'navigateByUrl']);
    router.navigate.and.returnValue(Promise.resolve(true));
    router.navigateByUrl.and.returnValue(Promise.resolve(true));
    toast = jasmine.createSpyObj<ToastService>('ToastService', ['show']);
    await TestBed.configureTestingModule({
      imports: [AppModule],
      providers: [{provide: FastenApiService, useValue: api}, {provide: Router, useValue: router}, {provide: ToastService, useValue: toast}],
    }).compileComponents();
  });

  function restore(): ComponentFixture<SetupEncryptionKeyComponent> {
    const fixture = TestBed.createComponent(SetupEncryptionKeyComponent);
    fixture.detectChanges();
    fixture.componentInstance.encryptionKeyForm.setValue({encryptionKey: key});
    return fixture;
  }

  function generate(): ComponentFixture<GetEncryptionKeyWizardComponent> {
    const fixture = TestBed.createComponent(GetEncryptionKeyWizardComponent);
    fixture.detectChanges();
    fixture.componentInstance.encryptionKeyDownloaded = true;
    return fixture;
  }

  it('does not declare restored setup successful until independent unlocked confirmation', fakeAsync(() => {
    const fixture = restore();
    const component = fixture.componentInstance;
    void component.validateAndSetKey();
    tick(1000);
    expect(api.validateEncryptionKey).toHaveBeenCalledOnceWith(key);
    expect(api.setupEncryptionKey).toHaveBeenCalledOnceWith(key);
    expect(api.confirmEncryptionReady).toHaveBeenCalledTimes(1);
    expect(component.isValidated).toBeFalse();
    expect(router.navigate).not.toHaveBeenCalled();
    confirmation.next(undefined);
    confirmation.complete();
    flushMicrotasks();
    expect(component.isValidated).toBeTrue();
    fixture.destroy();
  }));

  it('does not announce or count down generated-key success on acknowledgment alone', fakeAsync(() => {
    const fixture = generate();
    const component = fixture.componentInstance;
    const countdown = spyOn(component, 'startCountdown');
    void component.proceedToSignup();
    flushMicrotasks();
    expect(api.confirmEncryptionReady).toHaveBeenCalledTimes(1);
    expect(countdown).not.toHaveBeenCalled();
    expect(toast.show).not.toHaveBeenCalled();
    confirmation.next(undefined);
    confirmation.complete();
    flushMicrotasks();
    expect(countdown).toHaveBeenCalledTimes(1);
    fixture.destroy();
  }));

  it('discards a generated key arriving after its view was destroyed', () => {
    const pending = new Subject<string>();
    api.getEncryptionKey.and.returnValue(pending.asObservable());
    const fixture = generate();
    fixture.destroy();
    pending.next(key);
    pending.complete();
    expect(fixture.componentInstance.encryptionKey).toBeNull();
  });

  it('clears the restored private key when its view is destroyed', () => {
    const fixture = restore();
    fixture.destroy();
    expect(fixture.componentInstance.encryptionKeyForm.get('encryptionKey')?.value).toBe('');
  });

  it('admits only one restore mutation while validation is pending', fakeAsync(() => {
    const pending = new Subject<void>();
    api.validateEncryptionKey.and.returnValue(pending.asObservable());
    const fixture = restore();
    void fixture.componentInstance.validateAndSetKey();
    void fixture.componentInstance.validateAndSetKey();
    tick(500);
    expect(api.validateEncryptionKey).toHaveBeenCalledTimes(1);
    fixture.destroy();
  }));

  it('admits only one generated-key setup mutation while its request is pending', fakeAsync(() => {
    const pending = new Subject<void>();
    api.setupEncryptionKey.and.returnValue(pending.asObservable());
    const fixture = generate();
    void fixture.componentInstance.proceedToSignup();
    void fixture.componentInstance.proceedToSignup();
    flushMicrotasks();
    expect(api.setupEncryptionKey).toHaveBeenCalledTimes(1);
    fixture.destroy();
    expect(fixture.componentInstance.encryptionKey).toBeNull();
  }));

  it('does not log arbitrary key-generation errors', () => {
    api.getEncryptionKey.and.returnValue(throwError(() => new Error(privateMarker)));
    const log = spyOn(console, 'error');
    const fixture = generate();
    expect(log).not.toHaveBeenCalled();
    expect(fixture.componentInstance.error).not.toContain(privateMarker);
    fixture.destroy();
  });

  it('keeps a restored private key masked in the actual form', () => {
    const fixture = restore();
    const input: unknown = fixture.debugElement.query(By.css('input[formControlName="encryptionKey"]')).nativeElement;
    expect(input instanceof HTMLInputElement && input.type === 'password').toBeTrue();
    fixture.destroy();
  });

  it('cancels owned validation on destruction without sending setup or exposing its captured key', fakeAsync(() => {
    const pending = new Subject<void>();
    api.validateEncryptionKey.and.returnValue(pending.asObservable());
    const fixture = restore();
    void fixture.componentInstance.validateAndSetKey();
    tick(500);
    expect(pending.observed).toBeTrue();
    fixture.destroy();
    expect(pending.observed).toBeFalse();
    pending.next(undefined);
    pending.complete();
    tick(1000);
    expect(api.setupEncryptionKey).not.toHaveBeenCalled();
    expect(router.navigate).not.toHaveBeenCalled();
    expect(fixture.componentInstance.encryptionKeyForm.controls.encryptionKey.value).toBe('');
  }));

  it('does not treat cancellation of an already-sent setup as completion or send health after destruction', fakeAsync(() => {
    const pending = new Subject<void>();
    api.setupEncryptionKey.and.returnValue(pending.asObservable());
    const fixture = generate();
    void fixture.componentInstance.proceedToSignup();
    flushMicrotasks();
    expect(api.setupEncryptionKey).toHaveBeenCalledOnceWith(key);
    fixture.destroy();
    expect(pending.observed).toBeFalse();
    pending.next(undefined);
    pending.complete();
    flushMicrotasks();
    expect(api.confirmEncryptionReady).not.toHaveBeenCalled();
    expect(router.navigate).not.toHaveBeenCalled();
    expect(toast.show).not.toHaveBeenCalled();
  }));

  it('cancels confirmation and ignores its late result after destruction', fakeAsync(() => {
    const fixture = generate();
    void fixture.componentInstance.proceedToSignup();
    flushMicrotasks();
    expect(confirmation.observed).toBeTrue();
    fixture.destroy();
    expect(confirmation.observed).toBeFalse();
    confirmation.next(undefined);
    confirmation.complete();
    tick(6000);
    expect(router.navigate).not.toHaveBeenCalled();
    expect(toast.show).not.toHaveBeenCalled();
    expect(fixture.componentInstance.encryptionKey).toBeNull();
  }));

  it('discards superseded generation and resets the old key acknowledgments', () => {
    const first = new Subject<string>();
    const second = new Subject<string>();
    api.getEncryptionKey.and.returnValues(first.asObservable(), second.asObservable());
    const fixture = generate();
    fixture.componentInstance.acknowledged = true;
    fixture.componentInstance.fetchEncryptionKey();
    expect(first.observed).toBeFalse();
    first.next('synthetic-stale-key');
    second.next(key);
    expect(fixture.componentInstance.encryptionKey).toBe(key);
    expect(fixture.componentInstance.encryptionKeyDownloaded).toBeFalse();
    expect(fixture.componentInstance.acknowledged).toBeFalse();
    fixture.destroy();
  });

  it('sets only the exact captured validated key even if form state changes while pending', fakeAsync(() => {
    const pending = new Subject<void>();
    api.validateEncryptionKey.and.returnValue(pending.asObservable());
    const fixture = restore();
    void fixture.componentInstance.validateAndSetKey();
    tick(500);
    fixture.componentInstance.encryptionKeyForm.controls.encryptionKey.setValue('synthetic-later-key');
    pending.next(undefined);
    pending.complete();
    tick(500);
    expect(api.setupEncryptionKey).toHaveBeenCalledOnceWith(key);
    confirmation.next(undefined);
    confirmation.complete();
    flushMicrotasks();
    fixture.destroy();
  }));

  it('retains explicit sent-setup uncertainty rather than success or a false invalid-key claim', fakeAsync(() => {
    api.confirmEncryptionReady.and.returnValue(throwError(() => new Error(privateMarker)));
    const fixture = restore();
    void fixture.componentInstance.validateAndSetKey();
    tick(1000);
    expect(fixture.componentInstance.isValidated).toBeFalse();
    expect(fixture.componentInstance.statusMessage).toContain('may already have changed');
    expect(fixture.componentInstance.statusMessage).not.toContain(privateMarker);
    expect(fixture.componentInstance.encryptionKeyForm.controls.encryptionKey.errors).toBeNull();
    expect(api.setupEncryptionKey).toHaveBeenCalledTimes(1);
    expect(router.navigate).not.toHaveBeenCalled();
    fixture.destroy();
  }));

  it('does not announce a late clipboard completion after destruction or undo the sent copy', fakeAsync(() => {
    let completeCopy: (() => void) | undefined;
    const pending = new Promise<void>(resolve => completeCopy = resolve);
    const clipboard = spyOn(navigator.clipboard, 'writeText').and.returnValue(pending);
    const fixture = generate();
    void fixture.componentInstance.copyToClipboard();
    expect(clipboard).toHaveBeenCalledOnceWith(key);
    fixture.destroy();
    completeCopy?.();
    flushMicrotasks();
    expect(toast.show).not.toHaveBeenCalled();
    expect(fixture.componentInstance.encryptionKey).toBeNull();
  }));

  it('requests an exact-key file download without treating a request as saved-file attestation', async () => {
    const captured: Blob[] = [];
    spyOn(URL, 'createObjectURL').and.callFake(value => {
      if (value instanceof Blob) captured.push(value);
      return 'blob:synthetic-key-only';
    });
    const revoke = spyOn(URL, 'revokeObjectURL');
    const click = spyOn(HTMLAnchorElement.prototype, 'click');
    const fixture = generate();
    fixture.componentInstance.downloadEncryptionKey();
    expect(click).toHaveBeenCalledTimes(1);
    expect(captured.length).toBe(1);
    const blob = captured.at(0);
    expect(blob ? await blob.text() : '').toBe(key);
    expect(revoke).toHaveBeenCalledOnceWith('blob:synthetic-key-only');
    expect(toast.show.calls.mostRecent().args[0].message).toContain('Download requested');
    fixture.destroy();
  });

  it('enables actual signup navigation only after confirmed readiness and exposes uncertainty warnings', fakeAsync(() => {
    const fixture = generate();
    const component = fixture.componentInstance;
    void component.proceedToSignup();
    flushMicrotasks();
    expect(component.encryptionReady).toBeFalse();
    confirmation.next(undefined);
    confirmation.complete();
    flushMicrotasks();
    fixture.detectChanges();
    expect(component.encryptionReady).toBeTrue();
    const button: unknown = fixture.debugElement.query(By.css('button.btn-primary')).nativeElement;
    expect(button instanceof HTMLButtonElement && !button.disabled).toBeTrue();
    if (button instanceof HTMLButtonElement) button.click();
    flushMicrotasks();
    expect(router.navigate).toHaveBeenCalledOnceWith(['/auth/signup/wizard']);
    expect(fixture.nativeElement.textContent).toContain('does not roll back setup');
    fixture.destroy();
  }));

  it('keeps sent generated-key setup uncertainty on failed confirmation without retry or success', fakeAsync(() => {
    const fixture = generate();
    void fixture.componentInstance.proceedToSignup();
    flushMicrotasks();
    confirmation.error(new Error(privateMarker));
    flushMicrotasks();
    expect(fixture.componentInstance.encryptionReady).toBeFalse();
    expect(fixture.componentInstance.error).toContain('may already have changed');
    expect(fixture.componentInstance.error).not.toContain(privateMarker);
    expect(api.setupEncryptionKey).toHaveBeenCalledTimes(1);
    expect(router.navigate).not.toHaveBeenCalled();
    expect(toast.show.calls.mostRecent().args[0].message).not.toContain(privateMarker);
    fixture.destroy();
  }));

  it('handles private clipboard rejection without raw logging or disclosure', fakeAsync(() => {
    spyOn(navigator.clipboard, 'writeText').and.returnValue(Promise.reject(new Error(privateMarker)));
    const log = spyOn(console, 'error');
    const fixture = generate();
    void fixture.componentInstance.copyToClipboard();
    flushMicrotasks();
    expect(log).not.toHaveBeenCalled();
    expect(toast.show.calls.mostRecent().args[0].message).not.toContain(privateMarker);
    fixture.destroy();
  }));

  it('releases a failed download URL without claiming the file was saved', () => {
    spyOn(URL, 'createObjectURL').and.returnValue('blob:synthetic-failed-download');
    const revoke = spyOn(URL, 'revokeObjectURL');
    spyOn(HTMLAnchorElement.prototype, 'click').and.throwError(privateMarker);
    const fixture = generate();
    fixture.componentInstance.encryptionKeyDownloaded = false;
    fixture.componentInstance.downloadEncryptionKey();
    expect(fixture.componentInstance.encryptionKeyDownloaded).toBeFalse();
    expect(revoke).toHaveBeenCalledOnceWith('blob:synthetic-failed-download');
    expect(toast.show.calls.mostRecent().args[0].message).not.toContain(privateMarker);
    fixture.destroy();
  });

  it('does not log arbitrary setup errors or present their private text', fakeAsync(() => {
    api.setupEncryptionKey.and.returnValue(throwError(() => new Error(privateMarker)));
    const log = spyOn(console, 'error');
    const fixture = generate();
    void fixture.componentInstance.proceedToSignup();
    flushMicrotasks();
    expect(log).not.toHaveBeenCalled();
    expect(fixture.componentInstance.error).not.toContain(privateMarker);
    fixture.destroy();
  }));
});
