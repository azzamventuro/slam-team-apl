import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { ApiResponse } from '../../core/models/api.model';
import { ApiService } from '../../core/services/api.service';
import { FileKategori, FileVarian, UploadOptions, UploadedFile } from './file.model';

/**
 * The one client for the centralised file layer (`POST /files`,
 * `GET /files/:uuid/:varian`, `DELETE /files/:uuid`).
 *
 * Everything goes through `ApiService`, so the base URL, the
 * `{success,message,data,errors}` envelope and the Bearer token (added by
 * `authInterceptor`) are handled in exactly one place — this service only knows
 * paths and shapes. See conventions-app §3 and §8.
 */
@Injectable({ providedIn: 'root' })
export class FileService {
  private api = inject(ApiService);

  /**
   * Fetches a private variant as a blob and returns an object URL.
   *
   * A plain `<img src>` cannot carry the Bearer token, which is why the bytes
   * are fetched instead of linked. **The caller owns the returned URL and MUST
   * `URL.revokeObjectURL` it** when the image is replaced or destroyed —
   * `<app-secure-image>` does this for you.
   */
  imageUrl(uuid: string, varian: FileVarian = 'medium'): Observable<string> {
    return this.api.blob(`/files/${uuid}/${varian}`).pipe(map((b) => URL.createObjectURL(b)));
  }

  /**
   * Uploads one file as `multipart/form-data`.
   *
   * `Content-Type` is deliberately never set: the browser has to generate the
   * multipart boundary itself (rancangan Bab 6.6). Empty extras are dropped so
   * an absent `reff_id` does not arrive as the string "null".
   *
   * ponytail: reports no byte-level progress — `ApiService` returns the parsed
   * body, not the event stream. The UI shows an indeterminate bar. Swap in an
   * `HttpRequest` with `reportProgress: true` if large documents ever need a
   * percentage.
   */
  upload(field: string, file: File, extra?: Record<string, string>): Observable<UploadedFile> {
    const form = new FormData();
    form.append(field, file);
    for (const [key, value] of Object.entries(extra ?? {})) {
      if (value !== undefined && value !== null && value !== '') form.append(key, value);
    }
    return this.api.upload<UploadedFile>('/files', form);
  }

  /** `upload` with the multipart field name and the option mapping already applied. */
  uploadFor(kategori: FileKategori, file: File, options?: UploadOptions): Observable<UploadedFile> {
    const extra: Record<string, string> = { kategori };
    if (options?.reffType) extra['reff_type'] = options.reffType;
    if (options?.reffId !== undefined && options.reffId !== null) {
      extra['reff_id'] = String(options.reffId);
    }
    if (options?.isPublik !== undefined && options.isPublik !== null) {
      extra['is_publik'] = String(options.isPublik);
    }
    return this.upload('file', file, extra);
  }

  /** Metadata + the variants written so far — how a client sees `status_proses` reach `selesai`. */
  detail(uuid: string): Observable<UploadedFile> {
    return this.api.get<UploadedFile>(`/files/${uuid}`);
  }

  /** Soft delete. The Go service enforces ownership; hiding the button is not the guard. */
  remove(uuid: string): Observable<void> {
    return this.api.delete<void>(`/files/${uuid}`);
  }
}

/**
 * Pulls the most useful sentence out of a failed call: the first field error,
 * then the envelope message. Returns '' when there is nothing better than the
 * caller's own i18n fallback.
 *
 * Two things can arrive here, and both carry the same envelope: an
 * `HttpErrorResponse` (non-2xx, envelope under `.error`) and the envelope
 * itself, which `ApiService.unwrap` throws for a `success: false` body served
 * with 200. `errors` is likewise either a string or a field map whose values
 * are a message or a list of messages.
 */
export function fileErrorText(err: unknown): string {
  const outer = err as { error?: unknown } | null | undefined;
  const body = (outer && typeof outer === 'object' && 'error' in outer ? outer.error : outer) as
    ApiResponse<unknown> | undefined;
  if (!body || typeof body !== 'object') return '';

  const errors = body.errors;
  if (typeof errors === 'string' && errors) return errors;
  if (errors && typeof errors === 'object') {
    for (const value of Object.values(errors)) {
      if (Array.isArray(value) && value.length) return String(value[0]);
      if (typeof value === 'string' && value) return value;
    }
  }

  // Only an actual envelope has a user-facing `message`. A transport failure
  // arrives as a plain Error whose message is technical ("Http failure
  // response for /api/v1/files: 0 Unknown Error") — the caller's i18n key is
  // the better thing to show, so return nothing.
  const isEnvelope = 'success' in body || 'errors' in body;
  return isEnvelope && typeof body.message === 'string' ? body.message : '';
}
