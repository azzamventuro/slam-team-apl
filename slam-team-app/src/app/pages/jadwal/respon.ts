import { Component, inject, input, output, signal } from '@angular/core';
import { TranslatePipe } from '@ngx-translate/core';

import { ToastService } from '../../core/services/toast.service';
import { SlamIconComponent } from '../../shared/components/slam-icon/slam-icon.component';
import { JadwalPeserta } from './penugasan.model';
import { PenugasanService } from './penugasan.service';

/**
 * Terima / Tolak for the assignee's OWN assignment (PATCH /peserta/:id/respon).
 * Bearer-only by design — no `*hasPermission`: the API scopes the write to
 * the participant behind the caller's anggota_id and answers 403/404 for
 * anyone else. Mount it only on rows the panel has already matched to the
 * current user; it renders nothing once the row is no longer `ditugaskan`.
 */
@Component({
  selector: 'app-peserta-respon',
  imports: [TranslatePipe, SlamIconComponent],
  template: `
    @if (peserta().status_tugas === 'ditugaskan') {
      <div class="respon" role="group" [attr.aria-label]="'PESERTA.RESPON_LABEL' | translate">
        <button
          type="button"
          class="btn btn-sm btn-success respon__btn"
          [disabled]="busy()"
          (click)="send('diterima')"
        >
          <slam-icon name="check-square" [size]="16" />
          {{ 'PESERTA.ACCEPT' | translate }}
        </button>
        <button
          type="button"
          class="btn btn-sm btn-outline-danger respon__btn"
          [disabled]="busy()"
          (click)="send('ditolak')"
        >
          <slam-icon name="x" [size]="16" />
          {{ 'PESERTA.REJECT' | translate }}
        </button>
      </div>
    }
  `,
  styles: `
    .respon {
      display: inline-flex;
      gap: var(--slam-space-2);
    }
    .respon__btn {
      display: inline-flex;
      align-items: center;
      gap: var(--slam-space-1);
      min-height: var(--slam-tap-min);
      white-space: nowrap;
    }
  `,
})
export class PesertaRespon {
  private svc = inject(PenugasanService);
  private toast = inject(ToastService);

  readonly peserta = input.required<JadwalPeserta>();
  /** Emits the updated row so the parent can patch its list without a refetch. */
  readonly responded = output<JadwalPeserta>();

  readonly busy = signal(false);

  send(status: 'diterima' | 'ditolak'): void {
    this.busy.set(true);
    this.svc.respon(this.peserta().id, { status_tugas: status }).subscribe({
      next: (row) => {
        this.busy.set(false);
        this.toast.success(status === 'diterima' ? 'PESERTA.ACCEPTED' : 'PESERTA.REJECTED');
        this.responded.emit(row);
      },
      error: () => {
        this.busy.set(false);
        this.toast.error('PESERTA.RESPON_ERROR');
      },
    });
  }
}
