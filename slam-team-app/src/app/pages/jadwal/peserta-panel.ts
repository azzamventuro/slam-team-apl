import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, effect, inject, signal } from '@angular/core';
import { toObservable, toSignal } from '@angular/core/rxjs-interop';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import {
  EMPTY,
  Observable,
  catchError,
  debounceTime,
  distinctUntilChanged,
  expand,
  map,
  of,
  reduce,
  switchMap,
} from 'rxjs';

import { Anggota } from '../../core/models/anggota.model';
import { emptyPage, Page } from '../../core/models/api.model';
import { Instansi } from '../../core/models/instansi.model';
import { AuthService } from '../../core/services/auth.service';
import { ToastService } from '../../core/services/toast.service';
import { SlamIconComponent } from '../../shared/components/slam-icon/slam-icon.component';
import { HasPermissionDirective } from '../../shared/directives/has-permission.directive';
import { formatDate, formatTimeWithZone } from '../../shared/util/tz';
import { AnggotaService } from '../anggota/anggota.service';
import { InstansiService } from '../instansi/instansi.service';
import {
  JadwalPeserta,
  JadwalRingkas,
  JadwalSesi,
  PERAN_PESERTA,
  PeranPeserta,
  STATUS_TUGAS,
  StatusTugas,
} from './penugasan.model';
import { PenugasanService } from './penugasan.service';
import { PesertaRespon } from './respon';

/** Page size of the bulk picker; the API caps per_page at 100. */
const PICK_PER_PAGE = 50;
/** Bulk endpoint accepts at most 500 ids per call. */
const BULK_MAX = 500;

/**
 * Inside a `switchMap`, an inner error would terminate the whole picker stream
 * for the rest of the component's life; degrade to an empty page instead (the
 * interceptor has already toasted the failure).
 */
function safePage<T>(perPage = 20) {
  return (src: Observable<Page<T>>) => src.pipe(catchError(() => of(emptyPage<T>(perPage))));
}

/**
 * Participants of one schedule (`/jadwal/:id/peserta`): the list, a single
 * assign form (anggota picker), a bulk assign form (multi-select with an
 * instansi expander), remove, and — on the caller's own row — Terima/Tolak.
 *
 * Assign/remove are gated `jadwal.assign` in the template (UX only; the Go
 * handler answers 403). Respon is bearer-only and matched to the caller by
 * `user.anggota_id`.
 */
@Component({
  selector: 'app-peserta-panel',
  imports: [
    FormsModule,
    RouterLink,
    TranslatePipe,
    HasPermissionDirective,
    SlamIconComponent,
    PesertaRespon,
  ],
  templateUrl: './peserta-panel.html',
  styleUrl: './peserta-panel.scss',
})
export class PesertaPanel {
  private route = inject(ActivatedRoute);
  private svc = inject(PenugasanService);
  private anggotaSvc = inject(AnggotaService);
  private instansiSvc = inject(InstansiService);
  private auth = inject(AuthService);
  private toast = inject(ToastService);
  private translate = inject(TranslateService);

  readonly peranOptions = PERAN_PESERTA;
  readonly statusOptions = STATUS_TUGAS;
  readonly bulkMax = BULK_MAX;

  // ── schedule ────────────────────────────────────────────────────────────
  readonly jadwalId = toSignal(this.route.paramMap.pipe(map((p) => Number(p.get('id')))), {
    initialValue: 0,
  });
  readonly jadwal = signal<JadwalRingkas | null>(null);
  readonly sesiList = signal<JadwalSesi[]>([]);

  // ── participant list ────────────────────────────────────────────────────
  readonly loading = signal(true);
  readonly rows = signal<JadwalPeserta[]>([]);
  readonly q = signal('');
  readonly statusFilter = signal<StatusTugas | ''>('');

  /** Ids already on this schedule — greyed out in the bulk picker. */
  readonly assignedIds = computed(() => new Set(this.rows().map((r) => r.anggota_id)));
  readonly counts = computed(() => {
    const c: Record<StatusTugas, number> = { ditugaskan: 0, diterima: 0, ditolak: 0, izin: 0 };
    for (const r of this.rows()) c[r.status_tugas]++;
    return c;
  });

  /** The caller's anggota_id — null for accounts not linked to a member. */
  private myAnggotaId = computed(() => this.auth.user()?.anggota_id ?? null);

  // ── assign: shared ──────────────────────────────────────────────────────
  readonly mode = signal<'single' | 'bulk'>('single');
  readonly peran = signal<PeranPeserta>('peserta');
  readonly wajibAbsen = signal(true);
  readonly sesiId = signal<number | null>(null);
  readonly saving = signal(false);

  // ── assign: single (anggota picker) ─────────────────────────────────────
  readonly pickerQ = signal('');
  readonly pickerOpen = signal(false);
  readonly selectedAnggota = signal<Anggota | null>(null);
  readonly keterangan = signal('');
  private pickerPage = toSignal(
    toObservable(this.pickerQ).pipe(
      debounceTime(250),
      distinctUntilChanged(),
      switchMap((q) =>
        q.trim().length < 2
          ? of(emptyPage<Anggota>())
          : this.anggotaSvc
              .list({ page: 1, per_page: 10, q: q.trim(), status: 'aktif' })
              .pipe(safePage<Anggota>()),
      ),
    ),
    { initialValue: emptyPage<Anggota>() },
  );
  readonly pickerResults = computed(() => this.pickerPage().items);

