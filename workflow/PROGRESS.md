# PROGRESS — laporan pengerjaan modul

Tracker per-prompt untuk build SLAM Team. Satu baris = satu prompt yang di-paste ke Claude Code (API atau APP). Urutan = build order (`EXECUTION.md` §3). Tandai selesai setiap kali sebuah prompt tuntas + terverifikasi + ter-commit.

**Progress: 6 / 49 selesai** · 🟡 0 proses · ⬜ 43 belum

Status: ⬜ Belum · 🟡 Proses · ✅ Selesai · ⛔ Terblokir

### Cara update (perintah ke AI)
Saat sebuah prompt selesai, minta AI: **"update PROGRESS: `<modul>` `<api|app>` selesai, commit `<hash>`"**. AI akan:
1. Ubah **Status** baris itu ⬜ → ✅ (atau 🟡 saat sedang dikerjakan).
2. Isi **Tanggal** (YYYY-MM-DD), **Commit** (hash pendek), dan **Catatan** bila ada.
3. Perbarui angka **Progress: N / 49** di atas.
4. Tambahkan satu baris ke **Riwayat** di bawah.
5. (Opsional) tandai **V** modul terkait di tabel `00-general-tasks.md`.

Aturan tetap: API dulu, baru APP. Verifikasi = checklist prompt + `make build && go vet ./...` (API) / `npm run build:local` (APP) + satu round-trip **form → API → DB** (row masuk Postgres, envelope benar, role terlarang → 403). Commit per prompt.

---

## Fase 0 — Fondasi

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 1 | 01-foundation | API | Opus · high | ✅ | 2026-09-09 | `46e69f6` | envelope+`apperr`+`FromError`, `model.Audit`, `pagination`, tzdata, CORS config, `slamctl` + migrator |
| 2 | 01-foundation | APP | Opus · high | ✅ | 2026-09-09 | `46e69f6` | `ApiService` + `Page<T>`, interceptor 401/403, design tokens (dark + merah), i18n `COMMON.*`/`VALIDATION.*` |
| 3 | 03-pengaturan-log | API | Sonnet · high | ✅ | 2026-09-09 | `c90d083` | migrasi `0002` + seeder 24 baris idempoten, `shared/audit` (writer lintas modul), gate `RequireAdmin` + `is_terkunci` SA-only di service |
| 4 | 03-pengaturan-log | APP | Sonnet · high | ✅ | 2026-09-09 | `62f1a0d` | halaman self-generating per `tipe_nilai`, `adminGuard` meniru `RequireAdmin` (belum ada `pengaturan.*`), `User.role` ditambah (opsional — auth Fase 0 belum menerbitkan klaim peran) |
| 5 | 02-file-management | API | Opus · high | ✅ | 2026-09-09 | `01d98d9` | migrasi `0003` identik `.dbml`; `imaging` (murni Go, WebP ditunda → JPEG); varian px dari `mst_pengaturan`; serve ber-izin via `JWTOptional`; seam `perm.Require` siap untuk 05 |
| 6 | 02-file-management | APP | Opus · high | ✅ | 2026-09-09 | `e0a52a5` | `FileService` di atas `ApiService`; `<app-file-upload>` (dropzone → emit `{id,uuid}`) + `<app-secure-image>` (blob privat, revoke object URL); `UploadResp` belum punya `id` ⇒ `FileRef.id` masih undefined |
| 7 | 04-auth-session | API | Opus · high | ⬜ |  |  |  |
| 8 | 04-auth-session | APP | Opus · high | ⬜ |  |  |  |

## Fase 1 — Hak akses dinamis

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 9 | 05-hak-akses | API | Opus · xhigh ★ | ⬜ |  |  |  |
| 10 | 05-hak-akses | APP | Opus · xhigh ★ | ⬜ |  |  | RBAC grid — layar tersulit |

