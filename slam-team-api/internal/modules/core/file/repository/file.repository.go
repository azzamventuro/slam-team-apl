// Package repository is the GORM data access for mst_file and mst_file_varian.
// It returns domain entities and never touches *gin.Context.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"slam-team-api/internal/modules/core/file/domain"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/model"

	"gorm.io/gorm"
)

// FileRepository reads and writes the file layer's two tables.
type FileRepository struct {
	db *gorm.DB
}

// NewFileRepository binds the repository to the pool.
func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

// Create inserts the mst_file row together with its first variants in ONE
// transaction: a file row without the original variant would advertise a byte
// stream nothing can serve.
//
// The enum columns are written through explicit casts because GORM hands a Go
// string to the driver with no type, and Postgres will not coerce it into
// varian_file / status_proses_file inside a multi-row insert.
func (r *FileRepository) Create(ctx context.Context, f *domain.File, varians []domain.FileVarian) error {
	// RETURNING is scanned into a narrow struct rather than into f: GORM's
	// Scan REPLACES the destination, so handing it the entity would blank
	// every column the statement does not return.
	var created struct {
		ID        int64     `gorm:"column:id"`
		UUID      string    `gorm:"column:uuid"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(`
			INSERT INTO mst_file
				(uuid, nama_asli, nama_slug, ekstensi, mime_type, ukuran_byte,
				 hash_sha256, lebar_px, tinggi_px, kategori, reff_type, reff_id,
				 storage_driver, path_dasar, is_publik, status_proses,
				 metadata_exif, created_at, created_by)
			VALUES (CAST(? AS uuid), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			        CAST(? AS status_proses_file), CAST(? AS jsonb), now(), ?)
			RETURNING id, uuid, created_at`,
			f.UUID, f.NamaAsli, f.NamaSlug, f.Ekstensi, f.MimeType, f.UkuranByte,
			f.HashSHA256, f.LebarPx, f.TinggiPx, f.Kategori, f.ReffType, f.ReffID,
			f.StorageDriver, f.PathDasar, f.IsPublik, string(f.StatusProses),
			f.MetadataEXIF, f.CreatedBy,
		).Scan(&created).Error; err != nil {
			return err
		}
		f.ID, f.UUID, f.CreatedAt = created.ID, created.UUID, created.CreatedAt

		for i := range varians {
			varians[i].FileID = f.ID
			if err := createVarian(tx, &varians[i]); err != nil {
				return err
			}
		}
		return nil
	})
}

// CreateVarian inserts one variant row. The async generator calls it per
// variant, so a slow "low" render does not hold "medium" back.
func (r *FileRepository) CreateVarian(ctx context.Context, v *domain.FileVarian) error {
	return createVarian(r.db.WithContext(ctx), v)
}

// createVarian is the shared insert, usable on the pool or inside a
// transaction. ON CONFLICT keeps a retried generator idempotent against the
// (file_id, varian) unique index instead of failing the whole run.
func createVarian(db *gorm.DB, v *domain.FileVarian) error {
	return db.Exec(`
		INSERT INTO mst_file_varian
			(file_id, varian, path, url_publik, lebar_px, tinggi_px,
			 ukuran_byte, mime_type, kualitas, created_at)
		VALUES (?, CAST(? AS varian_file), ?, ?, ?, ?, ?, ?, ?, now())
		ON CONFLICT (file_id, varian) DO UPDATE SET
			path        = EXCLUDED.path,
			url_publik  = EXCLUDED.url_publik,
			lebar_px    = EXCLUDED.lebar_px,
			tinggi_px   = EXCLUDED.tinggi_px,
			ukuran_byte = EXCLUDED.ukuran_byte,
			mime_type   = EXCLUDED.mime_type,
			kualitas    = EXCLUDED.kualitas`,
		v.FileID, string(v.Varian), v.Path, v.URLPublik, v.LebarPx, v.TinggiPx,
		v.UkuranByte, v.MimeType, v.Kualitas,
	).Error
}

// FindByUUID returns a live (not soft-deleted) file. A deleted or unknown uuid
// is ErrNotFound — the two are deliberately indistinguishable to a caller.
func (r *FileRepository) FindByUUID(ctx context.Context, uuid string) (domain.File, error) {
	var f domain.File
	err := r.db.WithContext(ctx).Model(&domain.File{}).
		Scopes(model.NotDeleted).
		Where("uuid = CAST(? AS uuid)", uuid).
		First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.File{}, fmt.Errorf("berkas %s: %w", uuid, apperr.ErrNotFound)
	}
	return f, err
}

// FindVarian returns one variant of a live file, looked up by the public uuid
// in a single join so the handler never has to trust two separate reads.
func (r *FileRepository) FindVarian(ctx context.Context, uuid string, varian domain.Varian) (domain.FileVarian, error) {
	var v domain.FileVarian
	err := r.db.WithContext(ctx).
		Table("mst_file_varian AS v").
		Select("v.*").
		Joins("JOIN mst_file AS f ON f.id = v.file_id").
		Where("f.uuid = CAST(? AS uuid) AND f.is_deleted = false AND v.varian = CAST(? AS varian_file)", uuid, string(varian)).
		First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.FileVarian{}, fmt.Errorf("varian %s berkas %s: %w", varian, uuid, apperr.ErrNotFound)
	}
	return v, err
}

// VariansOf returns every variant of one file, original first.
func (r *FileRepository) VariansOf(ctx context.Context, fileID int64) ([]domain.FileVarian, error) {
	var rows []domain.FileVarian
	err := r.db.WithContext(ctx).
		Where("file_id = ?", fileID).
		Order("CASE varian WHEN 'original' THEN 0 WHEN 'medium' THEN 1 ELSE 2 END").
		Find(&rows).Error
	return rows, err
}

// ExistsByHash is the dedup probe: the same bytes, in the same kategori,
// uploaded again by the same person. It returns ok=false rather than an error
// when there is no match, because "no previous copy" is the normal case.
//
// Dedup is deliberately scoped per uploader: sharing one row between two users
// would make one person's soft delete remove the other's file.
func (r *FileRepository) ExistsByHash(ctx context.Context, hash, kategori string, createdBy *int64) (domain.File, bool, error) {
	if hash == "" || createdBy == nil {
		return domain.File{}, false, nil
	}
	var f domain.File
	err := r.db.WithContext(ctx).Model(&domain.File{}).
		Scopes(model.NotDeleted).
		Where("hash_sha256 = ? AND kategori = ? AND created_by = ?", hash, kategori, *createdBy).
		Order("id ASC").
		First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.File{}, false, nil
	}
	if err != nil {
		return domain.File{}, false, err
	}
	return f, true, nil
}

// UpdateStatus moves status_proses once the async variant pass finishes —
// 'selesai' or, just as importantly, 'gagal'. A failure that left the row at
// 'menunggu' would look like work still in progress forever.
func (r *FileRepository) UpdateStatus(ctx context.Context, fileID int64, status domain.StatusProses) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE mst_file
		   SET status_proses = CAST(? AS status_proses_file), modified_at = now()
		 WHERE id = ?`, string(status), fileID).Error
}

// SoftDelete marks the row deleted. Bytes on disk are left alone — a separate
// sweeper reclaims them, so a mistaken delete is still recoverable.
//
// The is_deleted = false guard makes a second delete a no-op instead of
// overwriting who deleted it first.
func (r *FileRepository) SoftDelete(ctx context.Context, fileID int64, actor *int64) error {
	res := r.db.WithContext(ctx).Model(&domain.File{}).
		Where("id = ? AND is_deleted = false", fileID).
		Updates(model.SoftDeleteFields(actor))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("berkas %d: %w", fileID, apperr.ErrNotFound)
	}
	return nil
}
