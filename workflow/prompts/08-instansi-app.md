# PROMPT — Build APP UI for `instansi` (Master Data Instansi / Sekolah)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory, do this before writing UI)
1. **Invoke the Taste Skill.** Announce **"Using design-taste-frontend"**, then run its pre-flight / audit-first pass on the existing app shell before implementing.
2. **Enforce the SLAM design tokens** in `D:/xampp/htdocs/slam-team-apl/workflow/_shared/design-tokens.md` — dark near-black default (`--slam-bg #0B0B0D`), dark surfaces, **SLAM red** accent (`--slam-primary #E11D2A`), white-ish text, **UPPERCASE wide-tracked** headings, red focus ring (never removed), 44px min touch targets, `prefers-reduced-motion` honored, status never color-only (pair icon/label). Where tokens differ from taste-skill defaults, **the tokens file wins**.
3. **Map to Bootstrap 5** (already installed) — use its grid/utilities/components; apply tokens via the SCSS/CSS-variable overrides already documented in the tokens file. No new UI library, no new webfont.
4. **Note:** *If `blog-fe` source is restored, mirror its layout/menu for this screen; otherwise follow the design tokens.* (`blog-fe` is currently empty — tokens + taste-skill are the fallback.)

## 1. Read first
- `D:/xampp/htdocs/slam-team-apl/workflow/modules/08-instansi.md` — the module spec (fields, endpoints, flow).
- `D:/xampp/htdocs/slam-team-apl/workflow/_shared/conventions-app.md` — Angular 21 standalone + signals + zoneless patterns, `ApiService` envelope, `PermissionService`, `*hasPermission`, `permissionGuard`, `FileService`, the canonical signal-based list page, reactive-form/server-error mapping. Reuse these; do not re-implement HTTP/token/permission logic.
- Skim `slam-team-app/CLAUDE.md`.

## 2. Scope
Build the `instansi` (Master Data Instansi / Sekolah) admin UI: a filterable/paginated list and a create/edit form with two logo uploads. Backend contract (envelope `{success,message,data,errors}`):

| Method | Path | Permission | Body / Query |
|--------|------|-----------|--------------|
| GET | `/instansi` | `instansi.read` | `page, per_page, q, sort, status` → `Page<Instansi>` |
| GET | `/instansi/:id` | `instansi.read` | → `Instansi` |
| POST | `/instansi` | `instansi.create` | `InstansiForm` → `Instansi` |
| PUT | `/instansi/:id` | `instansi.update` | `InstansiForm` → `Instansi` |
| DELETE | `/instansi/:id` | `instansi.delete` | → void (soft delete) |

Permission matrix: **SA/Admin CRUD, Moderator/User read-only, Guest none.**

## 3. Files to create under `src/app/pages/instansi/`

### `core/models/instansi.model.ts` (or `pages/instansi/instansi.model.ts`)
```ts
export interface Instansi {
  id: number;
  kode: string;
  nama: string;
  nama_club: string;
  alamat: string;
  no_telepon: string;
  logo_utama_file_id: number | null;
  logo_utama_uuid: string | null;
  logo_tambahan_file_id: number | null;
  logo_tambahan_uuid: string | null;
  tanggal_bergabung: string | null;   // YYYY-MM-DD
  status: number;                       // 1 aktif, 0 nonaktif
  jumlah_anggota: number;
  created_at: string;
  modified_at: string | null;
}

export interface InstansiForm {
  kode: string;
  nama: string;
  nama_club: string;
  alamat: string;
  no_telepon: string;
  logo_utama_file_id: number | null;
  logo_tambahan_file_id: number | null;
  tanggal_bergabung: string | null;
  status: number;
}

export interface InstansiQuery {
  page: number; per_page: number; q: string; sort?: string; status?: number | null;
}
```

### `pages/instansi/instansi.service.ts` (`providedIn:'root'`, thin wrapper over `ApiService`)
```ts
@Injectable({ providedIn: 'root' })
export class InstansiService {
  private api = inject(ApiService);
  list(q: InstansiQuery)          { return this.api.get<Page<Instansi>>('/instansi', q); }
  detail(id: number)              { return this.api.get<Instansi>(`/instansi/${id}`); }
  create(b: InstansiForm)         { return this.api.post<Instansi>('/instansi', b); }
  update(id: number, b: InstansiForm) { return this.api.put<Instansi>(`/instansi/${id}`, b); }
  remove(id: number)              { return this.api.delete<void>(`/instansi/${id}`); }
}
```

### `pages/instansi/instansi-list.ts/.html/.scss`
- Standalone component; imports `TranslatePipe, RouterLink, HasPermissionDirective, FormsModule`.
- Signal-based list (conventions-app §9): `q`, `page`, `perPage`, `status` signals → `computed` query → `toSignal(toObservable(query).pipe(debounceTime(250), switchMap(q => svc.list(q))))`. Show `total` + compute `lastPage`.
- Table columns: Kode, Nama, Nama Club, No. Telepon, **Status** (badge `.text-bg-success` aktif / `.text-bg-secondary` nonaktif — pair with label text, not color alone), **Jumlah Anggota**, actions. Table styled per tokens (dark rows, surface-2 uppercase muted header).
- Header **Tambah** button `*hasPermission="'instansi.create'"` → `routerLink="new"`.
- Row actions: **Edit** `*hasPermission="'instansi.update'"`, **Hapus** `*hasPermission="'instansi.delete'"` (confirm dialog using `INSTANSI.DELETE_CONFIRM`; on 409 show `INSTANSI.DELETE_IN_USE` toast — instansi still used by members).
- Search input bound to `q` (reset `page` to 1 on change) + a status filter `<select>`.
- Empty + loading states.

