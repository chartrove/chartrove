import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, provideRouter } from '@angular/router';
import { OIDCComponent } from './oidc.component';
import { OIDCService } from '../../services/oidc.service';
import { AuthService } from '../../services/auth.service';

describe('OIDC settings and recovery', () => {
  let fixture: ComponentFixture<OIDCComponent>;
  let api: jasmine.SpyObj<OIDCService>;
  beforeEach(async () => {
    api = jasmine.createSpyObj('OIDCService', ['enabled', 'identities', 'audit', 'start', 'revoke']);
    api.enabled.and.resolveTo(true); api.identities.and.resolveTo([{id: 'synthetic-identity', issuer: 'https://idp.example', revoked_at: null, created_at: '2026-11-04T00:00:00Z'}]); api.audit.and.resolveTo([]);
    await TestBed.configureTestingModule({imports: [OIDCComponent], providers: [provideRouter([]), {provide: OIDCService, useValue: api}, {provide: AuthService, useValue: {IsAuthenticated: () => Promise.resolve(true)}}, {provide: ActivatedRoute, useValue: {snapshot: {routeConfig: {path: 'auth/oidc'}}}}]}).compileComponents();
    fixture = TestBed.createComponent(OIDCComponent); fixture.detectChanges(); await fixture.whenStable(); fixture.detectChanges();
  });
  it('requires local password proof and clears it after a failed link', async () => {
    const view = fixture.nativeElement as HTMLElement;
    expect(view.textContent).toContain('Local password sign-in remains available');
    expect(view.querySelector<HTMLButtonElement>('button[type="submit"]')?.disabled).toBeTrue();
    fixture.componentInstance.password = 'synthetic-password'; api.start.and.rejectWith(new Error('unavailable'));
    await fixture.componentInstance.link(); fixture.detectChanges();
    expect(api.start).toHaveBeenCalledWith('synthetic-password'); expect(fixture.componentInstance.password).toBe('');
    expect(view.textContent).toContain('use local sign-in');
  });
  it('shows explicit revocation and does not claim success when it fails', async () => {
    fixture.componentInstance.password = 'synthetic-password'; api.revoke.and.rejectWith(new Error('denied'));
    await fixture.componentInstance.revoke('synthetic-identity'); fixture.detectChanges();
    expect(api.revoke).toHaveBeenCalledWith('synthetic-identity', 'synthetic-password');
    expect(fixture.nativeElement.textContent).toContain('Revocation was not confirmed');
    expect(fixture.componentInstance.identities.length).toBe(1);
  });
});
