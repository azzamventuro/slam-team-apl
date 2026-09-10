import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '../../core/services/api.service';
import { Instansi, InstansiForm, InstansiQuery } from '../../core/models/instansi.model';
import { Page } from '../../core/models/api.model';

@Injectable({ providedIn: 'root' })
export class InstansiService {
  private api = inject(ApiService);

  list(q: InstansiQuery): Observable<Page<Instansi>> {
    return this.api.get<Page<Instansi>>('/instansi', q as unknown as Record<string, unknown>);
  }

  detail(id: number): Observable<Instansi> {
    return this.api.get<Instansi>(`/instansi/${id}`);
  }

  create(body: InstansiForm): Observable<Instansi> {
    return this.api.post<Instansi>('/instansi', body);
  }

  update(id: number, body: InstansiForm): Observable<Instansi> {
    return this.api.put<Instansi>(`/instansi/${id}`, body);
  }

  remove(id: number): Observable<void> {
    return this.api.delete<void>(`/instansi/${id}`);
  }
}
