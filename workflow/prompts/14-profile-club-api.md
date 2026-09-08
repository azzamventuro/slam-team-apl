# PROMPT — Build API module `profile_club` (singleton club profile) (Fase 2)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Build the **`profile_club`** module — the club's public profile, effectively a **singleton** (one active row).

## 0. Read first
1. `slamteam_db.dbml` — **`profile_club`** (line 1085): `nama`, `singkatan`, `banner_file_id`, `logo_simple_file_id`, `logo_besar_file_id`, `alamat`, `keterangan`, soft-delete + audit. DBML WINS.
2. `workflow/modules/14-profile-club.md`; `_shared/conventions-api.md`; `_shared/api-endpoints.md` §13 + §12 (public via `/public/profil`).

## 1. Scope + endpoints
| Method | Path | Permission |
|--------|------|-----------|
| GET | `/profile_club` | `profile_club.read` (returns the single active row) |
| PUT | `/profile_club` | `profile_club.update` (**upsert**: create if none, else update the row) |
Band: SA/Admin CRUD, Mod/User/Guest R. Public read via `/public/profil` (landing module).

## 2. Files under `internal/modules/core/profileclub/`
- domain `ProfileClub` (`profile_club`, `TableName`); dto (`UpsertProfileReq`: `nama max=150`; `singkatan max=50`; three file ids `omitempty,gt=0`); repository (`GetActive`, `Upsert` — one row semantics; join three `*_uuid`); service (create-if-none-else-update; `audit.Write`); handler; main+router.
- Since it is a singleton, no list/pagination; `PUT` upserts. (POST/DELETE optional — keep to GET+PUT unless the doc says otherwise.)

## 3. Migration
`profile_club` EXACTLY per `.dbml` (three `*_file_id → mst_file`). After `mst_file`. Optional seeder: one default row (`nama` from `mst_pengaturan.umum.nama_klub`). Confirm `profile_club.*` seeded.

## 4. Verification checklist
- [ ] build/vet pass; migration up/down clean.
- [ ] `GET /profile_club` returns the single row; `PUT` creates-if-none then updates in place (no duplicate rows).
- [ ] Three uploads (banner/logo_simple/logo_besar) stored; detail returns uuids.
- [ ] Mod/User read-only (writes → 403); readable publicly via landing; `audit.Write` on changes.
