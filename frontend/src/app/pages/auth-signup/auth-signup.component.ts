import { Component, OnInit, ChangeDetectionStrategy } from '@angular/core';
import {User} from '../../models/fasten/user';
import {Router} from '@angular/router';
import {ToastNotification, ToastType} from '../../models/fasten/toast';
import {ToastService} from '../../services/toast.service';
import {AuthService} from '../../services/auth.service';

@Component({
    selector: 'app-auth-signup',
    templateUrl: './auth-signup.component.html',
    styleUrls: ['./auth-signup.component.scss'],
    changeDetection: ChangeDetectionStrategy.Eager,
    standalone: false
})
export class AuthSignupComponent implements OnInit {
  loading: boolean = false

  passwordConfirmation: string = ""

  submitted: boolean = false
  newUser: User = new User()
  errorMsg: string = ""

  constructor(
    private authService: AuthService,
    private router: Router,
    private toastService: ToastService
  ) { }

  ngOnInit(): void {
  }

  signupSubmit(){
    if (this.loading) return;
    this.submitted = true;
    if (!this.passwordConfirmation || this.passwordConfirmation !== this.newUser.password) {
      this.errorMsg = "Passwords do not match.";
      return;
    }
    this.errorMsg = "";
    this.loading = true;

    this.authService.Signup(this.newUser).then((tokenResp: any) => {
        this.loading = false
        this.router.navigateByUrl('/dashboard');
      },
      (err)=>{
        this.loading = false
        console.error("an error occured while signup",err)
        if(err.name === 'conflict') {
          // "batman" already exists, choose another username
          this.errorMsg = "username already exists"
        } else if (err.name === 'forbidden') {
          // invalid username
          this.errorMsg = "invalid username"
        } else {
          this.errorMsg = "an unknown error occurred during sign-up"
        }

        const toastNotificaiton = new ToastNotification()
        toastNotificaiton.type = ToastType.Error
        toastNotificaiton.message = this.errorMsg
        this.toastService.show(toastNotificaiton)
    })
  }

}
