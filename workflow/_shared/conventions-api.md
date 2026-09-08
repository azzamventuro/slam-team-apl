# conventions-api.md — Go backend conventions (SLAM Team API)

Distilled, load-bearing rules every module MUST follow. Consistent with
`slam-team-api/CLAUDE.md`. Authoritative sources on conflict: `slamteam_db.dbml`
(schema) > rancangan Bab 9/10 (routes/migrations) > this file > old app (never).

Stack: Go 1.24 · Gin · GORM (PostgreSQL, db `slamteam_db`) · JWT HS256 · Zap ·
go-playground/validator. Entry: `cmd/api/main.go`. Routes mount under `/api/v1`
in `internal/router/router.go`.

---

## 1. Module layout

One folder per feature under `internal/modules/core/<mod>/`:

```
<mod>/
├── domain/            GORM entities + enum value types
├── dto/               request/response structs (binding + json tags)
├── repository/        GORM data access, returns domain entities
├── service/           business logic, composes repositories, enforces rules
├── handler/           Gin handlers: bind → call service → response envelope
└── main.<mod>.go      Initialize(deps) *Module + (m *Module) SetupRoutes(rg)
```

Request flow is one-directional: **router → middleware → handler → service →
repository → GORM**. Handlers never touch `*gorm.DB`; repositories never touch
`*gin.Context`. Services hold all business rules (RBAC scoping, NRA numbering,
Haversine, status assignment).

`main.<mod>.go` wires the module and is registered in `router.go`:

```go
// internal/modules/core/lokasi/main.lokasi.go
package lokasi

type Module struct {
	h      *handler.LokasiHandler
	jwtMgr *jwt.Manager
	perm   *middleware.PermGuard // shared RBAC guard (Fase 1)
}

func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard) *Module {
	repo := repository.NewLokasiRepository(db)
	svc := service.NewLokasiService(repo)
	return &Module{h: handler.NewLokasiHandler(svc), jwtMgr: jwtMgr, perm: perm}
}

func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/lokasi", middleware.JWTAuth(m.jwtMgr))
	g.GET("", m.perm.Require("lokasi.read"), m.h.List)
	g.POST("", m.perm.Require("lokasi.create"), m.h.Create)
	g.PUT("/:id", m.perm.Require("lokasi.update"), m.h.Update)
	g.DELETE("/:id", m.perm.Require("lokasi.delete"), m.h.Delete)
}
```

Register in `router.go`:

```go
lokasi.Initialize(db, jwtMgr, permGuard).SetupRoutes(apiV1)
```

---

## 2. Response envelope

Always use `internal/shared/response`. Never hand-roll `c.JSON`. Envelope:

```json
{ "success": true|false, "message": "...", "data": {...}?, "errors": {...}? }
```

Helpers already in the scaffold: `OK` (200), `Created` (201), `BadRequest`
(400), `Unauthorized` (401), `Unprocess` (422, validation), `Internal` (500,
logs real error + returns generic message — never leak internals). Add the two
still missing to the same package (do NOT fork it): `Forbidden` (403, RBAC
denials) and `NotFound` (404). Envelope stays `{success,message,data?,errors?}`.

Handler shape:

```go
func (h *LokasiHandler) Create(c *gin.Context) {
	var req dto.CreateLokasiReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	out, err := h.svc.Create(c.Request.Context(), middleware.Claims(c), req)
	if err != nil {
		response.FromError(c, err) // maps sentinel errors → status; else Internal
		return
	}
	response.Created(c, out)
}
```

---

## 3. Error handling

- Services return typed sentinel errors, handlers map them. Add a small
  `response.FromError(c, err)` that switches on sentinels and falls back to
  `Internal`. Do NOT scatter status codes across services.

```go
// service-level sentinels (shared/apperr or per package)
var (
	ErrNotFound     = errors.New("data tidak ditemukan")
	ErrForbidden    = errors.New("tidak diizinkan")
	ErrConflict     = errors.New("data bentrok")
	ErrValidation   = errors.New("validasi gagal")
)
```

