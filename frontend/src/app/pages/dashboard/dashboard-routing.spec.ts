import { CommonModule } from '@angular/common';
import { Component, isStandalone } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { FormsModule } from '@angular/forms';
import { RouteConfigLoadEnd, RouteConfigLoadStart, Router } from '@angular/router';
import { RouterTestingHarness } from '@angular/router/testing';
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { of } from 'rxjs';
import { AppRoutingModule } from '../../app-routing.module';
import { IsAuthenticatedAuthGuard } from '../../auth-guards/is-authenticated-auth-guard';
import { GridstackComponent } from '../../components/gridstack/gridstack.component';
import { DirectivesModule } from '../../directives/directives.module';
import { AuthService } from '../../services/auth.service';
import { FastenApiService } from '../../services/fasten-api.service';
import { LighthouseService } from '../../services/lighthouse.service';
import { DashboardComponent } from './dashboard.component';

@Component({ template: '<p>Synthetic sign-in destination</p>' })
class SyntheticSignIn {}

describe('Dashboard route loading', () => {
  let auth: jasmine.SpyObj<AuthService>;
  let api: jasmine.SpyObj<FastenApiService>;
  let router: Router;
  let harness: RouterTestingHarness;
  let starts: number;
  let ends: number;

  beforeEach(async () => {
    auth = jasmine.createSpyObj('AuthService', ['IsAuthenticated']);
    auth.IsAuthenticated.and.returnValue(Promise.resolve(true));
    api = jasmine.createSpyObj('FastenApiService', ['getSummary', 'getDashboards']);
    api.getSummary.and.returnValue(of({ sources: [], patients: [], resource_type_counts: [] }));
    api.getDashboards.and.returnValue(of([]));
    const standalone = isStandalone(DashboardComponent);
    await TestBed.configureTestingModule({
      imports: [AppRoutingModule, CommonModule, FormsModule, NgbModule, DirectivesModule,
        GridstackComponent, SyntheticSignIn, ...(standalone ? [DashboardComponent] : [])],
      declarations: standalone ? [] : [DashboardComponent],
      providers: [
        { provide: AuthService, useValue: auth },
        { provide: FastenApiService, useValue: api },
        { provide: LighthouseService, useValue: {} },
        IsAuthenticatedAuthGuard,
      ],
    }).compileComponents();
    router = TestBed.inject(Router);
    // Exercise the production dashboard route/guard. Only its public redirect
    // destination is replaced so the test needs no sign-in backend.
    const dashboard = router.config.find(route => route.path === 'dashboard');
    expect(dashboard).toBeDefined();
    router.resetConfig([dashboard, { path: 'auth/signin', component: SyntheticSignIn }]);
    starts = ends = 0;
    router.events.subscribe(event => {
      if (event instanceof RouteConfigLoadStart && event.route.path === 'dashboard') starts++;
      if (event instanceof RouteConfigLoadEnd && event.route.path === 'dashboard') ends++;
    });
    harness = await RouterTestingHarness.create();
  });

  it('loads the dashboard feature on authenticated navigation and renders it', async () => {
    await harness.navigateByUrl('/dashboard', DashboardComponent);
    expect(router.url).toBe('/dashboard');
    expect(starts).toBe(1);
    expect(ends).toBe(1);
    expect(api.getDashboards).toHaveBeenCalledTimes(1);
    expect(harness.routeNativeElement.textContent).toContain('Welcome back!');
  });

  it('redirects signed-out users before dashboard loading or data requests', async () => {
    auth.IsAuthenticated.and.returnValue(Promise.resolve(false));
    await harness.navigateByUrl('/dashboard', SyntheticSignIn);
    expect(router.url).toBe('/auth/signin');
    expect(starts).toBe(0);
    expect(ends).toBe(0);
    expect(api.getSummary).not.toHaveBeenCalled();
    expect(api.getDashboards).not.toHaveBeenCalled();
  });
});
