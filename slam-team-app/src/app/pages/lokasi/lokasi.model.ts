// Mirrors internal/modules/core/lokasi/dto — keep field names identical to the
// JSON on the wire (rancangan Bab 5.5 geofence + Bab 4.5 IANA timezone).

/** `dto.LokasiResp` — a training/activity location with a geofence + timezone. */
export interface Lokasi {
  id: number;
  kode: string;
  nama: string;
  jenis_lokasi: string;
  alamat: string;
  latitude: number;
  longitude: number;
  radius_meter: number;
  timezone: string;
  foto_file_id: number | null;
  /** Resolved by a LEFT JOIN on mst_file — feed this to <app-secure-image>/<app-file-upload>. */
  foto_uuid: string | null;
  keterangan: string;
  is_aktif: boolean;
  /** Active jadwal referencing this lokasi. While > 0, DELETE is refused with 409. */
  jumlah_jadwal: number;
  created_at: string;
  modified_at: string | null;
}

/** `dto.CreateLokasiReq` / `UpdateLokasiReq` — identical shape, PUT is a full replace. */
export interface LokasiForm {
  kode: string;
  nama: string;
  jenis_lokasi: string;
  alamat: string;
  latitude: number;
  longitude: number;
  radius_meter: number;
  timezone: string;
  foto_file_id: number | null;
  keterangan: string;
  is_aktif: boolean;
}

/** `dto.ListLokasiQuery`. */
export interface LokasiQuery {
  page: number;
  per_page: number;
  q: string;
  sort?: string;
  is_aktif?: boolean | null;
}

/** `jenis_lokasi` is a free-text varchar server-side; these are the common values used so far. */
export const JENIS_LOKASI_OPSI = ['lapangan', 'indoor', 'sekretariat'] as const;

/** Server default (mst_lokasi.timezone default) and CHECK-constraint bounds mirrored client-side. */
export const DEFAULT_TIMEZONE = 'Asia/Jakarta';
export const DEFAULT_RADIUS_METER = 100;
export const MIN_RADIUS_METER = 1;
export const MAX_RADIUS_METER = 100000;

/**
 * IANA zone names for the `<select>`. `Intl.supportedValuesOf` gives the
 * full IANA database on modern engines; the curated fallback keeps the form
 * usable on anything older, biased toward the zones an Indonesian club
 * actually needs (Bab 4.5).
 */
export function listTimezones(): string[] {
  try {
    const supported = (
      Intl as unknown as { supportedValuesOf?: (key: string) => string[] }
    ).supportedValuesOf?.('timeZone');
    if (supported?.length) return supported;
  } catch {
    // Fall through to the curated list below.
  }
  return [
    'Asia/Jakarta',
    'Asia/Pontianak',
    'Asia/Makassar',
    'Asia/Jayapura',
    'Asia/Singapore',
    'Asia/Kuala_Lumpur',
    'Asia/Bangkok',
    'Asia/Tokyo',
    'Asia/Dubai',
    'Europe/London',
    'Europe/Amsterdam',
    'UTC',
  ];
}
