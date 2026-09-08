# SLAM Team — Angular 21 App Conventions (`conventions-app.md`)

Distilled, authoritative front-end conventions for **slam-team-app**. Every APP
prompt in the workflow assumes these. They are consistent with
`slam-team-app/CLAUDE.md`; if that file and this one ever disagree, the code in
the scaffold wins — read it before inventing a pattern.

Target: `D:/xampp/htdocs/slam-team-apl/slam-team-app`. Build **into** the
existing scaffold (`core/`, `layouts/`, `pages/`, `shared/`, `environments/`).
Do not create a new layout.

---

## 1. Framework baseline (non-negotiable)

- **Angular 21, standalone components only.** No `NgModule`. Declare every
  dependency in the component's own `imports: [...]`.
- **Zoneless.** `zone.js` is NOT loaded. Change detection is driven by
  **signals** and the `async` pipe. Never mutate a plain field after async work
  and expect the view to update — put mutable view state in a `signal`.
- **Signals for state.** `signal`, `computed`, `effect`, and
  `toSignal(obs$)` for streams. Services expose read-only signals
  (`readonly x = this._x.asReadonly()`); components read them as `x()`.
- **New control flow only** in templates: `@if`, `@for` (with `track`),
  `@switch`, `@let`. No `*ngIf` / `*ngFor`.
- **File naming (no `.component` suffix):** `login.ts`, `login.html`,
  `login.scss`; class `Login`. Services `auth.service.ts` → `AuthService`.
  Guards `auth.guard.ts`, interceptors `auth.interceptor.ts`.
- **Lazy routes** via `loadComponent: () => import('...')`. Never eagerly import
  a page into `app.routes.ts`.
- TypeScript 5.9, RxJS 7.8, **Bootstrap 5** (already installed — use its
  classes, do not add a UI kit), **ngx-translate v18**. Tests: **vitest**.
- `npm install --legacy-peer-deps` (npm 11.3 crashes otherwise). Dev vs local
  API: `npm run serve:local`.

---

## 2. Folder placement (where things go)

```
src/app/
  core/
    guards/        functional CanActivateFn         (authGuard, permissionGuard)
    interceptors/  functional HttpInterceptorFn     (authInterceptor, errorInterceptor)
    services/      singletons providedIn:'root'     (AuthService, ApiService, PermissionService, NotifikasiService, ...)
    models/        interfaces / DTO types           (ApiResponse<T>, User, Modul, Permission, Anggota, ...)
  layouts/
    vertical/      shell: <app-sidebar> + <app-topbar> + <router-outlet> + <app-footer>
    topbar/ sidebar/ footer/
  pages/<feature>/ one folder per module (anggota/, jadwal/, absensi/, hak-akses/, ...)
                   sub-routes for list/detail/form live under the feature folder
  shared/          reusable pipes, directives (hasPermission), small dumb components
  environments/    environment.ts (apiURL:'/api/v1') + environment.local.ts (http://localhost:8080/api/v1)
public/i18n/       IND.json, ENG.json
```

Rule of thumb: **cross-cutting singleton → `core/services`**; **reused across
features → `shared/`**; **belongs to one module → `pages/<feature>/`**. Don't
promote to `shared/` until a second consumer actually exists (YAGNI).

---

## 3. The response envelope + typed API service

Backend always returns `{ success, message, data, errors }` (see
`internal/shared/response`). Model it once and route **all** HTTP through a
single generic `ApiService` so no page hand-writes `environment.apiURL` or
unwraps envelopes.

```ts
// core/models/api.model.ts
export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data?: T;
  errors?: Record<string, string[]> | string | null;
}

export interface Page<T> {          // list endpoints: data is a page
  items: T[];
  total: number;
  page: number;
  per_page: number;
}
```

