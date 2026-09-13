// Mirrors internal/modules/core/notifikasi/{domain,dto}. One row per
// recipient (fan-out); every read is already scoped to the caller, so there
// is no user_id on the wire.

export type NotifWarna = 'info' | 'sukses' | 'peringatan' | 'bahaya';
export type NotifPrioritas = 'rendah' | 'normal' | 'tinggi';

export interface Notifikasi {
  id: number;
  uuid: string;
  /** jadwal_ditugaskan | jadwal_pengingat | sesi_dibatalkan | sistem | … */
  tipe: string;
  judul: string;
  isi?: string | null;
  ikon?: string | null;
  warna?: NotifWarna | null;
  /** In-app destination, e.g. `/jadwal/12` or `/jadwal/12/sesi/34`. */
  route?: string | null;
  reff_type?: string | null;
  reff_id?: number | null;
  prioritas: NotifPrioritas;
  is_dibaca: boolean;
  dibaca_pada?: string | null;
  is_diarsipkan: boolean;
  kedaluwarsa_pada?: string | null;
  created_at: string;
}

/** `dto.ListNotifikasiQuery`. */
export interface NotifikasiQuery {
  page?: number;
  per_page?: number;
  is_dibaca?: boolean | null;
  tipe?: string;
  arsip?: boolean;
}

/** GET /notifikasi/jumlah-belum-dibaca */
export interface JumlahBelumDibaca {
  jumlah: number;
}

/** PATCH /notifikasi/baca-semua */
export interface BacaSemuaResp {
  ditandai: number;
}
