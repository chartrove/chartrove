import { Inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { HTTP_CLIENT_TOKEN } from '../dependency-injection';
import { environment } from '../../environments/environment';
import { GetEndpointAbsolutePath } from '../../lib/utils/endpoint_absolute_path';

export interface DocumentSource { id: string; display: string }
export interface DocumentReceipt { source_id: string; source_resource_type: string; source_resource_id: string; original_sha256: string; created: boolean }
interface Response<T> { success: boolean; data: T }

@Injectable({providedIn:'root'})
export class DocumentImportService {
  private readonly base = `${GetEndpointAbsolutePath(globalThis.location,environment.fasten_api_endpoint_base)}/secure/document-import`;
  constructor(@Inject(HTTP_CLIENT_TOKEN) private http: HttpClient) {}
  async sources(): Promise<DocumentSource[]> {
    const response = await firstValueFrom(this.http.get<Response<DocumentSource[]>>(`${this.base}/sources`));
    if (!response?.success) throw new Error('Import sources unavailable');
    return response.data;
  }
  async save(source: string, file: File, title: string, text: string, reviewed: boolean): Promise<DocumentReceipt> {
    if (!reviewed) throw new Error('Review the transcript before importing');
    const body = new FormData();
    body.append('file',file,file.name); body.append('title',title); body.append('text',text); body.append('reviewed','true');
    const response = await firstValueFrom(this.http.post<Response<DocumentReceipt>>(`${this.base}/${encodeURIComponent(source)}`,body));
    if (!response?.success) throw new Error('Document import was not confirmed');
    return response.data;
  }
}