```ts
// core/services/api.service.ts
import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';
import { environment } from '../../../environments/environment';
import { ApiResponse } from '../models/api.model';

/** Thin generic client: builds the URL, unwraps the envelope, returns data. */
@Injectable({ providedIn: 'root' })
export class ApiService {
  private http = inject(HttpClient);
  private base = environment.apiURL;

  get<T>(path: string, query?: Record<string, unknown>): Observable<T> {
    return this.http
      .get<ApiResponse<T>>(`${this.base}${path}`, { params: toParams(query) })
      .pipe(map(unwrap));
  }
  post<T>(path: string, body: unknown): Observable<T> {
    return this.http.post<ApiResponse<T>>(`${this.base}${path}`, body).pipe(map(unwrap));
  }
  put<T>(path: string, body: unknown): Observable<T> {
    return this.http.put<ApiResponse<T>>(`${this.base}${path}`, body).pipe(map(unwrap));
  }
  delete<T>(path: string): Observable<T> {
    return this.http.delete<ApiResponse<T>>(`${this.base}${path}`).pipe(map(unwrap));
  }
}

function unwrap<T>(res: ApiResponse<T>): T {
  if (!res.success) throw res;           // errorInterceptor / caller handles it
  return res.data as T;
}
function toParams(query?: Record<string, unknown>): HttpParams {
  let p = new HttpParams();
  for (const [k, v] of Object.entries(query ?? {}))
    if (v !== undefined && v !== null && v !== '') p = p.set(k, String(v));
  return p;
}
```

Feature services are thin wrappers that only know their paths:

```ts
// pages/anggota/anggota.service.ts  (feature-scoped, still providedIn:'root')
@Injectable({ providedIn: 'root' })
export class AnggotaService {
  private api = inject(ApiService);
  list(q: AnggotaQuery)      { return this.api.get<Page<Anggota>>('/anggota', q); }
  detail(id: number)         { return this.api.get<Anggota>(`/anggota/${id}`); }
  create(body: AnggotaForm)  { return this.api.post<Anggota>('/anggota', body); }
  update(id: number, b: AnggotaForm) { return this.api.put<Anggota>(`/anggota/${id}`, b); }
  remove(id: number)         { return this.api.delete<void>(`/anggota/${id}`); }
}
```

---

## 4. Interceptors (functional) + `app.config.ts`

Registered once, in order, via `withInterceptors`. Two exist in the scaffold —
reuse them, do not add per-request header logic in pages.

```ts
// core/interceptors/auth.interceptor.ts
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const token = inject(AuthService).token();
  return token
    ? next(req.clone({ setHeaders: { Authorization: `Bearer ${token}` } }))
    : next(req);
};
```

```ts
// core/interceptors/error.interceptor.ts — 401 => logout + redirect to login
export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  const router = inject(Router);
  return next(req).pipe(
    catchError((err: HttpErrorResponse) => {
      if (err.status === 401) {
        auth.logout();
        router.navigate(['/auth/login']);
      }
      // 403 = permission denied (backend RequirePermission). Surface a toast; do NOT logout.
      return throwError(() => err);
    }),
  );
};
```

```ts
// app.config.ts (shape)
provideHttpClient(withInterceptors([authInterceptor, errorInterceptor])),
provideRouter(routes),
provideTranslateService({ /* IND/ENG, loader prefix './i18n/' */ }),
```

- **`localStorage`:** JWT in `slam_token`, profile in `slam_user`. The
  `authInterceptor` is the ONLY place the token is attached.
- **`multipart/form-data` (file upload):** send a `FormData` and **do not set
  `Content-Type`** — the browser sets the boundary. The `authInterceptor` still
  adds the Bearer header (fine). See §8.

---

## 5. Auth + dynamic RBAC (guard, permissions, `*hasPermission`, sidebar)

RBAC is DB-driven. The JWT carries `role_id` + `perm_version`; the effective
permission list comes from `GET /me/permissions`. `is_super` roles bypass every
check. **Hiding a button is not security** — the Go handler enforces it too;
the front-end mirror is UX only.

### 5.1 AuthService (scaffold — extend, don't rewrite)

Owns `user` signal, `token()`, `login/logout`. Login hits
`POST /auth/login` with **username OR email** + password, stores token + user.

### 5.2 PermissionService

Loads `/me/permissions` once after login, holds them in a signal, exposes a
pure `can(code)` helper. Codes are `"<modul>.<aksi>"` (e.g. `anggota.create`).

