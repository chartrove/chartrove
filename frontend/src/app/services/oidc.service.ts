import { Inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { HTTP_CLIENT_TOKEN } from '../dependency-injection';
import { environment } from '../../environments/environment';
import { GetEndpointAbsolutePath } from '../../lib/utils/endpoint_absolute_path';
import { AuthService } from './auth.service';

export interface OIDCIdentity { id: string; issuer: string; revoked_at: string | null; created_at: string }
export interface OIDCAudit { id: string; identity_id: string; action: string; created_at: string }
interface Response<T> { success: boolean; data: T }
interface Start { authorization_url: string; state: string }

@Injectable({ providedIn: 'root' })
export class OIDCService {
  private get base() { return GetEndpointAbsolutePath(globalThis.location, environment.fasten_api_endpoint_base) }
  constructor(@Inject(HTTP_CLIENT_TOKEN) private http: HttpClient, private auth: AuthService) {}

  async enabled(): Promise<boolean> {
    if (environment.environment_desktop || environment.environment_cloud) return false;
    try { return (await firstValueFrom(this.http.get<Response<{enabled: boolean}>>(`${this.base}/auth/oidc`))).data.enabled === true }
    catch { return false }
  }

  async start(password?: string): Promise<string> {
    const endpoint = password === undefined ? '/auth/oidc/start' : '/secure/auth/oidc/link';
    const response = await firstValueFrom(this.http.post<Response<Start>>(this.base + endpoint, password === undefined ? {} : {password}));
    const flow = response.data;
    const url = new URL(flow.authorization_url);
    if (!response.success || url.protocol !== 'https:' || !flow.state) throw new Error('Invalid OIDC transaction');
    return url.href;
  }

  async complete(params: URLSearchParams): Promise<void> {
    if (params.has('error') || params.getAll('state').length !== 1 || params.getAll('code').length !== 1 || !params.get('state') || !params.get('code')) throw new Error('Invalid OIDC callback');
    const response = await firstValueFrom(this.http.post<Response<string>>(`${this.base}/auth/oidc/complete`, {state: params.get('state'), code: params.get('code')}));
    if (!response.success || typeof response.data !== 'string') throw new Error('OIDC sign-in failed');
    this.auth.setAuthToken(response.data);
  }

  async identities(): Promise<OIDCIdentity[]> { return (await firstValueFrom(this.http.get<Response<OIDCIdentity[]>>(`${this.base}/secure/auth/oidc/identities`))).data || [] }
  async audit(): Promise<OIDCAudit[]> { return (await firstValueFrom(this.http.get<Response<OIDCAudit[]>>(`${this.base}/secure/auth/oidc/audit`))).data || [] }
  async revoke(id: string, password: string): Promise<void> { await firstValueFrom(this.http.post(`${this.base}/secure/auth/oidc/identities/${encodeURIComponent(id)}/revoke`, {password})) }
}
