# PROMPT (paste ke Claude Code) — Build API modul **Medsos** (`slam-team-api`)

Kamu bekerja di `D:/xampp/htdocs/slam-team-apl/slam-team-api` (Go 1.24, Gin, GORM/Postgres `slamteam_db`, JWT HS256, Zap). Bangun modul **Medsos** (master data tautan media sosial per anggota) mengikuti module pattern yang sudah ada. Jangan mengubah layout, jangan pakai `AutoMigrate`, jangan bocorkan error GORM ke klien.

## Baca dulu (otoritatif — jangan menebak)
1. `D:/xampp/htdocs/slam-team-apl/slamteam_db.dbml` — tabel **`mst_medsos`** (baris ~1122–1137) dan relasi `Ref: mst_medsos.anggota_id > anggota.id` (baris ~1238).
2. `D:/xampp/htdocs/slam-team-apl/workflow/modules/12-medsos.md` — spesifikasi modul ini (kolom, endpoint, hak akses, aturan bisnis, alur).
3. `D:/xampp/htdocs/slam-team-apl/workflow/_shared/conventions-api.md` — konvensi WAJIB (module layout §1, envelope §2, error §3, validasi §4, pagination §5, GORM/soft-delete §6, migrasi/seeder §7, RBAC/cakupan §8, audit §12).
4. Modul sejenis yang sudah ada (mis. `internal/modules/core/instansi` atau `lokasi`) sebagai contoh gaya — **reuse** pola & helper yang sudah ada; jangan bikin abstraksi baru.

## Skema `mst_medsos` (persis dari DBML — jangan ganti nama kolom)
`id bigint pk`, `anggota_id bigint NOT NULL` (FK → `anggota.id`), `kode varchar(50)`, `tipe int`, `icon varchar(50)` (**nama ikon dari icon-set, mis. "instagram" — BUKAN berkas**), `jenis_medsos varchar(50)`, `konten_medsos varchar(255)`, `is_deleted bool default false`, `deleted_at timestamptz`, `deleted_by bigint`, `created_at timestamptz default now()`, `created_by bigint`, `modified_at timestamptz`, `modified_by bigint`.

Tidak ada `*_file_id` → modul ini TIDAK menyentuh file-layer.

## Endpoint (rancangan Bab 9 — CRUD standar, semua di bawah `/api/v1`, guard JWT + `medsos.<aksi>`)
- `GET    /medsos`      → `medsos.read`   (list+filter+paginate: `page, per_page, q, sort, anggota_id, jenis_medsos`)
- `GET    /medsos/{id}` → `medsos.read`   (detail)
- `POST   /medsos`      → `medsos.create`
- `PUT    /medsos/{id}` → `medsos.update`
- `DELETE /medsos/{id}` → `medsos.delete` (**soft delete**)

## Hak akses & cakupan (permission-matrix Bab 3.5)
SA/Admin/Moderator = **CRUD cakupan `semua`**. User = **CRUD cakupan `milik_sendiri`** (hanya medsos yang `anggota_id`-nya = `claims.AnggotaID`). Guest = read (tanpa rute tulis). `is_super` bypass. Terapkan cakupan di **service/repository** (baca DAN tulis/hapus), bukan sekadar UI — baca `c.Get("cakupan:medsos.<aksi>")` yang di-set `PermGuard`.

## File yang dibuat — `internal/modules/core/medsos/`
1. `domain/medsos.go`
   - struct `Medsos` (map ke `mst_medsos`) + `func (Medsos) TableName() string { return "mst_medsos" }`. Embed `Audit` (created/modified/deleted + `IsDeleted`) sesuai conventions §6. Kolom nullable (`kode`, `icon`, `modified_*`, `deleted_*`) → pointer.
2. `dto/medsos.go`
   - `CreateMedsosReq{ AnggotaID int64 `binding:"required"`; JenisMedsos string `binding:"required,max=50"`; KontenMedsos string `binding:"required,max=255"`; Icon string `binding:"omitempty,max=50"`; Kode string `binding:"omitempty,max=50"`; Tipe int `binding:"omitempty,min=0"` }`
   - `UpdateMedsosReq` — sama TANPA `AnggotaID` (immutable).
   - `ListMedsosQuery` — embed `ListQuery`(`page,per_page,q,sort`) + `AnggotaID int64 form:"anggota_id"` + `JenisMedsos string form:"jenis_medsos"`.
   - `MedsosResponse` (kolom §skema kecuali soft-delete internal) + `ToResponse(domain.Medsos)`.
