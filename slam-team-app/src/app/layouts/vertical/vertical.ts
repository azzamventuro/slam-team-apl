import { Component, signal } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';

import { Topbar } from '../topbar/topbar';
import { Sidebar } from '../sidebar/sidebar';
import { Footer } from '../footer/footer';

/** Authenticated shell: sidebar + topbar + routed content + footer. */
@Component({
  selector: 'app-vertical-layout',
  imports: [RouterOutlet, Topbar, Sidebar, Footer, TranslatePipe],
  templateUrl: './vertical.html',
  styleUrl: './vertical.scss',
})
export class VerticalLayout {
  /** Sidebar drawer state below `lg`; above it the sidebar is always visible. */
  readonly menuOpen = signal(false);

  toggleMenu(): void {
    this.menuOpen.update((open) => !open);
  }
}