```ts
// core/services/permission.service.ts
@Injectable({ providedIn: 'root' })
export class PermissionService {
  private api = inject(ApiService);
  private _perms = signal<Set<string>>(new Set());
  private _super = signal(false);
  readonly ready = signal(false);

  load() {
    return this.api.get<MePermissions>('/me/permissions').pipe(
      tap((p) => {
        this._super.set(p.is_super);
        this._perms.set(new Set(p.permissions));   // ['anggota.read', 'jadwal.create', ...]
        this.ready.set(true);
      }),
    );
  }
  can = (code: string) => this._super() || this._perms().has(code);
  clear() { this._perms.set(new Set()); this._super.set(false); this.ready.set(false); }
}
```

### 5.3 Guards — `authGuard` (present) + `permissionGuard` (data-driven)

```ts
// core/guards/permission.guard.ts
export const permissionGuard: CanActivateFn = (route) => {
  const perms = inject(PermissionService);
  const router = inject(Router);
  const need = route.data['permission'] as string | undefined;
  if (!need || perms.can(need)) return true;
  return router.createUrlTree(['/forbidden']);
};
// route: { path: 'anggota', loadComponent: ..., canActivate: [authGuard, permissionGuard],
//          data: { permission: 'anggota.read' } }
```

### 5.4 `*hasPermission` structural directive

```ts
// shared/directives/has-permission.directive.ts
@Directive({ selector: '[hasPermission]', standalone: true })
export class HasPermissionDirective {
  private tpl = inject(TemplateRef<unknown>);
  private vcr = inject(ViewContainerRef);
  private perms = inject(PermissionService);
  private shown = false;

  @Input() set hasPermission(code: string) {
    // effect() re-evaluates when perms signal changes (zoneless-safe)
    effect(() => {
      const allow = this.perms.can(code);
      if (allow && !this.shown) { this.vcr.createEmbeddedView(this.tpl); this.shown = true; }
      else if (!allow && this.shown) { this.vcr.clear(); this.shown = false; }
    });
  }
}
// usage:  <button *hasPermission="'anggota.create'" class="btn btn-danger">Tambah</button>
```

### 5.5 Dynamic sidebar from `/modul` + `/me/permissions`

The sidebar is generated, never hard-coded. Fetch `GET /modul` (all modules with
`kode`, `nama`, `grup`, `icon`, `urutan`) and filter to modules the user has at
least `.read` on. Group by `grup` (Master Data / Konten / Sistem / Operasional).

```ts
// layouts/sidebar/sidebar.ts (signal-computed menu)
export class Sidebar {
  private api = inject(ApiService);
  private perms = inject(PermissionService);
  private moduls = toSignal(this.api.get<Modul[]>('/modul'), { initialValue: [] });

  readonly menu = computed(() =>
    groupBy(
      this.moduls().filter((m) => this.perms.can(`${m.kode}.read`)),
      (m) => m.grup,
    ),
  );
}
```

### 5.6 Anti-escalation (RBAC admin page — Fase 1)

Mirror backend rules in the UI (backend still enforces): cannot grant a
permission you lack (disable those checkboxes), cannot edit a role whose
`level <= yours`, cannot delete the last super admin. The role×modul grid is one
of the three hardest screens — build it first with the taste-skill.

---

## 6. Reactive forms — validation mirrors the backend DTO

Use **typed Reactive Forms** (`FormBuilder.nonNullable`). Client validators
mirror the Go DTO `binding`/validator tags so users see errors before the round
trip — but the **server is authoritative**. On a `422`/`400`, map
`res.errors` (field → messages) back onto the matching controls.

