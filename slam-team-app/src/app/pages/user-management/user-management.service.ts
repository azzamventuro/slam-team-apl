import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { ApiService } from '../../core/services/api.service';
import { Page } from '../../core/models/api.model';
import {
  AnggotaOpsi,
  RoleOpsi,
  UserForm,
  UserListQuery,
  UserRow,
} from './user-management.model';

/**
 * UserService wraps the user-management CRUD endpoints.
 * Parameterized by `base` (`admin` | `moderator` | `user`) which maps to
 * the backend surface route prefix.
 */
@Injectable({ providedIn: 'root' })
export class UserService {
  private api = inject(ApiService);

  /** Build the base path for the given surface. */
  private base(base: string): string {
    return `/${base}`;
  }

  list(base: string, q: UserListQuery): Observable<Page<UserRow>> {
    return this.api.get<Page<UserRow>>(this.base(base), q as unknown as Record<string, unknown>);
  }

  detail(base: string, id: number): Observable<UserRow> {
    return this.api.get<UserRow>(`${this.base(base)}/${id}`);
  }

  create(base: string, body: UserForm): Observable<UserRow> {
    return this.api.post<UserRow>(this.base(base), body);
  }

  update(base: string, id: number, body: Partial<UserForm>): Observable<UserRow> {
    return this.api.put<UserRow>(`${this.base(base)}/${id}`, body);
  }

  remove(base: string, id: number): Observable<void> {
    return this.api.delete<void>(`${this.base(base)}/${id}`);
  }

  /** GET /{surface}/peran-tersedia — roles the actor can assign. */
  availableRoles(base: string): Observable<RoleOpsi[]> {
    return this.api.get<RoleOpsi[]>(`${this.base(base)}/peran-tersedia`);
  }

  /** GET /{surface}/anggota-tersedia — anggota without user accounts. */
  availableAnggota(base: string, q?: string, instansiId?: number): Observable<AnggotaOpsi[]> {
    const params: Record<string, unknown> = {};
    if (q) params['q'] = q;
    if (instansiId) params['instansi_id'] = instansiId;
    return this.api.get<AnggotaOpsi[]>(`${this.base(base)}/anggota-tersedia`, params);
  }
}
