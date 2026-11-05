import {ComponentFixture, TestBed} from '@angular/core/testing';
import {FormsModule} from '@angular/forms';
import {provideRouter,RouterModule} from '@angular/router';
import {AuthSignupComponent} from './auth-signup/auth-signup.component';
import {AuthSignupWizardComponent} from './auth-signup-wizard/auth-signup-wizard.component';
import {AuthService,RegistrationPolicy} from '../services/auth.service';

for (const componentType of [AuthSignupComponent,AuthSignupWizardComponent]) {
  describe(`${componentType.name} registration policy`,()=>{
    let fixture:ComponentFixture<AuthSignupComponent|AuthSignupWizardComponent>;
    let auth:jasmine.SpyObj<AuthService>;
    beforeEach(async()=>{
      auth=jasmine.createSpyObj('AuthService',['Signup','RegistrationStatus']);
      auth.RegistrationStatus.and.resolveTo({enabled:true,allowed:true,bootstrap:false});
      auth.Signup.and.returnValue(new Promise(()=>{}));
      await TestBed.configureTestingModule({declarations:[componentType],imports:[FormsModule,RouterModule],providers:[provideRouter([]),{provide:AuthService,useValue:auth}]}).compileComponents();
      fixture=TestBed.createComponent(componentType as typeof AuthSignupComponent);
    });
    async function render(policy:RegistrationPolicy):Promise<void>{auth.RegistrationStatus.and.resolveTo(policy);fixture.detectChanges();await fixture.whenStable();fixture.detectChanges();}
    it('shows an initialized closed policy without a credential form, and preserves sign-in navigation',async()=>{
      await render({enabled:false,allowed:false,bootstrap:false});
      expect(fixture.nativeElement.querySelector('form')).toBeNull();
      expect(fixture.nativeElement.textContent).toContain('Public registration is disabled');
      expect(fixture.nativeElement.querySelector('a[routerLink="/auth/signin"]')).not.toBeNull();
      expect(fixture.nativeElement.querySelector('a[routerLink="/auth/signin"]').getAttribute('href')).toContain('/auth/signin');
      fixture.componentInstance.signupSubmit();expect(auth.Signup).not.toHaveBeenCalled();
    });
    it('allows first-administrator bootstrap even when subsequent public registration is disabled',async()=>{
      await render({enabled:false,allowed:true,bootstrap:true});
      expect(fixture.nativeElement.querySelector('form')).not.toBeNull();
      const c=fixture.componentInstance;c.newUser.password='synthetic-password';c.passwordConfirmation='synthetic-password';c.signupSubmit();expect(auth.Signup).toHaveBeenCalledTimes(1);
    });
    it('fails closed while availability is pending or unavailable',async()=>{
      let reject:(reason:any)=>void;auth.RegistrationStatus.and.returnValue(new Promise((_resolve,fail)=>reject=fail));
      fixture.detectChanges();expect(fixture.nativeElement.querySelector('form')).toBeNull();
      fixture.componentInstance.signupSubmit();expect(auth.Signup).not.toHaveBeenCalled();
      reject(new Error('synthetic outage'));await fixture.whenStable();fixture.detectChanges();
      expect(fixture.nativeElement.querySelector('form')).toBeNull();expect(fixture.nativeElement.textContent).toContain('could not be checked');
    });
    it('handles a server policy change after loading and clears entered passwords',async()=>{
      await render({enabled:true,allowed:true,bootstrap:false});
      auth.Signup.and.rejectWith({status:403,error:{code:'registration_closed'}});
      const c=fixture.componentInstance;c.newUser.password='synthetic-password';c.passwordConfirmation='synthetic-password';c.signupSubmit();
      await fixture.whenStable();fixture.detectChanges();
      expect(c.registration.allowed).toBeFalse();expect(c.newUser.password).toBe('');expect(c.passwordConfirmation).toBe('');expect(c.loading).toBeFalse();
      expect(fixture.nativeElement.querySelector('form')).toBeNull();
    });
  });
}
