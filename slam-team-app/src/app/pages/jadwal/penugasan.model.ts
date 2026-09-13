// Mirrors internal/modules/core/penugasan/{domain,dto} — keep the field names
// identical. Assignment is keyed by anggota_id, NOT user_id: members without
// an account can be assigned; for them the API forces wajib_absen=false.

export type StatusTugas = 'ditugaskan' | 'diterima' | 'ditolak' | 'izin';
export type PeranPeserta = 'peserta' | 'pelatih' | 'panitia' | 'pengawas';

export const PERAN_PESERTA: PeranPeserta[] = ['peserta', 'pelatih', 'panitia', 'pengawas'];
export const STATUS_TUGAS: StatusTugas[] = ['ditugaskan', 'diterima', 'ditolak', 'izin'];

/** `domain.JadwalPeserta` — one row of GET /jadwal/:id/peserta. */
export interface JadwalPeserta {
  id: number;
  jadwal_id: number;
  /** null = every session of the schedule. */
  sesi_id: number | null;
  anggota_id: number;
  peran_peserta?: PeranPeserta | null;
  wajib_absen: boolean;
  status_tugas: StatusTugas;
  ditugaskan_oleh: number;
  ditugaskan_pada: string;
  direspon_pada?: string | null;
  keterangan?: string | null;
  // read-only projections from the repository join
  anggota_nama?: string | null;
  no_induk?: string | null;
  anggota_berakun: boolean;
  sesi_tanggal?: string | null;
}

/** `dto.AssignReq` — POST /jadwal/:id/peserta. */
export interface AssignReq {
  anggota_id: number;
  sesi_id?: number | null;
  peran_peserta?: PeranPeserta;
  wajib_absen?: boolean;
  keterangan?: string | null;
}

/** `dto.BulkAssignReq` — POST /jadwal/:id/peserta/bulk (max 500 ids). */
export interface BulkAssignReq {
  anggota_ids: number[];
  sesi_id?: number | null;
  peran_peserta?: PeranPeserta;
  wajib_absen?: boolean;
}

/** `dto.BulkAssignResp` — duplicates are skipped, not rejected. */
export interface BulkAssignResp {
  ditugaskan: number;
  dilewati: number;
}

/** `dto.ResponReq` — PATCH /peserta/:id/respon, by the assignee only. */
export interface ResponReq {
  status_tugas: 'diterima' | 'ditolak';
  keterangan?: string | null;
}

/** `dto.ListPesertaQuery`. */
export interface ListPesertaQuery {
  sesi_id?: number | null;
  status_tugas?: StatusTugas | '';
  wajib_absen?: boolean | null;
  q?: string;
}

/** Minimal `domain.Jadwal` projection the panel header needs (GET /jadwal/:id). */
export interface JadwalRingkas {
  id: number;
  kode: string;
  nama: string;
  jenis_jadwal: string;
  timezone: string;
  tanggal_mulai: string;
  tanggal_selesai?: string | null;
  jam_mulai: string;
  jam_selesai: string;
  status: string;
  lokasi_nama?: string | null;
  jumlah_sesi: number;
}

/** `domain.Sesi` projection for the session picker (GET /jadwal/:id/sesi). */
export interface JadwalSesi {
  id: number;
  jadwal_id: number;
  tanggal_lokal: string;
  mulai_utc: string;
  selesai_utc: string;
  timezone: string;
  status: string;
  jml_ditugaskan: number;
}
