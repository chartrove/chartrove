import {Inject, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {map, Observable} from 'rxjs';
import {HTTP_CLIENT_TOKEN} from '../dependency-injection';
import {GetEndpointAbsolutePath} from '../../lib/utils/endpoint_absolute_path';
import {environment} from '../../environments/environment';

export type SourceAccessLevel = 'VIEW' | 'EDIT';
export type SourceSearchMode = 'literal' | 'exact_id';
export interface AccessSource { id: string; display: string; platform_type?: string; owner_user_id?: string; access_level?: SourceAccessLevel; }
export interface SourceGrant { id: string; delegate_username: string; delegate_user_id: string; access_level: SourceAccessLevel; expires_at: string | null; revoked_at: string | null; }
export interface SourceAudit { id: string; created_at: string; actor_user_id: string; action: string; grant_id?: string; access_level?: SourceAccessLevel; resource_type?: string; resource_id?: string; }
export interface SharedResource { source_id?: string; source_resource_type: string; source_resource_id: string; resource_raw: Record<string, unknown>; resource_raw_text?: string; sort_title?: string; editable?: boolean; }
export type ResourceListItem = Pick<SharedResource, 'source_resource_type' | 'source_resource_id' | 'sort_title'>;
export interface SharedSummary { source: AccessSource; counts: Record<string, number>; }
interface AccessResponse<T> { success: boolean; data: T; }

@Injectable({providedIn: 'root'})
export class SourceAccessService {
  private readonly api = `${GetEndpointAbsolutePath(globalThis.location, environment.fasten_api_endpoint_base)}/secure`;
  private readonly base = `${this.api}/source-access`;
  constructor(@Inject(HTTP_CLIENT_TOKEN) private http: HttpClient) {}

  private data<T>(request: Observable<AccessResponse<T>>): Observable<T> {
    return request.pipe(map(response => {
      if (!response?.success) { throw new Error('Source access unavailable'); }
      return response.data;
    }));
  }
  private source(id: string): string { return `${this.base}/${encodeURIComponent(id)}`; }
  private resource(source: string, type: string, id?: string): string {
    return `${this.source(source)}/resources/${encodeURIComponent(type)}${id === undefined ? '' : `/${encodeURIComponent(id)}`}`;
  }
  owned(): Observable<AccessSource[]> {
    return this.data(this.http.get<AccessResponse<AccessSource[]>>(`${this.api}/source`)).pipe(
      // Owner credential responses are never retained by the sharing UI.
      map(sources => sources.map(({id, display, platform_type}) => ({id, display, platform_type})))
    );
  }
  shared(): Observable<AccessSource[]> { return this.data(this.http.get<AccessResponse<AccessSource[]>>(`${this.base}/shared`)); }
  grants(source: string): Observable<SourceGrant[]> { return this.data(this.http.get<AccessResponse<SourceGrant[]>>(`${this.source(source)}/grants`)); }
  audit(source: string): Observable<SourceAudit[]> { return this.data(this.http.get<AccessResponse<SourceAudit[]>>(`${this.source(source)}/audit`)); }
  grant(source: string, username: string, level: SourceAccessLevel, expires: string | null): Observable<SourceGrant> {
    return this.data(this.http.post<AccessResponse<SourceGrant>>(`${this.source(source)}/grants`, {delegate_username: username, access_level: level, expires_at: expires}));
  }
  revoke(source: string, grant: string): Observable<null> { return this.data(this.http.delete<AccessResponse<null>>(`${this.source(source)}/grants/${encodeURIComponent(grant)}`)); }
  summary(source: string): Observable<SharedSummary> { return this.data(this.http.get<AccessResponse<SharedSummary>>(`${this.source(source)}/summary`)); }
  list(source: string, type: string, page: number): Observable<ResourceListItem[]> { return this.data(this.http.get<AccessResponse<ResourceListItem[]>>(`${this.source(source)}/metadata/${encodeURIComponent(type)}?page=${page}`)); }
  search(source: string, type: string, text: string, page: number, mode: SourceSearchMode = 'literal'): Observable<ResourceListItem[]> {
    const body = mode === 'literal' ? {text, page} : {text, page, mode};
    return this.data(this.http.post<AccessResponse<ResourceListItem[]>>(`${this.source(source)}/search/${encodeURIComponent(type)}`, body));
  }
  detail(source: string, type: string, id: string): Observable<SharedResource> { return this.data(this.http.get<AccessResponse<SharedResource>>(this.resource(source, type, id))); }
  edit(source: string, type: string, id: string, raw: string): Observable<null> {
    return this.data(this.http.patch<AccessResponse<null>>(this.resource(source, type, id), `{"resource_raw":${raw}}`, {headers: {'Content-Type': 'application/json'}}));
  }
  createLocal(source: string, raw: string): Observable<SharedResource> {
    return this.data(this.http.post<AccessResponse<SharedResource>>(`${this.source(source)}/resources`, `{"resource_raw":${raw}}`, {headers: {'Content-Type': 'application/json'}}));
  }
}
