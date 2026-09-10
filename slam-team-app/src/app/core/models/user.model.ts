// Auth/domain models. The API envelope and paging shapes are in api.model.ts.

/** Mirrors pkg/jwt.Claims' role fields (mst_role). */
export interface Role {
  id: number;
  level: number;
  is_super: boolean;
}

export interface User {
  id: number;
  name: string;
  email: string;
  /** Populated from JWT claims via GET /auth/me. */
  role?: Role;
  /** Monotonic counter from mst_pengaturan "rbac.perm_version". */
  perm_version?: number;
}

export interface LoginResponse {
  token: string;
  user: User;
}

/** Response from GET /auth/me/permissions. */
export interface MePermissions {
  permissions: string[]; // ["anggota.read", "file.delete"] or ["*"] for super
  perm_version: number;
  is_super: boolean;
}
