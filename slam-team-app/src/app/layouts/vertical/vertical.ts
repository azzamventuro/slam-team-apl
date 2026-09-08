import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';

import { Topbar } from '../topbar/topbar';
import { Sidebar } from '../sidebar/sidebar';
import { Footer } from '../footer/footer';

/** Authenticated shell: sidebar + topbar + routed content + footer. */
@Component({
  selector: 'app-vertical-layout',
  imports: [RouterOutlet, Topbar, Sidebar, Footer],
  templateUrl: './vertical.html',
  styleUrl: './vertical.scss',
})
export class VerticalLayout {}
