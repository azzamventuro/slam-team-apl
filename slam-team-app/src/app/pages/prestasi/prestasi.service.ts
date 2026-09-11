import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '../../core/services/api.service';
import { Prestasi, PrestasiForm, PrestasiQuery, AnggotaOpsi } from './prestasi.model';
import { Page } from '../../core/models/api.model';

@Injectable({ providedIn: 'root' })
export class PrestasiService {
  private api = inject(ApiService);

  list(q: PrestasiQuery): Observable<Page<Prestasi>> {
    return this.api.get<Page<Prestasi>>('/prestasi', q as unknown as Record<string, unknown>);
  }

  detail(id: number): Observable<Prestasi> {
    return this.api.get<Prestasi>(`/prestasi/${id}`);
  }

  create(body: PrestasiForm): Observable<Prestasi> {
    return this.api.post<Prestasi>('/prestasi', body);
  }

  update(id: number, body: PrestasiForm): Observable<Prestasi> {
    return this.api.put<Prestasi>(`/prestasi/${id}`, body);
  }

  remove(id: number): Observable<void> {
    return this.api.delete<void>(`/prestasi/${id}`);
  }

  anggotaTersedia(): Observable<AnggotaOpsi[]> {
    return this.api.get<AnggotaOpsi[]>('/prestasi/anggota-tersedia');
  }
}
