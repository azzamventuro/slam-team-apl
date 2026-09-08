# SLAM Team — Design Tokens (single source)

**This is the authoritative design-token source for the whole app.** Every APP prompt in this workflow references this file. When these tokens differ from `design-taste-frontend` (taste-skill) defaults, **these tokens win** — taste-skill enforces the *method* (anti-generic layout, spacing discipline, dark-mode correctness), this file supplies the *values*. Bootstrap 5 is already installed; use the CSS-variable / SCSS overrides below rather than ad-hoc hex in components.

Derived from the KTA card (rancangan Bab 7.7 / 7.8 / 11.3): near-black background, dark-gray surfaces, SLAM red accent from the logo, white text, muted gray labels; bold, wide letter-spaced UPPERCASE headings; ID-1 rounded-corner feel (card radius ≈ 3.18 mm → ~12px at screen scale).

Rules for executors:
- Never hardcode a hex that exists here — use the CSS variable (`var(--slam-*)`) or the Bootstrap variable it maps to.
- Dark is the **default** theme (matches the card). Light is a secondary surface for print/preview and public pages.
- Headings are UPPERCASE with wide letter-spacing; body text is normal case.
- Red is an **accent**, not a fill — use it for primary actions, active state, dividers, and focus. Large red fills only on the KTA card / hero, never on data tables.

---

## 1. Color

### Core palette (dark, default)

