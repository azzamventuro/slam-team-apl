# Modul 25 — PWA + Penyempurnaan (`pwa`)

> Sumber otoritatif: rancangan Bab 11.4 (Progressive Web App) > `_shared/conventions-app.md` > `_shared/design-tokens.md`. Kerja platform front-end — **tidak ada endpoint backend baru** (memakai `/notifikasi` yang sudah ada).

---

## 1. Ringkasan & tujuan modul

**PWA** menjadikan aplikasi Angular dapat dipasang (installable), bekerja andal, dan siap kamera/lokasi di perangkat. **Fase 9** (polish), bergantung pada **Absensi (18)** karena kamera/GPS-nya butuh **secure context (HTTPS)**.

Cakupan: web app manifest, service worker (cache), ikon (termasuk maskable), redirect/panduan iOS Safari untuk kamera, dan pengingat jadwal (notifikasi lokal/push).

> Modul ini **hanya app** — tidak ada prompt/kode API. Fokus pada `angular.json`, manifest, `ngsw-config.json`/service worker, dan komponen guidance.

---

## 2. Berkas & konfigurasi (bukan tabel DB)

| Berkas | Isi |
|--------|-----|
| `public/manifest.webmanifest` | `name`, `short_name`, `theme_color=#E11D2A` (SLAM red), `background_color=#0B0B0D`, `display=standalone`, `start_url`, `icons[]` (termasuk maskable 512). |
| `ngsw-config.json` (bila pakai `@angular/pwa`) | strategi cache: **precache app shell**; **JANGAN** cache respons API ber-auth / berkas privat secara tidak aman. |
| `public/icons/*` | ikon 192/512 + maskable. |
| `angular.json` | daftarkan manifest & service worker (production). |
| `index.html` | link manifest, `theme-color`, apple-touch-icon, meta secure-context. |

---

## 3. Endpoint

Tidak ada endpoint baru. Pengingat jadwal membaca **`GET /notifikasi`** (modul 17). Push (opsional) butuh service worker + izin notifikasi browser.

---

## 4. Hak akses

N/A (fitur klien). Halaman/aksi tetap tunduk pada guard/permission masing-masing modul.

---

## 5. Kebutuhan BACKEND

**Tidak ada.** (Push nyata via server memerlukan endpoint langganan — di luar cakupan saat ini; pakai notifikasi lokal + `/notifikasi`.)

---

## 6. Kebutuhan FRONTEND (Angular)

**WAJIB taste-skill.** Warna tema PWA = SLAM red/near-black (design-tokens).

### Pekerjaan
- **Manifest + ikon** (installable). Pasang lewat `@angular/pwa` (ngsw) atau service worker manual.
- **Service worker + cache**: precache shell; runtime cache aset statis; **hindari** cache respons API ber-auth / `/files/*` privat (jangan simpan data sensitif offline tanpa perlindungan).
- **iOS Safari**: `getUserMedia`/`geolocation` butuh **secure context (HTTPS)** dan PWA standalone iOS punya keanehan kamera — tambah **panduan/redirect** pada layar Absensi (18): deteksi non-secure-context / izin diblok → tampilkan instruksi buka via HTTPS/Safari.
- **Pengingat jadwal**: notifikasi lokal untuk sesi mendatang (baca `/notifikasi`); minta izin notifikasi dengan sopan (bukan saat load pertama).
- **Prefers-reduced-motion & offline fallback** halaman.

### i18n
- Namespace `PWA` (`INSTALL`, `OFFLINE`, `IOS_CAMERA_HINT`, `NOTIF_PERMISSION`).

---

## 7. ALUR

| Aksi | Mekanisme |
|------|-----------|
| Install app | Manifest + service worker terdaftar → prompt install |
| Buka absen di iOS non-HTTPS | Deteksi secure-context → panduan/redirect |
| Pengingat sesi | Baca `/notifikasi` → notifikasi lokal |

Tidak ada tulisan DB dari modul ini.

---

## 8. Dependencies / prasyarat & Acceptance criteria

### Prasyarat
- **Absensi (18)** ada (target guidance kamera). **Notifikasi (17)** untuk pengingat. Deploy HTTPS.

### Acceptance criteria
- [ ] App installable (manifest valid, ikon + maskable, theme SLAM red).
- [ ] Service worker aktif (production); shell ter-precache; respons API ber-auth/berkas privat **tidak** di-cache tak aman.
- [ ] Layar absensi menampilkan panduan bila bukan secure context / izin kamera-GPS diblok (khususnya iOS Safari).
- [ ] Pengingat jadwal muncul dari `/notifikasi`; izin notifikasi diminta sopan.
- [ ] `prefers-reduced-motion` dihormati; taste-skill diterapkan; token tema diterapkan.
