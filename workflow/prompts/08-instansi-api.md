# PROMPT — Build API module `instansi` (Master Data Instansi / Sekolah)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

You are building the **`instansi`** (Master Data Instansi / Sekolah) module into the existing **slam-team-api** scaffold (Go 1.24 · Gin · GORM/PostgreSQL `slamteam_db` · JWT HS256 · Zap · go-playground/validator). Build INTO the existing layout — do not invent a new one.

## 0. Read first (authoritative, in this order)
1. `D:/xampp/htdocs/slam-team-apl/slamteam_db.dbml` — table **`mst_instansi`** (around line 983) and its relations (`anggota.instansi_id → mst_instansi.id` **restrict**, `mst_instansi.logo_utama_file_id`/`logo_tambahan_file_id → mst_file.id`). The DBML WINS on any conflict.
2. `D:/xampp/htdocs/slam-team-apl/workflow/modules/08-instansi.md` — this module's spec.
3. `D:/xampp/htdocs/slam-team-apl/workflow/_shared/conventions-api.md` — module layout, envelope, error sentinels, list/pagination, GORM/soft-delete, RBAC middleware, migrations/seeders, file-layer contract, logging. Follow it exactly.
4. Skim `slam-team-api/CLAUDE.md`. Use `auth` as the reference module.

## 1. Scope — what to build
Standard master-data CRUD for `mst_instansi`, RBAC-guarded, soft-delete, with two logo file references. **No background jobs.**

### Table `mst_instansi` (columns EXACTLY as `.dbml`)
`id` bigint pk · `kode` varchar(50) unique · `nama` varchar(150) · `nama_club` varchar(150) · `alamat` text · `no_telepon` varchar(30) · `logo_utama_file_id` bigint (FK mst_file) · `logo_tambahan_file_id` bigint (FK mst_file) · `tanggal_bergabung` date · `status` int (1=aktif, 0=nonaktif, default 1) · soft-delete triplet `is_deleted/deleted_at/deleted_by` · audit `created_at/created_by/modified_at/modified_by`.

### Endpoints (all `/api/v1`, JWT, `RequirePermission`)
| Method | Path | Permission |
|--------|------|-----------|
| GET | `/instansi` | `instansi.read` |
| GET | `/instansi/:id` | `instansi.read` |
| POST | `/instansi` | `instansi.create` |
| PUT | `/instansi/:id` | `instansi.update` |
| DELETE | `/instansi/:id` | `instansi.delete` |

## 2. Files to create under `internal/modules/core/instansi/`

### `domain/instansi.go`
```go
package domain

import "time"

type Instansi struct {
	ID                  int64      `gorm:"column:id;primaryKey"`
	Kode                string     `gorm:"column:kode"`
	Nama                string     `gorm:"column:nama"`
	NamaClub            string     `gorm:"column:nama_club"`
	Alamat              string     `gorm:"column:alamat"`
	NoTelepon           string     `gorm:"column:no_telepon"`
	LogoUtamaFileID     *int64     `gorm:"column:logo_utama_file_id"`
	LogoTambahanFileID  *int64     `gorm:"column:logo_tambahan_file_id"`
	TanggalBergabung    *time.Time `gorm:"column:tanggal_bergabung;type:date"`
	Status              int        `gorm:"column:status;default:1"`

	CreatedAt  time.Time  `gorm:"column:created_at"`
	CreatedBy  *int64     `gorm:"column:created_by"`
	ModifiedAt *time.Time `gorm:"column:modified_at"`
	ModifiedBy *int64     `gorm:"column:modified_by"`
	IsDeleted  bool       `gorm:"column:is_deleted;default:false"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"`
	DeletedBy  *int64     `gorm:"column:deleted_by"`
}

func (Instansi) TableName() string { return "mst_instansi" }
```
(If the scaffold already has a shared `Audit` embed as in conventions-api §6, use it instead of repeating the audit fields.)

### `dto/instansi.go`
```go
package dto

type CreateInstansiReq struct {
	Kode               string  `json:"kode"                 binding:"required,max=50"`
	Nama               string  `json:"nama"                 binding:"required,max=150"`
	NamaClub           string  `json:"nama_club"            binding:"max=150"`
	Alamat             string  `json:"alamat"`
	NoTelepon          string  `json:"no_telepon"           binding:"max=30"`
	LogoUtamaFileID    *int64  `json:"logo_utama_file_id"   binding:"omitempty,gt=0"`
	LogoTambahanFileID *int64  `json:"logo_tambahan_file_id" binding:"omitempty,gt=0"`
	TanggalBergabung   *string `json:"tanggal_bergabung"    binding:"omitempty,datetime=2006-01-02"`
	Status             *int    `json:"status"               binding:"omitempty,oneof=0 1"`
}

