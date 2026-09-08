# PROMPT — Build APP UI for `izin` (Fase 5)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`; izin status color = info), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
`workflow/modules/19-izin.md`; `_shared/conventions-app.md` (§9 list, §6 form, §8 FileService).

## 2. Scope
POST `/izin` (`izin.create`, all login), GET `/izin` (`izin.read`, own/all by cakupan), PATCH `/izin/:id/approve` & `/tolak` (`izin.approve`).

## 3. Files under `src/app/pages/izin/`
- `izin.model.ts`, `izin.service.ts`.
- `izin-form.ts/.html` — request form: pick **sesi**, `jenis` select (izin/sakit/dinas/pulang_cepat), `alasan` textarea (required), `waktu_pulang_diminta` (native time, **shown only when jenis=pulang_cepat**), optional `lampiran` upload via `<app-file-upload>`. Map `422` inline.
- `izin-list.ts/.html` — list (User = own; admins = all) + status filter.
- `izin-approval.ts/.html` — approver queue of `menunggu` with **Setujui / Tolak** (Tolak requires `catatan_peninjau`), `*hasPermission="'izin.approve'"`.

## 4. Routing
`{ path:'izin', canActivate:[authGuard, permissionGuard], data:{permission:'izin.read'}, children:[list, new(create), approval(izin.approve)] }`.

## 5. i18n
Namespace `IZIN` (`FORM.JENIS/ALASAN/WAKTU_PULANG/LAMPIRAN`, `STATUS.MENUNGGU/DISETUJUI/DITOLAK`, `APPROVE`, `REJECT`, `REJECT_NOTE`) + `COMMON.*`. IND/ENG in sync.

## 6. Verification checklist
- [ ] build compiles; `waktu_pulang_diminta` visible only for pulang_cepat; lampiran uploads (private).
- [ ] Create/list round-trip; approve/tolak gated `izin.approve` (Tolak requires note).
- [ ] User sees only own izin; buttons hidden without permission + backend 403; dark theme + red accent + focus ring; IND/ENG keys match; taste-skill announced.
