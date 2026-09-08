# PROMPT — Foundation APP shell (Angular 21) (Fase 0)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
1. **Invoke the Taste Skill.** Announce **"Using design-taste-frontend"**, run its pre-flight/audit on the existing shell first.
2. **Enforce the SLAM design tokens** in `D:/xampp/htdocs/slam-team-apl/workflow/_shared/design-tokens.md` — near-black bg `#0B0B0D`, dark surfaces, SLAM red `#E11D2A` accent, UPPERCASE wide-tracked headings, red focus ring (never removed), 44px touch targets, `prefers-reduced-motion`, status never color-only. **Tokens win** over taste-skill defaults.
3. **Map to Bootstrap 5** (already installed) via SCSS/CSS-variable overrides. No new UI kit, no new webfont.
4. *If `blog-fe` source is restored, mirror its layout/menu; otherwise follow the tokens.* (`blog-fe` is currently empty.)

## 1. Read first
- `D:/xampp/htdocs/slam-team-apl/workflow/modules/01-foundation.md`.
- `D:/xampp/htdocs/slam-team-apl/workflow/_shared/conventions-app.md` — §3 ApiService/envelope, §4 interceptors, §2 folder placement, §10 design method.
- Existing scaffold: `src/app/{core,layouts,pages,shared,environments}`, `slam-team-app/CLAUDE.md`.

## 2. What exists (confirm, don't rebuild)
Angular 21 zoneless scaffold: `core/{guards/auth.guard, interceptors/{auth,error}, services/auth.service, models}`, `layouts/{vertical,topbar,sidebar,footer}`, `pages/{auth/login,dashboard}`, `shared/pipes`, `environments/*`.

## 3. Gaps to fill / verify
- **`ApiService`** (`core/services/api.service.ts`): the generic envelope client from conventions-app §3 (`get/post/put/delete`, `unwrap`, `toParams`) + `ApiResponse<T>` / `Page<T>` models. Route ALL HTTP through it.
- **Interceptors** wired in `app.config.ts` via `withInterceptors([authInterceptor, errorInterceptor])` (token from `slam_token`; 401 → logout+redirect; 403 → toast, no logout).
- **Design tokens as SCSS**: apply the token palette to `styles.scss` (Bootstrap variable overrides) so dark is default and red is the accent.
- **Base layout** (`layouts/vertical`) = `<app-sidebar> + <app-topbar> + <router-outlet> + <app-footer>`; sidebar/permission wiring is delivered later (`05-hak-akses` app).
- **i18n baseline**: `public/i18n/{IND,ENG}.json` with `COMMON.*` and `VALIDATION.*` keys reused by all modules.
- Note: `FileService` (Fase 0 module 02), `PermissionService` + `*hasPermission` + dynamic sidebar (Fase 1 module 05) are delivered by those modules — leave clear seams, do not stub duplicates.

## 4. Verification checklist
- [ ] `npm install --legacy-peer-deps` then `npm run build:local` compiles.
- [ ] All HTTP goes through `ApiService`; envelope unwrapped once; `Page<T>` typed.
- [ ] `authInterceptor` attaches Bearer from `slam_token`; `errorInterceptor` handles 401/403 correctly.
- [ ] Dark theme default + SLAM red accent + UPPERCASE headings + red focus ring applied via tokens (no ad-hoc hex).
- [ ] `COMMON.*`/`VALIDATION.*` present in BOTH IND & ENG.
- [ ] Existing files extended, not rebuilt; "Using design-taste-frontend" announced.
