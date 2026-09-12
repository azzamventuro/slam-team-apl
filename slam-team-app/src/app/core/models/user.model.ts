// Auth/domain models. The API envelope and paging shapes are in api.model.ts.
// Mirrors internal/modules/core/auth/dto — keep the field names identical.

/** `dto.AuthRole` — the mst_role summary embedded in every user projection. */
export interface Role {
  id: number;
  nama?: string;
  level: number;
  is_super: boolean;
}

/**
 * `dto.AuthUser` — the account as returned with a token pair and persisted in
 * `slam_user`. `nama` comes from the linked anggota row (users has no name
 * column); `anggota_id` is null until an admin links the account.
 */
export interface User {
  id: number;
  username: string;
  email: string;
  anggota_id: number | null;
  nama: string;
  /** Authenticated serve route of the anggota's photo (`/api/v1/files/<uuid>/low`). */
  foto_url?: string | null;
  role: Role;
  /** Monotonic counter from mst_pengaturan "rbac.perm_version". */
  perm_version: number;
}

/** `dto.MeAnggota` — the member behind the account on GET /auth/me. */
export interface MeAnggota {
  id: number;
  nama_lengkap: string;
  no_induk?: string | null;
  foto_uuid?: string | null;
  foto_url?: string | null;
}

/** `dto.MeResponse` — GET /auth/me: the user projection plus profile extras. */
export interface Me extends User {
  timezone?: string;
  is_aktif?: boolean;
  anggota?: MeAnggota | null;
}

/** `dto.AuthResponse` — returned by POST /auth/login and POST /auth/refresh. */
export interface AuthResponse {
  access_token: string;
  /** Opaque, rotates on every refresh — the old one is revoked server-side. */
  refresh_token: string;
  /** Always "Bearer". */
  token_type: string;
  /** Access-token lifetime in seconds. */
  expires_in: number;
  user: User;
}

/** Response from GET /auth/me/permissions. */
export interface MePermissions {
  permissions: string[]; // ["anggota.read", "file.delete"] or ["*"] for super
  perm_version: number;
  is_super: boolean;
}
