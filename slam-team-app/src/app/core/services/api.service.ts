import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { environment } from '../../../environments/environment';
import { ApiResponse } from '../models/api.model';

/**
 * The single HTTP client for the app: prefixes `environment.apiURL`, builds
 * params, and unwraps the `{success, message, data}` envelope so callers get
 * `data` and nothing else. Feature services wrap this and only know paths —
 * no page ever touches HttpClient or the envelope directly.
 */
@Injectable({ providedIn: 'root' })
export class ApiService {
  private http = inject(HttpClient);
  private base = environment.apiURL;

  get<T>(path: string, query?: Record<string, unknown>): Observable<T> {
    return this.http
      .get<ApiResponse<T>>(this.url(path), { params: toParams(query) })
      .pipe(map(unwrap));
  }

  post<T>(path: string, body?: unknown): Observable<T> {
    return this.http.post<ApiResponse<T>>(this.url(path), body ?? {}).pipe(map(unwrap));
  }

  put<T>(path: string, body?: unknown): Observable<T> {
    return this.http.put<ApiResponse<T>>(this.url(path), body ?? {}).pipe(map(unwrap));
  }

  patch<T>(path: string, body?: unknown): Observable<T> {
    return this.http.patch<ApiResponse<T>>(this.url(path), body ?? {}).pipe(map(unwrap));
  }

  delete<T>(path: string, query?: Record<string, unknown>): Observable<T> {
    return this.http
      .delete<ApiResponse<T>>(this.url(path), { params: toParams(query) })
      .pipe(map(unwrap));
  }

  /**
   * Multipart POST. Never set Content-Type here — the browser must add the
   * boundary itself (upload checklist, rancangan Bab 6.6). The
   * authInterceptor still attaches the Bearer token.
   */
  upload<T>(path: string, form: FormData): Observable<T> {
    return this.http.post<ApiResponse<T>>(this.url(path), form).pipe(map(unwrap));
  }

  /** Raw blob GET for private files served by the authenticated file handler. */
  blob(path: string): Observable<Blob> {
    return this.http.get(this.url(path), { responseType: 'blob' });
  }

  private url(path: string): string {
    return `${this.base}${path.startsWith('/') ? path : `/${path}`}`;
  }
}

/**
 * Envelope -> data. A `success: false` body is thrown as-is so callers can read
 * `errors` (field -> messages) and map them onto form controls.
 */
export function unwrap<T>(res: ApiResponse<T>): T {
  if (!res.success) throw res;
  return res.data as T;
}

/** Drops empty/absent values so they don't become `?q=undefined`. */
export function toParams(query?: Record<string, unknown>): HttpParams {
  let params = new HttpParams();
  for (const [key, value] of Object.entries(query ?? {})) {
    if (value === undefined || value === null || value === '') continue;
    if (Array.isArray(value)) {
      for (const v of value) params = params.append(key, String(v));
    } else {
      params = params.set(key, String(value));
    }
  }
  return params;
}
