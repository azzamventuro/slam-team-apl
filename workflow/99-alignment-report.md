# 99 — Alignment Report (rencana ↔ realita)

> Catatan hidup: rekam **setiap kali realita menyimpang dari rencana** — perbaikan skema,
> koreksi prompt, keputusan yang diambil di tengah build. Diperbarui setiap langkah build
> yang menyimpang. Sumber otoritatif tetap: `slamteam_db.dbml` (skema) > `.rancangan.txt`
> (kebutuhan) > `_shared/conventions-*.md` > file ini. Jika file ini bertentangan dengan
> `.dbml`, **`.dbml` menang** — dan catat konfliknya di §3.

---

## 1. Status build saat ini (per 2026-09-08)

Kode nyata di repo, dibaca langsung dari `slam-team-api/` dan `slam-team-app/`:

| Area | Status kode | Catatan |
|------|-------------|---------|
| **API skeleton** (Fase 0) | ✅ ada | `cmd/api/main.go`, `internal/{config,database,middleware,router}`, `shared/response`, `shared/redis`, `pkg/{jwt,logger,validator,utils}`. |
| **Auth** (Fase 0) | ⚠️ sebagian | `internal/modules/core/auth` ada (login + `/me`) tapi **masih memakai tabel `users` "mainan"** — belum skema `.dbml` penuh (`anggota_id`, `role_id`, dst) dan **belum ada `sesi_login`/refresh**. Diselesaikan oleh `04-auth-session`. |
| **response envelope** | ⚠️ sebagian | `OK/Created/BadRequest/Unprocess/Unauthorized/Internal` ada. **`Forbidden`/`NotFound`/`FromError` BELUM ada** → ditambah di Fase 1 (`05-hak-akses`). |
| **File layer** (Fase 0) | ❌ belum | `02-file-management` belum di-build. |
| **`mst_pengaturan` + `log_aktivitas`** (Fase 0) | ❌ belum | `03-pengaturan-log` belum di-build. |
| **slamctl / migrations** | ❌ belum | Belum ada `cmd/slamctl`, belum ada folder `migrations/`, belum ada import `_ "time/tzdata"`. Diselesaikan `01-foundation`. |
| **RBAC dinamis** (Fase 1) | ❌ belum | 5 tabel, `PermGuard`, seed matriks belum ada. `router.go` baru mendaftarkan `auth`. |
| **APP skeleton** (Fase 0) | ✅ ada | Angular 21 zoneless, `core/{guards,interceptors,services,models}`, `layouts/{vertical,topbar,sidebar,footer}`, `pages/{auth/login,dashboard}`. |
| **APP RBAC** (`PermissionService`, `*hasPermission`, sidebar dinamis) | ❌ belum | Dibangun di `05-hak-akses` (APP). |
| **Semua modul Fase 2–9** | ❌ belum | Baru pada tahap perencanaan (`workflow/`). |

**Kesimpulan:** yang sudah ada di kode adalah **skeleton (01) + auth mentah (04 sebagian)**.
Sisa Fase 0 (file, pengaturan/log, slamctl, tzdata) + Fase 1–9 masih rencana.

---

## 2. Status kelengkapan `workflow/` (playbook)

Sesi ini menuntaskan playbook agar 25 work-item punya set lengkap **doc + prompt api + prompt app**.

| Bagian | Sebelum sesi ini | Ditambahkan sesi ini |
|--------|------------------|----------------------|
| `modules/*.md` (rencana) | 01,02,03,05,06,07,08,09,10,11,12,13,14 | **04, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25** |
| `prompts/*-api.md` | 08, 12 | 01,02,03,04,05,06,07,09,10,11,13,14,15,16,17,18,19,20,21,22,23,24 |
| `prompts/*-app.md` | 08 | 01,02,03,04,05,06,07,09,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25 |
| `99-alignment-report.md` | — | file ini |