type UpdateInstansiReq = CreateInstansiReq // full-replace PUT; same shape

type InstansiResp struct {
	ID                 int64   `json:"id"`
	Kode               string  `json:"kode"`
	Nama               string  `json:"nama"`
	NamaClub           string  `json:"nama_club"`
	Alamat             string  `json:"alamat"`
	NoTelepon          string  `json:"no_telepon"`
	LogoUtamaFileID    *int64  `json:"logo_utama_file_id"`
	LogoUtamaUUID      *string `json:"logo_utama_uuid"`
	LogoTambahanFileID *int64  `json:"logo_tambahan_file_id"`
	LogoTambahanUUID   *string `json:"logo_tambahan_uuid"`
	TanggalBergabung   *string `json:"tanggal_bergabung"` // YYYY-MM-DD
	Status             int     `json:"status"`
	JumlahAnggota      int64   `json:"jumlah_anggota"`
	CreatedAt          string  `json:"created_at"`
	ModifiedAt         *string `json:"modified_at"`
}

type ListInstansiQuery struct {
	Page    int    `form:"page,default=1"      binding:"min=1"`
	PerPage int    `form:"per_page,default=20" binding:"min=1,max=100"`
	Q       string `form:"q"`
	Sort    string `form:"sort"`   // whitelist: kode, nama, status, created_at (prefix - = DESC)
	Status  *int   `form:"status"  binding:"omitempty,oneof=0 1"`
}
```

### `repository/instansi_repository.go`
- `Create(ctx, *domain.Instansi) error`, `Update`, `SoftDelete(ctx, id, actor)`, `FindByID(ctx, id) (*domain.Instansi, joined uuids + jumlahAnggota, error)`, `List(ctx, ListInstansiQuery) (items, total, error)`.
- Every read filters `is_deleted = false`.
- `List`: build one base query, `Count`, then paged `Find` with `Offset/Limit`. `q` → `WHERE (kode ILIKE ? OR nama ILIKE ? OR nama_club ILIKE ?)`. Optional `status` filter. `sort` **whitelisted** to `{kode,nama,status,created_at}` — never interpolate raw input; default `-created_at`.
- Join logo uuids: `LEFT JOIN mst_file lu ON lu.id = mst_instansi.logo_utama_file_id` and `lt` for tambahan; select `lu.uuid`, `lt.uuid`. Add `jumlah_anggota` via correlated subquery `(SELECT COUNT(*) FROM anggota a WHERE a.instansi_id = mst_instansi.id AND a.is_deleted = false)`.
- `ExistsKode(ctx, kode string, exceptID int64) (bool, error)` — `WHERE kode = ? AND is_deleted = false AND id <> ?`.
- `CountAnggotaAktif(ctx, instansiID int64) (int64, error)`.
- `FileExistsLogo(ctx, id int64) (bool, error)` — optional: `WHERE id=? AND kategori='logo' AND is_deleted=false` on `mst_file`.
- Translate `gorm.ErrRecordNotFound` → service sentinel `ErrNotFound`; let other DB errors bubble.

### `service/instansi_service.go`
Business rules (see conventions-api §3 sentinels — reuse shared `ErrNotFound/ErrConflict/...`):
- **Create**: reject if `ExistsKode(kode, 0)` → `ErrConflict`. If a logo id is provided, verify it via `FileExistsLogo` (else `ErrValidation`). Default `status` = 1. Parse `tanggal_bergabung` (`2006-01-02`) → `*time.Time`. Set `CreatedBy = claims.UserID`. Persist. Write `log_aktivitas` (modul `instansi`, aksi `create`, reff_id, nilai_baru).
- **Update**: load by id (`ErrNotFound`). Reject if `ExistsKode(kode, id)`. Same logo/date handling. Set `ModifiedBy`, `ModifiedAt = now()`. Persist. Log with nilai_lama/nilai_baru.
- **Delete**: load by id. If `CountAnggotaAktif(id) > 0` → `ErrConflict` ("instansi masih dipakai oleh N anggota"). Else soft delete (`IsDeleted=true, DeletedAt=now(), DeletedBy=claims.UserID`). Log.
- **Map** domain → `InstansiResp` (uuids + jumlah_anggota + formatted dates). RBAC is enforced by middleware; do not re-check permission here, but DO enforce the business guards.

### `handler/instansi_handler.go`
- `List`, `Detail`, `Create`, `Update`, `Delete`. Bind (`ShouldBindQuery`/`ShouldBindJSON`), read actor with `middleware.Claims(c)`, call service, return via `response.OK/Created/FromError`. On bind error → `response.Unprocess(c, "validasi gagal", validator.Explain(err))`. List returns the paginated envelope `{items,page,per_page,total,last_page}`.

### `main.instansi.go`
```go
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard) *Module { ... }

