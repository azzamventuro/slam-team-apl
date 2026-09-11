import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '../../core/services/api.service';
import { Medsos, MedsosForm, MedsosQuery, AnggotaOpsi } from './medsos.model';
import { Page } from '../../core/models/api.model';

@Injectable({ providedIn: 'root' })
export class MedsosService {
  private api = inject(ApiService);

  list(q: MedsosQuery): Observable<Page<Medsos>> {
    return this.api.get<Page<Medsos>>('/medsos', q as unknown as Record<string, unknown>);
  }

  detail(id: number): Observable<Medsos> {
    return this.api.get<Medsos>(`/medsos/${id}`);
  }

  create(body: MedsosForm): Observable<Medsos> {
    return this.api.post<Medsos>('/medsos', body);
  }

  update(id: number, body: MedsosForm): Observable<Medsos> {
    return this.api.put<Medsos>(`/medsos/${id}`, body);
  }

  remove(id: number): Observable<void> {
    return this.api.delete<void>(`/medsos/${id}`);
  }

  anggotaTersedia(): Observable<AnggotaOpsi[]> {
    return this.api.get<AnggotaOpsi[]>('/medsos/anggota-tersedia');
  }
}
