import { Component, inject } from '@angular/core';
import { Router } from '@angular/router';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';

import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'app-topbar',
  imports: [TranslatePipe],
  templateUrl: './topbar.html',
})
export class Topbar {
  private auth = inject(AuthService);
  private router = inject(Router);
  private translate = inject(TranslateService);

  readonly user = this.auth.user;

  switchLang(lang: string): void {
    this.translate.use(lang);
  }

  logout(): void {
    this.auth.logout();
    this.router.navigate(['/auth/login']);
  }
}
