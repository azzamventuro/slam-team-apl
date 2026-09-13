import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { Page } from '../../core/models/api.model';
import { ApiService } from '../../core/services/api.service';
import { CreateIzinReq, Izin, IzinQuery, JadwalOpsi, SesiOpsi, TolakReq } from './izin.model';

/** Thin wrapper over api-endpoints §9 (izin) + the two jadwal reads the sesi picker needs. */
@Injectable({ providedIn: 'root' })
export class IzinService {
  private api = inject(ApiService);

  /** Cakupan is applied server-side: User gets their own rows, Admin/Mod get all. */
  list(q: IzinQuery): Observable<Page<Izin>> {
    return this.api.get<Page<Izin>>('/izin', q as unknown as Record<string, unknown>);
  }

  create(body: CreateIzinReq): Observable<Izin> {
    return this.api.post<Izin>('/izin', body);
  }

  approve(id: number): Observable<Izin> {
    return this.api.patch<Izin>(`/izin/${id}/approve`);
  }

  tolak(id: number, body: TolakReq): Observable<Izin> {
    return this.api.patch<Izin>(`/izin/${id}/tolak`, body);
  }

  // ── Sesi picker (jadwal.read — every role has it) ──

  jadwalOpsi(): Observable<Page<JadwalOpsi>> {
    return this.api.get<Page<JadwalOpsi>>('/jadwal', {
      per_page: 100,
      sort: '-tanggal_mulai',
    });
  }

  sesiOpsi(jadwalId: number): Observable<SesiOpsi[]> {
    return this.api.get<SesiOpsi[]>(`/jadwal/${jadwalId}/sesi`);
  }
}
