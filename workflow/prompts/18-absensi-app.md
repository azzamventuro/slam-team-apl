# PROMPT — Build APP absensi screen (camera + GPS) (Fase 5)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`: dark, SLAM red, UPPERCASE headings, red focus ring, 44px targets), Bootstrap 5. **The absen screen is one of the 3 hardest — build it first.** Status colors per tokens (hadir=success, terlambat/pulang_cepat/hadir_luar_radius=warning, izin/sakit/dinas=info, alfa=danger). *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
`workflow/modules/18-absensi.md`; `_shared/conventions-app.md` (§8 FileService/multipart, §9 list).

## 2. Scope
| Method | Path |
|--------|------|
| GET | `/absensi/sesi-aktif` (`absensi.create`) |
| POST | `/absensi/check-in`, `/absensi/check-out` (multipart) |
| GET | `/absensi` (`absensi.read`) |
| PATCH | `/absensi/:id/override` (`absensi.override`) |
Band: create = all login; read = semua (SA/Admin/Mod) / milik_sendiri (User); override = SA/Admin; delete = SA.

## 3. Files under `src/app/pages/absensi/`
- `absensi.model.ts`, `absensi.service.ts` (sesiAktif, checkIn, checkOut, list, override).
- `absen.ts/.html/.scss` — **the attendance screen**:
  - Requires a **secure context (HTTPS)** — detect and warn if not (getUserMedia/geolocation fail on plain HTTP).
  - Live **front-camera** preview (`getUserMedia({video:{facingMode:'user'}})`), capture a selfie to a `File`/`Blob`.
  - Read **GPS** (`navigator.geolocation`, accuracy); show distance to the session location.
  - **Out-of-radius confirm** dialog → sets `konfirmasi_luar_radius`.
  - **Permission-denied guidance** for blocked camera/GPS (how to enable).
  - Submit `FormData` (selfie + fields) to check-in/check-out **without** setting `Content-Type`; show a result card with `waktu_server_utc`, status, distance. Stop camera tracks + revoke object URLs on destroy.
- `absensi-list.ts/.html` — admin list (cakupan) with **override** action (`*hasPermission="'absensi.override'"`, `alasan_override` required).

## 4. Routing
`{ path:'absensi', canActivate:[authGuard, permissionGuard], data:{permission:'absensi.create'}, children:[absen, list(data:{permission:'absensi.read'})] }`.

## 5. i18n
Namespace `ABSENSI` (`CHECK_IN/OUT`, `STATUS.*`, `OUT_OF_RADIUS_CONFIRM`, `CAMERA_DENIED`, `GPS_DENIED`, `INSECURE_CONTEXT`, `OVERRIDE`, `OVERRIDE_REASON`) + `COMMON.*`. IND/ENG in sync.

## 6. Verification checklist
- [ ] `npm run build:local` compiles.
- [ ] Camera preview + selfie capture + GPS read work in a secure context; insecure context shows guidance.
- [ ] Out-of-radius confirm sets the flag; permission-denied guidance shown when blocked.
- [ ] Check-in/out post multipart **without** manual Content-Type; result card shows server time + status (correct status color).
- [ ] Override gated `absensi.override` with required reason; User list = own only.
- [ ] Camera tracks stopped + object URLs revoked on destroy; dark theme + red accent + focus ring; screen built first; IND/ENG keys match; taste-skill announced.
