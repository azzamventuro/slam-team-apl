// Auth/domain models. The API envelope and paging shapes live in api.model.ts.

export interface User {
  id: number;
  name: string;
  email: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}
