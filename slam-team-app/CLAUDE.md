# slam-team-app — Claude guide

Guidance for AI agents (and humans). Keep it lean: only what isn't obvious from
the code.

## Stack

- **Angular 21** — standalone components + **signals**, **zoneless** (no
  `zone.js`). New file naming: `app.ts` / `login.ts` (no `.component` suffix),
  classes are `App`, `Login`, `Dashboard`, etc.
- TypeScript 5.9, RxJS 7.8, **Bootstrap 5**, **ngx-translate v18** (IND/ENG).
- Builder: `@angular/build` (esbuild). Tests: **vitest** + jsdom.

## Commands

```bash
npm install --legacy-peer-deps   # REQUIRED — see gotcha below
npm run serve:local              # dev vs local API (:4200, environment.local.ts → :8080)
npm start                        # dev with default environment.ts
npm run build                    # production build   (npm run build:local for local env)
npm test                         # vitest
```

## Structure — `src/app`

```
core/
  guards/        functional CanActivateFn (authGuard)
  interceptors/  functional HttpInterceptorFn (authInterceptor, errorInterceptor)
  services/      singletons (AuthService)
  models/        interfaces (User, ApiResponse<T>, ...)
layouts/
  vertical/      shell = <app-sidebar> + <app-topbar> + <router-outlet> + <app-footer>
  topbar/ sidebar/ footer/
pages/
  auth/login/    login form → AuthService.login → /dashboard
  dashboard/     example protected page
shared/
  pipes/         reusable pipes (safe.pipe)
app.routes.ts    login is public; '' → VerticalLayout guarded by authGuard, children lazy
app.config.ts    providers: router, httpClient(withInterceptors), translate
```

## Conventions

- **Standalone only** — no `NgModule`. Declare deps in the component's
  `imports: [...]`.
- **Signals for state** — e.g. `AuthService.user` is a signal; components read
  `user()`. Templates use new control flow: `@if`, `@for`.
- **Lazy routes** — pages are loaded with `loadComponent: () => import(...)`.
- **i18n** — import `TranslatePipe` into a component's `imports`; use
  `{{ 'KEY' | translate }}`. Keys live in `public/i18n/{IND,ENG}.json`. Switch
  language with `TranslateService.use('ENG')`.
- **HTTP** — `authInterceptor` adds the Bearer token; `errorInterceptor` logs
  out and redirects to `/auth/login` on 401.

## Environments

`src/environments/environment.ts` (default/prod, `apiURL: '/api/v1'`),
`environment.local.ts` (`http://localhost:8080/api/v1`), `environment.example.ts`.
File replacement is wired **only** for the `local` configuration (see
`angular.json`); default dev/prod use `environment.ts`.

## Gotchas

- **`npm install` must use `--legacy-peer-deps`** — npm 11.3 crashes with
  `Cannot read properties of null (reading 'edgesOut')` during peer resolution
  otherwise.
- **Zoneless** — don't assume `zone.js`. Trigger change detection via signals
  or the `async` pipe, not by mutating fields imperatively after async work.
- **Static assets** — Angular 21 serves `public/` at the web root, so
  translations are at `/i18n/*.json` (loader prefix `./i18n/`).
- **Backend** — slam-team-api on `:8080`. JWT is kept in `localStorage`
  (`slam_token`); the user profile in `slam_user`.
