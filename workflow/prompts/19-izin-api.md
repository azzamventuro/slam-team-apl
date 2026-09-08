# PROMPT — Build API module `izin` (`absensi_izin`) (Fase 5)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Build the **`izin`** module (`absensi_izin` — leave/permission requests tied to a session).

## 0. Read first
1. `slamteam_db.dbml` — **`absensi_izin`** (line 929); relations (`sesi_id`,`anggota_id` **restrict**; `lampiran_file_id`; `diproses_oleh`; ← `absensi.izin_id`). DBML WINS.
2. `workflow/modules/19-izin.md`; `_shared/api-endpoints.md` §9; `_shared/conventions-api.md` §8 (cakupan).

## 1. Scope + endpoints
| Method | Path | Permission |
|--------|------|-----------|
| POST | `/izin` | `izin.create` (all login) |
| GET | `/izin` | `izin.read` (own vs all by cakupan) |
| PATCH | `/izin/:id/approve` | `izin.approve` |
| PATCH | `/izin/:id/tolak` | `izin.approve` |

## 2. Files under `internal/modules/core/izin/`
- domain `Izin` (`absensi_izin`); dto (`CreateIzinReq{sesi_id required; jenis oneof=izin sakit dinas pulang_cepat; alasan required; waktu_pulang_diminta omitempty; lampiran_file_id omitempty,gt=0}`, `TolakReq{catatan_peninjau required}`, `IzinResp`).
- repository (`Create`, `FindByID`, `List(cakupan + filter status/sesi)`, `Approve`, `Tolak`; `is_deleted=false`).
- service:
  - **Create:** sesi exists; `waktu_pulang_diminta` required **only** for `jenis=pulang_cepat` (reject if set for other jenis); cakupan `milik_sendiri` forces `anggota_id=claims.AnggotaID`; validate `lampiran_file_id` if present; `audit.Write`.
  - **Approve:** `status=disetujui`, `diproses_oleh/pada`; **sync to absensi** — create/update an `absensi` row with `status_kehadiran` = jenis (izin/sakit/dinas) and set `izin_id` so the member isn't marked alfa (or have the session-closer read approved izin); fan-out notifikasi `izin_disetujui`.
  - **Tolak:** `status=ditolak` + `catatan_peninjau` (required); notifikasi `izin_ditolak`.
- handler/main/router: `approve`/`tolak` → `izin.approve`.

## 3. Migration
`absensi_izin` EXACTLY per `.dbml` (enum `jenis_izin`/`status_izin`; FK restrict). After `jadwal_sesi`, `anggota`; before/with `absensi` (`izin_id`). Confirm `izin.*` seeded.

## 4. Verification checklist
- [ ] build/vet pass; migration up/down clean.
- [ ] `waktu_pulang_diminta` valid only for `pulang_cepat`; other jenis with it → 422.
- [ ] Approve → disetujui + member not marked alfa (absensi sync); Tolak requires `catatan_peninjau`.
- [ ] `izin.approve` blocked for User (403); User sees/creates only own (cakupan).
- [ ] Private lampiran via `/files`; `audit.Write` + notifikasi on approve/tolak.
