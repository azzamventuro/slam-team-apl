import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';
import { environment } from '../../../environments/environment';
import { ApiResponse } from '../models/api.model';

/** Response shape from POST /files. */
export interface FileUploadResult {
  uuid: string;
  id: number;
  kategori: string;
}

/**
 * Thin wrapper around the file-layer endpoints. Every private image
 * must be fetched as a blob (authInterceptor attaches the token);
 * plain `<img src>` will NOT carry the Bearer header.
 */
@Injectable({ providedIn: 'root' })
export class FileService {
  private http = inject(HttpClient);
  private base = environment.apiURL;

  /** Blob fetch for a private file. Caller MUST revokeObjectURL. */
  imageUrl(uuid: string, varian: 'original' | 'medium' | 'low' = 'medium'): Observable<string> {
    return this.http
      .get(`${this.base}/files/${uuid}/${varian}`, { responseType: 'blob' })
      .pipe(map((b) => URL.createObjectURL(b)));
  }

  /**
   * Multipart upload. NEVER set Content-Type — the browser adds the
   * boundary itself. authInterceptor attaches the Bearer token.
   */
  upload(field: string, file: File, extra?: Record<string, string>): Observable<FileUploadResult> {
    const fd = new FormData();
    fd.append(field, file);
    for (const [k, v] of Object.entries(extra ?? {})) fd.append(k, v);
    return this.http
      .post<ApiResponse<FileUploadResult>>(`${this.base}/files`, fd)
      .pipe(map((res) => res.data as FileUploadResult));
  }
}
