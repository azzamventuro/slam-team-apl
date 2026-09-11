/** Admin response shape from GET /profile-club. */
export interface ProfileClub {
  id: number;
  nama: string;
  singkatan: string;
  alamat: string;
  keterangan: string;
  banner_file_id: number | null;
  logo_simple_file_id: number | null;
  logo_besar_file_id: number | null;
  banner_file_uuid: string | null;
  logo_simple_file_uuid: string | null;
  logo_besar_file_uuid: string | null;
  created_at: string;
  created_by: number | null;
  modified_at: string | null;
  modified_by: number | null;
}

/** Payload for POST /profile-club or PUT /profile-club/:id. */
export interface ProfileClubForm {
  nama: string;
  singkatan: string;
  alamat: string;
  keterangan: string;
  banner_file_id: number | null;
  logo_simple_file_id: number | null;
  logo_besar_file_id: number | null;
}
