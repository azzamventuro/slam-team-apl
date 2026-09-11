/** Embedded file reference (mst_file) — no numeric id exposed. */
export interface FileRef {
  uuid: string;
  nama_asli: string;
  ekstensi: string;
  mime_type: string;
  ukuran_byte: number;
  is_publik: boolean;
}

export interface Dokumen {
  id: number;
  kode: string | null;
  tipe: string | null;
  format: string | null;
  reff_type: string;
  reff_id: number;
  jenis: number | null;
  keterangan: string | null;
  file: FileRef | null;
  created_at: string;
  created_by: number | null;
  modified_at: string | null;
  modified_by: number | null;
}

export interface DokumenForm {
  file_uuid: string;
  reff_type: string;
  reff_id: number;
  kode?: string | null;
  tipe?: string | null;
  jenis?: number | null;
  keterangan?: string | null;
}

export interface DokumenQuery {
  page: number;
  per_page: number;
  q: string;
  sort?: string;
  reff_type?: string | null;
  reff_id?: number | null;
  tipe?: string | null;
  jenis?: number | null;
  format?: string | null;
}

/** Valid polymorphic reff_type values. */
export const VALID_REFF_TYPES = [
  'anggota',
  'instansi',
  'inorga',
  'unit',
  'prestasi',
  'kegiatan',
  'artikel',
] as const;