- Repository: translate `gorm.ErrRecordNotFound` → `ErrNotFound`; let other DB
  errors bubble as-is (handler → `Internal`, which logs). Never log-and-return
  the same error twice; `Internal` is the single log point for 500s.
- Never return raw GORM/PG error text to the client (leaks schema).

---

## 4. Validation (Gin binding)

DTOs carry `binding` tags; custom rules registered in `pkg/validator`.

```go
type CreateLokasiReq struct {
	Nama       string  `json:"nama"       binding:"required,max=150"`
	Latitude   float64 `json:"latitude"   binding:"required,latitude"`
	Longitude  float64 `json:"longitude"  binding:"required,longitude"`
	RadiusMtr  int     `json:"radius_mtr" binding:"required,min=10,max=5000"`
}
```

- Bind with `ShouldBindJSON` (JSON) / `ShouldBind` (multipart/query). On error →
  `response.Unprocess` with a field→message map from `validator.Explain(err)`.
- Business validation (uniqueness, cross-field, permission-dependent) lives in
  the **service**, not binding tags.
- Never bind directly into a domain entity; DTO → map into entity in the service.

---

## 5. List endpoints — pagination & filter shape

Standard master-data list: `GET /<mod>?page=1&per_page=20&q=&sort=-created_at&<filters>`.

Query DTO + envelope:

```go
type ListQuery struct {
	Page    int    `form:"page,default=1"     binding:"min=1"`
	PerPage int    `form:"per_page,default=20" binding:"min=1,max=100"`
	Q       string `form:"q"`
	Sort    string `form:"sort"` // e.g. "-created_at" (prefix - = DESC)
}

type Paginated[T any] struct {
	Items    []T   `json:"items"`
	Page     int   `json:"page"`
	PerPage  int   `json:"per_page"`
	Total    int64 `json:"total"`
	LastPage int   `json:"last_page"`
}
```

Returned inside `data`: `{ "success": true, "data": { "items": [...],
"page":1, "per_page":20, "total":123, "last_page":7 } }`. Repository runs a
`Count` then a paged `Find` on the same scoped query (see §8 cakupan). Cap
`per_page` at 100. Whitelist sortable columns — never interpolate `sort` into
SQL.

---

## 6. GORM model conventions

Every persistent table maps to a domain struct. Rules from the DBML header:

- **PK**: `bigint identity` → `ID int64 gorm:"primaryKey"`.
- **Time**: all timestamps are `timestamptz` (UTC). Use `time.Time`. Never
  `datetime`, never store local time. Column defaults `now()` handled by DB.
- **Soft delete**: schema uses explicit `is_deleted bool + deleted_at + deleted_by`
  (NOT GORM's `gorm.DeletedAt` magic — the DBML columns differ). Embed:

```go
type Audit struct {
	CreatedAt  time.Time  `gorm:"column:created_at"           json:"created_at"`
	CreatedBy  *int64     `gorm:"column:created_by"           json:"created_by,omitempty"`
	ModifiedAt *time.Time `gorm:"column:modified_at"          json:"modified_at,omitempty"`
	ModifiedBy *int64     `gorm:"column:modified_by"          json:"modified_by,omitempty"`
	IsDeleted  bool       `gorm:"column:is_deleted;default:false" json:"-"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"           json:"-"`
	DeletedBy  *int64     `gorm:"column:deleted_by"           json:"-"`
}
```

  All list/read queries add `Where("is_deleted = false")` (a reusable scope).
  DELETE sets `is_deleted=true, deleted_at=now(), deleted_by=<actor>` — never a
  hard `DELETE`.

- **Public identity**: never expose sequential `id` in public URLs. Use the
  separate `uuid`/`token` column (`gen_random_uuid()`). Internal admin routes may
  use `id`.
- **File refs**: image/document columns are `*_file_id bigint` FKs to `mst_file.id`
  (nullable → `*int64`). Never store paths/URLs on the owning row; resolve via the
  file layer (§9).
- **Enums**: DBML defines Postgres enums (`jenis_anggota`, `cakupan_permission`,
  `varian_file`, `status_absensi`, …). Model them as Go string types with
  constants; the DB column is the PG enum. Validate allowed values in the service
  (or a binding `oneof`). Do not use int codes.

```go
type JenisAnggota string
const (
	JenisSiswa        JenisAnggota = "siswa"
	JenisDewasa       JenisAnggota = "dewasa"
	JenisSiswaKeDewasa JenisAnggota = "siswa_ke_dewasa"
)
```

- `TableName()` explicit when it differs from GORM's guess (e.g. `mst_lokasi`,
  `mst_file`, `mst_file_varian`, `mst_pengaturan`).

---

## 7. Migrations & seeders

**No `AutoMigrate`.** Schema is numbered SQL migration files from day one
(rancangan Bab 10.1), generated to match the DBML — DBML wins on conflict.

```
migrations/
  0001_enums.up.sql            0001_enums.down.sql
  0002_mst_pengaturan.up.sql   0002_mst_pengaturan.down.sql
  0003_file_layer.up.sql       0003_file_layer.down.sql
  0004_rbac.up.sql             0004_rbac.down.sql
  ...
