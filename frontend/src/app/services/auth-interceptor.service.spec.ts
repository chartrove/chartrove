import { HttpErrorResponse, HttpHandler, HttpRequest, HttpResponse } from '@angular/common/http';
import {Router} from '@angular/router';
import {of,throwError} from 'rxjs';
import {AuthInterceptorService} from './auth-interceptor.service';
import {AuthService} from './auth.service';

describe('AuthInterceptorService Patient privacy', () => {
  for(const path of ['/api/auth/signup','/api/secure/users']) {
    it(`keeps session invalidation limited to authentication failures on ${path}`,()=>{
      const auth=jasmine.createSpyObj<AuthService>('AuthService',['GetAuthToken','Logout']);auth.GetAuthToken.and.returnValue('synthetic-existing-session');
      const router=jasmine.createSpyObj<Router>('Router',['navigateByUrl']);const handler=jasmine.createSpyObj<HttpHandler>('HttpHandler',['handle']);
      handler.handle.and.returnValue(throwError(()=>new HttpErrorResponse({status:403,error:{code:'registration_closed'}})));
      let rejected=false;new AuthInterceptorService(auth,router).intercept(new HttpRequest('POST',path,{}),handler).subscribe({error:()=>rejected=true});
      if(path==='/api/auth/signup') {expect(rejected).toBeTrue();expect(auth.Logout).not.toHaveBeenCalled();expect(router.navigateByUrl).not.toHaveBeenCalled();}
      else {expect(auth.Logout).toHaveBeenCalledWith(false);expect(router.navigateByUrl).toHaveBeenCalledWith('/auth/signin');}
    });
  }
  it('does not log a serialized Patient import or its authorization header', () => {
    const log = spyOn(console, 'log');
    const auth = jasmine.createSpyObj<AuthService>('AuthService', ['GetAuthToken']);
    auth.GetAuthToken.and.returnValue('synthetic-test-token');
    const handler = jasmine.createSpyObj<HttpHandler>('HttpHandler', ['handle']);
    handler.handle.and.returnValue(of(new HttpResponse({status: 200})));
    const interceptor = new AuthInterceptorService(auth, jasmine.createSpyObj<Router>('Router', ['navigateByUrl']));
    const body = '--synthetic-boundary\r\n{"resourceType":"Patient","id":"synthetic-only","name":[{"text":"Synthetic Patient"}]}';
    interceptor.intercept(new HttpRequest('POST', '/api/secure/source/manual', body), handler).subscribe();
    expect(handler.handle).toHaveBeenCalledTimes(1);
    expect(log).not.toHaveBeenCalled();
  });
});
