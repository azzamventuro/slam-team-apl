import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '../../core/services/api.service';
import { FileService } from '../../core/services/file.service';
import { Dokumen, DokumenForm, DokumenQuery } from './dokumen.model';
import { Page } from '../../core/models/api.model';

@Injectable({ providedIn: 'root' })
export class DokumenService {
  private api = inject(ApiService);
  private fileSvc = inject(FileService);

  list(q: DokumenQuery): Observable<Page<Dokumen>> {
    return this.api.get<Page<Dokumen>>('/dokumen', q as unknown as Record<string, unknown>);
  }

  detail(id: number): Observable<Dokumen> {
    return this.api.get<Dokumen>(`/dokumen/${id}`);
  }

  create(body: DokumenForm): Observable<Dokumen> {
    return this.api.post<Dokumen>('/dokumen', body);
  }

  update(id: number, body: Partial<DokumenForm>): Observable<Dokumen> {
    return this.api.put<Dokumen>(`/dokumen/${id}`, body);
  }

  remove(id: number): Observable<void> {
    return this.api.delete<void>(`/dokumen/${id}`);
  }

  /** Upload a file to the file layer (kategori='dokumen'). */
  uploadFile(file: File): Observable<{ uuid: string; id: number; kategori: string }> {
    return this.fileSvc.upload('file', file, { kategori: 'dokumen' });
  }

  /** Download a file blob via the authenticated file endpoint. */
  downloadBlob(uuid: string): Observable<Blob> {
    const path = `/files/${uuid}/original`;
    return this.api.blob(path);
  }
}
