// The response envelope every slam-team-api endpoint returns, plus the paged
// payload list endpoints put in `data`. Mirrors internal/shared/response.Body
// and internal/shared/pagination.Paginated — keep the field names identical.

/** `{success, message, data?, errors?}` — the one shape the API always returns. */
export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data?: T;
  /** Validation failures as field -> messages, or a plain string. */
  errors?: Record<string, string[]> | string | null;
}

/** `data` of a list endpoint. */
export interface Page<T> {
  items: T[];
  page: number;
  per_page: number;
  total: number;
  last_page: number;
}

/** Alias kept because the backend type is named `Paginated[T]`. */
export type Paginated<T> = Page<T>;

/** Query accepted by every list endpoint (pagination.ListQuery). */
export interface ListQuery {
  page?: number;
  per_page?: number;
  q?: string;
  /** Column name, `-` prefix for DESC, e.g. `-created_at`. */
  sort?: string;
}

/** An empty page — a safe `initialValue` for a `toSignal` list stream. */
export function emptyPage<T>(perPage = 20): Page<T> {
  return { items: [], page: 1, per_page: perPage, total: 0, last_page: 1 };
}
