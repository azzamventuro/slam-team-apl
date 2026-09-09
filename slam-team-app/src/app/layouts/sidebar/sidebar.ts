import { Component, computed, inject, input, output } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';

import { isAdmin } from '../../core/guards/admin.guard';
import { AuthService } from '../../core/services/auth.service';

/** One entry in the vertical navigation. */
export interface MenuItem {
  /** i18n key, e.g. `MENU.DASHBOARD`. */
  label: string;
  route: string;
  /** Single glyph placeholder until an icon set is chosen. */
  glyph: string;
}

@Component({
  selector: 'app-sidebar',
  imports: [RouterLink, RouterLinkActive, TranslatePipe],
  templateUrl: './sidebar.html',
  styleUrl: './sidebar.scss',
})
export class Sidebar {
  private auth = inject(AuthService);

  /** Drawer state below the `lg` breakpoint; the sidebar is always shown above it. */
  readonly open = input(false);
  /** Asks the shell to close the drawer (backdrop click, or a link was followed). */
  readonly dismiss = output<void>();

  /**
   * SEAM (modul 05 — hak akses): this list becomes a `computed()` over
   * `GET /modul` filtered by `PermissionService.can('<kode>.read')` and grouped
   * by `grup`. Pengaturan is the one exception — it has no seeded permission
   * and is gated by role instead (see adminGuard), so it is added here
   * directly rather than waiting on that seam. Hiding the link is UX only.
   */
  readonly items = computed<MenuItem[]>(() => {
    const items: MenuItem[] = [{ label: 'MENU.DASHBOARD', route: '/dashboard', glyph: '◧' }];
    if (isAdmin(this.auth.user()?.role)) {
      items.push({ label: 'MENU.PENGATURAN', route: '/pengaturan', glyph: '⚙' });
    }
    return items;
  });
}
