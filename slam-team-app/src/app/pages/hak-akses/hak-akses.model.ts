/** Hak Akses domain models — mirrors the Go DTOs from internal/modules/core/hakakses/dto/. */

/** Role from GET /hakakses/roles. */
export interface Role {
  id: number;
  kode: string;
  nama: string;
  level: number;
  is_super: boolean;
  keterangan: string;
  created_at: string;
  updated_at: string;
}

/** Role with its full permission set from GET /hakakses/roles/:id/permissions. */
export interface RoleDetail extends Role {
  permissions: RolePermissionEntry[];
}

/** One permission grant on a role. */
export interface RolePermissionEntry {
  permission_id: number;
  modul_kode: string;
  aksi: string;
  cakupan: string; // "semua" | "instansi_sendiri" | "milik_sendiri"
}

/** Module from GET /hakakses/moduls. */
export interface ModulCatalogue {
  id: number;
  kode: string;
  nama: string;
  icon: string;
  grup: string;
  urutan: number;
}

/** Permission from GET /hakakses/permissions. */
export interface Permission {
  id: number;
  modul_id: number;
  modul_kode: string;
  aksi: string;
  label: string; // "anggota.create"
}

/** Payload for POST /hakakses/roles. */
export interface CreateRoleReq {
  kode: string;
  nama: string;
  level: number;
  keterangan?: string;
}

/** Payload for PUT /hakakses/roles/:id. */
export interface UpdateRoleReq {
  nama?: string;
  level?: number;
  keterangan?: string;
}

/** Payload for PUT /hakakses/roles/:id/permissions. */
export interface SetPermissionsReq {
  permissions: PermissionGrant[];
}

/** One permission grant in the set-permissions payload. */
export interface PermissionGrant {
  permission_id: number;
  cakupan: string; // "semua" | "instansi_sendiri" | "milik_sendiri"
}
