# PROMPT — Make the app a PWA + iOS camera guidance (Fase 9)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`) — PWA theme colors = SLAM red `#E11D2A` / near-black `#0B0B0D`. Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
`workflow/modules/25-pwa.md`; `_shared/conventions-app.md`. NOTE: **no API prompt** for this module — it is platform/build work.

## 2. Scope (app-only, Fase 9 polish; depends on Absensi module 18)
Make the Angular app an installable PWA and harden camera/GPS on iOS.

## 3. Work items
- **Manifest** `public/manifest.webmanifest`: `name`, `short_name`, `theme_color:#E11D2A`, `background_color:#0B0B0D`, `display:standalone`, `start_url`, `icons[]` (192, 512, + **maskable** 512). Link it in `index.html` with `<meta name="theme-color">` + apple-touch-icon.
- **Service worker**: add via `@angular/pwa` (ngsw) OR a hand-rolled SW. Cache strategy: **precache the app shell**; runtime-cache static assets; **do NOT insecurely cache authed API responses or `/files/*` private responses**. Wire in `angular.json` (production) + `ngsw-config.json`.
- **Icons**: generate 192/512 + maskable in `public/icons/`.
- **iOS Safari camera guidance**: on the Absensi screen (module 18), detect **non-secure-context** (`window.isSecureContext===false`) or blocked camera/GPS and show a redirect/guidance (getUserMedia/geolocation require HTTPS; iOS standalone PWAs have camera quirks).
- **Schedule reminders**: local notifications for upcoming sessions (read `/notifikasi`); request notification permission **politely** (on user action, not first load).
- Honor `prefers-reduced-motion`; provide an offline fallback page.

## 4. i18n
Namespace `PWA` (`INSTALL`, `OFFLINE`, `IOS_CAMERA_HINT`, `NOTIF_PERMISSION`) + `COMMON.*`. IND/ENG in sync.

## 5. Verification checklist
- [ ] `npm run build` (production) compiles; app is installable (valid manifest + icons incl. maskable; theme = SLAM red).
- [ ] Service worker active in production; app shell precached; authed API/`/files` responses NOT cached insecurely.
- [ ] Absensi screen shows guidance when not a secure context / camera or GPS blocked (esp. iOS Safari).
- [ ] Schedule reminders appear from `/notifikasi`; notification permission requested politely.
- [ ] `prefers-reduced-motion` honored; offline fallback works; taste-skill announced; theme tokens applied.
