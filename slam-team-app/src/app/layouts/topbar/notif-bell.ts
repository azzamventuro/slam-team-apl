import {
  Component,
  DestroyRef,
  ElementRef,
  HostListener,
  computed,
  effect,
  inject,
  signal,
} from '@angular/core';
import { Router } from '@angular/router';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';

import { Notifikasi, NotifWarna } from '../../core/models/notifikasi.model';
import { AuthService } from '../../core/services/auth.service';
import { NotifikasiService } from '../../core/services/notifikasi.service';
import { SlamIconComponent } from '../../shared/components/slam-icon/slam-icon.component';
import { formatDateTime } from '../../shared/util/tz';

/** Unread badge refresh cadence. Also refreshed after every action here. */
const POLL_MS = 60_000;
/** How many rows the dropdown shows — it is an inbox peek, not the archive. */
const PEEK = 15;

/**
 * Topbar bell: unread badge (`/notifikasi/jumlah-belum-dibaca`, polled),
 * dropdown of the latest notifikasi, mark-read / mark-all-read, and click →
 * `router.navigate(notif.route)`. Bearer-only: the inbox is the caller's, no
 * permission gate. Signal-driven so the badge updates without zone.js.
 */
@Component({
  selector: 'app-notif-bell',
  imports: [TranslatePipe, SlamIconComponent],
  templateUrl: './notif-bell.html',
  styleUrl: './notif-bell.scss',
})
export class NotifBell {
  private svc = inject(NotifikasiService);
  private auth = inject(AuthService);
  private router = inject(Router);
  private translate = inject(TranslateService);
  private host = inject(ElementRef<HTMLElement>);
  private destroyRef = inject(DestroyRef);

  readonly open = signal(false);
  readonly unread = signal(0);
  readonly items = signal<Notifikasi[]>([]);
  readonly loading = signal(false);
  readonly busy = signal(false);

  /** `99+` past two digits so the badge never grows the bell. */
  readonly badge = computed(() => (this.unread() > 99 ? '99+' : String(this.unread())));
  readonly hasUnread = computed(() => this.unread() > 0);

  private timer: ReturnType<typeof setInterval> | null = null;

  constructor() {
    // Poll only while a session exists; stop (and clear) on logout so a stale
    // badge never survives into the next login.
    effect(() => {
      if (this.auth.isLoggedIn()) this.startPolling();
      else this.stopPolling();
    });
    this.destroyRef.onDestroy(() => this.stopPolling());
  }

  private startPolling(): void {
    if (this.timer) return;
    this.refreshCount();
    this.timer = setInterval(() => this.refreshCount(), POLL_MS);
  }

  private stopPolling(): void {
    if (this.timer) clearInterval(this.timer);
    this.timer = null;
    this.unread.set(0);
    this.items.set([]);
    this.open.set(false);
  }

  refreshCount(): void {
    this.svc.unreadCount().subscribe({
      next: (r) => this.unread.set(r.jumlah),
      error: () => {}, // the interceptor already reports auth/network failures
    });
  }

  private loadList(): void {
    this.loading.set(true);
    this.svc.list({ page: 1, per_page: PEEK }).subscribe({
      next: (p) => {
        this.items.set(p.items);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });
  }

  toggle(): void {
    const next = !this.open();
    this.open.set(next);
    if (next) {
      this.loadList();
      this.refreshCount();
    }
  }

  close(): void {
    this.open.set(false);
  }

  /** Click outside the bell closes the dropdown. */
  @HostListener('document:click', ['$event'])
  onDocumentClick(ev: MouseEvent): void {
    if (this.open() && !this.host.nativeElement.contains(ev.target as Node)) this.close();
  }

  @HostListener('document:keydown.escape')
  onEscape(): void {
    if (this.open()) this.close();
  }

  /** Open a notification: mark it read (if needed), close, and follow `route`. */
  openItem(n: Notifikasi): void {
    if (!n.is_dibaca) this.markRead(n, false);
    this.close();
    if (n.route) this.router.navigateByUrl(n.route);
  }

  markRead(n: Notifikasi, ev: Event | false): void {
    if (ev) ev.stopPropagation();
    if (n.is_dibaca) return;
    // Optimistic: flip locally, then reconcile the count from the server.
    this.items.update((list) =>
      list.map((x) => (x.id === n.id ? { ...x, is_dibaca: true } : x)),
    );
    this.unread.update((c) => Math.max(0, c - 1));
    this.svc.markRead(n.id).subscribe({
      next: () => this.refreshCount(),
      error: () => this.refreshCount(),
    });
  }

  markAll(): void {
    if (this.busy() || !this.hasUnread()) return;
    this.busy.set(true);
    this.svc.markAllRead().subscribe({
      next: () => {
        this.items.update((list) => list.map((x) => ({ ...x, is_dibaca: true })));
        this.unread.set(0);
        this.busy.set(false);
        this.refreshCount();
      },
      error: () => this.busy.set(false),
    });
  }

  /** Backend `warna` → SLAM status token; the icon is the other half of the signal. */
  warnaClass(w: NotifWarna | null | undefined): string {
    switch (w) {
      case 'sukses':
        return 'bell__item--sukses';
      case 'peringatan':
        return 'bell__item--peringatan';
      case 'bahaya':
        return 'bell__item--bahaya';
      default:
        return 'bell__item--info';
    }
  }

  icon(n: Notifikasi): string {
    switch (n.tipe) {
      case 'jadwal_ditugaskan':
      case 'jadwal_pengingat':
        return 'calendar';
      case 'sesi_dibatalkan':
        return 'x';
      default:
        return 'bell';
    }
  }

  when(n: Notifikasi): string {
    return formatDateTime(n.created_at, undefined, this.translate.currentLang() === 'ENG' ? 'en-GB' : 'id-ID');
  }
}
