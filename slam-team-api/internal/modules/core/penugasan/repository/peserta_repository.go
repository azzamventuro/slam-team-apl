// Package repository is the GORM data access for jadwal_peserta plus the
// small read-only lookups the service needs from neighbouring tables
// (jadwal, jadwal_sesi, anggota, users). It never reaches into those
// modules' repositories: a few narrow raw reads keep the coupling explicit.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"slam-team-api/internal/modules/core/penugasan/domain"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/model"

	"gorm.io/gorm"
)

// pesertaBatchChunk caps rows per INSERT (14 columns per row) well under
// PostgreSQL's 65535-parameter limit; bulk is already capped at 500 ids.
const pesertaBatchChunk = 500

// PesertaRepository is the data-access layer for jadwal_peserta.
type PesertaRepository struct {
	db *gorm.DB
}

// NewPesertaRepository creates a repository bound to the pool.
func NewPesertaRepository(db *gorm.DB) *PesertaRepository {
	return &PesertaRepository{db: db}
}

// Transaction runs fn inside one DB transaction. The service uses it to make
// "insert assignment + fan-out notifikasi + refresh counter" atomic.
func (r *PesertaRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// WithTx returns a copy bound to a transaction handle.
func (r *PesertaRepository) WithTx(tx *gorm.DB) *PesertaRepository {
	return &PesertaRepository{db: tx}
}

// ── Neighbour lookups ──

// JadwalRingkas is the slice of a jadwal row the service needs.
type JadwalRingkas struct {
	ID     int64  `gorm:"column:id"`
	Kode   string `gorm:"column:kode"`
	Nama   string `gorm:"column:nama"`
	Status string `gorm:"column:status"`
}

// FindJadwal returns the non-deleted schedule, or ErrNotFound.
func (r *PesertaRepository) FindJadwal(ctx context.Context, id int64) (*JadwalRingkas, error) {
	var j JadwalRingkas
	err := r.db.WithContext(ctx).Table("jadwal").
		Select("id, kode, nama, status").
		Where("id = ? AND is_deleted = false", id).
		Take(&j).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: jadwal", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &j, nil
}

// LockJadwal takes a row lock on the schedule for the rest of the
// transaction, serialising concurrent assignments to the same jadwal so the
// duplicate check below cannot race (the DBML index is not unique).
func (r *PesertaRepository) LockJadwal(ctx context.Context, id int64) error {
	var lockedID int64
	return r.db.WithContext(ctx).Raw(
		"SELECT id FROM jadwal WHERE id = ? FOR UPDATE", id,
	).Scan(&lockedID).Error
}

// SesiRingkas is the slice of a jadwal_sesi row the service needs.
type SesiRingkas struct {
	ID           int64     `gorm:"column:id"`
	JadwalID     int64     `gorm:"column:jadwal_id"`
	TanggalLokal time.Time `gorm:"column:tanggal_lokal"`
	Status       string    `gorm:"column:status"`
}

// FindSesi returns one session, or ErrNotFound. The service checks that it
// belongs to the schedule in the URL.
func (r *PesertaRepository) FindSesi(ctx context.Context, id int64) (*SesiRingkas, error) {
	var s SesiRingkas
	err := r.db.WithContext(ctx).Table("jadwal_sesi").
		Select("id, jadwal_id, tanggal_lokal, status").
		Where("id = ?", id).
		Take(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: sesi", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &s, nil
}

// AnggotaRingkas is the slice of an anggota row the service needs.
type AnggotaRingkas struct {
	ID          int64  `gorm:"column:id"`
	NamaLengkap string `gorm:"column:nama_lengkap"`
}

// FindAnggota returns the non-deleted anggota among ids, keyed by id. Missing
// ids are simply absent — the service decides whether that is an error.
func (r *PesertaRepository) FindAnggota(ctx context.Context, ids []int64) (map[int64]AnggotaRingkas, error) {
	out := map[int64]AnggotaRingkas{}
	if len(ids) == 0 {
		return out, nil
	}
	var rows []AnggotaRingkas
	err := r.db.WithContext(ctx).Table("anggota").
		Select("id, nama_lengkap").
		Where("id IN ? AND is_deleted = false", ids).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, a := range rows {
		out[a.ID] = a
	}
	return out, nil
}

// AnggotaBerakun resolves which of the anggota have a live (non-deleted)
// users row, returning anggota_id → users.id. An anggota absent from the
// result has no account: it gets wajib_absen=false and no notification.
func (r *PesertaRepository) AnggotaBerakun(ctx context.Context, anggotaIDs []int64) (map[int64]int64, error) {
	out := map[int64]int64{}
	if len(anggotaIDs) == 0 {
		return out, nil
	}
	var rows []struct {
		AnggotaID int64 `gorm:"column:anggota_id"`
		UserID    int64 `gorm:"column:id"`
	}
	err := r.db.WithContext(ctx).Table("users").
		Select("anggota_id, id").
		Where("anggota_id IN ? AND is_deleted = false", anggotaIDs).
		Order("id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if _, dup := out[row.AnggotaID]; !dup {
			out[row.AnggotaID] = row.UserID
		}
	}
	return out, nil
}

// ── jadwal_peserta ──

// baseQuery filters soft-deleted rows, table-qualified because the reads
// JOIN anggota (which carries the same column).
func (r *PesertaRepository) baseQuery() *gorm.DB {
	return r.db.Model(&domain.JadwalPeserta{}).Where("jadwal_peserta.is_deleted = false")
}

// readJoin resolves the read-only projections (anggota_nama, no_induk,
// anggota_berakun, sesi_tanggal). Call before Find/First, never before Count.
func (r *PesertaRepository) readJoin(q *gorm.DB) *gorm.DB {
	return q.
		Select(`
			jadwal_peserta.*,
			a.nama_lengkap AS anggota_nama,
			a.no_induk     AS no_induk,
			EXISTS (SELECT 1 FROM users u WHERE u.anggota_id = jadwal_peserta.anggota_id AND u.is_deleted = false) AS anggota_berakun,
			s.tanggal_lokal AS sesi_tanggal
		`).
		Joins("LEFT JOIN anggota a ON a.id = jadwal_peserta.anggota_id").
		Joins("LEFT JOIN jadwal_sesi s ON s.id = jadwal_peserta.sesi_id")
}

// ExistingAnggota returns anggota_id → jadwal_peserta.id for the anggota
// already (actively) assigned to the schedule. This is the duplicate key the
// spec names: one live row per (anggota_id, jadwal_id).
func (r *PesertaRepository) ExistingAnggota(ctx context.Context, jadwalID int64, anggotaIDs []int64) (map[int64]int64, error) {
	out := map[int64]int64{}
	if len(anggotaIDs) == 0 {
		return out, nil
	}
	var rows []struct {
		ID        int64 `gorm:"column:id"`
		AnggotaID int64 `gorm:"column:anggota_id"`
	}
	err := r.db.WithContext(ctx).Model(&domain.JadwalPeserta{}).
		Select("id, anggota_id").
		Where("jadwal_id = ? AND anggota_id IN ? AND is_deleted = false", jadwalID, anggotaIDs).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.AnggotaID] = row.ID
	}
	return out, nil
}

