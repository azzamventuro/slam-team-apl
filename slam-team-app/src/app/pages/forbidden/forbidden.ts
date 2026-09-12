import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';

@Component({
  selector: 'app-forbidden',
  imports: [RouterLink, TranslatePipe],
  template: `
    <div class="d-flex flex-column align-items-center justify-content-center vh-100 text-center">
      <h1 class="display-1 text-danger mb-3">403</h1>
      <p class="lead mb-4">{{ 'COMMON.FORBIDDEN' | translate }}</p>
      <a routerLink="/dashboard" class="btn btn-outline-light">{{ 'COMMON.BACK' | translate }}</a>
    </div>
  `,
})
export class Forbidden {}
