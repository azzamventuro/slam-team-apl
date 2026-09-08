import { Component, inject } from '@angular/core';
import { TranslatePipe } from '@ngx-translate/core';

import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'app-dashboard',
  imports: [TranslatePipe],
  templateUrl: './dashboard.html',
  styleUrl: './dashboard.scss',
})
export class Dashboard {
  readonly user = inject(AuthService).user;
}
