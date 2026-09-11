import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';
import { ApiService } from '../../core/services/api.service';
import { FileService } from '../../core/services/file.service';
import { ProfileClub, ProfileClubForm } from './profile-club.model';
import { Page } from '../../core/models/api.model';

/** Shape of GET /profile-club list (always ≤1 row). */
export interface ProfileClubPage {
  items: ProfileClub[];
  total: number;
  page: number;
  per_page: number;
  last_page: number;
}

@Injectable({ providedIn: 'root' })
export class ProfileClubService {
  private api = inject(ApiService);
  private fileSvc = inject(FileService);

  /** GET /profile-club — list (0 or 1 rows). */
  list(): Observable<Page<ProfileClub>> {
    return this.api.get<Page<ProfileClub>>('/profile-club');
  }

  /** GET /profile-club/:id — single row detail. */
  detail(id: number): Observable<ProfileClub> {
    return this.api.get<ProfileClub>(`/profile-club/${id}`);
  }

  /** POST /profile-club — create (only if no active row). */
  create(body: ProfileClubForm): Observable<ProfileClub> {
    return this.api.post<ProfileClub>('/profile-club', body);
  }

  /** PUT /profile-club/:id — update existing. */
  update(id: number, body: Partial<ProfileClubForm>): Observable<ProfileClub> {
    return this.api.put<ProfileClub>(`/profile-club/${id}`, body);
  }

  /** Upload a file to the file layer. */
  uploadFile(file: File, kategori: string): Observable<{ uuid: string; id: number; kategori: string }> {
    return this.fileSvc.upload('file', file, { kategori });
  }

  /** Blob URL for an image file (caller must revokeObjectURL). */
  imageUrl(uuid: string): Observable<string> {
    return this.fileSvc.imageUrl(uuid, 'medium');
  }
}
