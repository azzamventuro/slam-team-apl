// Package audit writes the business audit trail (log_aktivitas). It lives in
// shared/ rather than in a feature module because every state-changing action
// in the system records one row: RBAC changes, settings updates, absensi
// overrides, KTA cabut/cetak, login, and ordinary create/ubah/hapus.
//
// The table is append-only — no update, no delete, no soft-delete columns.
package audit

import (
	"context"
	"encoding/json"
	"time"

	"slam-team-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// LogAktivitas maps the log_aktivitas table. Rows are written through Write;
// the entity exists for the read side (a future audit viewer) and to keep the
// column set documented in Go.
type LogAktivitas struct {
	ID          int64           `gorm:"column:id;primaryKey" json:"id"`
	AktorUserID *int64          `gorm:"column:aktor_user_id" json:"aktor_user_id,omitempty"`
	Modul       string          `gorm:"column:modul" json:"modul"`
	Aksi        string          `gorm:"column:aksi" json:"aksi"`
	ReffType    *string         `gorm:"column:reff_type" json:"reff_type,omitempty"`
	ReffID      *int64          `gorm:"column:reff_id" json:"reff_id,omitempty"`
	Ringkasan   *string         `gorm:"column:ringkasan" json:"ringkasan,omitempty"`
	NilaiLama   json.RawMessage `gorm:"column:nilai_lama;type:jsonb" json:"nilai_lama,omitempty"`
	NilaiBaru   json.RawMessage `gorm:"column:nilai_baru;type:jsonb" json:"nilai_baru,omitempty"`
	IPAddress   *string         `gorm:"column:ip_address;type:inet" json:"ip_address,omitempty"`
	UserAgent   *string         `gorm:"column:user_agent" json:"user_agent,omitempty"`
	CreatedAt   time.Time       `gorm:"column:created_at" json:"created_at"`
}

// TableName pins the table; GORM's pluraliser would guess wrong.
func (LogAktivitas) TableName() string { return "log_aktivitas" }

// Entry is what a caller describes: what happened, to which row, by whom.
// Only Modul and Aksi are required.
//
//	audit.Write(ctx, db, audit.Entry{
//	    Modul: "kta", Aksi: "cabut", ReffType: "kta", ReffID: &id,
//	    Ringkasan: "Mencabut KTA 3573.0001",
//	    NilaiLama: before, NilaiBaru: after,
//	}.WithRequest(c, actorID))
type Entry struct {
	// AktorUserID is nil for system/scheduled-job actions.
	AktorUserID *int64
	// Modul is the owning feature: hak_akses / pengaturan / absensi / kta / …
	Modul string
	// Aksi is the verb: buat / ubah / hapus / override / cabut / cetak / login.
	Aksi string
	// ReffType/ReffID point at the affected row (polymorphic, both optional).
	ReffType string
	ReffID   *int64
	// Ringkasan is a ready-to-display sentence, max 255 chars.
	Ringkasan string
	// NilaiLama/NilaiBaru are snapshots; any JSON-marshalable value, or nil.
	NilaiLama any
	NilaiBaru any
	// IPAddress must be a bare IP (c.ClientIP()); empty stores NULL.
	IPAddress string
	UserAgent string
}

// WithRequest fills the actor and the request fingerprint from a Gin context.
// Callers that have neither (jobs, the CLI) simply skip it.
func (e Entry) WithRequest(c *gin.Context, aktorUserID *int64) Entry {
	e.AktorUserID = aktorUserID
	if c != nil {
		e.IPAddress = c.ClientIP()
		if c.Request != nil {
			e.UserAgent = c.Request.UserAgent()
		}
	}
	return e
}

// Writer is the injectable form of Write, so a module takes its audit
// dependency at Initialize() like every other collaborator.
type Writer struct{ db *gorm.DB }

// NewWriter binds a Writer to the process-wide pool.
func NewWriter(db *gorm.DB) *Writer { return &Writer{db: db} }

// Log records one entry. It never returns an error — see Write.
func (w *Writer) Log(ctx context.Context, e Entry) {
	if w == nil || w.db == nil {
		return
	}
	Write(ctx, w.db, e)
}

// Write inserts one log_aktivitas row.
//
// It deliberately returns nothing: a failed audit write must never fail the
// business action that succeeded. Failures are reported to Zap and swallowed.
//
// Call it AFTER the business transaction commits, with the plain pool rather
// than the transaction handle, so a rolled-back change leaves no audit row.
func Write(ctx context.Context, db *gorm.DB, e Entry) {
	if db == nil || e.Modul == "" || e.Aksi == "" {
		logger.Error("audit: entry tidak lengkap",
			zap.String("modul", e.Modul), zap.String("aksi", e.Aksi))
		return
	}

	lama, err := marshalSnapshot(e.NilaiLama)
	if err != nil {
		logger.Error("audit: nilai_lama tidak bisa di-marshal", zap.Error(err))
	}
	baru, err := marshalSnapshot(e.NilaiBaru)
	if err != nil {
		logger.Error("audit: nilai_baru tidak bisa di-marshal", zap.Error(err))
	}

	// Raw SQL rather than GORM Create: the casts are what make an empty IP land
	// as NULL instead of being rejected by the inet type, and jsonb columns
	// need the text-to-jsonb cast to accept a parameter.
	err = db.WithContext(ctx).Exec(`
		INSERT INTO log_aktivitas
			(aktor_user_id, modul, aksi, reff_type, reff_id, ringkasan,
			 nilai_lama, nilai_baru, ip_address, user_agent, created_at)
		VALUES (?, ?, ?, NULLIF(?, ''), ?, NULLIF(?, ''),
		        CAST(? AS jsonb), CAST(? AS jsonb), CAST(NULLIF(?, '') AS inet), NULLIF(?, ''), now())
	`,
		e.AktorUserID, e.Modul, e.Aksi, e.ReffType, e.ReffID, truncate(e.Ringkasan, 255),
		lama, baru, e.IPAddress, e.UserAgent,
	).Error
	if err != nil {
		logger.Error("audit: gagal menulis log_aktivitas",
			zap.String("modul", e.Modul), zap.String("aksi", e.Aksi), zap.Error(err))
	}
}

// marshalSnapshot turns a snapshot into a *string for the jsonb parameter; nil
// (and a nil-valued interface) become a NULL column rather than the JSON text
// "null", which would be indistinguishable from a real recorded null.
func marshalSnapshot(v any) (*string, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if string(b) == "null" {
		return nil, nil
	}
	s := string(b)
	return &s, nil
}

// truncate keeps ringkasan inside varchar(255) — an over-long summary must not
// turn into a failed audit write.
func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}