> `25-pwa` sengaja **tidak** punya prompt `-api` (kerja platform front-end, tanpa endpoint baru).
> `12-medsos` sudah punya doc + prompt api sebelumnya; sesi ini hanya menambah prompt app.

Verifikasi: prompt-prompt baru dibuat oleh agen generator lalu **diaudit adversarial**
terhadap `.dbml` + `_shared/api-endpoints.md` + `_shared/permission-matrix.md`. Temuan audit
(jika ada) dan perbaikannya dicatat di §4.

---

## 3. Keputusan & konflik skema (rencana ↔ `.dbml`)

Rekam di sini setiap tempat prosa/rancangan berbeda dari `.dbml`, dan bagaimana diselesaikan
(default: `.dbml` menang).

- **`artikel.status`** — kolom status pada `artikel` bernama **`status_kegiatan` (varchar)**,
  sama seperti `kegiatan`, BUKAN `status_artikel`. Ikuti nama `.dbml` apa adanya.
- **`users` vs `user`** — tabel bernama **`users`** (bukan `user`) karena `USER` reserved di Postgres.
- **NRA milik kartu** — NRA (`kta.no_kta`) milik **kartu**, bukan orang; `anggota.no_induk`
  hanya salinan NRA aktif. `siswa_ke_dewasa` ⇒ 2 kartu, 2 NRA berbeda.
- **`absensi_rekap` opsional** — awalnya hitung rekap langsung dari `absensi`; materialisasi
  tabel `absensi_rekap` hanya bila laporan bulanan melambat.
- _(tambah entri baru di bawah saat build menemukan penyimpangan)_

---

## 4. Koreksi prompt / plan (log audit)

Diisi dari hasil audit dan temuan saat eksekusi build.

- **2026-09-08 — Metode generasi playbook.** 12 doc modul (04, 15–25) + 46 prompt
  (api+app) ditulis **langsung (inline)** dengan grounding dari `slamteam_db.dbml`
  yang dibaca penuh. (Percobaan awal via workflow 30-agen paralel gagal karena
  **session usage limit** — 0 file dihasilkan; diganti metode inline yang lebih
  hemat dan terkendali.)
- **Self-audit (grep) — bersih:** tidak ada nama tabel salah (mis. tak ada
  `mst_anggota`/`mst_kta`); `status_artikel` hanya muncul sebagai peringatan
  "jangan ganti nama" (kolom sebenarnya `artikel.status_kegiatan varchar`);
  seluruh 25 prompt app memanggil `design-taste-frontend`; seluruh 24 prompt api
  merujuk `slamteam_db.dbml`. Prompt api yang tidak memakai kata literal
  `RequirePermission` (01 fondasi, 04 auth, 24 publik) memang **tanpa gerbang izin
  by design**; sisanya mencantumkan kode `modul.aksi` per-endpoint + merujuk
  `conventions-api.md §8`.
- `25-pwa` sengaja tanpa prompt `-api` (kerja platform front-end).

---

## 5. Pertanyaan terbuka / keputusan yang tertunda

- **Library peta** (`15-lokasi`) — Leaflet + OSM (tanpa API key) vs input lat/long numerik.
  Prompt menawarkan keduanya; finalisasi saat build UI.
- **Library render KTA PNG** (`20-kta-nra-qr`) — pustaka Go untuk generate CR80 300dpi
  (mis. `fogleman/gg` / `golang.org/x/image`) + QR; finalisasi saat build.
- **Library export** (`21-laporan-rekap`) — Excel (`excelize`) + PDF (mis. `gofpdf` /
  `maroto`); finalisasi saat build.
- **Editor konten artikel** (`23-artikel`) — textarea vs rich-text ringan; hindari dependency berat.

---

## 6. Cara pakai file ini

Setiap kali sebuah langkah build **menyimpang** dari rencana (skema diperbaiki, prompt salah,
keputusan library, workaround), tambahkan satu baris di §3/§4/§5 dengan tanggal absolut.
Jangan hapus entri lama — file ini adalah jejak audit rencana ↔ realita.
