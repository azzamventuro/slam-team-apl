// Package repository is the GORM data access for absensi_izin plus the
// narrow read-only lookups the service needs from neighbouring tables
// (jadwal_sesi, mst_file, jadwal_peserta). It never reaches into those
// modules' repositories: a few explicit raw reads keep the coupling visible.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"slam-team-api/internal/modules/core/izin/domain"
	"slam-team-api/internal/shared/apperr"

	"gorm.io/gorm"
)

// IzinRepository is the data-access layer for absensi_izin.
type IzinRepository struct {
	db *gorm.DB
}

// NewIzinRepository creates a repository bound to the pool.
func NewIzinRepository(db *gorm.DB) *IzinRepository {
	return &IzinRepository{db: db}
}

// Transaction runs fn inside one DB transaction. The service uses it to make
// "decide the request + fan-out notifikasi" atomic.
func (r *IzinRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// WithTx returns a copy bound to a transaction handle.
func (r *IzinRepository) WithTx(tx *gorm.DB) *IzinRepository {
	return &IzinRepository{db: tx}
}

// ── Neighbour lookups ──

// SesiRingkas is the slice of a jadwal_sesi row (plus its schedule) the
// service needs to validate a request and word the notifications.
type SesiRingkas struct {
	ID           int64     `gorm:"column:id"`
	JadwalID     int64     `gorm:"column:jadwal_id"`
	TanggalLokal time.Time `gorm:"column:tanggal_lokal"`
	Status       string    `gorm:"column:status"`
	JadwalNama   string    `gorm:"column:jadwal_nama"`
	JadwalKode   string    `gorm:"column:jadwal_kode"`
}

// FindSesi returns one session with its schedule's name, or ErrNotFound.
func (r *IzinRepository) FindSesi(ctx context.Context, id int64) (*SesiRingkas, error) {
	var s SesiRingkas
	err := r.db.WithContext(ctx).Table("jadwal_sesi AS s").
		Select("s.id, s.jadwal_id, s.tanggal_lokal, s.status, j.nama AS jadwal_nama, j.kode AS jadwal_kode").
		Joins("JOIN jadwal j ON j.id = s.jadwal_id").
		Where("s.id = ?", id).
		Take(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: sesi", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &s, nil
}

// FileExists checks whether a mst_file row exists and is not soft-deleted.
func (r *IzinRepository) FileExists(ctx context.Context, fileID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("mst_file").
		Where("id = ? AND is_deleted = false", fileID).
		Count(&count).Error
	return count > 0, err
}

// HasOpenRequest reports whether the member already has a live request on
// the session that is pending or approved — a second one would only create
// two answers to the same question. A rejected request does not count: the
// member may try again with a better reason or an attachment.
func (r *IzinRepository) HasOpenRequest(ctx context.Context, sesiID, anggotaID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Izin{}).
		Where("sesi_id = ? AND anggota_id = ? AND is_deleted = false AND status IN ?",
			sesiID, anggotaID, []domain.StatusIzin{domain.StatusMenunggu, domain.StatusDisetujui}).
		Count(&count).Error
	return count > 0, err
}

// MarkPesertaIzin flips the member's live session-specific assignment row
// to status_tugas = 'izin', so the roster shows the excuse. Schedule-wide
// rows (sesi_id NULL) are left alone: they cover every session and cannot
// carry a one-session excuse. Returns the number of rows changed; zero is
// normal when the member was assigned schedule-wide or not at all.
func (r *IzinRepository) MarkPesertaIzin(ctx context.Context, sesiID, anggotaID int64) (int64, error) {
	res := r.db.WithContext(ctx).Exec(`
		UPDATE jadwal_peserta
		SET status_tugas = 'izin'
		WHERE sesi_id = ? AND anggota_id = ? AND is_deleted = false
		  AND status_tugas <> 'izin'`, sesiID, anggotaID)
	return res.RowsAffected, res.Error
}

// ── absensi_izin ──

// baseQuery filters soft-deleted rows, table-qualified because the reads
// JOIN tables that carry the same column.
func (r *IzinRepository) baseQuery() *gorm.DB {
	return r.db.Model(&domain.Izin{}).Where("absensi_izin.is_deleted = false")
}

// readJoin resolves the read-only projections. Call before Find/First,
// never before Count.
func (r *IzinRepository) readJoin(q *gorm.DB) *gorm.DB {
	return q.
		Select(`
			absensi_izin.*,
			a.nama_lengkap   AS anggota_nama,
			a.no_induk       AS no_induk,
			s.tanggal_lokal  AS sesi_tanggal,
			s.status         AS sesi_status,
			j.nama           AS jadwal_nama,
			j.kode           AS jadwal_kode,
			f.uuid           AS lampiran_uuid,
			pa.nama_lengkap  AS diproses_oleh_nama
		`).
		Joins("LEFT JOIN anggota a ON a.id = absensi_izin.anggota_id").
		Joins("LEFT JOIN jadwal_sesi s ON s.id = absensi_izin.sesi_id").
		Joins("LEFT JOIN jadwal j ON j.id = absensi_izin.jadwal_id").
		Joins("LEFT JOIN mst_file f ON f.id = absensi_izin.lampiran_file_id AND f.is_deleted = false").
		Joins("LEFT JOIN users pu ON pu.id = absensi_izin.diproses_oleh").
		Joins("LEFT JOIN anggota pa ON pa.id = pu.anggota_id")
}

// Create inserts one request and populates its ID.
func (r *IzinRepository) Create(ctx context.Context, z *domain.Izin) error {
	return r.db.WithContext(ctx).Omit(domain.ProjectionColumns...).Create(z).Error
}

// FindByID returns one live request with projections, or ErrNotFound.
func (r *IzinRepository) FindByID(ctx context.Context, id int64) (*domain.Izin, error) {
	var z domain.Izin
	err := r.readJoin(r.baseQuery().Where("absensi_izin.id = ?", id)).
		WithContext(ctx).First(&z).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: izin", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &z, nil
}

// ListFilter narrows List. AnggotaID is the cakupan owner filter the service
// sets for milik_sendiri callers; the rest are query-string filters.
type ListFilter struct {
	AnggotaID *int64
	Status    string
	SesiID    *int64
	JadwalID  *int64
	Jenis     string
	Q         string
	// Sort is a column already whitelisted by the service (never raw input).
	Sort   string
	Desc   bool
	Offset int
	Limit  int
}

// List returns one page of live requests with projections, newest first by
// default (the approval queue reads oldest-first via sort=created_at).
func (r *IzinRepository) List(ctx context.Context, f ListFilter) ([]domain.Izin, int64, error) {
	q := r.baseQuery().WithContext(ctx)
	if f.AnggotaID != nil {
		q = q.Where("absensi_izin.anggota_id = ?", *f.AnggotaID)
	}
	if f.Status != "" {
		q = q.Where("absensi_izin.status = ?", f.Status)
	}
	if f.SesiID != nil {
		q = q.Where("absensi_izin.sesi_id = ?", *f.SesiID)
	}
	if f.JadwalID != nil {
		q = q.Where("absensi_izin.jadwal_id = ?", *f.JadwalID)
	}
	if f.Jenis != "" {
		q = q.Where("absensi_izin.jenis = ?", f.Jenis)
	}
	if f.Q != "" {
		like := "%" + f.Q + "%"
		q = q.Where("absensi_izin.anggota_id IN (SELECT id FROM anggota WHERE nama_lengkap ILIKE ? OR no_induk ILIKE ?)", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "absensi_izin.created_at DESC, absensi_izin.id DESC"
	if f.Sort != "" {
		dir := "ASC"
		if f.Desc {
			dir = "DESC"
		}
		order = fmt.Sprintf("absensi_izin.%s %s, absensi_izin.id %s", f.Sort, dir, dir)
	}

	var items []domain.Izin
	err := r.readJoin(q).
		Order(order).
		Offset(f.Offset).Limit(f.Limit).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// Approve moves a pending request to disetujui. The status guard in the
// WHERE makes two concurrent reviewers safe: the second one changes nothing
// and gets ErrConflict.
func (r *IzinRepository) Approve(ctx context.Context, id, oleh int64) error {
	return r.decide(ctx, id, domain.StatusDisetujui, oleh, nil)
}

// Tolak moves a pending request to ditolak with the reviewer's note.
func (r *IzinRepository) Tolak(ctx context.Context, id, oleh int64, catatan string) error {
	return r.decide(ctx, id, domain.StatusDitolak, oleh, &catatan)
}

func (r *IzinRepository) decide(ctx context.Context, id int64, status domain.StatusIzin, oleh int64, catatan *string) error {
	now := time.Now().UTC()
	fields := map[string]any{
		"status":        string(status),
		"diproses_oleh": oleh,
		"diproses_pada": now,
		"modified_by":   oleh,
		"modified_at":   now,
	}
	if catatan != nil {
		fields["catatan_peninjau"] = *catatan
	}
	res := r.db.WithContext(ctx).Model(&domain.Izin{}).
		Where("id = ? AND is_deleted = false AND status = ?", id, domain.StatusMenunggu).
		Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("izin sudah diproses: %w", apperr.ErrConflict)
	}
	return nil
}
