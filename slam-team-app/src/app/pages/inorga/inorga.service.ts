import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '../../core/services/api.service';
import { Inorga, InorgaForm, InorgaQuery } from './inorga.model';
import { Page } from '../../core/models/api.model';

@Injectable({ providedIn: 'root' })
export class InorgaService {
  private api = inject(ApiService);

  list(q: InorgaQuery): Observable<Page<Inorga>> {
    return this.api.get<Page<Inorga>>('/inorga', q as unknown as Record<string, unknown>);
  }

  detail(id: number): Observable<Inorga> {
    return this.api.get<Inorga>(`/inorga/${id}`);
  }

  create(body: InorgaForm): Observable<Inorga> {
    return this.api.post<Inorga>('/inorga', body);
  }

  update(id: number, body: InorgaForm): Observable<Inorga> {
    return this.api.put<Inorga>(`/inorga/${id}`, body);
  }

  remove(id: number): Observable<void> {
    return this.api.delete<void>(`/inorga/${id}`);
  }
}
