# PROMPT — Build API module `kta` (NRA sequence, PNG, public QR) (Fase 6)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Build the **`kta`** module — member cards that own the NRA. The most domain-specific module; read the DBML Note on `kta` carefully.

## 0. Read first
1. `slamteam_db.dbml` — **`kta`** (line 537, WITH its long Note on NRA format) and **`mst_wilayah`** (line 469); relations (`kta.anggota_id → anggota.id` **restrict**, `file_kta_id`, `dicetak_oleh`/`dicabut_oleh`). DBML WINS.
2. `workflow/modules/20-kta-nra-qr.md` (full spec + NRA rules).
3. `_shared/api-endpoints.md` §11 (requests/responses verbatim) + `_shared/conventions-api.md` §10 (file layer).

## 1. Scope + endpoints
| Method | Path | Permission |
|--------|------|-----------|
| POST | `/kta` | `kta.create` |
| POST | `/kta/bulk` | `kta.create` |
| GET | `/kta/anggota/:anggotaId` | `kta.read` |
| POST | `/kta/:id/cetak-ulang` | `kta.print` |
| PATCH | `/kta/:id/cabut` | `kta.cabut` |
| GET | `/kta/batch/:batchId` | `kta.read` |
| GET | `/public/kta/:token` | — (**public**, rate-limited) |

Band: create/print/cabut/read = SA/Admin. Public page = Guest.

## 2. Files under `internal/modules/core/kta/`
- domain `Kta` (all columns, enums), plus a `mst_wilayah` read.
- dto: `CreateKtaReq{anggota_id required; jenis_kta oneof=pelajar dewasa; alasan_cetak; berlaku_dari?; berlaku_sampai?}`, `BulkKtaReq`, `CabutReq{alasan_dicabut required}`, `KtaResp`, `PublicKtaResp`.
- repository: `NextNoUrut()` (SEQUENCE `nextval`; **setval guard** when a manual value overtakes), `Create`, `FindByID`, `ListByAnggota`, `ListByBatch`, `BumpVersiCetak`, `Cabut`, `FindByToken`, `ArsipkanPelajar(anggotaID)`.
- service:
  - **Issue (POST /kta):** load anggota (needs `tanggal_lahir`); build NRA = `wilayah_paten(3573) + bulan(2) + tahun(2) + no_urut` (from SEQUENCE). Random 32-char `token_publik`; `qr_url`. `berlaku_sampai` = `berlaku_dari` + (pelajar 12 / dewasa 24) months from `mst_pengaturan`. Generate CR80/ID-1 PNG (300 dpi, trim+bleed) → file-service → `file_kta_id`. If issuing a **dewasa** card for a `siswa_ke_dewasa` member: set the existing pelajar card `status=arsip`, set `anggota.no_induk` = dewasa NRA. `audit.Write` (aksi cetak).
  - **Bulk:** loop `anggota_ids` in one `batch_cetak_id`; each card its own sequential NRA; return `{batch_cetak_id, diterbitkan, gagal, kta[]}`.
  - **cetak-ulang:** `versi_cetak+1`, new `token_publik`, **keep `no_kta`**, regenerate PNG.
  - **cabut:** `status=dicabut` + `alasan_dicabut` + `dicabut_oleh/pada`; public token stops validating.
  - **public:** `FindByToken`; arsip/dicabut/kadaluarsa → status notice; return ONLY `{foto_url, nama, nra, status, instansi, berlaku_sampai, digantikan}` — never address/id/DOB/contact; arsip → `{digantikan:true, pesan}` without the replacement token.
- PNG lib: choose (e.g. `fogleman/gg` + `golang.org/x/image`, QR `skip2/go-qrcode`) — record in `workflow/99-alignment-report.md`.
- handler/main/router: `/public/kta/:token` **no JWT** + **rate limit**; others RequirePermission.

## 3. Migration + seeder
- `mst_wilayah` (seed `3573` Kota Malang) + `kta` EXACTLY per `.dbml` (enum `jenis_kta`/`status_kta`; unique `no_kta`/`token_publik`; `UNIQUE(anggota_id,jenis_kta,versi_cetak)`; FK restrict). Create a **SEQUENCE** for `no_urut_kartu`. After `anggota`, `mst_file`. Confirm `kta.*` seeded.

## 4. Verification checklist
- [ ] build/vet pass; migration up/down clean; SEQUENCE created; matches `.dbml`.
- [ ] NRA built correctly (paten wilayah + birth month/year + urut); expands to 12 digits after urut 999.
- [ ] `siswa_ke_dewasa` → two cards (pelajar arsip + dewasa aktif), different NRAs; `no_induk` → dewasa.
- [ ] setval guard prevents collisions when a manual number overtakes the sequence.
- [ ] cetak-ulang bumps `versi_cetak` + new token but **keeps `no_kta`**.
- [ ] `GET /public/kta/:token` returns only whitelisted fields; arsip → digantikan notice without leaking the new token; no auth + rate-limited.
- [ ] `kta.*` blocked for Mod/User (403); PNG CR80 300dpi generated; `audit.Write` on cetak/cabut.
