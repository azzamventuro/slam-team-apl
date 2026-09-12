import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '../../core/services/api.service';
import { Anggota, AnggotaForm, AnggotaQuery } from '../../core/models/anggota.model';
import { Page } from '../../core/models/api.model';

@Injectable({ providedIn: 'root' })
export class AnggotaService {
  private api = inject(ApiService);

  list(q: AnggotaQuery): Observable<Page<Anggota>> {
    return this.api.get<Page<Anggota>>('/anggota', q as unknown as Record<string, unknown>);
  }

  detail(id: number): Observable<Anggota> {
    return this.api.get<Anggota>(`/anggota/${id}`);
  }

  create(body: AnggotaForm): Observable<Anggota> {
    return this.api.post<Anggota>('/anggota', body);
  }

  update(id: number, body: AnggotaForm): Observable<Anggota> {
    return this.api.put<Anggota>(`/anggota/${id}`, body);
  }

  remove(id: number): Observable<void> {
    return this.api.delete<void>(`/anggota/${id}`);
  }
}
