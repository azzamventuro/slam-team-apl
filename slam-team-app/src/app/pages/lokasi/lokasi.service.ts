import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { Page } from '../../core/models/api.model';
import { ApiService } from '../../core/services/api.service';
import { Lokasi, LokasiForm, LokasiQuery } from './lokasi.model';

/** Thin wrapper over ApiService — only knows the /lokasi paths and shapes. */
@Injectable({ providedIn: 'root' })
export class LokasiService {
  private api = inject(ApiService);

  list(query: LokasiQuery): Observable<Page<Lokasi>> {
    return this.api.get<Page<Lokasi>>('/lokasi', query as unknown as Record<string, unknown>);
  }

  detail(id: number): Observable<Lokasi> {
    return this.api.get<Lokasi>(`/lokasi/${id}`);
  }

  create(body: LokasiForm): Observable<Lokasi> {
    return this.api.post<Lokasi>('/lokasi', body);
  }

  update(id: number, body: LokasiForm): Observable<Lokasi> {
    return this.api.put<Lokasi>(`/lokasi/${id}`, body);
  }

  remove(id: number): Observable<void> {
    return this.api.delete<void>(`/lokasi/${id}`);
  }
}
