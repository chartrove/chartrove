import { Inject, Injectable, InjectionToken } from '@angular/core';
import { HttpErrorResponse, HttpEvent, HttpHandler, HttpInterceptor, HttpRequest } from '@angular/common/http';
import {Router} from '@angular/router';
import {from, Observable, of, throwError} from 'rxjs';
import {catchError, mergeMap} from 'rxjs/operators';
import {AuthService} from './auth.service';
import {GetEndpointAbsolutePath} from '../../lib/utils/endpoint_absolute_path';
import {environment} from '../../environments/environment';
import {prepareNativeMultipart} from '../../assets/desktop/multipart.mjs';

export interface ApplicationHttpLocation {
  readonly protocol: string;
  readonly host: string;
  readonly pathname: string;
}

export const APPLICATION_HTTP_LOCATION = new InjectionToken<ApplicationHttpLocation>('APPLICATION_HTTP_LOCATION', {
  providedIn: 'root', factory: () => globalThis.location,
});

@Injectable({
  providedIn: 'root'
})

// based on https://stackoverflow.com/questions/46017245/how-to-handle-unauthorized-requestsstatus-with-401-or-403-with-new-httpclient
export class AuthInterceptorService implements HttpInterceptor {

  constructor(
    private authService: AuthService,
    private router: Router,
    @Inject(APPLICATION_HTTP_LOCATION) private readonly requestLocation: ApplicationHttpLocation = globalThis.location,
  ) { }

  private handleAuthError(err: HttpErrorResponse): Observable<any> {
    //handle your auth error or rethrow
    if (err.status === 401 || err.status === 403) {
      //navigate /delete cookies or whatever
      this.authService.Logout(false)
      this.router.navigateByUrl(`/auth/signin`);
      // if you've caught / handled the error, you don't want to rethrow it unless you also want downstream consumers to have to handle it as well.
      return of(err.message); // or EMPTY may be appropriate here
    }
    return throwError(err);
  }

  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {

    // Native custom schemes serialize origin as "null". Resolve against the actual
    // browser authority, then compare scheme/host rather than conflating opaque origins.
    const browserBase = `${this.requestLocation.protocol}//${this.requestLocation.host}/`;
    const reqUrl = new URL(req.url, browserBase);
    const apiUrl = new URL(GetEndpointAbsolutePath(this.requestLocation, environment.fasten_api_endpoint_base));
    const apiPath = apiUrl.pathname.replace(/\/+$/, '');
    const withinApiPath = reqUrl.pathname === apiPath || reqUrl.pathname.startsWith(apiPath + '/');

    // Optional Lighthouse/provider services are not local session authorities.
    // Preserve their requests/explicit credentials and errors without reading a native JWT.
    if (!apiUrl.host || reqUrl.protocol !== apiUrl.protocol || reqUrl.host !== apiUrl.host ||
        reqUrl.username !== '' || reqUrl.password !== '' || !withinApiPath) {
      return next.handle(req);
    }

    const authReq = req.clone({headers: req.headers.set('Authorization', 'Bearer ' + this.authService.GetAuthToken())});
    const send = (request: HttpRequest<unknown>) => next.handle(request).pipe(catchError(x => {
      // A public signup policy denial is not invalidation of an existing
      // native session. All protected-route 401/403 handling remains intact.
      if (reqUrl.pathname === apiPath + '/auth/signup' &&
          x.status === 403 && x.error?.code === 'registration_closed') return throwError(() => x);
      return this.handleAuthError(x);
    }));
    if (reqUrl.protocol === 'wails:' && req.body instanceof FormData) {
      // Pin the already selected token before asynchronous encoding. The server
      // still checks expected_owner_id; do not reselect a replacement actor here.
      return from(prepareNativeMultipart(req.body)).pipe(mergeMap(prepared => send(authReq.clone({
        body: prepared.body, setHeaders: {'Content-Type': prepared.contentType},
      }))));
    }
    return send(authReq);

  }
}