| Token | Hex | Use |
|-------|-----|-----|
| `--slam-bg` | `#0B0B0D` | App background (near-black, from card latar) |
| `--slam-surface` | `#141418` | Cards, panels, sidebar, navbar |
| `--slam-surface-2` | `#1E1E24` | Raised surface: modals, dropdowns, table header, active row |
| `--slam-surface-3` | `#2A2A32` | Inputs, hover state on surface-2 |
| `--slam-primary` | `#E11D2A` | SLAM red — primary action, active nav, links, focus ring |
| `--slam-primary-hover` | `#C2151F` | Hover/pressed red |
| `--slam-primary-soft` | `rgba(225,29,42,.14)` | Red tint bg: badges, active-row highlight, selection |
| `--slam-text` | `#F5F5F7` | Primary text (white-ish, not pure #fff — softer on black) |
| `--slam-text-muted` | `#9A9AA6` | Labels, secondary text, placeholder, disabled |
| `--slam-text-dim` | `#6B6B76` | Captions, timestamps, table meta |
| `--slam-border` | `#2C2C34` | Card/table/input borders, dividers |
| `--slam-border-strong` | `#3A3A44` | Focused/active borders |
| `--slam-red-line` | `#E11D2A` | The card's thin red divider / accent bar |

### Status

| Token | Hex | Bootstrap |
|-------|-----|-----------|
| `--slam-success` | `#2FB56B` | `$success` — hadir, approved, active |
| `--slam-warning` | `#E7A417` | `$warning` — terlambat, pending, expiring |
| `--slam-danger` | `#E11D2A` | `$danger` — alfa, revoked, error (shares SLAM red) |
| `--slam-info` | `#3B82C4` | `$info` — izin, dinas, informational |

Attendance status colors (rancangan absensi): hadir → success, terlambat/pulang_cepat/hadir_luar_radius → warning, izin/sakit/dinas → info, alfa → danger, manual override → text-muted badge.

### Light theme (public pages, KTA on-screen preview, print)

| Token | Hex |
|-------|-----|
| `--slam-bg` | `#F5F6F8` |
| `--slam-surface` | `#FFFFFF` |
| `--slam-surface-2` | `#F0F1F4` |
| `--slam-surface-3` | `#E7E9EE` |
| `--slam-text` | `#141418` |
| `--slam-text-muted` | `#5A5A66` |
| `--slam-text-dim` | `#8A8A96` |
| `--slam-border` | `#DCDEE4` |
| `--slam-border-strong` | `#C4C7D0` |

Primary/status colors are unchanged between themes. **The printed KTA itself always uses the dark card palette regardless of app theme** (near-black bg, red accents) — the light theme is only the surrounding app chrome / on-screen preview frame.

---

## 2. Typography

**Font family** — system stack (no webfont dependency, fast, no license):
```
--slam-font-sans: "Inter", system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
--slam-font-heading: "Inter", system-ui, sans-serif; /* same face; headings distinguished by weight+tracking+case */
--slam-font-mono: "JetBrains Mono", ui-monospace, "SF Mono", "Cascadia Code", Consolas, monospace; /* NRA, tokens, codes */
```
If a display face is later wanted for hero/KTA headings, add one webfont only — otherwise Inter Black at wide tracking already reads as the card's bold-caps voice.

**Type scale** (1.250 major-third, base 16px):

| Token | Size | Line-height | Weight | Letter-spacing | Case |
|-------|------|-------------|--------|----------------|------|
| `--slam-fs-display` | 40px / 2.5rem | 1.1 | 800 | 0.06em | UPPERCASE |
| `--slam-fs-h1` | 32px / 2rem | 1.15 | 800 | 0.05em | UPPERCASE |
| `--slam-fs-h2` | 26px / 1.625rem | 1.2 | 700 | 0.04em | UPPERCASE |
| `--slam-fs-h3` | 21px / 1.3125rem | 1.25 | 700 | 0.03em | UPPERCASE |
| `--slam-fs-h4` | 18px / 1.125rem | 1.3 | 600 | 0.02em | Title Case ok |
| `--slam-fs-lg` | 18px / 1.125rem | 1.5 | 400 | 0 | body |
| `--slam-fs-base` | 16px / 1rem | 1.55 | 400 | 0 | body |
| `--slam-fs-sm` | 14px / 0.875rem | 1.5 | 400 | 0 | secondary/table |
| `--slam-fs-xs` | 12px / 0.75rem | 1.4 | 500 | 0.04em | labels/badges, often UPPERCASE |
| `--slam-fs-mono` | 15px | 1.4 | 500 | 0.02em | NRA / token / QR payload |

Section eyebrow / label style (card "NAMA / TTL / ALAMAT / NRA" labels): `--slam-fs-xs`, weight 600, `letter-spacing: 0.08em`, UPPERCASE, color `--slam-text-muted`.

Weights available: 400 (body), 500 (label/emphasis), 600 (subhead), 700 (heading), 800 (display/hero).

---

## 3. Spacing

4px base grid. Use these steps, don't invent in-betweens.

| Token | px |
|-------|----|
| `--slam-space-0` | 0 |
| `--slam-space-1` | 4 |
| `--slam-space-2` | 8 |
| `--slam-space-3` | 12 |
| `--slam-space-4` | 16 |
| `--slam-space-5` | 24 |
| `--slam-space-6` | 32 |
| `--slam-space-7` | 48 |
| `--slam-space-8` | 64 |
| `--slam-space-9` | 96 |

Defaults: card/panel padding `--slam-space-5` (24); section vertical rhythm `--slam-space-7` (48); form control gap `--slam-space-4` (16); inline gap `--slam-space-2` (8); page gutter `--slam-space-5` on mobile / `--slam-space-6`+ on desktop.

---

## 4. Radius

ID-1 card feel — rounded but not pill.

| Token | px | Use |
|-------|----|-----|
| `--slam-radius-sm` | 6 | Inputs, small buttons, badges |
| `--slam-radius` | 10 | Default: buttons, cards, dropdowns |
| `--slam-radius-lg` | 12 | Panels, modals, KTA card frame (≈3.18mm @ screen) |
| `--slam-radius-xl` | 16 | Hero / large feature cards |
| `--slam-radius-pill` | 999 | Status chips, avatar, toggle |

---

## 5. Elevation / shadow

On near-black, elevation comes from **surface lightness first, shadow second** (shadows are subtle on dark).

| Token | Value |
|-------|-------|
| `--slam-shadow-sm` | `0 1px 2px rgba(0,0,0,.4)` |
| `--slam-shadow` | `0 4px 12px rgba(0,0,0,.45)` |
| `--slam-shadow-lg` | `0 12px 32px rgba(0,0,0,.55)` |
| `--slam-shadow-focus` | `0 0 0 3px rgba(225,29,42,.35)` (focus ring, uses SLAM red) |
| `--slam-glow-primary` | `0 0 0 1px rgba(225,29,42,.5), 0 4px 16px rgba(225,29,42,.25)` (active primary / KTA accent) |

Focus is **always** the red ring (`--slam-shadow-focus`) — accessibility baseline, never remove outline without replacing it.

---

## 6. Control heights & sizing

| Token | Value | Use |
|-------|-------|-----|
| `--slam-control-h-sm` | 32px | Compact table actions, filters |
| `--slam-control-h` | 40px | Default input / button / select |
| `--slam-control-h-lg` | 48px | Primary CTA, login, mobile touch targets |
| `--slam-tap-min` | 44px | Minimum touch target (absen screen, mobile) — a11y |
| `--slam-icon` | 20px | Default inline icon |
| `--slam-icon-lg` | 24px | Nav / action icons |
| `--slam-border-w` | 1px | Default border |
| `--slam-focus-w` | 3px | Focus ring |
| `--slam-sidebar-w` | 264px | Vertical sidebar |
| `--slam-topbar-h` | 60px | Topbar |
| `--slam-container-max` | 1320px | Content max-width |

Input/button vertical padding derives from control height; do not set both height and large padding.

---

## 7. Motion

| Token | Value |
|-------|-------|
| `--slam-transition-fast` | 120ms ease |
| `--slam-transition` | 180ms ease |
| `--slam-transition-slow` | 280ms cubic-bezier(.4,0,.2,1) |

Respect `prefers-reduced-motion: reduce` → drop transitions/animations. Keep motion functional (hover, focus, enter/leave), not decorative.

---

## 8. CSS custom properties (drop into `:root`)

Put this in `src/styles.scss`. `[data-bs-theme="light"]` overrides the surface/text tokens for public + print.

```scss
:root {
  // color — dark (default)
  --slam-bg: #0B0B0D;
  --slam-surface: #141418;
  --slam-surface-2: #1E1E24;
  --slam-surface-3: #2A2A32;
  --slam-primary: #E11D2A;
  --slam-primary-hover: #C2151F;
  --slam-primary-soft: rgba(225,29,42,.14);
  --slam-text: #F5F5F7;
  --slam-text-muted: #9A9AA6;
  --slam-text-dim: #6B6B76;
  --slam-border: #2C2C34;
  --slam-border-strong: #3A3A44;
  --slam-success: #2FB56B;
  --slam-warning: #E7A417;
  --slam-danger: #E11D2A;
  --slam-info: #3B82C4;

  // type
  --slam-font-sans: "Inter", system-ui, -apple-system, "Segoe UI", Roboto, Arial, sans-serif;
  --slam-font-heading: var(--slam-font-sans);
  --slam-font-mono: "JetBrains Mono", ui-monospace, "SF Mono", Consolas, monospace;

  // radius / shadow / motion
  --slam-radius-sm: 6px;
  --slam-radius: 10px;
  --slam-radius-lg: 12px;
  --slam-radius-xl: 16px;
  --slam-shadow-sm: 0 1px 2px rgba(0,0,0,.4);
  --slam-shadow: 0 4px 12px rgba(0,0,0,.45);
  --slam-shadow-lg: 0 12px 32px rgba(0,0,0,.55);
  --slam-shadow-focus: 0 0 0 3px rgba(225,29,42,.35);
  --slam-transition: 180ms ease;

  // sizing
  --slam-control-h: 40px;
  --slam-control-h-lg: 48px;
  --slam-tap-min: 44px;
  --slam-sidebar-w: 264px;
  --slam-topbar-h: 60px;
}

[data-bs-theme="light"] {
  --slam-bg: #F5F6F8;
  --slam-surface: #FFFFFF;
  --slam-surface-2: #F0F1F4;
  --slam-surface-3: #E7E9EE;
  --slam-text: #141418;
  --slam-text-muted: #5A5A66;
  --slam-text-dim: #8A8A96;
  --slam-border: #DCDEE4;
  --slam-border-strong: #C4C7D0;
}
```

---

## 9. Bootstrap 5 mapping (SCSS overrides)

Bootstrap 5.3+ ships a dark mode via `data-bs-theme`. Override Bootstrap's SCSS `$variables` **before** importing Bootstrap so components inherit SLAM values, then reconcile with the CSS vars above. Set the app shell to dark by default: `<html data-bs-theme="dark">`.

`src/styles.scss` (before `@import "bootstrap/scss/bootstrap"`):

```scss
// 1. SLAM theme colors -> Bootstrap theme map
$primary:   #E11D2A;
$success:   #2FB56B;
$warning:   #E7A417;
$danger:    #E11D2A;
$info:      #3B82C4;

// 2. dark surfaces (Bootstrap dark-mode body/surface)
$body-bg-dark:        #0B0B0D;
$body-color-dark:     #F5F5F7;
$border-color-dark:   #2C2C34;

// 3. typography
$font-family-sans-serif: "Inter", system-ui, -apple-system, "Segoe UI", Roboto, Arial, sans-serif;
$font-family-monospace:  "JetBrains Mono", ui-monospace, "SF Mono", Consolas, monospace;
$font-size-base:      1rem;      // 16px
$headings-font-weight: 800;
$headings-line-height: 1.15;

// 4. shape / sizing
$border-radius:       .625rem;   // 10px  -> --slam-radius
$border-radius-sm:    .375rem;   // 6px
$border-radius-lg:    .75rem;    // 12px
$input-btn-padding-y: .5rem;
$input-btn-padding-x: 1rem;
$input-btn-font-size: 1rem;
$input-height:        40px;      // --slam-control-h

// 5. focus ring = SLAM red
$focus-ring-width:    3px;
$focus-ring-color:    rgba(225,29,42,.35);
$focus-ring-opacity:  1;
$btn-focus-width:     3px;

@import "bootstrap/scss/bootstrap";
```

After the import, bridge Bootstrap's dark CSS vars to the SLAM tokens so custom components and Bootstrap components agree:

```scss
[data-bs-theme="dark"] {
  --bs-body-bg: var(--slam-bg);
  --bs-body-color: var(--slam-text);
  --bs-border-color: var(--slam-border);
  --bs-secondary-color: var(--slam-text-muted);
  --bs-tertiary-bg: var(--slam-surface);
  --bs-emphasis-color: var(--slam-text);
  --bs-link-color: var(--slam-primary);
  --bs-link-hover-color: var(--slam-primary-hover);
}

// cards / panels
.card, .dropdown-menu, .modal-content {
  background-color: var(--slam-surface);
  border-color: var(--slam-border);
  border-radius: var(--slam-radius-lg);
}
.table > :not(caption) > * > * { background-color: transparent; }
.table thead th {
  background-color: var(--slam-surface-2);
  text-transform: uppercase;
  letter-spacing: .04em;
  font-size: .75rem;
  color: var(--slam-text-muted);
}
.form-control, .form-select {
  background-color: var(--slam-surface-3);
  border-color: var(--slam-border);
  color: var(--slam-text);
}
.form-control::placeholder { color: var(--slam-text-dim); }

// headings = card voice
h1,h2,h3,.h1,.h2,.h3 { text-transform: uppercase; letter-spacing: .04em; }

// SLAM red divider (the card's thin red line)
.slam-divider { height: 2px; background: var(--slam-primary); border: 0; opacity: 1; }

// active nav / active row
.nav-link.active, .list-group-item.active { color: var(--slam-primary); background: var(--slam-primary-soft); }
```

Bootstrap classes you get for free that already match: `.btn-primary` (SLAM red), `.badge text-bg-success/warning/danger/info`, `.rounded-3`, `.shadow`, `.text-uppercase`. Prefer these over custom CSS. Use `.text-bg-*` and `.border` utilities rather than re-declaring colors.

---

## 10. Component quick-reference

- **Primary button**: `.btn .btn-primary`, height 40 (48 for hero/login), UPPERCASE optional for CTAs, red fill, `--slam-shadow-focus` on focus.
- **Secondary button**: `.btn .btn-outline-light` (dark) / `.btn-outline-secondary` (light).
- **Card / panel**: surface `#141418`, border `--slam-border`, radius 12, padding 24.
- **Table**: transparent rows on surface, header = surface-2 uppercase muted xs; active/selected row = `--slam-primary-soft`.
- **Status chip**: `.badge .rounded-pill .text-bg-{success|warning|danger|info}`, `--slam-fs-xs`.
- **Input**: surface-3 bg, border `--slam-border`, radius 6–10, height 40, red focus ring.
- **NRA / token / QR payload**: mono font, letter-spacing 0.02em (e.g. `3573 10 02 021`).
- **KTA card frame** (on-screen preview): radius 12, dark palette locked, thin red divider, wide-tracked uppercase heading — mirrors the printed card.
- **Sidebar** (generated from `mst_modul`): surface bg, 264px, active item = red text + red-soft bg + 3px left red bar.

---

## 11. Note for every APP prompt

Each APP prompt must instruct the executor to:
1. Announce **"Using design-taste-frontend"** and run its pre-flight/audit.
2. Load **this file** and enforce these tokens; **where they differ from taste-skill defaults, this file wins.**
3. Use the CSS variables / Bootstrap overrides above — no ad-hoc hex, no new font dependency, dark default.
4. Meet a11y baselines: red focus ring never removed, 44px min touch target on mobile/absen, `prefers-reduced-motion` honored, status never conveyed by color alone (pair with icon/label).
5. "If `blog-fe` source is restored, mirror its layout/menu for this screen; otherwise follow these design tokens."