3. `repository/medsos_repository.go`
   - `List` (Count lalu paged Find pada query ter-scope sama; selalu `is_deleted=false`; filter `anggota_id`/`jenis_medsos`; `q` ILIKE `jenis_medsos|konten_medsos|kode`; whitelist sort `created_at|jenis_medsos|tipe`, default `-created_at`; cap per_page 100).
   - `FindByID` (`is_deleted=false`; NotFound→`ErrNotFound`).
   - `Create`, `Update`, `SoftDelete` (`is_deleted=true, deleted_at=now(), deleted_by=?`).
   - `AnggotaExists(anggotaID)` (anggota ada & `is_deleted=false`).
   - Terima fungsi/scope cakupan dari service.
4. `service/medsos_service.go`
   - Create: validasi `AnggotaID` ada; jika cakupan `milik_sendiri` paksa `AnggotaID==claims.AnggotaID` else `ErrForbidden`; default `tipe=0`; trim `konten_medsos`; `icon` kosong boleh diisi dari `jenis_medsos`; set `created_by`.
   - Update: `FindByID` ter-scope; cek kepemilikan untuk `milik_sendiri`; `anggota_id` tak diubah; set `modified_by/at`.
   - Delete: `FindByID` ter-scope; soft delete; set `deleted_by`.
   - List/Detail: terapkan scope cakupan (`milik_sendiri` → `Where("anggota_id = ?", claims.AnggotaID)`; `semua`/super → tanpa filter).
   - Tulis `log_aktivitas` untuk create/update/delete (modul `medsos`, `reff_type=mst_medsos`, `reff_id`, `nilai_lama`/`nilai_baru`).
   - Sentinel errors: `ErrNotFound/ErrForbidden/ErrValidation` (pakai yang sudah ada di shared bila tersedia).
5. `handler/medsos_handler.go`
   - `List/Detail/Create/Update/Delete`: bind (`ShouldBindQuery`/`ShouldBindJSON`) → `validator.Explain` pada error (`response.Unprocess`) → panggil service dgn `middleware.Claims(c)` → `response.OK`/`Created`, error via `response.FromError`. Handler tak menyentuh `*gorm.DB`.
6. `main.medsos.go`
   - `Initialize(db, jwtMgr, perm)` wiring repo→service→handler; `SetupRoutes(rg)` seperti blok di modul MD.
7. Registrasi di `internal/router/router.go`: `medsos.Initialize(db, jwtMgr, permGuard).SetupRoutes(apiV1)`.

## Migrasi & seeder
- `migrations/00XX_mst_medsos.up.sql` + `.down.sql` (nomor berikutnya yang tersedia, setelah `anggota`): `CREATE TABLE mst_medsos (...)` persis kolom DBML; `FOREIGN KEY (anggota_id) REFERENCES anggota(id)`; `CREATE INDEX idx_mst_medsos_anggota_id ON mst_medsos(anggota_id) WHERE is_deleted = false`. `.down.sql` drop index + table.
- Pastikan seeder RBAC memuat permission `medsos.create/read/update/delete` untuk modul `medsos` dan baris `role_permission` sesuai matriks (SA/Admin/Mod=`semua`, User=`milik_sendiri`). Idempoten (`ON CONFLICT`). Jika seeder RBAC global sudah menghasilkannya, cukup verifikasi.

## Edge cases (wajib ditangani)
- `anggota_id` tak ada/terhapus → 422/404 (tanpa insert).
- User `milik_sendiri` POST `anggota_id` orang lain → 403.
- Update/Delete id di luar scope user → 404 (jangan bocorkan keberadaan).
- `jenis_medsos` kosong / `konten_medsos` > 255 → 422 via binding.
- Tidak ada background job, tidak ada file.

## Checklist verifikasi (jalankan sebelum selesai)
- [ ] `make build` (atau `go build ./...`) sukses; `go vet ./...` bersih.
- [ ] `slamctl migrate up` membuat `mst_medsos`; `migrate down` merollback bersih.
- [ ] Login dapat token, lalu keempat endpoint merespons dengan envelope `{success,message,data,errors}`:
  - `POST /medsos` (payload valid) → 201, baris muncul di DB (`SELECT * FROM mst_medsos`).
  - `GET /medsos?anggota_id=<id>&per_page=5` → page benar, hanya `is_deleted=false`.
  - `PUT /medsos/{id}` → field berubah, `modified_by/at` terisi.
  - `DELETE /medsos/{id}` → `is_deleted=true`, hilang dari list.
- [ ] Role tanpa `medsos.*` → 403. User `milik_sendiri`: CRUD medsos sendiri OK; medsos anggota lain → 404/403. `is_super` bypass.
- [ ] Tidak ada teks error GORM/PG mentah yang bocor ke klien.
- [ ] `log_aktivitas` bertambah untuk create/update/delete.

Ikuti nama rute, payload key, dan nama kolom PERSIS seperti DBML + peta endpoint. Jangan meniru bentuk API app lama.
