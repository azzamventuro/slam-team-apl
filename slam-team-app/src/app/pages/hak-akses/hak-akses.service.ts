import { Injectable, inject } from '@angular/core';

import { ApiService } from '../../core/services/api.service';
import {
  CreateRoleReq,
  ModulCatalogue,
  Permission,
  Role,
  RoleDetail,
  SetPermissionsReq,
  UpdateRoleReq,
} from './hak-akses.model';

/**
 * HakAksesService wraps the 7 RBAC endpoints under `/hakakses/*`.
 * Thin wrapper — only knows paths, delegates HTTP to ApiService.
 */
@Injectable({ providedIn: 'root' })
export class HakAksesService {
  private api = inject(ApiService);

  // ── Roles ──

  listRoles() {
    return this.api.get<Role[]>('/hakakses/roles');
  }

  getRole(id: number) {
    return this.api.get<RoleDetail>(`/hakakses/roles/${id}`);
  }

  createRole(req: CreateRoleReq) {
    return this.api.post<Role>('/hakakses/roles', req);
  }

  updateRole(id: number, req: UpdateRoleReq) {
    return this.api.put<Role>(`/hakakses/roles/${id}`, req);
  }

  deleteRole(id: number) {
    return this.api.delete<void>(`/hakakses/roles/${id}`);
  }

  // ── Permissions ──

  /** Get all permissions for a role. */
  getRolePermissions(id: number) {
    return this.api.get<RoleDetail>(`/hakakses/roles/${id}/permissions`);
  }

  /** Set (replace) all permissions for a role. */
  setRolePermissions(id: number, req: SetPermissionsReq) {
    return this.api.put<void>(`/hakakses/roles/${id}/permissions`, req);
  }

  // ── Catalogue ──

  /** List all modules. */
  listModuls() {
    return this.api.get<ModulCatalogue[]>('/hakakses/moduls');
  }

  /** List all permissions (grouped by module). */
  listPermissions() {
    return this.api.get<Permission[]>('/hakakses/permissions');
  }
}
