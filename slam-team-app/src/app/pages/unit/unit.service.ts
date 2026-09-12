import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '../../core/services/api.service';
import { Unit, UnitForm, UnitQuery, AnggotaOpsi } from './unit.model';
import { Page } from '../../core/models/api.model';

@Injectable({ providedIn: 'root' })
export class UnitService {
  private api = inject(ApiService);

  list(q: UnitQuery): Observable<Page<Unit>> {
    return this.api.get<Page<Unit>>('/unit', q as unknown as Record<string, unknown>);
  }

  detail(id: number): Observable<Unit> {
    return this.api.get<Unit>(`/unit/${id}`);
  }

  create(body: UnitForm): Observable<Unit> {
    return this.api.post<Unit>('/unit', body);
  }

  update(id: number, body: UnitForm): Observable<Unit> {
    return this.api.put<Unit>(`/unit/${id}`, body);
  }

  remove(id: number): Observable<void> {
    return this.api.delete<void>(`/unit/${id}`);
  }

  anggotaTersedia(): Observable<AnggotaOpsi[]> {
    return this.api.get<AnggotaOpsi[]>('/unit/anggota-tersedia');
  }
}
