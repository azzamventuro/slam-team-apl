import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { ApiService } from '../../core/services/api.service';
import {
  AssignReq,
  BulkAssignReq,
  BulkAssignResp,
  JadwalPeserta,
  JadwalRingkas,
  JadwalSesi,
  ListPesertaQuery,
  ResponReq,
} from './penugasan.model';

/** Thin wrapper over api-endpoints §7 (penugasan) + the two jadwal reads the panel needs. */
@Injectable({ providedIn: 'root' })
export class PenugasanService {
  private api = inject(ApiService);

  jadwal(jadwalId: number): Observable<JadwalRingkas> {
    return this.api.get<JadwalRingkas>(`/jadwal/${jadwalId}`);
  }

  sesi(jadwalId: number): Observable<JadwalSesi[]> {
    return this.api.get<JadwalSesi[]>(`/jadwal/${jadwalId}/sesi`);
  }

  list(jadwalId: number, q: ListPesertaQuery = {}): Observable<JadwalPeserta[]> {
    return this.api.get<JadwalPeserta[]>(`/jadwal/${jadwalId}/peserta`, q as Record<string, unknown>);
  }

  assign(jadwalId: number, body: AssignReq): Observable<JadwalPeserta> {
    return this.api.post<JadwalPeserta>(`/jadwal/${jadwalId}/peserta`, body);
  }

  bulkAssign(jadwalId: number, body: BulkAssignReq): Observable<BulkAssignResp> {
    return this.api.post<BulkAssignResp>(`/jadwal/${jadwalId}/peserta/bulk`, body);
  }

  remove(jadwalId: number, anggotaId: number): Observable<void> {
    return this.api.delete<void>(`/jadwal/${jadwalId}/peserta/${anggotaId}`);
  }

  /** Bearer-only: the assignee accepts/declines their own assignment. */
  respon(pesertaId: number, body: ResponReq): Observable<JadwalPeserta> {
    return this.api.patch<JadwalPeserta>(`/peserta/${pesertaId}/respon`, body);
  }
}
