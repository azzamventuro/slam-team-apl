# PROMPT — Build APP public landing pages (Fase 8)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`) — **public pages use the LIGHT theme tokens** (the dark chrome is for the authed admin area); red accent + UPPERCASE headings still apply. Bootstrap 5. *If `blog-fe` restored mirror its landing/menu; else follow tokens.*

## 1. Read first
`workflow/modules/24-landing-publik.md`; `_shared/conventions-app.md` (§3 ApiService; §10 design).

## 2. Scope (all PUBLIC, no auth)
GET `/public/profil`, `/public/kegiatan`, `/public/artikel` (list + `?slug=`), `/public/pengaturan`.

## 3. Files under `src/app/pages/public/`
- `public.service.ts` — profil, kegiatan, artikel(list/detail), pengaturan.
- `landing.ts/.html` — hero (club name/logo from `/public/pengaturan`), club profile (`/public/profil`), latest kegiatan + artikel, prestasi showcase. Images from **public variant URLs** (not blob/token).
- `artikel-publik.ts/.html` + `artikel-detail.ts/.html` — public blog list + article detail (via slug).
- `kegiatan-publik.ts/.html` + `profil-publik.ts/.html` as needed.

## 4. Routing (OUTSIDE authGuard)
Add public routes NOT wrapped by `authGuard`, e.g.:
```
{ path: '', loadComponent: () => import('./pages/public/landing').then(m => m.Landing) },
{ path: 'artikel', loadComponent: ... }, { path: 'artikel/:slug', loadComponent: ... },
{ path: 'kegiatan', ... }, { path: 'profil', ... },
```
Keep SEO/OG friendliness (title/meta per page).

## 5. i18n
Namespace `LANDING`/`PUBLIC` (`HERO`, `LATEST_ARTIKEL`, `LATEST_KEGIATAN`, `PRESTASI`, `READ_MORE`) + `COMMON.*`. IND/ENG in sync.

## 6. Verification checklist
- [ ] build compiles; landing renders profile + latest content + prestasi from public endpoints (no auth).
- [ ] Public routes are OUTSIDE authGuard; light theme; images via public URLs (no token/blob).
- [ ] Article detail resolves by `?slug=`; SEO/OG meta set; IND/ENG keys match; taste-skill announced.
