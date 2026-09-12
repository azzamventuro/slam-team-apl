import { Component, inject, input, output } from '@angular/core';
import { Router } from '@angular/router';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';

import { AuthService } from '../../core/services/auth.service';
import { SlamIconComponent } from '../../shared/components/slam-icon/slam-icon.component';

@Component({
  selector: 'app-topbar',
  imports: [TranslatePipe, SlamIconComponent],
  templateUrl: './topbar.html',
  styleUrl: './topbar.scss',
})
export class Topbar {
  private auth = inject(AuthService);
  private router = inject(Router);
  private translate = inject(TranslateService);

  /** Drawer state, owned by the shell; used for `aria-expanded`. */
  readonly menuOpen = input(false);
  readonly toggleMenu = output<void>();

  readonly user = this.auth.user;
  readonly languages = ['IND', 'ENG'] as const;
  /** ngx-translate exposes the active language as a signal — read it directly. */
  readonly currentLang = this.translate.currentLang;

  switchLang(lang: string): void {
    this.translate.use(lang);
  }

  /** Avatar fallback — first letter of the display name. */
  initial(name: string): string {
    return name.trim().charAt(0).toUpperCase() || '?';
  }

  logout(): void {
    this.auth.logout();
    this.router.navigate(['/auth/login']);
  }
}