// Create inserts one assignment and populates its ID.
func (r *PesertaRepository) Create(ctx context.Context, p *domain.JadwalPeserta) error {
	return r.db.WithContext(ctx).Omit(projectionColumns...).Create(p).Error
}

// BulkCreate inserts many assignments in chunks and populates their IDs.
// Duplicate filtering is the service's job (via ExistingAnggota under
// LockJadwal); this method inserts exactly what it is given.
func (r *PesertaRepository) BulkCreate(ctx context.Context, rows []domain.JadwalPeserta) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Omit(projectionColumns...).CreateInBatches(rows, pesertaBatchChunk).Error
}

// projectionColumns are the "->" fields; named explicitly on writes so GORM
// never tries to insert them.
var projectionColumns = []string{"anggota_nama", "no_induk", "anggota_berakun", "sesi_tanggal"}

// ListFilter narrows List.
type ListFilter struct {
	// SesiID, when set, returns rows for that session PLUS the schedule-wide
	// rows (sesi_id IS NULL), because those apply to every session.
	SesiID      *int64
	StatusTugas string
	WajibAbsen  *bool
	// Q matches anggota nama_lengkap / no_induk (ILIKE).
	Q string
}

// List returns every live assignment of a schedule with projections.
func (r *PesertaRepository) List(ctx context.Context, jadwalID int64, f ListFilter) ([]domain.JadwalPeserta, error) {
	q := r.baseQuery().Where("jadwal_peserta.jadwal_id = ?", jadwalID)
	if f.SesiID != nil {
		q = q.Where("(jadwal_peserta.sesi_id = ? OR jadwal_peserta.sesi_id IS NULL)", *f.SesiID)
	}
	if f.StatusTugas != "" {
		q = q.Where("jadwal_peserta.status_tugas = ?", f.StatusTugas)
	}
	if f.WajibAbsen != nil {
		q = q.Where("jadwal_peserta.wajib_absen = ?", *f.WajibAbsen)
	}
	if f.Q != "" {
		like := "%" + f.Q + "%"
		q = q.Where("(a.nama_lengkap ILIKE ? OR a.no_induk ILIKE ?)", like, like)
	}

	var items []domain.JadwalPeserta
	err := r.readJoin(q).WithContext(ctx).
		Order("a.nama_lengkap ASC, jadwal_peserta.id ASC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

// FindByID returns one live assignment with projections, or ErrNotFound.
func (r *PesertaRepository) FindByID(ctx context.Context, id int64) (*domain.JadwalPeserta, error) {
	var p domain.JadwalPeserta
	err := r.readJoin(r.baseQuery().Where("jadwal_peserta.id = ?", id)).
		WithContext(ctx).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: peserta", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &p, nil
}

// SoftDeleteByAnggota marks every live assignment of one anggota on one
// schedule deleted and returns how many rows changed (0 = nothing to remove).
func (r *PesertaRepository) SoftDeleteByAnggota(ctx context.Context, jadwalID, anggotaID int64, actor *int64) (int64, error) {
	res := r.db.WithContext(ctx).Model(&domain.JadwalPeserta{}).
		Where("jadwal_id = ? AND anggota_id = ? AND is_deleted = false", jadwalID, anggotaID).
		Updates(model.SoftDeleteFields(actor))
	return res.RowsAffected, res.Error
}

// Respon records the assignee's answer. Ownership is checked by the service.
func (r *PesertaRepository) Respon(ctx context.Context, id int64, status domain.StatusTugas, keterangan *string) error {
	fields := map[string]any{
		"status_tugas":  string(status),
		"direspon_pada": time.Now().UTC(),
	}
	if keterangan != nil {
		fields["keterangan"] = *keterangan
	}
	return r.db.WithContext(ctx).Model(&domain.JadwalPeserta{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(fields).Error
}

// SyncJmlDitugaskan recomputes jadwal_sesi.jml_ditugaskan for every session
// of a schedule from the live assignment rows: a session counts the rows
// bound to it plus the schedule-wide rows (sesi_id IS NULL). Recomputing
// rather than incrementing keeps the counter right after removals and
// re-assignments, without ever drifting.
func (r *PesertaRepository) SyncJmlDitugaskan(ctx context.Context, jadwalID int64) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE jadwal_sesi s
		SET jml_ditugaskan = (
			SELECT count(*) FROM jadwal_peserta p
			WHERE p.jadwal_id = s.jadwal_id
			  AND p.is_deleted = false
			  AND (p.sesi_id IS NULL OR p.sesi_id = s.id)
		)
		WHERE s.jadwal_id = ?`, jadwalID).Error
}
