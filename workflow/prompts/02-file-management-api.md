# PROMPT — Build API module `file` (3-variant file layer) (Fase 0)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

You are building the **file layer** into **slam-team-api** (Go 1.24 · Gin · GORM/PostgreSQL · Zap). Built ONCE in Fase 0 — every later module depends on it. Build INTO the existing layout.

## 0. Read first (authoritative)
1. `D:/xampp/htdocs/slam-team-apl/slamteam_db.dbml` — tables **`mst_file`** (line 639) and **`mst_file_varian`** (line 685). DBML WINS.
2. `D:/xampp/htdocs/slam-team-apl/workflow/modules/02-file-management.md` — this module's spec.
3. `D:/xampp/htdocs/slam-team-apl/workflow/_shared/conventions-api.md` — §10 file-layer contract + §2 envelope + §7 migrations. `_shared/api-endpoints.md` §10.
4. Use `auth` as the reference module.

## 1. Scope
`mst_file` (uuid, nama_asli, nama_slug, ekstensi, mime_type, ukuran_byte, hash_sha256, lebar_px/tinggi_px, kategori, reff_type/reff_id, storage_driver, path_dasar, is_publik, status_proses, metadata_exif, soft-delete) + `mst_file_varian` (file_id, varian, path, url_publik, lebar_px/tinggi_px, ukuran_byte, mime_type, kualitas; UNIQUE(file_id, varian)). Columns EXACTLY per `.dbml`.

### Endpoints (`/api/v1`)
| Method | Path | Permission |
|--------|------|-----------|
| POST | `/files` | `file.create` (all logged-in) — **multipart** |
| GET | `/files/:uuid/:varian` | bearer (permission+owner checked in handler) |
| DELETE | `/files/:uuid` | `file.delete` (SA/Admin all, Mod/User `milik_sendiri`) |

## 2. Files under `internal/modules/core/file/`
- **domain**: `File` (`mst_file`, `TableName`), `FileVarian` (`mst_file_varian`).
- **dto**: `UploadResp` (`uuid, nama_asli, ekstensi, mime_type, ukuran_byte, variants[], url`). Upload fields bound from multipart: `file` (the file), `modul`/`kategori`, `reff_id`.
- **repository**: `Create` (file + variants in a tx), `FindByUUID`, `FindVarian(uuid, varian)`, `ExistsByHash(sha256)` (dedupe), `SoftDelete`.
- **service** — the processing pipeline:
  - **Multipart handling (Bab 6.6):** `ParseMultipartForm(maxBytes)` BEFORE `FormFile`; check `io.Copy`/`os.Create` errors.
  - Detect **type from content** (not extension); compute **sha256**; **auto-orient EXIF**; read `lebar_px/tinggi_px`.
  - For images produce **3 variants** from `mst_pengaturan` grup `file` (`file.varian_original_maks_px=4000`/q90 source fmt, `medium_px=1200`/q80 WebP+JPEG, `low_px=400`/q70 WebP+JPEG) — **never upscale**. Non-image → `original` only.
  - Disk layout `storage/{kategori}/{tahun}/{bulan}/{uuid}/{varian}.{ext}` (writable dir). Set `is_publik` per category (private by default).
  - Image lib: pick one (e.g. `disintegration/imaging` or `h2non/bimg`) — record the choice in `workflow/99-alignment-report.md`.
- **serve handler** `GET /files/:uuid/:varian`: load file; if `is_publik=false`, check the caller's permission + **cakupan against the file's `reff_type/reff_id` owner**; stream from disk with correct `Content-Type`. **Never** serve private files from a static folder.
- **main.file.go + router**: `RequirePermission("file.create")` on POST; bearer on GET; `file.delete` on DELETE.

## 3. Migration + seeder
- `mst_file` + `mst_file_varian` EXACTLY per `.dbml` (uuid default `gen_random_uuid()`, `UNIQUE(file_id, varian)`, FK `mst_file_varian.file_id → mst_file.id` **cascade**, index `(reff_type, reff_id)`, `hash_sha256`, `kategori`). Depends on `mst_pengaturan` (variant sizes) — order after `0002_mst_pengaturan`.

## 4. Conventions
Envelope; no AutoMigrate; soft-delete; multipart checklist verbatim; never store paths on owning rows (they store `mst_file.id`); `log_aktivitas` on delete.

## 5. Verification checklist
- [ ] `go build`/`go vet` pass; migration up/down clean; matches `.dbml`.
- [ ] `POST /files` with an image → 1 `mst_file` + 3 `mst_file_varian`; returns uuid + variants; **no manual Content-Type** required from client.
- [ ] `POST /files` with a PDF → `original` only.
- [ ] `GET /files/:uuid/:varian` streams; a **private** file is refused (403) to a caller who fails permission/cakupan against the owner.
- [ ] EXIF auto-oriented; sha256 stored; identical re-upload deduped (by hash) if implemented.
- [ ] Variant sizes read from `mst_pengaturan` (not hardcoded); images never upscaled.
- [ ] `DELETE /files/:uuid` soft-deletes; Mod/User limited to own files; `log_aktivitas` written.