  // ── assign: bulk (multi-select) ─────────────────────────────────────────
  readonly bulkInstansi = signal<number | null>(null);
  readonly bulkQ = signal('');
  readonly bulkPage = signal(1);
  readonly selected = signal<Set<number>>(new Set());
  readonly expanding = signal(false);
  readonly instansiList = toSignal(
    this.instansiSvc.list({ page: 1, per_page: 100, q: '' }).pipe(
      map((p) => p.items),
      catchError(() => of([] as Instansi[])),
    ),
    { initialValue: [] as Instansi[] },
  );
  private bulkQuery = computed(() => ({
    page: this.bulkPage(),
    per_page: PICK_PER_PAGE,
    q: this.bulkQ().trim(),
    instansi_id: this.bulkInstansi(),
    status: 'aktif',
  }));
  readonly bulkResult = toSignal(
    toObservable(this.bulkQuery).pipe(
      debounceTime(250),
      switchMap((q) => this.anggotaSvc.list(q).pipe(safePage<Anggota>(PICK_PER_PAGE))),
    ),
    { initialValue: emptyPage<Anggota>(PICK_PER_PAGE) },
  );
  readonly bulkItems = computed(() => this.bulkResult().items);
  readonly selectedCount = computed(() => this.selected().size);
  /** Every selectable candidate on the current page is already ticked. */
  readonly pageAllSelected = computed(() => {
    const sel = this.selected();
    const pick = this.bulkItems().filter((a) => !this.assignedIds().has(a.id));
    return pick.length > 0 && pick.every((a) => sel.has(a.id));
  });

  constructor() {
    // Reload whenever the route's :id changes (the component is reused).
    effect(() => {
      const id = this.jadwalId();
      if (!id) return;
      this.loadJadwal(id);
      this.load();
    });
  }

  // ── loading ─────────────────────────────────────────────────────────────
  private loadJadwal(id: number): void {
    this.svc.jadwal(id).subscribe({
      next: (j) => this.jadwal.set(j),
      error: () => this.toast.error('PESERTA.JADWAL_LOAD_ERROR'),
    });
    this.svc.sesi(id).subscribe({
      next: (s) => this.sesiList.set(s),
      error: () => this.sesiList.set([]),
    });
  }

  load(): void {
    this.loading.set(true);
    this.svc
      .list(this.jadwalId(), { q: this.q().trim(), status_tugas: this.statusFilter() })
      .subscribe({
        next: (rows) => {
          this.rows.set(rows);
          this.loading.set(false);
        },
        error: () => {
          this.toast.error('PESERTA.LOAD_ERROR');
          this.loading.set(false);
        },
      });
  }

  onStatusFilter(val: string): void {
    this.statusFilter.set(val as StatusTugas | '');
    this.load();
  }

  // ── row helpers ─────────────────────────────────────────────────────────
  isMine(row: JadwalPeserta): boolean {
    const mine = this.myAnggotaId();
    return mine !== null && row.anggota_id === mine;
  }

  onResponded(updated: JadwalPeserta): void {
    this.rows.update((list) => list.map((r) => (r.id === updated.id ? { ...r, ...updated } : r)));
  }

  remove(row: JadwalPeserta): void {
    const nama = row.anggota_nama ?? `#${row.anggota_id}`;
    if (!confirm(this.translate.instant('PESERTA.CONFIRM_REMOVE', { nama }))) return;
    this.svc.remove(this.jadwalId(), row.anggota_id).subscribe({
      next: () => {
        this.toast.success('PESERTA.REMOVED');
        this.load();
      },
      error: (err: HttpErrorResponse) => {
        if (err.status !== 403) this.toast.error('PESERTA.REMOVE_ERROR');
      },
    });
  }

  /** Status is never colour-only — the template pairs this with a label. */
  statusClass(s: StatusTugas): string {
    switch (s) {
      case 'diterima':
        return 'text-bg-success';
      case 'ditolak':
        return 'text-bg-danger';
      case 'izin':
        return 'text-bg-info';
      default:
        return 'text-bg-warning';
    }
  }

  sesiLabel(row: JadwalPeserta): string {
    if (row.sesi_id === null) return this.translate.instant('PESERTA.ALL_SESI');
    return row.sesi_tanggal ? formatDate(row.sesi_tanggal, 'UTC') : `#${row.sesi_id}`;
  }

  /** Session option label: local calendar date + start time in the schedule's zone. */
  sesiOption(s: JadwalSesi): string {
    return `${formatDate(s.tanggal_lokal, 'UTC')} · ${formatTimeWithZone(s.mulai_utc, s.timezone)}`;
  }

