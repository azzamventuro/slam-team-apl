import { Component, input } from '@angular/core';

/**
 * Lightweight inline SVG icon component.
 * Uses hardcoded SVG paths for the small set of icons this app needs.
 * Zero external dependencies — icons are rendered directly in the template.
 *
 * Usage:
 *   <slam-icon name="edit" />
 *   <slam-icon name="trash" size="16" />
 */
@Component({
  selector: 'slam-icon',
  standalone: true,
  template: `
      <svg
        xmlns="http://www.w3.org/2000/svg"
        [attr.width]="size()"
        [attr.height]="size()"
        viewBox="0 0 256 256"
        fill="none"
        stroke="currentColor"
        stroke-width="16"
        stroke-linecap="round"
        stroke-linejoin="round"
        [attr.aria-hidden]="ariaHidden() ? 'true' : null"
        [class]="'slam-icon ' + (cssClass() || '')"
      >
        @switch (name()) {
          @case ('list') {
            <line x1="80" y1="36" x2="216" y2="36" />
            <line x1="80" y1="100" x2="216" y2="100" />
            <line x1="80" y1="164" x2="216" y2="164" />
            <line x1="40" y1="36" x2="40" y2="36" />
            <line x1="40" y1="100" x2="40" y2="100" />
            <line x1="40" y1="164" x2="40" y2="164" />
          }
          @case ('list-dashes') {
            <line x1="104" y1="36" x2="216" y2="36" />
            <line x1="104" y1="100" x2="216" y2="100" />
            <line x1="104" y1="164" x2="216" y2="164" />
            <line x1="40" y1="36" x2="40.01" y2="36" />
            <line x1="40" y1="100" x2="40.01" y2="100" />
            <line x1="40" y1="164" x2="40.01" y2="164" />
          }
          @case ('pencil') {
            <path d="M96 56H48a8 8 0 0 0-8 8v144a8 8 0 0 0 8 8h144a8 8 0 0 0 8-8v-48" />
            <path d="M160 24l72 72-96 96H64v-72Z" />
            <line x1="136" y1="48" x2="208" y2="120" />
          }
          @case ('trash') {
            <path d="M48 56V216a8 8 0 0 0 8 8H200a8 8 0 0 0 8-8V56" />
            <line x1="24" y1="56" x2="232" y2="56" />
            <path d="M104 24h48a8 8 0 0 1 8 8v24H96V32a8 8 0 0 1 8-8Z" />
            <line x1="104" y1="96" x2="104" y2="184" />
            <line x1="152" y1="96" x2="152" y2="184" />
          }
          @case ('grid-four') {
            <rect x="40" y="40" width="72" height="72" rx="8" />
            <rect x="144" y="40" width="72" height="72" rx="8" />
            <rect x="40" y="144" width="72" height="72" rx="8" />
            <rect x="144" y="144" width="72" height="72" rx="8" />
          }
          @case ('x') {
            <line x1="168" y1="96" x2="88" y2="160" />
            <line x1="88" y1="96" x2="168" y2="160" />
          }
          @case ('magnifying-glass') {
            <circle cx="116" cy="116" r="84" />
            <line x1="176" y1="176" x2="232" y2="232" />
          }
          @case ('list-numbers') {
            <line x1="104" y1="40" x2="216" y2="40" />
            <line x1="104" y1="100" x2="216" y2="100" />
            <line x1="104" y1="160" x2="216" y2="160" />
            <path d="M56 52.8V40H40v24h24" />
            <path d="M40 88h16l-16 20h16" />
            <path d="M56 148.8V140H40v24h24" />
          }
          @case ('star') {
            <polygon points="128,16 156,96 240,96 172,148 196,228 128,180 60,228 84,148 16,96 100,96" />
          }
          @case ('gear-six') {
            <path d="M128 176a48 48 0 1 0 0-96 48 48 0 0 0 0 96Z" />
            <path d="M224 128h-24" />
            <path d="M56 128H32" />
            <path d="M196.8 196.8l-16.97-16.97" />
            <path d="M76.17 76.17L59.2 59.2" />
            <path d="M196.8 59.2l-16.97 16.97" />
            <path d="M76.17 179.83L59.2 196.8" />
          }
          @case ('compass') {
            <circle cx="128" cy="128" r="96" />
            <polygon points="168,128 128,92 88,128 128,164" fill="currentColor" />
          }
          @case ('monitor') {
            <rect x="24" y="24" width="208" height="160" rx="8" />
            <line x1="128" y1="200" x2="128" y2="232" />
            <line x1="96" y1="232" x2="160" y2="232" />
          }
          @case ('user-circle') {
            <circle cx="128" cy="96" r="48" />
            <path d="M216 192c0 35.37-35.82 56-88 56s-88-20.63-88-56" />
          }
          @case ('sliders') {
            <line x1="48" y1="80" x2="208" y2="80" />
            <line x1="48" y1="176" x2="208" y2="176" />
            <circle cx="104" cy="80" r="24" fill="var(--slam-surface-2)" stroke="currentColor" />
            <circle cx="168" cy="176" r="24" fill="var(--slam-surface-2)" stroke="currentColor" />
          }
          @case ('file') {
            <path d="M144 24H48a8 8 0 0 0-8 8V224a8 8 0 0 0 8 8H208a8 8 0 0 0 8-8V88Z" />
            <path d="M144 24v64h64" />
          }
          @case ('map-pin') {
            <path d="M128 208s-72-56-72-112a72 72 0 0 1 144 0c0 56-72 112-72 112Z" />
            <circle cx="128" cy="96" r="24" />
          }
          @case ('calendar') {
            <rect x="32" y="48" width="192" height="176" rx="8" />
            <line x1="32" y1="96" x2="224" y2="96" />
            <line x1="80" y1="24" x2="80" y2="56" />
            <line x1="176" y1="24" x2="176" y2="56" />
          }
          @case ('check-square') {
            <rect x="32" y="32" width="192" height="192" rx="8" />
            <path d="M80 128l32 32 64-64" />
          }
          @case ('clock') {
            <circle cx="128" cy="128" r="96" />
            <polyline points="128,72 128,128 176,128" />
          }
          @case ('link') {
            <path d="M100 156a48 48 0 0 0 56 0l40-40a48 48 0 0 0-67.88-67.88l-11.13 11.12" />
            <path d="M156 100a48 48 0 0 0-56 0l-40 40a48 48 0 0 0 67.88 67.88l11.12-11.12" />
          }
          @case ('newspaper') {
            <path d="M32 56a8 8 0 0 1 8-8h168a8 8 0 0 1 8 8v160a8 8 0 0 1-8 8H40a8 8 0 0 1-8-8Z" />
            <line x1="72" y1="32" x2="72" y2="48" />
            <line x1="112" y1="32" x2="112" y2="48" />
            <line x1="152" y1="32" x2="152" y2="48" />
            <line x1="72" y1="96" x2="184" y2="96" />
            <line x1="72" y1="128" x2="184" y2="128" />
            <line x1="72" y1="160" x2="120" y2="160" />
          }
          @case ('chart-bar') {
            <rect x="32" y="128" width="48" height="96" rx="4" />
            <rect x="104" y="64" width="48" height="160" rx="4" />
            <rect x="176" y="96" width="48" height="128" rx="4" />
          }
          @case ('id') {
            <rect x="24" y="48" width="208" height="160" rx="8" />
            <circle cx="88" cy="112" r="28" />
            <path d="M48 176c0-24 24-32 40-32s32 8 32 32" />
            <line x1="144" y1="104" x2="200" y2="104" />
            <line x1="144" y1="128" x2="200" y2="128" />
            <line x1="144" y1="152" x2="180" y2="152" />
          }
          @case ('bell') {
            <path d="M56 96a72 72 0 0 1 144 0c0 32 16 64 16 96H40c0-32 16-64 16-96Z" />
            <path d="M88 208v16a40 40 0 0 0 80 0v-16" />
          }
          @case ('map-trifold') {
            <path d="M4 6l96-4v200l-96 4Z" />
            <path d="M100 2l96-4v200l-96 4Z" />
            <path d="M196 2l56-4v200l-56 4Z" />
          }
          @case ('shield') {
            <path d="M128 24L32 64v64c0 56 40 96 96 112 56-16 96-56 96-112V64Z" />
          }
          @case ('chart-line') {
            <polyline points="32,184 88,120 136,152 224,56" />
            <polyline points="176,56 224,56 224,104" />
          }
          @case ('note-pencil') {
            <path d="M176 24l32 32L88 176H56v-32Z" />
            <path d="M144 56l32 32" />
          }
          @case ('house') {
            <path d="M24 104l104-80 104 80v112a8 8 0 0 1-8 8H32a8 8 0 0 1-8-8Z" />
            <polyline points="96,224 96,128 160,128 160,224" />
          }
          @case ('sign-out') {
            <path d="M56 200H40a8 8 0 0 1-8-8V64a8 8 0 0 1 8-8h16" />
            <polyline points="112,160 152,128 112,96" />
            <line x1="152" y1="128" x2="68" y2="128" />
          }
        }
      </svg>
  `,
  styles: `
    :host {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      line-height: 0;
    }

    .slam-icon {
      vertical-align: middle;
      flex-shrink: 0;
    }
  `,
})
export class SlamIconComponent {
  /** Icon name — maps to a hardcoded SVG path. */
  readonly name = input.required<string>();

  /** Icon size in pixels (default: 20). */
  readonly size = input(20);

  /** Optional extra CSS class. */
  readonly cssClass = input<string>('');

  /** Whether the icon is purely decorative (aria-hidden). */
  readonly ariaHidden = input(true);
}
