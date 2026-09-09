// Auth/domain models. The API envelope and paging shapes live in api.model.ts.

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
  /**
   * Present once the auth module issues real RBAC claims — the current
   * scaffold's login response carries no role yet, so this is undefined for
   * everyone. Every role-based check must treat "no role" as denied, the same
   * way the backend's RequireAdmin denies a claims-less token.
   */
  role?: Role;
}

export interface LoginResponse {
  token: string;
  user: User;
}
