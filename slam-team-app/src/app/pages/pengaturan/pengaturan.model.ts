// Mirrors internal/modules/core/pengaturan/dto — GET /pengaturan groups every
// row by grup server-side, so the page never regroups them itself.

export type TipeNilai = 'string' | 'integer' | 'boolean' | 'json' | 'date';

/** One mst_pengaturan row, already cast (nilai/nilai_bawaan) for its tipe_nilai. */
export interface Setting {
  id: number;
  kunci: string;
  grup: string;
  label: string;
  nilai: unknown;
  tipe_nilai: TipeNilai;
  nilai_bawaan: unknown;
  opsi: string[] | null;
  satuan: string | null;
  keterangan: string | null;
  urutan: number;
  is_publik: boolean;
  is_terkunci: boolean;
  modified_at: string | null;
  modified_by: number | null;
}

/** One tab of the settings page. */
export interface SettingGroup {
  grup: string;
  items: Setting[];
}

/** One entry of a PUT /pengaturan bulk update. */
export interface SettingKV {
  kunci: string;
  nilai: unknown;
}
