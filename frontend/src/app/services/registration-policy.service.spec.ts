import {TestBed} from '@angular/core/testing';
import {HttpClient,provideHttpClient} from '@angular/common/http';
import {HttpTestingController,provideHttpClientTesting} from '@angular/common/http/testing';
import {HTTP_CLIENT_TOKEN} from '../dependency-injection';
import {AuthService} from './auth.service';
import {User} from '../models/fasten/user';

describe('Registration policy API contract',()=>{
  let service:AuthService,http:HttpTestingController,previous:string|null;
  beforeEach(()=>{
    TestBed.configureTestingModule({providers:[provideHttpClient(),provideHttpClientTesting(),{provide:HTTP_CLIENT_TOKEN,useExisting:HttpClient}]});
    service=TestBed.inject(AuthService);http=TestBed.inject(HttpTestingController);previous=localStorage.getItem(service.FASTEN_JWT_LOCALSTORAGE_KEY);
  });
  afterEach(()=>{http.verify();if(previous===null)localStorage.removeItem(service.FASTEN_JWT_LOCALSTORAGE_KEY);else localStorage.setItem(service.FASTEN_JWT_LOCALSTORAGE_KEY,previous);});
  it('reads only safe availability metadata and rejects malformed responses',async()=>{
    const pending=service.RegistrationStatus();const request=http.expectOne(r=>r.url.endsWith('/auth/registration'));expect(request.request.method).toBe('GET');
    request.flush({success:true,data:{enabled:false,allowed:false,bootstrap:false}});expect((await pending).allowed).toBeFalse();
    const invalid=service.RegistrationStatus();http.expectOne(r=>r.url.endsWith('/auth/registration')).flush({success:true,data:{allowed:'true'}});await expectAsync(invalid).toBeRejected();
  });
  it('never installs a token after an unconfirmed or policy-denied signup',async()=>{
    localStorage.setItem(service.FASTEN_JWT_LOCALSTORAGE_KEY,'synthetic-existing-session');
    const user=new User();user.username='synthetic-user';user.password='synthetic-password';
    const unconfirmed=service.Signup(user);http.expectOne(r=>r.url.endsWith('/auth/signup')).flush({success:false,data:'untrusted'});await expectAsync(unconfirmed).toBeRejected();
    const denied=service.Signup(user);http.expectOne(r=>r.url.endsWith('/auth/signup')).flush({success:false,code:'registration_closed'},{status:403,statusText:'Forbidden'});await expectAsync(denied).toBeRejected();
    expect(service.GetAuthToken()).toBe('synthetic-existing-session');
  });
});
