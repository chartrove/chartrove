import type { Type } from '@angular/core';
import type { User } from '../models/fasten/user';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../services/auth.service';
import { AuthSignupComponent } from './auth-signup/auth-signup.component';
import { AuthSignupWizardComponent } from './auth-signup-wizard/auth-signup-wizard.component';

type SignupComponent = AuthSignupComponent | AuthSignupWizardComponent;
const components: Type<SignupComponent>[] = [AuthSignupComponent, AuthSignupWizardComponent];

for (const componentType of components) {
  describe(`${componentType.name} password confirmation`, () => {
    let fixture: ComponentFixture<SignupComponent>;
    let signup: jasmine.Spy;

    beforeEach(async () => {
      signup = jasmine.createSpy('Signup').and.returnValue(new Promise(() => {}));
      await TestBed.configureTestingModule({
        declarations: [componentType],
        imports: [FormsModule],
        providers: [
          { provide: AuthService, useValue: { Signup: signup, RegistrationStatus: () => Promise.resolve({enabled:true,allowed:true,bootstrap:false}) } },
          { provide: Router, useValue: { navigateByUrl: jasmine.createSpy() } },
        ],
      }).compileComponents();
      fixture = TestBed.createComponent<SignupComponent>(componentType);
      fixture.componentInstance.newUser.full_name = 'Synthetic User';
      fixture.componentInstance.newUser.username = 'synthetic';
      fixture.componentInstance.newUser.email = 'synthetic@example.invalid';
      if (fixture.componentInstance instanceof AuthSignupWizardComponent) {
        fixture.componentInstance.newUser.agree_terms = true;
      }
      fixture.componentInstance.newUser.password = 'synthetic-password';
      fixture.detectChanges();
      await fixture.whenStable();
      fixture.detectChanges();
    });

    it('renders a required masked confirmation field', () => {
      const root: HTMLElement = fixture.nativeElement;
      const input = root.querySelector<HTMLInputElement>('input[name="passwordConfirmation"]');
      expect(input).not.toBeNull();
      if (input) {
        expect(input.type).toBe('password');
        expect(input.required).toBeTrue();
        expect(input.autocomplete).toBe('new-password');
      }
    });

    for (const confirmation of ['', 'different-password']) {
      it(`blocks form submission with ${confirmation ? 'mismatched' : 'missing'} confirmation`, async () => {
        fixture.componentInstance.passwordConfirmation = confirmation;
        fixture.detectChanges();
        await fixture.whenStable();
        fixture.detectChanges();
        expect(fixture.nativeElement.querySelector('button[type="submit"]').disabled).toBeTrue();
        fixture.nativeElement.querySelector('form').dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
        expect(signup).not.toHaveBeenCalled();
        expect(fixture.componentInstance.loading).toBeFalse();
      });
    }

    it('sends the existing user payload without the confirmation', async () => {
      fixture.componentInstance.passwordConfirmation = 'synthetic-password';
      fixture.detectChanges();
      await fixture.whenStable();
      fixture.detectChanges();
      expect(fixture.nativeElement.querySelector('button[type="submit"]').disabled).toBeFalse();
      fixture.nativeElement.querySelector('form').dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
      expect(signup).toHaveBeenCalledTimes(1);
      const payload: User = signup.calls.mostRecent().args[0];
      expect(payload.password).toBe('synthetic-password');
      expect(Object.keys(payload)).not.toContain('passwordConfirmation');
      expect(Object.keys(payload)).not.toContain('password_confirm');
    });

    it('does not create a second account request while signup is pending', () => {
      fixture.componentInstance.passwordConfirmation = 'synthetic-password';
      fixture.componentInstance.signupSubmit();
      fixture.componentInstance.signupSubmit();
      expect(signup).toHaveBeenCalledTimes(1);
    });
  });
}