  /** `08:00:00` → `08:00` (the API serialises the time column with seconds). */
  hhmm(t: string): string {
    return t.slice(0, 5);
  }

  /** `type:date` columns arrive as midnight UTC — format them in UTC to keep the calendar day. */
  dateOnly(value: string | null | undefined): string {
    return value ? formatDate(value, 'UTC') : '—';
  }

  // ── single assign ───────────────────────────────────────────────────────
  pick(a: Anggota): void {
    this.selectedAnggota.set(a);
    this.pickerQ.set('');
    this.pickerOpen.set(false);
  }

  clearPick(): void {
    this.selectedAnggota.set(null);
  }

  submitSingle(): void {
    const a = this.selectedAnggota();
    if (!a || this.saving()) return;
    this.saving.set(true);
    this.svc
      .assign(this.jadwalId(), {
        anggota_id: a.id,
        sesi_id: this.sesiId(),
        peran_peserta: this.peran(),
        wajib_absen: this.wajibAbsen(),
        keterangan: this.keterangan().trim() || null,
      })
      .subscribe({
        next: () => {
          this.saving.set(false);
          this.toast.success('PESERTA.ASSIGNED');
          this.selectedAnggota.set(null);
          this.keterangan.set('');
          this.load();
        },
        error: (err: HttpErrorResponse) => {
          this.saving.set(false);
          // 409 = already on this schedule; 403 is toasted by the interceptor.
          if (err.status === 409) this.toast.warning('PESERTA.DUPLICATE');
          else if (err.status !== 403) this.toast.error('PESERTA.ASSIGN_ERROR');
        },
      });
  }

  // ── bulk assign ─────────────────────────────────────────────────────────
  onBulkFilter(): void {
    this.bulkPage.set(1);
  }

  onBulkInstansi(val: string): void {
    this.bulkInstansi.set(val ? Number(val) : null);
    this.bulkPage.set(1);
  }

  isSelected(id: number): boolean {
    return this.selected().has(id);
  }

  toggle(id: number): void {
    this.selected.update((s) => {
      const next = new Set(s);
      if (next.has(id)) next.delete(id);
      else if (next.size < BULK_MAX) next.add(id);
      return next;
    });
  }

  togglePage(): void {
    const all = this.pageAllSelected();
    const ids = this.bulkItems()
      .filter((a) => !this.assignedIds().has(a.id))
      .map((a) => a.id);
    this.selected.update((s) => {
      const next = new Set(s);
      for (const id of ids) {
        if (all) next.delete(id);
        else if (next.size < BULK_MAX) next.add(id);
      }
      return next;
    });
  }

  /**
   * By-instansi bulk: walk every page of that instansi's active anggota and
   * add them all to the selection (the API pages at 100 max).
   */
  selectWholeInstansi(): void {
    const instansiId = this.bulkInstansi();
    if (!instansiId || this.expanding()) return;
    this.expanding.set(true);
    const fetch = (page: number) =>
      this.anggotaSvc.list({ page, per_page: 100, q: '', instansi_id: instansiId, status: 'aktif' });
    fetch(1)
      .pipe(
        expand((p: Page<Anggota>) => (p.page < p.last_page ? fetch(p.page + 1) : EMPTY)),
        reduce((ids, p) => {
          for (const a of p.items) ids.push(a.id);
          return ids;
        }, [] as number[]),
      )
      .subscribe({
        next: (ids) => {
          const assigned = this.assignedIds();
          this.selected.update((s) => {
            const next = new Set(s);
            for (const id of ids) {
              if (next.size >= BULK_MAX) break;
              if (!assigned.has(id)) next.add(id);
            }
            return next;
          });
          this.expanding.set(false);
          this.toast.info(
            this.translate.instant('PESERTA.INSTANSI_EXPANDED', { jumlah: this.selected().size }),
          );
        },
        error: () => {
          this.expanding.set(false);
          this.toast.error('PESERTA.LOAD_ERROR');
        },
      });
  }

  clearSelection(): void {
    this.selected.set(new Set());
  }

  submitBulk(): void {
    const ids = [...this.selected()];
    if (!ids.length || this.saving()) return;
    this.saving.set(true);
    this.svc
      .bulkAssign(this.jadwalId(), {
        anggota_ids: ids,
        sesi_id: this.sesiId(),
        peran_peserta: this.peran(),
        wajib_absen: this.wajibAbsen(),
      })
      .subscribe({
        next: (res) => {
          this.saving.set(false);
          // The result is the message: how many landed, how many were duplicates.
          this.toast.success(this.translate.instant('PESERTA.BULK_RESULT', res));
          this.selected.set(new Set());
          this.load();
        },
        error: (err: HttpErrorResponse) => {
          this.saving.set(false);
          if (err.status !== 403) this.toast.error('PESERTA.ASSIGN_ERROR');
        },
      });
  }

  bulkPrev(): void {
    if (this.bulkPage() > 1) this.bulkPage.update((p) => p - 1);
  }

  bulkNext(): void {
    if (this.bulkPage() < this.bulkResult().last_page) this.bulkPage.update((p) => p + 1);
  }
}
