import { HttpErrorResponse } from '@angular/common/http';
import {
  AfterViewInit,
  Component,
  ElementRef,
  OnDestroy,
  OnInit,
  ViewChild,
  inject,
  signal,
} from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';
import * as L from 'leaflet';

import { ApiResponse } from '../../core/models/api.model';
import { ToastService } from '../../core/services/toast.service';
import { FileUpload } from '../../shared/components/file-upload/file-upload';
import { FileRef } from '../../shared/services/file.model';
import {
  DEFAULT_RADIUS_METER,
  DEFAULT_TIMEZONE,
  JENIS_LOKASI_OPSI,
  Lokasi,
  LokasiForm as LokasiFormBody,
  MAX_RADIUS_METER,
  MIN_RADIUS_METER,
  listTimezones,
} from './lokasi.model';
import { LokasiService } from './lokasi.service';

/** Malang, Jawa Timur — SLAM's home base; a sane default map center for a brand-new lokasi. */
const DEFAULT_CENTER: L.LatLngTuple = [-7.9666, 112.6326];
const DEFAULT_ZOOM = 15;
/** design-tokens.md --slam-primary, as a literal (see the circle style comment below). */
const SLAM_RED = '#e11d2a';

/**
 * Create/edit form for a training/activity location — kode, nama, geofence
 * (lat/long + radius) and an IANA timezone. The map is a **Leaflet + OSM**
 * picker (no API key): click anywhere to drop the point, or drag the marker;
 * the circle around it always reflects `radius_meter` in both directions
 * (typing the number, or dragging the range slider next to it).
 *
 * The marker (not the circle) is what Leaflet can drag natively — a directly
 * draggable circle boundary needs a plugin (Leaflet.Path.Drag) that
 * duplicates capability the marker already gives us, so it was left out; the
 * circle simply recenters wherever the marker goes, which reads as one
 * draggable geofence to the user.
 */
