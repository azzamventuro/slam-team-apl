// Mirrors internal/modules/core/izin/{domain,dto} — keep the field names
// identical to the JSON on the wire. A request is keyed by the anggota behind
// the caller's account (there is no anggota_id in the create body) and tied
// to exactly one jadwal_sesi.

/** PG enum `jenis_izin`. */
export type JenisIzin = 'izin' | 'sakit' | 'dinas' | 'pulang_cepat';
/** PG enum `status_izin`. */
export type StatusIzin = 'menunggu' | 'disetujui' | 'ditolak';

export const JENIS_IZIN: readonly JenisIzin[] = ['izin', 'sakit', 'dinas', 'pulang_cepat'] as const;
export const STATUS_IZIN: readonly StatusIzin[] = ['menunggu', 'disetujui', 'ditolak'] as const;

/** `dto.LampiranInfo` — a PRIVATE file; only ever fetched through the authenticated /files route. */
export interface LampiranInfo {
  file_id: number;
  uuid: string;
  url: string;
}

/** `dto.IzinResp` — one row of GET /izin and the body of every write. */
export interface Izin {
  id: number;
  sesi_id: number;
  sesi_tanggal?: string | null;
  sesi_status?: string | null;
  jadwal_id?: number | null;
  jadwal_nama?: string | null;
  jadwal_kode?: string | null;
  anggota_id: number;
  anggota_nama?: string | null;
  no_induk?: string | null;
  jenis: JenisIzin;
  alasan: string;
  /** "HH:MM:SS" — only present for jenis = pulang_cepat. */
  waktu_pulang_diminta?: string | null;
  lampiran?: LampiranInfo | null;
  status: StatusIzin;
  diproses_oleh?: number | null;
  diproses_oleh_nama?: string | null;
  diproses_pada?: string | null;
  catatan_peninjau?: string | null;
  created_at: string;
  modified_at?: string | null;
}

/** `dto.CreateIzinReq` — POST /izin. */
export interface CreateIzinReq {
  sesi_id: number;
  jenis: JenisIzin;
  alasan: string;
  /** Required for pulang_cepat, rejected (422) for every other jenis. */
  waktu_pulang_diminta?: string | null;
  lampiran_file_id?: number | null;
}

/** `dto.TolakReq` — PATCH /izin/:id/tolak. */
export interface TolakReq {
  catatan_peninjau: string;
}

/** `dto.ListIzinQuery`. Ownership is not a parameter — it comes from the caller's cakupan. */
export interface IzinQuery {
  page: number;
  per_page: number;
  q?: string;
  sort?: string;
  status?: StatusIzin | '';
  sesi_id?: number | null;
  jadwal_id?: number | null;
  jenis?: JenisIzin | '';
}

/** The slice of `dto.JadwalResp` the sesi picker's first step needs (GET /jadwal). */
export interface JadwalOpsi {
  id: number;
  kode: string;
  nama: string;
  tanggal_mulai: string;
  tanggal_selesai?: string | null;
  status: string;
  timezone: string;
}

/** The slice of `dto.SesiResp` the picker's second step needs (GET /jadwal/:id/sesi). */
export interface SesiOpsi {
  id: number;
  jadwal_id: number;
  tanggal_lokal: string;
  /** "HH:MM:SS" in the schedule's own timezone. */
  mulai_lokal: string;
  selesai_lokal: string;
  timezone: string;
  status: string;
}

/** Bootstrap contextual colour per approval state; the label/glyph carries the meaning, never the colour alone. */
export const STATUS_IZIN_VARIANT: Record<StatusIzin, 'info' | 'success' | 'danger'> = {
  menunggu: 'info',
  disetujui: 'success',
  ditolak: 'danger',
};

/** "HH:MM:SS" (PG time) → "HH:MM" for display and for a native <input type="time">. */
export function hhmm(t: string | null | undefined): string {
  return t ? t.slice(0, 5) : '';
}

/** "KODE — Nama (2026-09-06 – 2026-12-20)" for the jadwal step of the picker. */
export function jadwalLabel(j: JadwalOpsi): string {
  const span = j.tanggal_selesai ? `${j.tanggal_mulai} – ${j.tanggal_selesai}` : j.tanggal_mulai;
  return `${j.kode} — ${j.nama} (${span})`;
}
