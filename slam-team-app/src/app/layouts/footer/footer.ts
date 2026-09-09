import { Component } from '@angular/core';

import { environment } from '../../../environments/environment';

@Component({
  selector: 'app-footer',
  template: `
    <footer class="app-footer">
      <span>&copy; {{ year }} {{ appName }}</span>
    </footer>
  `,
  styles: `
    .app-footer {
      padding: var(--slam-space-4) var(--slam-space-5);
      border-top: 1px solid var(--slam-border);
      background: var(--slam-surface);
      color: var(--slam-text-dim);
      font-size: 0.75rem;
      letter-spacing: 0.04em;
      text-transform: uppercase;
      text-align: center;
    }
  `,
})
export class Footer {
  readonly year = new Date().getFullYear();
  readonly appName = environment.appName;
}
