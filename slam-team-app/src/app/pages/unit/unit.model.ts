export interface Unit {
  id: number;
  anggota_id: number;
  anggota_nama: string | null;
  anggota_no_induk: string | null;
  kode: string;
  model: string;
  panjang: number | null;
  panjang_inbar: number | null;
  lebar: number | null;
  berat: number | null;
  berat_bb: number | null;
  fps: number | null;
  deskripsi_warna: string | null;
  foto_sampul: FotoSampulInfo | null;
  disetujui: boolean;
  disetujui_oleh: number | null;
  disetujui_oleh_nama: string | null;
  disetujui_pada: string | null;
  created_at: string;
  modified_at: string | null;
}

export interface FotoSampulInfo {
  uuid: string;
  url: string;
}

export interface UnitForm {
  anggota_id: number;
  kode?: string;
  model?: string;
  panjang?: number | null;
  panjang_inbar?: number | null;
  lebar?: number | null;
  berat?: number | null;
  berat_bb?: number | null;
  fps?: number | null;
  deskripsi_warna?: string;
  foto_sampul_uuid?: string | null;
  disetujui?: boolean;
}

export interface UnitQuery {
  page: number;
  per_page: number;
  q: string;
  sort?: string;
  anggota_id?: number | null;
  disetujui?: boolean | null;
}

export interface AnggotaOpsi {
  id: number;
  nama_lengkap: string;
  no_induk: string;
}