```ts
// pages/anggota/anggota-form.ts
export class AnggotaForm {
  private fb = inject(FormBuilder);
  private svc = inject(AnggotaService);

  saving = signal(false);
  form = this.fb.nonNullable.group({
    nama_lengkap: ['', [Validators.required, Validators.maxLength(120)]],
    email:        ['', [Validators.required, Validators.email]],
    no_hp:        ['', [Validators.pattern(/^[0-9+]{8,20}$/)]],
    instansi_id:  [null as number | null, Validators.required],
  });

  submit() {
    if (this.form.invalid) { this.form.markAllAsTouched(); return; }
    this.saving.set(true);
    this.svc.create(this.form.getRawValue()).subscribe({
      next: () => { /* toast + navigate */ },
      error: (e) => { this.applyServerErrors(e?.errors); this.saving.set(false); },
    });
  }

  /** Map backend {field:[msg]} onto controls so messages render inline. */
  private applyServerErrors(errors: Record<string, string[]> | undefined) {
    for (const [field, msgs] of Object.entries(errors ?? {}))
      this.form.get(field)?.setErrors({ server: msgs[0] });
  }
}
```

Validation parity checklist per field: `required`, length (`min/max`),
`email`, numeric/`pattern`, enum membership (dropdown = the DB enum values),
and any cross-field rule. Copy the exact column/field names from the Go DTO and
the `.dbml` — do not rename on the way in.

---

## 7. i18n

- Keys live in `public/i18n/IND.json` and `public/i18n/ENG.json` (loader prefix
  `./i18n/`). **IND is the default.**
- Namespace keys by module: `ANGGOTA.TITLE`, `ANGGOTA.FORM.NAMA`,
  `COMMON.SAVE`, `COMMON.CANCEL`, `VALIDATION.REQUIRED`. Keep both files in sync
  (same key set) — a missing key renders the raw key, which is the bug tell.
- Import `TranslatePipe` into the component's `imports`; use
  `{{ 'ANGGOTA.TITLE' | translate }}`. Switch with
  `TranslateService.use('ENG')`.
- **No hard-coded user-facing strings** in templates. Technical identifiers
  (column names, route paths, permission codes) stay verbatim in code — never
  translate those.

---

## 8. Private images via the file endpoint

Private files are served ONLY through the authenticated Go handler
`GET /api/v1/files/{uuid}/{varian}` (`varian` = `original|medium|low`). A plain
`<img src>` won't carry the Bearer token, so fetch as a blob (interceptor adds
the header) and bind an object URL.

```ts
// shared/services/file.service.ts
@Injectable({ providedIn: 'root' })
export class FileService {
  private http = inject(HttpClient);
  private base = environment.apiURL;

  /** Blob fetch — authInterceptor attaches the token. Caller must revokeObjectURL. */
  imageUrl(uuid: string, varian: 'original' | 'medium' | 'low' = 'medium') {
    return this.http
      .get(`${this.base}/files/${uuid}/${varian}`, { responseType: 'blob' })
      .pipe(map((b) => URL.createObjectURL(b)));
  }

  upload(field: string, file: File, extra?: Record<string, string>) {
    const fd = new FormData();
    fd.append(field, file);                       // DO NOT set Content-Type
    for (const [k, v] of Object.entries(extra ?? {})) fd.append(k, v);
    return this.http.post<ApiResponse<{ uuid: string }>>(`${this.base}/files`, fd).pipe(map(unwrap));
  }
}
```

- Prefer `medium` in lists/cards, `low` for thumbnails, `original` only when
  full resolution is needed. **Revoke** object URLs on destroy to avoid leaks
  (a `computed`/`effect` that revokes the previous URL, or handle in
  `ngOnDestroy`).
- **Public** KTA QR page uses `GET /public/kta/{token}` (no auth) and shows only
  the whitelisted fields — never fetch private files there.

---

## 9. A signal-based list page (the canonical CRUD list)

`toSignal` on a params-driven stream. Filters/pagination are signals; the list
recomputes when they change. Zoneless-clean, no manual `subscribe`+field.

```ts
// pages/anggota/anggota-list.ts
@Component({
  selector: 'app-anggota-list',
  standalone: true,
  imports: [TranslatePipe, RouterLink, HasPermissionDirective, FormsModule],
  templateUrl: './anggota-list.html',
})
export class AnggotaList {
  private svc = inject(AnggotaService);

  // filter state (signals)
  q = signal('');
  page = signal(1);
  perPage = signal(20);

  private query = computed(() => ({ q: this.q(), page: this.page(), per_page: this.perPage() }));

  // stream -> signal; switchMap so a new filter cancels the in-flight request
  private result = toSignal(
    toObservable(this.query).pipe(
      debounceTime(250),
      switchMap((query) => this.svc.list(query)),
    ),
    { initialValue: { items: [], total: 0, page: 1, per_page: 20 } as Page<Anggota> },
  );

  rows = computed(() => this.result().items);
  total = computed(() => this.result().total);
  lastPage = computed(() => Math.max(1, Math.ceil(this.total() / this.perPage())));

  remove(id: number) { this.svc.remove(id).subscribe(() => this.page.set(this.page())); } // refetch
}
```

