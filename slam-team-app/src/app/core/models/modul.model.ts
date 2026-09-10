/** Mirrors mst_modul — the module catalogue that drives the sidebar. */
export interface Modul {
  id: number;
  kode: string;   // e.g. "anggota", "jadwal", "file"
  nama: string;   // e.g. "Anggota"
  icon: string;
  grup: string;   // e.g. "Master Data", "Konten", "Sistem", "Operasional"
  urutan: number;
  is_aktif: boolean;
}
