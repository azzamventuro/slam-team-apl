import { Component, computed, inject, input, output } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { toSignal } from '@angular/core/rxjs-interop';
import { TranslatePipe } from '@ngx-translate/core';

import { isAdmin } from '../../core/guards/admin.guard';
import { Modul } from '../../core/models/modul.model';
import { ApiService } from '../../core/services/api.service';
import { AuthService } from '../../core/services/auth.service';
import { PermissionService } from '../../core/services/permission.service';
import { SlamIconComponent } from '../../shared/components/slam-icon/slam-icon.component';

/** One entry in the vertical navigation. */
export interface MenuItem {
  /** i18n key, e.g. `MENU.DASHBOARD`. */
  label: string;
  route: string;
  /** Icon name for <slam-icon>. */
  icon: string;
}

/** A group of menu items in the sidebar, keyed by grup name. */
export interface MenuGroup {
  name: string;
  items: MenuItem[];
}

/** Maps a module kode to its sidebar glyph. */
const GLYPHS: Record<string, string> = {
  anggota: 'user-circle',
  instansi: 'shield',
  unit: 'grid-four',
  prestasi: 'star',
  inorga: 'chart-bar',
  medsos: 'monitor',
  dokumen: 'file',
  profile_club: 'newspaper',
  jadwal: 'calendar',
  absensi: 'check-square',
  izin: 'clock',
  kegiatan: 'calendar-star',
  artikel: 'note-pencil',
  lokasi: 'map-pin',
  kta_nra_qr: 'id',
  laporan_rekap: 'chart-line',
  pengaturan: 'gear-six',
  user_management: 'user-circle',
  file: 'file',
};

/**
 * Sidebar is dynamically generated from `GET /modul` filtered by the user's
 * effective permissions. Modules where the user has `<kode>.read` are shown,
 * grouped by `grup`. Dashboard and Pengaturan are always included (Pengaturan
 * is gated by role, not permission).
 *
 * Hiding menu items is UX only — the backend enforces permissions.
 */
@Component({
  selector: 'app-sidebar',
  imports: [RouterLink, RouterLinkActive, TranslatePipe, SlamIconComponent],
  templateUrl: './sidebar.html',
  styleUrl: './sidebar.scss',
})
export class Sidebar {
  private api = inject(ApiService);
  private auth = inject(AuthService);
  private perms = inject(PermissionService);

  /** Drawer state below the `lg` breakpoint; the sidebar is always shown above it. */
  readonly open = input(false);
  /** Asks the shell to close the drawer (backdrop click, or a link was followed). */
  readonly dismiss = output<void>();

  /** All modules from the backend. */
  private readonly moduls = toSignal(this.api.get<Modul[]>('/modul'), {
    initialValue: [] as Modul[],
  });

  /** Sidebar groups: dashboard + permission-filtered modules + pengaturan (admin only). */
  readonly groups = computed<MenuGroup[]>(() => {
    const groups: MenuGroup[] = [];
    const items: MenuItem[] = [];

    // Dashboard is always first.
    items.push({ label: 'MENU.DASHBOARD', route: '/dashboard', icon: 'house' });

    // Modules the user can read, grouped by `grup`.
    const grouped = new Map<string, MenuItem[]>();
    for (const m of this.moduls()) {
      if (!m.is_aktif) continue;
      if (!this.perms.can(`${m.kode}.read`)) continue;
      const group = m.grup || 'Lainnya';
      if (!grouped.has(group)) grouped.set(group, []);
      grouped.get(group)!.push({
        label: m.nama,
        route: `/${m.kode}`,
        icon: GLYPHS[m.kode] || 'list',
      });
    }
    for (const [name, groupItems] of grouped) {
      groups.push({ name, items: groupItems });
    }

    // Pengaturan: admin-only, no seeded permission (gated by role).
    if (isAdmin(this.auth.user()?.role)) {
      items.push({ label: 'MENU.PENGATURAN', route: '/pengaturan', icon: 'sliders' });
    }

    groups.unshift({ name: 'MENU.GROUP.MAIN', items });
    return groups;
  });
}