```html
<!-- anggota-list.html -->
<div class="d-flex justify-content-between align-items-center mb-3">
  <h1 class="h4 text-uppercase fw-bold">{{ 'ANGGOTA.TITLE' | translate }}</h1>
  <a *hasPermission="'anggota.create'" routerLink="new" class="btn btn-danger">
    {{ 'COMMON.ADD' | translate }}
  </a>
</div>

<input class="form-control mb-3" [ngModel]="q()" (ngModelChange)="q.set($event); page.set(1)"
       [placeholder]="'COMMON.SEARCH' | translate" />

@if (rows().length) {
  <table class="table table-dark table-hover align-middle">
    <thead><tr><th>NRA</th><th>{{ 'ANGGOTA.NAMA' | translate }}</th><th></th></tr></thead>
    <tbody>
      @for (a of rows(); track a.id) {
        <tr>
          <td>{{ a.no_induk }}</td>
          <td>{{ a.nama_lengkap }}</td>
          <td class="text-end">
            <a *hasPermission="'anggota.read'" [routerLink]="[a.id]" class="btn btn-sm btn-outline-light">{{ 'COMMON.DETAIL' | translate }}</a>
            <button *hasPermission="'anggota.delete'" (click)="remove(a.id)" class="btn btn-sm btn-outline-danger">{{ 'COMMON.DELETE' | translate }}</button>
          </td>
        </tr>
      }
    </tbody>
  </table>
  <!-- pagination: page.set(...) up to lastPage() -->
} @else {
  <p class="text-muted">{{ 'COMMON.EMPTY' | translate }}</p>
}
```

---

## 10. Design method — MANDATORY (taste-skill + tokens)

Every APP prompt/screen MUST:

1. **Invoke the taste-skill.** Announce **"Using design-taste-frontend"**, run
   its pre-flight/audit (audit-first on any existing screen), then implement.
2. **Enforce the SLAM design tokens** in
   `workflow/_shared/design-tokens.md` — one token set so 19 modules don't
   become 19 styles. Palette derives from the KTA card: near-black background,
   dark-gray surfaces, **SLAM red** accent, white text; bold, wide
   letter-spaced **UPPERCASE** headings.
3. **Map to Bootstrap 5** (already installed) — use its grid, utilities, and
   components; apply tokens via SCSS variables/utility overrides, not a new UI
   library.
4. Note in each prompt: *"If `blog-fe` source is restored, mirror its
   layout/menu for this screen; otherwise follow the design tokens."* (`blog-fe`
   is currently empty — tokens + taste-skill are the fallback.)

Hardest, no-precedent screens to design first: **(1)** RBAC role×modul grid,
**(2)** schedule calendar, **(3)** attendance/absen screen (camera + GPS).
Camera/GPS require a **secure context** (HTTPS) — `getUserMedia`/`geolocation`
fail on plain HTTP; the PWA fase adds an iOS Safari redirect note.

---

## 11. Quick do / don't

- DO reuse `ApiService`, `AuthService`, the two interceptors, `HasPermissionDirective`. DON'T re-implement HTTP/token/permission logic per page.
- DO put mutable view state in signals; DO use `@if/@for` with `track`. DON'T rely on zone.js change detection.
- DO mirror backend DTO validators; DON'T trust the client — the Go handler is authoritative (permissions AND validation).
- DO fetch private images as blobs through `/files/{uuid}/{varian}`; DON'T put a token-less URL in `<img src>`.
- DO keep IND/ENG keys in sync and namespaced by module; DON'T hard-code UI strings.
- DO invoke `design-taste-frontend` and enforce the tokens on every screen.
