// Package domain holds the file-layer entities and their value types.
//
// Two tables, one logical file: mst_file is the row every other module points
// at with a `*_file_id`, and mst_file_varian holds the rendered sizes — three
// rows for an image (original / medium / low), one for a document.
package domain

import (
	"time"

	"slam-team-api/internal/shared/model"
)

// Varian is the varian_file enum: which rendered size a mst_file_varian row is.
type Varian string

// The complete enum, in the order 0001_enums declares it.
const (
	VarianOriginal Varian = "original"
	VarianMedium   Varian = "medium"
	VarianLow      Varian = "low"
)

// Valid reports whether v is one of the enum values. The serve endpoint takes
// the variant straight from the URL, so this is what keeps an arbitrary path
// segment out of the query.
func (v Varian) Valid() bool {
	switch v {
	case VarianOriginal, VarianMedium, VarianLow:
		return true
	default:
		return false
	}
}

// StatusProses is the status_proses_file enum. Variants are generated after
// the upload responds, so a file is "menunggu" until they land — and a failure
// becomes "gagal" rather than silently looking finished.
type StatusProses string

const (
	StatusMenunggu StatusProses = "menunggu"
	StatusSelesai  StatusProses = "selesai"
	StatusGagal    StatusProses = "gagal"
)

// The kategori values mst_file.kategori accepts. Kategori decides both the
// first path segment on disk and the default is_publik.
const (
	KategoriFotoProfil = "foto_profil"
	KategoriFotoFormal = "foto_formal"
	KategoriAbsensi    = "absensi"
	KategoriBanner     = "banner"
	KategoriLogo       = "logo"
	KategoriDokumen    = "dokumen"
	KategoriFlyer      = "flyer"
	KategoriKTA        = "kta"
)

// Kategori is the complete allowed set, mirrored by the `oneof` binding tag on
// the upload DTO. Keep the two in step.
var Kategori = []string{
	KategoriFotoProfil, KategoriFotoFormal, KategoriAbsensi, KategoriBanner,
	KategoriLogo, KategoriDokumen, KategoriFlyer, KategoriKTA,
}

// KategoriValid reports whether k is an accepted kategori.
func KategoriValid(k string) bool {
	for _, v := range Kategori {
		if v == k {
			return true
		}
	}
	return false
}

// kategoriPublik lists the categories whose files are public by default: club
// artwork that is meant to be seen. Everything else — anything carrying a face,
// an identity document or attendance evidence — defaults to private and is
// readable only through the authenticated serve handler.
//
// `kta` is here because the QR code on a card points at a public membership
// page that shows the photo; the card image itself is not secret.
var kategoriPublik = map[string]bool{
	KategoriBanner: true,
	KategoriLogo:   true,
	KategoriFlyer:  true,
	KategoriKTA:    true,
}

// PublikDefault returns the is_publik a kategori gets when the client does not
// send one. Unknown categories default to private — the safe direction.
func PublikDefault(kategori string) bool { return kategoriPublik[kategori] }

// DokumenMIME reports whether a kategori accepts non-image content. Only
// `dokumen` does (PDF); every other kategori is image-only, checked against the
// detected content type rather than the filename.
func DokumenMIME(kategori string) bool { return kategori == KategoriDokumen }

// File maps mst_file: one row per logical uploaded file.
//
// The public identity is UUID, never ID — `id` is what owning rows store in
// their `*_file_id`, and it never appears in a URL.
type File struct {
	ID   int64  `gorm:"column:id;primaryKey" json:"-"`
	UUID string `gorm:"column:uuid;type:uuid" json:"uuid"`
	// NamaAsli is for display only; the name on disk is "{varian}.{ext}".
	NamaAsli string `gorm:"column:nama_asli" json:"nama_asli"`
	NamaSlug string `gorm:"column:nama_slug" json:"nama_slug"`
	Ekstensi string `gorm:"column:ekstensi" json:"ekstensi"`
	// MimeType is detected from the content, never from the extension.
	MimeType   string `gorm:"column:mime_type" json:"mime_type"`
	UkuranByte int64  `gorm:"column:ukuran_byte" json:"ukuran_byte"`
	HashSHA256 string `gorm:"column:hash_sha256" json:"hash_sha256"`
	// LebarPx/TinggiPx are nil for non-images.
	LebarPx  *int   `gorm:"column:lebar_px" json:"lebar_px"`
	TinggiPx *int   `gorm:"column:tinggi_px" json:"tinggi_px"`
	Kategori string `gorm:"column:kategori" json:"kategori"`
	// ReffType/ReffID are the polymorphic owner; reff_id may be linked later,
	// once the owning row exists.
	ReffType      *string `gorm:"column:reff_type" json:"reff_type,omitempty"`
	ReffID        *int64  `gorm:"column:reff_id" json:"reff_id,omitempty"`
	StorageDriver string  `gorm:"column:storage_driver" json:"storage_driver"`
	// PathDasar is the folder holding every variant, relative to the storage
	// root: {kategori}/{tahun}/{bulan}/{uuid}
	PathDasar string `gorm:"column:path_dasar" json:"path_dasar"`
	IsPublik  bool   `gorm:"column:is_publik" json:"is_publik"`
	// StatusProses is written as text and cast in SQL — GORM cannot infer the
	// Postgres enum type from a Go string.
	StatusProses StatusProses `gorm:"column:status_proses;type:status_proses_file" json:"status_proses"`
	// MetadataEXIF is the original EXIF, kept as evidence. Re-encoding strips
	// EXIF from every written variant, so this row is the only copy left.
	MetadataEXIF *string `gorm:"column:metadata_exif;type:jsonb" json:"-"`

	model.Audit
}

// TableName pins the table; GORM's pluraliser would guess wrong.
func (File) TableName() string { return "mst_file" }

// FileVarian maps mst_file_varian: one rendered size of one file.
type FileVarian struct {
	ID     int64  `gorm:"column:id;primaryKey" json:"-"`
	FileID int64  `gorm:"column:file_id" json:"-"`
	Varian Varian `gorm:"column:varian;type:varian_file" json:"varian"`
	// Path is relative to the storage root, so moving the root does not
	// rewrite every row.
	Path string `gorm:"column:path" json:"-"`
	// URLPublik is filled only for public files; private ones are readable
	// solely through GET /files/{uuid}/{varian}.
	URLPublik  *string `gorm:"column:url_publik" json:"url_publik,omitempty"`
	LebarPx    *int    `gorm:"column:lebar_px" json:"lebar_px"`
	TinggiPx   *int    `gorm:"column:tinggi_px" json:"tinggi_px"`
	UkuranByte int64   `gorm:"column:ukuran_byte" json:"ukuran_byte"`
	// MimeType records what was actually written, which is not always what the
	// source was: medium/low are re-encoded.
	MimeType  string    `gorm:"column:mime_type" json:"mime_type"`
	Kualitas  *int      `gorm:"column:kualitas" json:"kualitas,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName pins the table; GORM's pluraliser would guess wrong.
func (FileVarian) TableName() string { return "mst_file_varian" }