@Component({
  selector: 'app-lokasi-form',
  imports: [ReactiveFormsModule, RouterLink, TranslatePipe, FileUpload],
  templateUrl: './lokasi-form.html',
  styleUrl: './lokasi-form.scss',
})
export class LokasiFormComponent implements OnInit, AfterViewInit, OnDestroy {
  private fb = inject(FormBuilder);
  private svc = inject(LokasiService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private toast = inject(ToastService);

  @ViewChild('mapEl') private mapEl!: ElementRef<HTMLDivElement>;

  readonly saving = signal(false);
  readonly isEdit = signal(false);
  readonly jenisOpsi = JENIS_LOKASI_OPSI;
  readonly timezones = listTimezones();
  readonly minRadius = MIN_RADIUS_METER;
  readonly maxRadius = MAX_RADIUS_METER;

  lokasiId: number | null = null;

  /** uuid of the stored foto — bound to <app-file-upload>'s [value]. */
  readonly fotoUuid = signal<string | null>(null);
  /** Numeric FK resolved from the upload's {id, uuid} — what actually goes on the wire. */
  private fotoFileId: number | null = null;

  private map: L.Map | null = null;
  private marker: L.Marker | null = null;
  private circle: L.Circle | null = null;
  /** True while a map interaction is writing the form, so the form→map sync doesn't echo it back. */
  private syncingFromMap = false;

  form: FormGroup = this.fb.nonNullable.group({
    kode: ['', [Validators.required, Validators.maxLength(50)]],
    nama: ['', [Validators.required, Validators.maxLength(150)]],
    jenis_lokasi: ['', [Validators.maxLength(50)]],
    alamat: [''],
    latitude: [DEFAULT_CENTER[0], [Validators.required, Validators.min(-90), Validators.max(90)]],
    longitude: [
      DEFAULT_CENTER[1],
      [Validators.required, Validators.min(-180), Validators.max(180)],
    ],
    radius_meter: [
      DEFAULT_RADIUS_METER,
      [Validators.required, Validators.min(MIN_RADIUS_METER), Validators.max(MAX_RADIUS_METER)],
    ],
    timezone: [DEFAULT_TIMEZONE, [Validators.required]],
    keterangan: [''],
    is_aktif: [true],
  });

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.lokasiId = +id;
      this.isEdit.set(true);
      this.svc.detail(this.lokasiId).subscribe({
        next: (l) => this.applyDetail(l),
        error: () => {
          this.toast.error('LOKASI.LOAD_ERROR');
          this.router.navigate(['/lokasi']);
        },
      });
    }
  }

  ngAfterViewInit(): void {
    this.initMap();
  }

  ngOnDestroy(): void {
    // Leaflet owns DOM listeners and tile requests outside Angular's view —
    // it must be disposed explicitly or both leak past this component's life.
    this.map?.remove();
    this.map = null;
  }

  private applyDetail(l: Lokasi): void {
    this.form.patchValue({
      kode: l.kode,
      nama: l.nama,
      jenis_lokasi: l.jenis_lokasi ?? '',
      alamat: l.alamat ?? '',
      latitude: l.latitude,
      longitude: l.longitude,
      radius_meter: l.radius_meter,
      timezone: l.timezone,
      keterangan: l.keterangan ?? '',
      is_aktif: l.is_aktif,
    });
    this.fotoFileId = l.foto_file_id;
    this.fotoUuid.set(l.foto_uuid);
    // The detail request and the map's own init race depending on how fast
    // the API answers; recentring here covers the case where the map already
    // exists by the time the record arrives.
    this.recenterMap();
  }

  // ── Map ──

  private initMap(): void {
    const { latitude, longitude, radius_meter } = this.form.getRawValue();
    const center: L.LatLngTuple = [latitude, longitude];

    const map = L.map(this.mapEl.nativeElement).setView(center, DEFAULT_ZOOM);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      maxZoom: 19,
      attribution:
        '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
    }).addTo(map);

    // A plain divIcon sidesteps Leaflet's default marker PNGs, whose paths
    // break under a bundler unless copied/aliased separately — this needs no
    // extra asset step and picks up the SLAM red from CSS instead of the
    // stock blue pin.
    const marker = L.marker(center, {
      draggable: true,
      // Styled inline rather than via a stylesheet rule: Leaflet injects this
      // markup with plain DOM APIs, outside Angular's view, so a component
      // SCSS selector would never reach it without ::ng-deep.
      icon: L.divIcon({
        className: 'lokasi-marker',
        html: `<span style="display:block;width:18px;height:18px;margin:2px;border-radius:50%;background:${SLAM_RED};border:3px solid #fff;box-shadow:0 0 0 2px rgba(225,29,42,.35);"></span>`,
        iconSize: [22, 22],
        iconAnchor: [11, 11],
      }),
    }).addTo(map);

    const circle = L.circle(center, {
      radius: radius_meter,
      // Leaflet's SVG renderer sets these via a raw attribute, not the style
      // cascade, so a `var(--slam-primary)` reference would not resolve here
      // — this literal is that token's value (design-tokens.md §1).
      color: SLAM_RED,
      fillColor: SLAM_RED,
      fillOpacity: 0.14,
      weight: 2,
    }).addTo(map);

    map.on('click', (e: L.LeafletMouseEvent) => this.setPoint(e.latlng.lat, e.latlng.lng));
    marker.on('dragend', () => {
      const pos = marker.getLatLng();
      this.setPoint(pos.lat, pos.lng);
    });

    this.map = map;
    this.marker = marker;
    this.circle = circle;

    // Numeric lat/long stay the fallback path: typing a value (or a loaded
    // record patching the form) moves the marker + circle the same as a
    // click would.
    this.form.controls['latitude'].valueChanges.subscribe(() => this.syncMapFromForm());
    this.form.controls['longitude'].valueChanges.subscribe(() => this.syncMapFromForm());
    // The range slider and the number input both write radius_meter, so one
    // subscription covers "dragging" and "typing" alike.
    this.form.controls['radius_meter'].valueChanges.subscribe((v: number) => {
      if (typeof v === 'number' && v > 0) circle.setRadius(v);
    });

    // A container whose size settles after layout (sidebar transition, slow
    // web font) can leave Leaflet holding a stale size — nudge it once after
    // the current frame.
    setTimeout(() => map.invalidateSize(), 0);
  }

  /** Map click / marker drag → the numeric fields, which stay the single source of truth. */
  private setPoint(lat: number, lng: number): void {
    this.syncingFromMap = true;
    this.form.patchValue({ latitude: round(lat), longitude: round(lng) });
    this.syncingFromMap = false;
    this.marker?.setLatLng([lat, lng]);
    this.circle?.setLatLng([lat, lng]);
  }

  /** Numeric fields → map. Skipped while the map itself is the one writing them (see setPoint). */
  private syncMapFromForm(): void {
    if (this.syncingFromMap || !this.map) return;
    const { latitude, longitude } = this.form.getRawValue();
    if (!isFiniteCoord(latitude, 90) || !isFiniteCoord(longitude, 180)) return;
    this.marker?.setLatLng([latitude, longitude]);
    this.circle?.setLatLng([latitude, longitude]);
  }

  private recenterMap(): void {
    if (!this.map) return;
    const { latitude, longitude, radius_meter } = this.form.getRawValue();
    this.map.setView([latitude, longitude], DEFAULT_ZOOM);
    this.marker?.setLatLng([latitude, longitude]);
    this.circle?.setLatLng([latitude, longitude]);
    this.circle?.setRadius(radius_meter);
  }

  // ── Foto ──

  onFotoUploaded(ref: FileRef): void {
    this.fotoFileId = ref.id ?? null;
    this.fotoUuid.set(ref.uuid);
  }

  onFotoCleared(): void {
    this.fotoFileId = null;
    this.fotoUuid.set(null);
  }

  // ── Submit ──

  submit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.saving.set(true);
    const raw = this.form.getRawValue();
    const body: LokasiFormBody = {
      kode: raw.kode,
      nama: raw.nama,
      jenis_lokasi: raw.jenis_lokasi,
      alamat: raw.alamat,
      latitude: raw.latitude,
      longitude: raw.longitude,
      radius_meter: raw.radius_meter,
      timezone: raw.timezone,
      foto_file_id: this.fotoFileId,
      keterangan: raw.keterangan,
      is_aktif: raw.is_aktif,
    };

    const obs = this.isEdit() ? this.svc.update(this.lokasiId!, body) : this.svc.create(body);
    obs.subscribe({
      next: () => {
        this.toast.success(this.isEdit() ? 'LOKASI.UPDATE_SUCCESS' : 'LOKASI.CREATE_SUCCESS');
        this.router.navigate(['/lokasi']);
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.applyServerErrors(err);
      },
    });
  }

  /** Maps {field: msg[]} onto the matching control; 409 (duplicate kode) and a bare 4xx each get their own toast. */
  private applyServerErrors(err: HttpErrorResponse): void {
    const body = err.error as ApiResponse<unknown> | undefined;
    const errors = body?.errors;
    let mapped = false;
    if (errors && typeof errors === 'object' && !Array.isArray(errors)) {
      for (const [field, msg] of Object.entries(errors as Record<string, unknown>)) {
        const ctrl = this.form.get(field);
        if (!ctrl) continue;
        ctrl.setErrors({ server: Array.isArray(msg) ? String(msg[0]) : String(msg) });
        ctrl.markAsTouched();
        mapped = true;
      }
    }
    if (err.status === 409) {
      this.toast.error('LOKASI.KODE_DUPLICATE');
    } else if (!mapped) {
      this.toast.error('LOKASI.SAVE_ERROR');
    }
  }

  hasError(field: string, error: string): boolean {
    const ctrl = this.form.get(field);
    if (!ctrl) return false;
    return ctrl.hasError(error) && (ctrl.dirty || ctrl.touched);
  }
}

/** Coordinates are decimal(10,7) server-side; round to match so a saved value round-trips unchanged. */
function round(n: number): number {
  return Math.round(n * 1e7) / 1e7;
}

function isFiniteCoord(n: unknown, max: number): n is number {
  return typeof n === 'number' && Number.isFinite(n) && Math.abs(n) <= max;
}
