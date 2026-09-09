// Package domain holds the pengaturan entity and its value types.
package domain

import (
	"encoding/json"
	"time"
)

// TipeNilai is the tipe_nilai_pengaturan enum. It decides two things: which
// control the settings page renders, and how nilai (always stored as text) is
// cast on the way out and parsed on the way in.
type TipeNilai string

// The complete enum, in the order 0001_enums declares it.
const (
	TipeString  TipeNilai = "string"
	TipeInteger TipeNilai = "integer"
	TipeBoolean TipeNilai = "boolean"
	TipeJSON    TipeNilai = "json"
	TipeDate    TipeNilai = "date"
)

// Valid reports whether t is one of the enum values.
func (t TipeNilai) Valid() bool {
	switch t {
	case TipeString, TipeInteger, TipeBoolean, TipeJSON, TipeDate:
		return true
	default:
		return false
	}
}

// Pengaturan maps mst_pengaturan: a typed key-value row carrying enough
// metadata (label, satuan, opsi, keterangan, urutan) for the frontend to
// generate its own settings page — adding a setting is one INSERT, no code.
//
// It does NOT embed model.Audit: the table has no created_at and no
// soft-delete columns. Settings are seeded once and only ever edited.
type Pengaturan struct {
	ID    int64  `gorm:"column:id;primaryKey" json:"id"`
	Kunci string `gorm:"column:kunci" json:"kunci"`
	Grup  string `gorm:"column:grup" json:"grup"`
	Label string `gorm:"column:label" json:"label"`
	// Nilai is nullable and always text; read it through TipeNilai.
	Nilai       *string   `gorm:"column:nilai" json:"nilai"`
	TipeNilai   TipeNilai `gorm:"column:tipe_nilai" json:"tipe_nilai"`
	NilaiBawaan *string   `gorm:"column:nilai_bawaan" json:"nilai_bawaan"`
	// Opsi is the allowed set when the control is a dropdown, e.g. ["png"].
	Opsi       json.RawMessage `gorm:"column:opsi;type:jsonb" json:"opsi"`
	Satuan     *string         `gorm:"column:satuan" json:"satuan"`
	Keterangan *string         `gorm:"column:keterangan" json:"keterangan"`
	Urutan     int             `gorm:"column:urutan" json:"urutan"`
	IsPublik   bool            `gorm:"column:is_publik" json:"is_publik"`
	IsTerkunci bool            `gorm:"column:is_terkunci" json:"is_terkunci"`
	ModifiedAt *time.Time      `gorm:"column:modified_at" json:"modified_at"`
	ModifiedBy *int64          `gorm:"column:modified_by" json:"modified_by"`
}

// TableName pins the table; GORM's pluraliser would guess wrong.
func (Pengaturan) TableName() string { return "mst_pengaturan" }
