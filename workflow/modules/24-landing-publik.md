# Modul 24 — Landing Page + Konten Publik (`landing`)

> Sumber otoritatif: `slamteam_db.dbml` (skema) > rancangan Bab 11.4 / Bab 8.5 (pengaturan publik) > `_shared/api-endpoints.md` §12 > `_shared/conventions-*.md`. Jika prosa dan `.dbml` berbeda, `.dbml` menang.

---

## 1. Ringkasan & tujuan modul

**Landing** adalah halaman publik + agregasi konten (tanpa autentikasi). Modul **Fase 8**. Menyajikan **hanya field aman** (whitelist) dari `profile_club`, `kegiatan` (terbit), `artikel` (terbit), `prestasi`, dan `mst_pengaturan` publik. Tidak pernah data privat/identitas.

CRUD admin untuk konten ini dimiliki modul masing-masing (14 profile_club, 22 kegiatan, 23 artikel, 10 prestasi). Modul ini **mewiring rute publik** + halaman landing.

---

## 2. Tabel & kolom (read-only publik)

| Tabel | Dipakai untuk | Field aman |
|-------|---------------|-----------|
| `profile_club` (DBML 1085) | profil klub | `nama, singkatan, banner/logo_* (uuid), alamat, keterangan` |
| `kegiatan` (DBML 1003) | kegiatan terbit | `judul, gambar (uuid), tanggal_*, konten` (status_kegiatan=1) |
| `artikel` (DBML 1022) | blog | `judul, gambar (uuid), konten` (status_kegiatan='terbit'), detail via slug |
| `prestasi` (DBML 1064) | showcase prestasi | `peringkat, tingkat, judul_kompetisi, flyer/foto (uuid), tanggal` |
| `mst_pengaturan` (DBML 270) | nama/logo/kontak | hanya `is_publik=true` |

**Aturan:** gambar dilayani lewat **varian publik** (bukan endpoint privat ber-auth). Tidak pernah alamat anggota/no identitas.

---

## 3. Endpoint (api-endpoints §12)

| Method | Path | Permission | Auth | Response |
|--------|------|-----------|------|----------|
| `GET` | `/public/profil` | — | public | `profile_club` + daftar `prestasi` publik |
| `GET` | `/public/kegiatan` | — | public | kegiatan `status_kegiatan=1` |
| `GET` | `/public/artikel` | — | public | artikel `terbit` (list + detail `?slug=`) |
| `GET` | `/public/pengaturan` | — | public | hanya `is_publik=true` (nama klub/logo/kontak) |

(`GET /public/kta/{token}` dimiliki modul KTA (20).) Semua PUBLIK, whitelist, boleh rate-limit.

---

## 4. Hak akses

Semua rute publik (Guest). Tidak ada permission — tetapi **hanya field whitelist**; `mst_pengaturan` non-publik & rahasia (ambang absensi, dll) **tidak pernah** dibocorkan.

---

## 5. Kebutuhan BACKEND (Go)

Modul `internal/modules/core/publik/` (agregator) — memanggil repository modul konten atau query read-only sendiri.

### service / handler
- `GET /public/profil` → satu `profile_club` (baris aktif) + `prestasi` (whitelist). `GET /public/kegiatan`/`artikel` → hanya terbit, field aman, gambar uuid → URL varian publik. `GET /public/pengaturan` → filter `is_publik=true` saja.
- Tidak menulis DB. Tanpa JWT. Rate-limit ringan.

### main / router
- Rute di grup publik (tanpa `JWTAuth`). Daftar di router setelah modul konten.

### edge cases
- Draft/non-publik tak pernah muncul. `mst_pengaturan` terkunci/rahasia tak bocor. Slug artikel tak ada → 404.

---

## 6. Kebutuhan FRONTEND (Angular)

**WAJIB taste-skill.** Halaman publik memakai **tema TERANG** (design-tokens light); chrome admin tetap gelap. SEO/OG friendly.

### Rute & komponen (`pages/public/`)
- Rute **di luar** `authGuard` (publik): `''` landing, `/artikel`, `/artikel/:slug`, `/kegiatan`, `/profil`.
- `landing.*` — hero (nama/logo dari `/public/pengaturan`), profil klub, kegiatan & artikel terbaru, showcase prestasi.
- `artikel-publik.*` / `artikel-detail.*` — blog list + detail (via slug).
- Gambar via URL varian publik (bukan blob ber-token).

### Service, i18n
- `public.service.ts` (profil, kegiatan, artikel, pengaturan). i18n `PUBLIC`/`LANDING`.

---

## 7. ALUR aksi → API → DATABASE

| Aksi | Endpoint | Baca DB |
|------|----------|---------|
| Buka landing | `GET /public/pengaturan` + `/public/profil` | SELECT whitelist |
| Baca artikel | `GET /public/artikel?slug=` | SELECT terbit by slug |
| Lihat kegiatan | `GET /public/kegiatan` | SELECT terbit |

Tidak ada tulisan DB (read-only).

---

## 8. Dependencies / prasyarat & Acceptance criteria

### Prasyarat
- **Profile Club (14)**, **Kegiatan (22)**, **Artikel (23)**, **Prestasi (10)**, **Pengaturan (03)**.

### Acceptance criteria
- [ ] Semua rute publik tanpa JWT; hanya field whitelist.
- [ ] Draft/non-publik & `mst_pengaturan` rahasia tidak pernah muncul.
- [ ] Gambar via varian publik (bukan endpoint privat).
- [ ] Rute landing di luar `authGuard`; tema terang; SEO/OG.
- [ ] Detail artikel via `?slug=`; taste-skill diterapkan.