func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/instansi", middleware.JWTAuth(m.jwtMgr))
	g.GET("",        m.perm.Require("instansi.read"),   m.h.List)
	g.GET("/:id",    m.perm.Require("instansi.read"),   m.h.Detail)
	g.POST("",       m.perm.Require("instansi.create"), m.h.Create)
	g.PUT("/:id",    m.perm.Require("instansi.update"), m.h.Update)
	g.DELETE("/:id", m.perm.Require("instansi.delete"), m.h.Delete)
}
```

### Register in `internal/router/router.go`
Add near the other module registrations under `/api/v1`:
```go
instansi.Initialize(db, jwtMgr, permGuard).SetupRoutes(apiV1)
```

## 3. Migration + seeder
Create numbered SQL (match the existing `migrations/` numbering; place it **after RBAC and BEFORE `anggota`** — `anggota.instansi_id` FK restrict points here):

`migrations/00XX_mst_instansi.up.sql`
- `CREATE TABLE mst_instansi (...)` columns/types EXACTLY per `.dbml`; `status int DEFAULT 1`; `created_at timestamptz DEFAULT now()`.
- `CREATE UNIQUE INDEX ux_mst_instansi_kode ON mst_instansi(kode) WHERE is_deleted = false;`
- `CREATE INDEX ix_mst_instansi_status ON mst_instansi(status);`
- FKs (nullable): `logo_utama_file_id`/`logo_tambahan_file_id` → `mst_file(id)`; `created_by`/`modified_by`/`deleted_by` → `users(id)`.

`migrations/00XX_mst_instansi.down.sql` — drop indexes + `DROP TABLE mst_instansi;`.

Seeder (idempotent, under `cmd/slamctl` seed set, e.g. `slamctl seed instansi`):
```sql
INSERT INTO mst_instansi (kode, nama, nama_club, status, created_at)
VALUES ('SLAM', 'Scouting Legion Airsofter Malang', 'SLAM Team', 1, now())
ON CONFLICT (kode) DO NOTHING;
```
(Keyed on `kode`; safe to re-run.)

Ensure the RBAC permission seed already contains `instansi.create/read/update/delete` mapped to the matrix (SA/Admin CRUD, Mod/User R). If Fase 1's seeder owns the permission rows, just confirm — do not duplicate.

## 4. Conventions to honor (non-negotiable)
- Envelope only via `internal/shared/response`; sentinel errors mapped by `response.FromError` (add `Forbidden`/`NotFound` to the package if still missing — do NOT fork it).
- NO `AutoMigrate`. Soft-delete via explicit columns (not `gorm.DeletedAt`). All reads filter `is_deleted=false`.
- Times are `timestamptz` UTC; `tanggal_bergabung` is a pure `date`.
- File ids reference `mst_file.id`; never store paths on `mst_instansi`.
- Whitelist `sort`; cap `per_page` at 100.
- Write `log_aktivitas` for create/update/delete (actor, modul, aksi, reff, nilai_lama/nilai_baru, ip, user_agent).

## 5. Verification checklist (run before declaring done)
- [ ] `make build` (or `go build ./...`) and `go vet ./...` pass.
- [ ] Migration up/down applies cleanly; table matches `.dbml`; partial-unique index present.
- [ ] Seeder run twice → exactly one `SLAM` row (idempotent).
- [ ] `POST /api/v1/instansi` with a valid body creates a row (`created_by` set); returns 201 + `InstansiResp`.
- [ ] Duplicate `kode` → 409; missing `nama`/`kode` → 422 with field errors.
- [ ] `GET /instansi?page=1&per_page=10&q=&status=1&sort=nama` returns paginated envelope with `total`+`last_page`; `logo_*_uuid` and `jumlah_anggota` populated.
- [ ] `PUT /instansi/:id` updates (`modified_at` set); `DELETE /instansi/:id` soft-deletes (`is_deleted=true`), and is **blocked 409** when an active `anggota` references it.
- [ ] As Moderator/User token: `POST/PUT/DELETE /instansi` → 403; `GET /instansi` → 200. Super Admin bypasses all.
- [ ] `log_aktivitas` rows written for create/update/delete.
- [ ] No raw GORM/PG error text leaks to the client; 500s logged once via `response.Internal`.
