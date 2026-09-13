import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { Page } from '../models/api.model';
import {
  BacaSemuaResp,
  JumlahBelumDibaca,
  Notifikasi,
  NotifikasiQuery,
} from '../models/notifikasi.model';
import { ApiService } from './api.service';

/**
 * Thin wrapper over api-endpoints §4. Bearer-only — the inbox is implicitly
 * the caller's, so there is no permission gate anywhere in here.
 */
@Injectable({ providedIn: 'root' })
export class NotifikasiService {
  private api = inject(ApiService);

  list(q: NotifikasiQuery = {}): Observable<Page<Notifikasi>> {
    return this.api.get<Page<Notifikasi>>('/notifikasi', q as Record<string, unknown>);
  }

  unreadCount(): Observable<JumlahBelumDibaca> {
    return this.api.get<JumlahBelumDibaca>('/notifikasi/jumlah-belum-dibaca');
  }

  markRead(id: number): Observable<Notifikasi> {
    return this.api.patch<Notifikasi>(`/notifikasi/${id}/baca`);
  }

  markAllRead(): Observable<BacaSemuaResp> {
    return this.api.patch<BacaSemuaResp>('/notifikasi/baca-semua');
  }
}
