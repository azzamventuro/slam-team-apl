// Shared API/domain models.

/** Standard response envelope returned by slam-team-api. */
export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data?: T;
  errors?: unknown;
}

export interface User {
  id: number;
  name: string;
  email: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}
