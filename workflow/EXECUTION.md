# EXECUTION — running the workflow with tiered models

Operational companion to the planning docs. This is **how to build** the app from the prompts in `prompts/`: which model per module, in what order, with what verification. Planning lives in `modules/` and `00-general-tasks.md`; this file drives the actual build.

Grounding: **Fable 5** is the most capable model but the priciest tier ($10/$50 per 1M — 2× Opus 4.8, 3.3× Sonnet 5, 10× Haiku 4.5). It excels at first-shot implementation of well-specified systems (these prompts) and long-horizon work, but its turns run long and over-prescriptive prompts *reduce* its quality — hence the Fable preamble. Use it only where capability pays; use **Sonnet 5** for templated CRUD.

---

## 1. Model tiering

**Fable 5** — `/model` → Fable, `/effort` → `xhigh` for ★, else `high`. Paste `_shared/fable-preamble.md` **above** the prompt.

| Module | ★ | Why Fable |
|--------|---|-----------|
| `01-foundation` | | Migration runner, slamctl, tzdata, `response.FromError` — everything builds on it |
| `02-file-management` | | Image pipeline: EXIF orient, sha256 dedupe, 3 variants, multipart Bab 6.6 |
| `04-auth-session` | | JWT claims, `sesi_login` refresh rotation, lockout, replaces toy table |
| `05-hak-akses` | ★ | RBAC grid (hardest screen) + PermGuard cache/perm_version + anti-escalation |
| `07-user-management` | | Anti-escalation over `users`/`user_role` |
| `16-jadwal` | ★ | Timezone-aware calendar + recurrence + idempotent `generate-sesi` |
| `17-penugasan-notifikasi` | | Fan-out notifikasi + bulk + `wajib_absen` auto-false rule |
| `18-absensi` | ★ | Camera+GPS, server-time authority, Haversine, status derivation, session-closer |
| `19-izin` | | Couples into absensi (`izin_id`, alfa suppression) |
| `20-kta-nra-qr` | ★ | NRA SEQUENCE + setval guard, CR80 PNG, public-token whitelist |
| `21-laporan-rekap` | | GROUP-BY recap correctness + xlsx/pdf export |

**Sonnet 5** — `/model` → Sonnet, `/effort` → `high` (`medium` OK for the simplest). Prompt as-is (no preamble):
`03-pengaturan-log`, `06-anggota`, `08-instansi`, `09-unit`, `10-prestasi`, `11-inorga`, `12-medsos`, `13-dokumen`, `14-profile-club`, `15-lokasi`, `22-kegiatan`, `23-artikel`, `24-landing-publik`, `25-pwa`.

> Budget tight? `19-izin` and `21-laporan-rekap` are the safe drops from Fable → Sonnet.

---

## 2. Per-module loop

For each module, **API prompt first, then APP prompt**:

1. **Fresh session** in the right dir: `slam-team-api` for `*-api.md`, `slam-team-app` for `*-app.md`.
2. `/model` + `/effort` per the tier table.
3. Paste — **Fable**: `_shared/fable-preamble.md` + the prompt; **Sonnet**: the prompt as-is. Let it run (Fable turns are long — don't interrupt).
4. **Verify**: the prompt's own checklist + `make build && go vet ./...` (API) or `npm run build:local` (APP). Then the **coherence gate** — one live **form → API → DB** round trip: submit the form, confirm the row in Postgres and the `{success,message,data,errors}` envelope, and that a disallowed role gets **403**. `/verify` or `/code-review` on the diff.
5. **Commit**: `git add -A && git commit -m "<module>: <api|app>"` (per project repo).
6. Mark **V** in the `00-general-tasks.md` progress table.

---

## 3. Build order (prerequisites first)

| Fase | Order |
|------|-------|
| 0 | `01-foundation` → `02-file-management` → `03-pengaturan-log` → `04-auth-session` |
| 1 | `05-hak-akses` |
| 2 | `08-instansi` → `06-anggota` → `07-user-management` → `09-unit` → `10-prestasi` → `11-inorga` → `12-medsos` → `13-dokumen` → `14-profile-club` |
| 3 | `15-lokasi` → `16-jadwal` |
| 4 | `17-penugasan-notifikasi` |
| 5 | `19-izin` → `18-absensi` |
| 6 | `20-kta-nra-qr` |
| 7 | `21-laporan-rekap` |
| 8 | `22-kegiatan` → `23-artikel` → `24-landing-publik` |
| 9 | `25-pwa` |

Key ordering reasons: `instansi` before `anggota` (`anggota.instansi_id` FK restrict); `izin` before `absensi` (`absensi.izin_id`).

---

## 4. Pacing & safety

- **Git is the rollback net** — both `slam-team-api` and `slam-team-app` are initialized with a baseline commit; commit after every verified module.
- **Interleave for budget**: when `/status` shows low headroom, do cheap Sonnet CRUD modules; save Fable modules for headroom.
- **Effort discipline**: `high` is the Fable default; reserve `xhigh` for the 4 ★ items.
- **One module per session** caps context growth and blast radius.

---

## 5. Milestone check (after Fase 1)

Log in → dynamic sidebar renders from `/modul` + `/me/permissions` → open the RBAC grid → change a permission → `perm_version` bumps and the UI re-gates. If that round trip holds, the foundation is coherent and the CRUD modules can fan out safely.