### `pages/instansi/instansi-form.ts/.html`
- One component for create AND edit (mode from route param `:id`). On edit, load `svc.detail(id)` and `patchValue`.
- **Typed reactive form** (`FormBuilder.nonNullable`) mirroring the API DTO:
```ts
form = this.fb.nonNullable.group({
  kode:      ['', [Validators.required, Validators.maxLength(50)]],
  nama:      ['', [Validators.required, Validators.maxLength(150)]],
  nama_club: ['', [Validators.maxLength(150)]],
  alamat:    [''],
  no_telepon:['', [Validators.maxLength(30)]],
  logo_utama_file_id:    [null as number | null],
  logo_tambahan_file_id: [null as number | null],
  tanggal_bergabung:     [null as string | null],   // <input type="date">
  status:    [1, [Validators.required]],             // select 1/0
});
```
- **Tanggal Bergabung** = native `<input type="date">` (emits `YYYY-MM-DD`); no date-picker library.
- **Status** = `<select>` with options 1 (Aktif) / 0 (Nonaktif).
- On submit: `if (form.invalid) { markAllAsTouched(); return; }`; call create/update; on `422/400` map `res.errors` (field→messages) onto controls via `setErrors({ server: msgs[0] })`; on success toast + navigate back to list. Disable submit while `saving()` signal is true.

### Logo uploads (two, via `FileService`)
- For each logo (utama, tambahan): a file `<input>` → `FileService.upload('file', file, { kategori: 'logo', reff_type: 'instansi' })` returns `{ id, uuid }`. Store `id` into the form control (`logo_utama_file_id` / `logo_tambahan_file_id`); render a preview from `FileService.imageUrl(uuid, 'low')` (blob — token added by interceptor). **Do NOT set `Content-Type`** (browser sets the multipart boundary). **Revoke** each object URL on destroy / when replaced (leak guard).
- In the list/detail, render the logo from `logo_utama_uuid` via `FileService.imageUrl(uuid, 'medium')`; fallback placeholder when null. Never put a token-less URL in `<img src>` — private files need the blob fetch.

## 4. Routing (add to the feature/children routes)
```ts
{
  path: 'instansi',
  canActivate: [authGuard, permissionGuard],
  data: { permission: 'instansi.read' },
  children: [
    { path: '',    loadComponent: () => import('./pages/instansi/instansi-list').then(m => m.InstansiList) },
    { path: 'new', loadComponent: () => import('./pages/instansi/instansi-form').then(m => m.InstansiForm),
      canActivate: [permissionGuard], data: { permission: 'instansi.create' } },
    { path: ':id/edit', loadComponent: () => import('./pages/instansi/instansi-form').then(m => m.InstansiForm),
      canActivate: [permissionGuard], data: { permission: 'instansi.update' } },
  ],
}
```
The dynamic sidebar (generated from `/modul` + `/me/permissions`) will surface this module automatically once the user has `instansi.read` — do not hard-code a menu entry.

## 5. i18n — add to `public/i18n/IND.json` AND `public/i18n/ENG.json` (keep in sync)
Namespace `INSTANSI`:
`TITLE`, `ADD`, `FORM.KODE`, `FORM.NAMA`, `FORM.NAMA_CLUB`, `FORM.ALAMAT`, `FORM.NO_TELEPON`, `FORM.LOGO_UTAMA`, `FORM.LOGO_TAMBAHAN`, `FORM.TANGGAL_BERGABUNG`, `FORM.STATUS`, `STATUS.AKTIF`, `STATUS.NONAKTIF`, `JUMLAH_ANGGOTA`, `DELETE_CONFIRM`, `DELETE_IN_USE`. Reuse `COMMON.*` (SAVE/CANCEL/ADD/EDIT/DELETE/SEARCH/DETAIL/EMPTY) and `VALIDATION.*`. No hard-coded user-facing strings; keep permission codes / column names verbatim.

## 6. Permission-gating reminder
`*hasPermission` and `permissionGuard` are **UX only** — the Go handler enforces `instansi.create/update/delete` server-side. Moderator/User see the list but no C/U/D buttons; even if they craft a request, the backend returns 403.

## 7. Verification checklist (run before declaring done)
- [ ] `npm install --legacy-peer-deps` then `npm run build:local` (or `serve:local`) compiles with no errors.
- [ ] As Super Admin/Admin: create an instansi → row persists (verify via `GET /instansi` / DB); edit updates it; delete soft-deletes and refreshes the list.
- [ ] Deleting an instansi still referenced by an active anggota → backend 409 surfaced as `INSTANSI.DELETE_IN_USE` toast (no crash).
- [ ] Logo upload posts `FormData` to `/files` (no manual `Content-Type`), stores the returned id, and preview renders via `/files/{uuid}/low`; list logo renders via `/files/{uuid}/medium`.
- [ ] Client validation mirrors the DTO (`kode`/`nama` required + max lengths, `status` 1/0); a backend `422` maps errors inline onto the right controls.
- [ ] As Moderator/User: Tambah/Edit/Hapus buttons hidden; list still loads; direct POST/PUT/DELETE would be blocked by backend (403).
- [ ] Responsive (mobile → desktop); dark theme default; SLAM red accent + UPPERCASE headings; red focus ring intact; status conveyed by label + badge (not color only).
- [ ] IND & ENG i18n files have the same `INSTANSI.*` key set; language switch shows no raw keys.
- [ ] "Using design-taste-frontend" announced and tokens enforced (no ad-hoc hex, no new font).
