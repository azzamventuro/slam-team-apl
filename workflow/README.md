# SLAM Team — Build Playbook (`workflow/`)

This folder is the **build playbook** for the SLAM Team (Scouting Legion Airsofter Malang)
greenfield rebuild. It does **not** contain application code. It contains the planning
and prompt documents that drive the rebuild: one set of Markdown files per module, plus
shared references that every module depends on.

The actual code lives elsewhere:

- **API** — `D:/xampp/htdocs/slam-team-apl/slam-team-api` (Go 1.24, Gin, GORM/Postgres, JWT HS256, zap)
- **APP** — `D:/xampp/htdocs/slam-team-apl/slam-team-app` (Angular 21 PWA, standalone, signals, zoneless, Bootstrap 5, ngx-translate)

You build **into** those existing scaffolds. You do not invent a new layout.

---

## The method (rancangan Bab 11)

The rebuild follows one disciplined loop. Do them in this order; do not skip ahead.

1. **Schema locked first.** The single source of truth for every table, column, type,
   and relation is the DBML: `../slamteam_db.dbml`. If prose (the rancangan, an old app,
   a comment) ever disagrees with the `.dbml`, **the `.dbml` wins**. Nothing gets built
   until the schema for that area is settled in the `.dbml`.

2. **One module at a time, via Claude Code.** Each of the 19 modules is planned and then
   executed as an isolated unit. You read the module's planning doc for understanding,
   then paste that module's API prompt and APP prompt into Claude Code. You do not build
   two modules at once, and you do not build a module before its prerequisites (see
   **Phase order** below).

3. **UI via the taste-skill.** Every APP prompt invokes the **design-taste-frontend**
   skill (installed via `npx skills add Leonxlnx/taste-skill`). The executor announces
   "Using design-taste-frontend", runs its pre-flight/audit, maps to the already-installed
   Bootstrap 5, and **enforces the SLAM design tokens** in
   [`_shared/design-tokens.md`](./_shared/design-tokens.md) so that 19 modules do not
   turn into 19 different styles.

---

## Folder map

| Path | What it is |
|------|-----------|
| [`00-general-tasks.md`](./00-general-tasks.md) | Cross-cutting, do-once setup and global conventions that are **not** owned by any single module (timestamptz/UTC + IANA tz, soft-delete, file layer + upload checklist, RBAC middleware, `slamctl create-superadmin`, response envelope). Read this before any module. |
| [`_shared/`](./_shared/) | References every module prompt points at. Includes [`_shared/design-tokens.md`](./_shared/design-tokens.md) (SLAM palette + typography derived from the KTA card), plus shared schema/endpoint/RBAC extracts. |
| [`modules/`](./modules/) | One planning doc per module (`modules/<modul>.md`). This is the **understanding** layer: what the module does, its tables, routes, permissions, edge cases. Read it before executing. |
| [`prompts/`](./prompts/) | One pair of executable prompts per module: `prompts/<modul>-api.md` (Go/Gin/GORM) and `prompts/<modul>-app.md` (Angular). These are pasted verbatim into Claude Code. |
| [`99-alignment-report.md`](./99-alignment-report.md) | Running record of where reality drifted from plan — schema fixes applied, prompt corrections, decisions made mid-build. Update it whenever a build step deviates. |
| `.rancangan.txt` | Extracted text of the original design document (~1424 lines). Authoritative for routes, payloads, and the permission matrix (Bab 9, Bab 3.5). |

---

## How to execute a single module

1. **Understand.** Open `modules/<modul>.md` and read it end to end. Cross-check any
   schema detail against `../slamteam_db.dbml` — the `.dbml` wins on conflict.
2. **Build the API.** Paste `prompts/<modul>-api.md` into Claude Code. This produces the
   Go module under `slam-team-api/internal/modules/core/<modul>/` and registers it in the
   router. Every route gets `RequirePermission("modul.aksi")`.
3. **Build the APP.** Paste `prompts/<modul>-app.md` into Claude Code. This produces the
   Angular feature under `slam-team-app/pages/<feature>/`. The prompt invokes
   **design-taste-frontend** and enforces the design tokens.
4. **Record drift.** If anything didn't match the plan, note it in
   [`99-alignment-report.md`](./99-alignment-report.md).

Always run the API prompt before the APP prompt for the same module — the frontend
consumes the endpoints the backend just defined.

---

## Phase order (rancangan Bab 12) — prerequisites in parentheses

Do not build a module before its prerequisites exist.

- **Fase 0 — Fondasi** (—): Go + Angular skeleton, `mst_pengaturan`, `log_aktivitas`,
  file layer (3 variants + upload checklist), timestamptz conventions,
  `slamctl create-superadmin`, auth (login + refresh + logout, `sesi_login`).
- **Fase 1 — Hak akses dinamis** (0): 5 RBAC tables, seed permission matrix, Go
  middleware, Angular guard + `*hasPermission` directive + dynamic sidebar, Hak Akses admin page.
- **Fase 2 — Anggota + master data** (1): anggota, instansi, unit, prestasi, inorga,
  medsos, dokumen, profile_club, user-management.
- **Fase 3 — Lokasi + jadwal** (1): `mst_lokasi` (map picker), jadwal, `jadwal_sesi`
  generator (recurrence), calendar view.
- **Fase 4 — Penugasan** (3): `jadwal_peserta` single + bulk, `wajib_absen`,
  accept/reject, in-app notifikasi.
- **Fase 5 — Absensi** (0, 3, 4): absen screen (camera + GPS), backend validation, izin,
  override, session-closer job.
- **Fase 6 — KTA / NRA / QR** (0, 2): `mst_wilayah` seed, NRA sequence, kta, PNG generate,
  bulk print, public page, reprint/revoke.
- **Fase 7 — Laporan / rekap** (5): `absensi_rekap`, per anggota/jadwal/periode,
  Excel + PDF export, dashboard.
- **Fase 8 — Landing + public content** (2): kegiatan, artikel, public
  profil/artikel/kegiatan/prestasi.
- **Fase 9 — PWA** (5): manifest, service worker, icons, iOS Safari camera redirect,
  schedule reminders.

The three hardest, no-precedent screens — design these first: (1) RBAC grid role × modul,
(2) schedule calendar, (3) attendance/absen screen (camera + GPS).

---

## Caveat: `blog-fe` is empty

The old frontend `D:/xampp/htdocs/slam-team/blog-fe` currently has **no source** (only
`node_modules/` and `.angular/`). Visual mirroring of the old login/admin/landing is
therefore **not possible right now**.

**Fallback:** use the [`_shared/design-tokens.md`](./_shared/design-tokens.md) plus the
**design-taste-frontend** skill for all UI. Every APP prompt carries this note:

> If `blog-fe` source is restored, mirror its layout/menu for this screen; otherwise
> follow the design tokens.

The old app is a **visual reference only**. API routes, payloads, and column names always
follow the rancangan (Bab 9) and the `.dbml` — never the old app.

---

## Start here

1. [`00-general-tasks.md`](./00-general-tasks.md) — global conventions, read first.
2. [`_shared/design-tokens.md`](./_shared/design-tokens.md) — the one visual language.
3. `../slamteam_db.dbml` — the schema that wins every argument.