## Fase 2 — Anggota + master data

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 11 | 08-instansi | API | Sonnet · high | ⬜ |  |  | migrasi sebelum anggota |
| 12 | 08-instansi | APP | Sonnet · high | ⬜ |  |  |  |
| 13 | 06-anggota | API | Sonnet · high | ⬜ |  |  |  |
| 14 | 06-anggota | APP | Sonnet · high | ⬜ |  |  |  |
| 15 | 07-user-management | API | Opus · high | ⬜ |  |  | anti-eskalasi |
| 16 | 07-user-management | APP | Opus · high | ⬜ |  |  |  |
| 17 | 09-unit | API | Sonnet · high | ⬜ |  |  |  |
| 18 | 09-unit | APP | Sonnet · high | ⬜ |  |  |  |
| 19 | 10-prestasi | API | Sonnet · high | ⬜ |  |  |  |
| 20 | 10-prestasi | APP | Sonnet · high | ⬜ |  |  |  |
| 21 | 11-inorga | API | Sonnet · high | ⬜ |  |  |  |
| 22 | 11-inorga | APP | Sonnet · high | ⬜ |  |  |  |
| 23 | 12-medsos | API | Sonnet · high | ⬜ |  |  |  |
| 24 | 12-medsos | APP | Sonnet · high | ⬜ |  |  |  |
| 25 | 13-dokumen | API | Sonnet · high | ⬜ |  |  |  |
| 26 | 13-dokumen | APP | Sonnet · high | ⬜ |  |  |  |
| 27 | 14-profile-club | API | Sonnet · high | ⬜ |  |  | singleton |
| 28 | 14-profile-club | APP | Sonnet · high | ⬜ |  |  |  |

## Fase 3 — Lokasi + jadwal

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 29 | 15-lokasi | API | Sonnet · high | ⬜ |  |  | migrasi sebelum jadwal |
| 30 | 15-lokasi | APP | Sonnet · high | ⬜ |  |  | map picker |
| 31 | 16-jadwal | API | Opus · xhigh ★ | ⬜ |  |  | generate-sesi |
| 32 | 16-jadwal | APP | Opus · xhigh ★ | ⬜ |  |  | kalender sadar timezone |

## Fase 4 — Penugasan + notifikasi

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 33 | 17-penugasan-notifikasi | API | Opus · high | ⬜ |  |  | fan-out + wajib_absen |
| 34 | 17-penugasan-notifikasi | APP | Opus · high | ⬜ |  |  | lonceng notifikasi |

## Fase 5 — Absensi + izin

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 35 | 19-izin | API | Opus · high | ⬜ |  |  | sebelum absensi (izin_id) |
| 36 | 19-izin | APP | Opus · high | ⬜ |  |  |  |
| 37 | 18-absensi | API | Opus · xhigh ★ | ⬜ |  |  | server-time authority |
| 38 | 18-absensi | APP | Opus · xhigh ★ | ⬜ |  |  | kamera + GPS |

## Fase 6 — KTA / NRA / QR

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 39 | 20-kta-nra-qr | API | Opus · xhigh ★ | ⬜ |  |  | NRA sequence + PNG |
| 40 | 20-kta-nra-qr | APP | Opus · xhigh ★ | ⬜ |  |  | halaman QR publik |

## Fase 7 — Laporan / rekap

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 41 | 21-laporan-rekap | API | Opus · high | ⬜ |  |  | export xlsx/pdf |
| 42 | 21-laporan-rekap | APP | Opus · high | ⬜ |  |  |  |

## Fase 8 — Landing + konten publik

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 43 | 22-kegiatan | API | Sonnet · high | ⬜ |  |  |  |
| 44 | 22-kegiatan | APP | Sonnet · high | ⬜ |  |  |  |
| 45 | 23-artikel | API | Sonnet · high | ⬜ |  |  |  |
| 46 | 23-artikel | APP | Sonnet · high | ⬜ |  |  |  |
| 47 | 24-landing-publik | API | Sonnet · high | ⬜ |  |  | endpoint publik |
| 48 | 24-landing-publik | APP | Sonnet · high | ⬜ |  |  | rute di luar authGuard |

