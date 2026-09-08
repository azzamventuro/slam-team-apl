# PROMPT — Build APP UI for `kta` (issue, preview, bulk, public QR) (Fase 6)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`). **The KTA card ALWAYS uses the dark card palette** (near-black + red) regardless of app theme; the surrounding admin chrome is dark, the public page is light. Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
`workflow/modules/20-kta-nra-qr.md`; `_shared/conventions-app.md` (§8 FileService, §9 list).

## 2. Scope
| Method | Path |
|--------|------|
| POST | `/kta`, `/kta/bulk` (`kta.create`) |
| GET | `/kta/anggota/:id`, `/kta/batch/:id` (`kta.read`) |
| POST | `/kta/:id/cetak-ulang` (`kta.print`) |
| PATCH | `/kta/:id/cabut` (`kta.cabut`) |
| GET | `/public/kta/:token` (public) |

## 3. Files
- `src/app/pages/kta/kta-issue.ts/.html` — issue form: pick anggota + `jenis_kta` + `alasan_cetak`; optional `berlaku_*` override; **on-screen card preview** (light frame, dark card) showing foto formal, nama, NRA, QR.
- `src/app/pages/kta/kta-anggota.ts/.html` — a member's cards (aktif/arsip) + **cetak ulang** / **cabut** actions (gated `kta.print`/`kta.cabut`).
- `src/app/pages/kta/kta-bulk.ts/.html` — bulk issue (multi-select anggota → `batch_cetak_id`); download batch PNGs (cards-per-sheet from settings).
- `src/app/pages/public/kta-verify.ts/.html` — **PUBLIC page** (outside authGuard, **light theme**) as the QR target: show only whitelisted fields; arsip → "digantikan" notice; **never** fetch private files (use the public foto URL from the response).
- Services `kta.service.ts` (authed) + `public-kta.service.ts` (public).

## 4. Routing
- Authed: `{ path:'kta', canActivate:[authGuard, permissionGuard], data:{permission:'kta.read'}, children:[issue(kta.create), anggota/:id, bulk(kta.create)] }`.
- Public: a route OUTSIDE authGuard, e.g. `{ path:'kta/:token', loadComponent: KtaVerify }`.

## 5. i18n
Namespace `KTA` (`ISSUE`, `REPRINT`, `REVOKE`, `NRA`, `VALID_UNTIL`, `BATCH`, `PUBLIC.TITLE`, `PUBLIC.REPLACED`, `PUBLIC.STATUS`) + `COMMON.*`. IND/ENG in sync.

## 6. Verification checklist
- [ ] build compiles; issue form previews a card (dark palette) with NRA + QR; issue persists a card.
- [ ] cetak-ulang / cabut gated (`kta.print`/`kta.cabut`); bulk issues a batch + downloads PNGs.
- [ ] Public verify page (light theme, outside authGuard) shows only whitelisted fields; arsip → replaced notice; no private file fetch.
- [ ] `kta.*` blocked for Mod/User (buttons hidden + backend 403); dark card palette regardless of theme; IND/ENG keys match; taste-skill announced.