```

- Every migration has an `.up.sql` and a paired `.down.sql` (rollback). Numbered,
  monotonic, never edited after commit — add a new one.
- Partial unique indexes as the DBML says (e.g. `users.username UNIQUE WHERE
  is_deleted = false`).
- Enums created once (`CREATE TYPE ... AS ENUM (...)`).
- Use a migration runner (e.g. golang-migrate) invoked via `slamctl migrate up`
  and `make migrate`.

**Seeders** (reference data: mst_pengaturan, RBAC matrix, mst_wilayah, mst_modul,
permissions) MUST be idempotent — running twice never duplicates rows. Use
`INSERT ... ON CONFLICT (kunci/kode) DO UPDATE/DO NOTHING`, keyed on the natural
unique column. Seeders live under `cmd/slamctl` (`slamctl seed`) and are safe to
re-run.

---

## 8. RBAC — RequirePermission, perm cache, perm_version, cakupan

Permission matrix lives in DB: `role_permission(role_id, permission_id, cakupan)`
where `cakupan ∈ {semua, instansi_sendiri, milik_sendiri}`. Permissions are
`modul.aksi` strings (e.g. `lokasi.read`, `absensi.override`).

**Middleware** on every guarded route:

```go
func (g *PermGuard) Require(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := middleware.Claims(c)
		if cl.IsSuper {           // is_super roles BYPASS all checks
			c.Next(); return
		}
		set := g.effective(c, cl.RoleID, cl.PermVersion) // cached
		cak, ok := set[perm]
		if !ok {
			response.Forbidden(c, "tidak diizinkan"); return
		}
		c.Set("cakupan:"+perm, cak) // handler/service reads scope
		c.Next()
	}
}
```

- **Cache**: effective permissions cached per role, key `perm:role:{id}` (Redis if
  present, else in-process). Cache stores the whole `map[perm]cakupan`.
- **perm_version**: JWT carries `role_id` + `perm_version`. On any RBAC change
  bump the role's version → cached entry keyed with the old version is ignored,
  forcing a reload. Old tokens with a stale `perm_version` are treated as stale
  and re-fetched (or rejected → refresh).
- **Anti-escalation** (enforced in service, not just UI): cannot grant a
  permission you lack; cannot edit a role whose `level <= yours`; the last
  `is_super` role/user cannot be deleted. `hak_akses` module: Super Admin only.

**Cakupan scoping in queries** — the handler/service reads the resolved cakupan
and narrows the repository query:

```go
switch cakupan {
case "semua":
	// no extra filter
case "instansi_sendiri":
	q = q.Where("instansi_id = ?", cl.InstansiID)
case "milik_sendiri":
	q = q.Where("anggota_id = ?", cl.AnggotaID) // or created_by, per module
}
```

Applies to reads AND writes/deletes: a `milik_sendiri` actor updating a row must
be filtered to their own rows server-side. Hiding a button is NOT security —
every UI permission has a matching handler check here. Example: `absensi.read`
= `semua` for SA/Admin/Mod, `milik_sendiri` for User.

---

## 9. Timezone rules

- All stored times are `timestamptz` in UTC. `SERVER time is authority`
  (absensi never trusts device clock; device time stored as evidence + `selisih`).
- Go MUST import blank tzdata so IANA zones resolve without OS tz files:

```go
import _ "time/tzdata"
```

- Each schedule/session carries an IANA `timezone` column (e.g. `Asia/Jakarta`).
  Display/compute session-local times (07:00 WIB) using that zone, never the
  server or browser zone. Users also have a `timezone` (default `Asia/Jakarta`).

```go
loc, _ := time.LoadLocation(jadwal.Timezone)
localStart := sesi.MulaiUTC.In(loc)
```

---

## 10. File-layer contract (built once, Fase 0)

Tables: `mst_file` (one row per logical file, `uuid`, `kategori`, `reff_type`,
`reff_id`, `is_publik`, `status_proses`) + `mst_file_varian` (3 rows per image:
`original` / `medium` / `low`). Owning rows reference `mst_file.id` via
`*_file_id`.

Every image → 3 variants sized from `mst_pengaturan` grup `file`
(`file.varian_original_maks_px=4000`, `medium_px=1200`, `low_px=400`; never
hardcode; never upscale). Non-image docs store `original` only.

Endpoints:

- `POST /files` — multipart upload. Go MUST `ParseMultipartForm(maxBytes)` before
  `FormFile`; check `io.Copy`/`os.Create` errors; generate variants; write
  `mst_file` + `mst_file_varian`. Returns `{ uuid, kategori, variants:[...] }`.
- `GET /files/{uuid}/{varian}` — the ONLY way to read a **private** file
  (`is_publik=false`). Authenticated Go handler checks the caller's permission +
  cakupan against the file's `reff_type/reff_id` owner, then streams from disk.
  Never serve private files from an nginx static folder.
- `DELETE /files/{uuid}` — soft delete; `file.delete` = SA/Admin all,
  Mod/User `milik_sendiri`.

Upload checklist (rancangan Bab 6.6) — other modules depend on it:
Angular sends `FormData` **without** a manual `Content-Type` (browser sets the
boundary); Go `ParseMultipartForm` before `FormFile`; check `io.Copy`/`os.Create`
errors; nginx `client_max_body_size`; writable storage dir; CORS preflight allows
multipart. Disk layout: `storage/{kategori}/{tahun}/{bulan}/{uuid}/{varian}.{ext}`.

Modules never write files directly — they call the file service, store the
returned `file_id`, and resolve variants through it.

---

## 11. slamctl CLI

Separate binary/subcommands under `cmd/slamctl` — operational tasks that must NOT
be HTTP endpoints.

```
slamctl migrate up|down [n]     # run/rollback numbered migrations
slamctl seed [name]             # idempotent reference-data seeders
slamctl create-superadmin       # bootstrap the first super admin
```

`create-superadmin` creates one `anggota` + linked `users` row bound to the
Super Admin role. It **refuses to run if any user with an is_super role already
exists** (no backdoor after launch) and is never exposed over HTTP. There is no
self-registration anywhere: `users`/`anggota` rows are created only via the admin
panel by holders of `user.create` — the CLI exists solely to make the first one.

---

## 12. Logging & audit

- App logging: `pkg/logger` (Zap) — `logger.Info("msg", zap.String("k", v))`.
  `response.Internal` is the single place 500s are logged.
- Business audit: write a `log_aktivitas` row for state-changing actions
  (RBAC change, pengaturan update, absensi override, KTA cabut/cetak, login,
  create/ubah/hapus). Store `aktor_user_id`, `modul`, `aksi`, `reff_type/reff_id`,
  `ringkasan`, `nilai_lama`/`nilai_baru` (jsonb), `ip_address`, `user_agent`.
  System/job actions leave `aktor_user_id` null.
</content>
</invoke>
