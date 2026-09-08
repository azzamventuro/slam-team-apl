# PROMPT — Build API public endpoints (landing/content) (Fase 8)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Build the **public read-only** endpoints that feed the landing page. NO auth; whitelist fields only.

## 0. Read first
1. `slamteam_db.dbml` — **`profile_club`** (1085), **`kegiatan`** (1003), **`artikel`** (1022), **`prestasi`** (1064), **`mst_pengaturan`** (270, `is_publik`). DBML WINS.
2. `workflow/modules/24-landing-publik.md`; `_shared/api-endpoints.md` §12.

## 1. Scope + endpoints (all PUBLIC, no JWT, rate-limited)
| Method | Path | Returns |
|--------|------|---------|
| GET | `/public/profil` | `profile_club` (active row) + public `prestasi` list |
| GET | `/public/kegiatan` | `kegiatan` where `status_kegiatan=1` |
| GET | `/public/artikel` | `artikel` where `status_kegiatan='terbit'` (list + detail via `?slug=`) |
| GET | `/public/pengaturan` | `mst_pengaturan` where `is_publik=true` only |

(`GET /public/kta/:token` is owned by module 20.)

## 2. Files under `internal/modules/core/publik/`
- service/handler that read (no writes) from the content repositories or dedicated read queries; expose ONLY safe columns; resolve images through **public variant URLs** (never the authenticated private endpoint). `/public/pengaturan` returns whitelisted `is_publik` rows only — **never** thresholds/secrets (absensi radius/tolerance, NRA config are NOT public).
- Mount in a public route group (no `JWTAuth`); apply a light rate limit. Register after the content modules.

## 3. Migration
None (read-only over existing tables).

## 4. Verification checklist
- [ ] build/vet pass.
- [ ] All four routes respond WITHOUT a token; only whitelisted fields returned.
- [ ] Draft/non-`terbit` content never appears; `mst_pengaturan` non-public/locked/secret rows never leak.
- [ ] Images use public variant URLs (no private file endpoint); artikel detail via `?slug=`; rate-limited.
