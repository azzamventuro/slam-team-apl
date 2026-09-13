import { Component, inject, input, signal } from '@angular/core';
import { TranslatePipe } from '@ngx-translate/core';

import { FileService } from '../../core/services/file.service';
import { ToastService } from '../../core/services/toast.service';
import { SlamIconComponent } from '../../shared/components/slam-icon/slam-icon.component';
import { JenisIzin, STATUS_IZIN_VARIANT, StatusIzin } from './izin.model';

/**
 * Small presentational pieces shared by the list and the approval queue.
 * Status is never colour-only (design-tokens §11): every chip pairs its
 * colour with a glyph and a translated label.
 */

/** Glyph per approval state — the meaning the colour alone must not carry. */
const STATUS_GLYPH: Record<StatusIzin, string> = {
  menunggu: 'clock',
  disetujui: 'check-square',
  ditolak: 'x',
};

@Component({
  selector: 'app-izin-status',
  imports: [TranslatePipe, SlamIconComponent],
  template: `
    <span class="badge rounded-pill izin-chip" [class]="'text-bg-' + variant()">
      <slam-icon [name]="glyph()" [size]="12" />
      {{ 'IZIN.STATUS.' + status().toUpperCase() | translate }}
    </span>
  `,
  styles: `
    :host { display: inline-flex; }
    .izin-chip {
      display: inline-flex;
      align-items: center;
      gap: var(--slam-space-1);
      font-size: 0.75rem;
      font-weight: 600;
      letter-spacing: 0.04em;
      text-transform: uppercase;
    }
  `,
})
export class IzinStatusChip {
  readonly status = input.required<StatusIzin>();
  variant(): string {
    return STATUS_IZIN_VARIANT[this.status()];
  }
  glyph(): string {
    return STATUS_GLYPH[this.status()];
  }
}

/** Jenis chip — always the info colour (izin/sakit/dinas/pulang_cepat are informational, not attendance faults). */
@Component({
  selector: 'app-izin-jenis',
  imports: [TranslatePipe],
  template: `
    <span class="badge rounded-pill izin-chip izin-chip--jenis">
      {{ 'IZIN.JENIS.' + jenis().toUpperCase() | translate }}
    </span>
  `,
  styles: `
    :host { display: inline-flex; }
    .izin-chip--jenis {
      font-size: 0.75rem;
      font-weight: 600;
      letter-spacing: 0.04em;
      text-transform: uppercase;
      color: var(--slam-info);
      background: rgba(59, 130, 196, 0.16);
      border: 1px solid rgba(59, 130, 196, 0.45);
    }
  `,
})
export class IzinJenisChip {
  readonly jenis = input.required<JenisIzin>();
}

/**
 * Opens the PRIVATE attachment in a new tab. A bare <a href> would not carry
 * the Bearer token, so the bytes are fetched as a blob first (the interceptor
 * adds the header) and shown through an object URL, which is released once
 * the tab has taken it.
 */
@Component({
  selector: 'app-izin-lampiran',
  imports: [TranslatePipe, SlamIconComponent],
  template: `
    <button
      type="button"
      class="btn btn-sm btn-outline-light izin-lampiran"
      [disabled]="busy()"
      (click)="open()"
    >
      <slam-icon name="file" [size]="14" />
      {{ 'IZIN.LAMPIRAN' | translate }}
    </button>
  `,
  styles: `
    :host { display: inline-flex; }
    .izin-lampiran {
      display: inline-flex;
      align-items: center;
      gap: var(--slam-space-1);
      white-space: nowrap;
    }
  `,
})
export class IzinLampiranButton {
  private files = inject(FileService);
  private toast = inject(ToastService);

  readonly uuid = input.required<string>();
  readonly busy = signal(false);

  open(): void {
    this.busy.set(true);
    this.files.imageUrl(this.uuid(), 'original').subscribe({
      next: (url) => {
        this.busy.set(false);
        window.open(url, '_blank', 'noopener');
        // The new tab holds its own reference to the blob once loaded; give
        // it a moment before the URL is revoked from this document.
        setTimeout(() => URL.revokeObjectURL(url), 60_000);
      },
      error: () => {
        this.busy.set(false);
        this.toast.error('IZIN.LAMPIRAN_ERROR');
      },
    });
  }
}