## Fase 9 — PWA

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 49 | 25-pwa | APP | Sonnet · high | ⬜ |  |  | manifest + SW (tanpa API) |

---

## Riwayat

- **2026-09-09** — #1 `01-foundation` API selesai (Opus · high) — commit `46e69f6`.
- **2026-09-09** — #2 `01-foundation` APP selesai (Opus · high) — commit `46e69f6` (satu commit bersama API, karena kerja API belum ter-commit saat APP dikerjakan).
- **2026-09-09** — #3 `03-pengaturan-log` API selesai (Opus · high) — commit `c90d083`. Gate Fase 0 memakai `middleware.RequireAdmin` (peran), bukan `pengaturan.read/update`: tidak ada izin `pengaturan.*` yang di-seed karena pengaturan bukan salah satu dari 19 `mst_modul`. FK `modified_by`/`aktor_user_id` → `users.id` ditunda ke migrasi yang membuat tabel `users`.
- **2026-09-09** — #4 `03-pengaturan-log` APP selesai (Sonnet · high) — commit `62f1a0d`. Mengikuti API: bukan `permissionGuard`, tapi `adminGuard` baru yang meniru `middleware.RequireAdmin` persis (`is_super OR level<=10`). Auth module (scaffold Fase 0) belum menerbitkan `role_id`/`role_level`/`is_super` di `LoginResponse` maupun `GET /auth/me` walau `pkg/jwt.Claims` sudah punya field-nya — jadi `User.role` di frontend opsional dan setiap akun hari ini otomatis ditolak `adminGuard`, sama seperti `RequireAdmin` menolak token tanpa klaim peran. Diverifikasi manual di Chrome dengan sesi & respons API dipalsukan (bukan lewat login sungguhan, karena belum ada user admin yang bisa lolos gate).
- **2026-09-09** — #5 `02-file-management` API selesai (Opus · high) — commit `01d98d9`. Migrasi `0003_file_layer` diverifikasi byte-per-byte terhadap skema `.dbml` yang sudah ada di DB lokal (kolom, tipe, default, indeks, FK cascade — diff kosong); FK audit → `users.id` ditunda seperti di `0002`. Pustaka gambar **`disintegration/imaging`** (murni Go, tanpa cgo) + `rwcarlsen/goexif` untuk `metadata_exif`; **WebP ditunda sadar** → varian `medium`/`low` JPEG q80/q70, kecuali sumber ber-transparansi (PNG/GIF) yang tetap PNG agar logo tidak berlatar hitam. Ukuran piksel dibaca dari `mst_pengaturan` (diuji: setelan diubah → varian ikut) dan gambar tidak pernah diperbesar. `variants[]` berisi objek (doc modul §3), bukan string (`api-endpoints.md` §10). Ditambah `GET /files/{uuid}` untuk polling `status_proses` (opsional menurut doc §3). Route stream memakai `middleware.JWTOptional` baru supaya logo publik bisa dimuat tanpa token sementara berkas privat tetap 401/403. Gerbang Fase 0: `JWTAuth` saja + kepemilikan di service; seam `perm.Require("file.create"/"file.delete")` ditandai di `main.file.go` dan `service.authorizeRead` untuk disambung di `05-hak-akses` — **seeder RBAC di 05 wajib membuat kedua izin itu** walau `file` bukan salah satu dari 19 `mst_modul`. Config baru `STORAGE_ROOT` (root tak bisa ditulis ⇒ gagal saat boot, `router.Setup` kini mengembalikan error). Verifikasi: `go build`/`go vet`/`go test` bersih; `migrate up → down 1 → up → down all` bersih; 55 pemeriksaan end-to-end lulus (gambar ber-EXIF Orientation 6 tersimpan tegak 100×200 → 200×100, 1 `mst_file` + 3 `mst_file_varian`, PDF hanya `original`, Range 206, dedup, 403 lintas pemilik, 401 tanpa token, soft delete + `log_aktivitas`). Uji dijalankan di DB bersih `slamteam_migtest` karena `slamteam_db` lokal berisi skema hasil ekspor dbdiagram yang dimuat manual dan `schema_migrations`-nya tersangkut `version=1, dirty=true`.
- **2026-09-09** — #6 `02-file-management` APP selesai (Opus · high) — commit `e0a52a5`. Lapisan berkas di frontend adalah **service bersama + 2 komponen reusable**, bukan halaman menu: tidak ada rute, tidak ada item sidebar, tidak ada guard (izin `file.create` dimiliki semua peran berakun). `FileService` dibangun **di atas `ApiService`** (modul 01), bukan `HttpClient` langsung, supaya base URL, envelope `{success,message,data,errors}` dan Bearer dari `authInterceptor` tetap di satu tempat — `ApiService.upload()`/`.blob()` sudah menyediakan persis yang dibutuhkan. `FormData` dikirim **tanpa `Content-Type` manual** (Bab 6.6) dan `extra` kosong dibuang agar `reff_id` yang absen tidak sampai ke server sebagai string `"null"`. `<app-secure-image>` mengambil varian sebagai **blob** lalu mem-bind object URL, dan **me-revoke** saat `uuid`/`varian` berubah maupun saat destroy (URL lama disimpan di field biasa, bukan signal, supaya `effect` tidak bergantung pada keluarannya sendiri); permintaan in-flight dibatalkan lewat `onCleanup` agar scroll cepat tidak menimpa gambar baru dengan blob lama. `<app-file-upload>` menampilkan pratinjau lokal (`URL.createObjectURL` dari `File`) selama unggah, lalu menyerahkannya ke varian `low` dari server, dan meng-emit `{id, uuid}`; input `value` (uuid) memuat berkas tersimpan lewat `GET /files/{uuid}` supaya form **edit** terbuka sudah terisi. **Seam:** `dto.UploadResp` Fase 0 hanya mengembalikan `uuid`, tidak `id`, padahal kolom pemilik bertipe `*_file_id bigint` — `UploadedFile.id`/`FileRef.id` dibuat opsional dan ditandai di `file.model.ts`; modul pemilik pertama yang butuh menulis `*_file_id` harus menambah `ID` ke `UploadResp` di API (satu baris di `service.respond`), tanpa perubahan di frontend. `fileErrorText` menangani dua bentuk kegagalan sekaligus — `HttpErrorResponse` (non-2xx) dan envelope mentah yang dilempar `ApiService.unwrap` untuk `success:false` ber-HTTP-200 — dan **menolak** memakai `message` dari `Error` transport biasa supaya kalimat teknis tidak bocor ke pengguna. Desain per `design-tokens.md`: dropzone `--slam-surface-2` + border putus-putus `--slam-border` → `--slam-primary` saat hover/drag, cincin fokus merah dipindahkan ke `<label>` karena `<input type=file>`-nya `visually-hidden` (tetap fokusabel, bukan `display:none`), status selalu berpasangan glyph (⚠/⏳) bukan warna saja, target sentuh 44px ikut media query global. i18n `FILE.*` + `COMMON.CHOOSE_FILE`/`REMOVE`, IND & ENG diverifikasi sinkron (112 kunci). Verifikasi: `npm run build:local` lolos; `npm test` **21/21** lolos — termasuk bukti revoke object URL saat ganti uuid & saat destroy (tidak ada blob menggantung), `FormData` tanpa `Content-Type`, tolak >15 MB dan tipe salah tanpa request, soft delete + emit `cleared`, dan pemuatan uuid tersimpan. Progress unggah masih **indeterminate** (ditandai `ponytail:`): `ApiService` mengembalikan body yang sudah di-parse, bukan aliran `HttpEvent`; ganti ke `HttpRequest` + `reportProgress` bila dokumen besar kelak butuh persentase.
