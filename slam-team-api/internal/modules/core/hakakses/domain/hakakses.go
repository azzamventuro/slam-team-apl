// Package hakakses holds the GORM entities for the dynamic RBAC module.
// Tables: mst_modul, mst_permission, mst_role, role_permission, user_role.
package hakakses

import (
	"time"

	"slam-team-api/internal/shared/model"
)

// ── Enum types (mirror Postgres enums from 0001_enums) ──

type AksiPermission string

const (
	AksiCreate     AksiPermission = "create"
	AksiRead       AksiPermission = "read"
	AksiUpdate     AksiPermission = "update"
	AksiDelete     AksiPermission = "delete"
	AksiApprove    AksiPermission = "approve"
	AksiAssign     AksiPermission = "assign"
	AksiExport     AksiPermission = "export"
	AksiOverride   AksiPermission = "override"
	AksiPrint      AksiPermission = "print"
	AksiCabut      AksiPermission = "cabut"
	AksiBatalSesi  AksiPermission = "batal_sesi"
)

type CakupanPermission string

const (
	CakupanSemua          CakupanPermission = "semua"
	CakupanInstansiSendiri CakupanPermission = "instansi_sendiri"
	CakupanMilikSendiri   CakupanPermission = "milik_sendiri"
)

// ── Entities ──

// Modul maps to mst_modul — one of the 19 (+file) functional areas.
type Modul struct {
	ID            int64      `gorm:"column:id;primaryKey" json:"id"`
	Kode          string     `gorm:"column:kode;uniqueIndex" json:"kode"`
	Nama          string     `gorm:"column:nama" json:"nama"`
	Grup          string     `gorm:"column:grup" json:"grup"`
	Route         string     `gorm:"column:route" json:"route"`
	Icon          string     `gorm:"column:icon" json:"icon"`
	Urutan        int        `gorm:"column:urutan" json:"urutan"`
	TampilDiMenu  bool       `gorm:"column:tampil_di_menu" json:"tampil_di_menu"`
	IsAktif       bool       `gorm:"column:is_aktif" json:"is_aktif"`
	CreatedAt     time.Time  `gorm:"column:created_at" json:"created_at"`
	CreatedBy     *int64     `gorm:"column:created_by" json:"created_by,omitempty"`
	ModifiedAt    *time.Time `gorm:"column:modified_at" json:"modified_at,omitempty"`
	ModifiedBy    *int64     `gorm:"column:modified_by" json:"modified_by,omitempty"`
}

func (Modul) TableName() string { return "mst_modul" }

// Permission maps to mst_permission — a modul × aksi pair.
type Permission struct {
	ID           int64           `gorm:"column:id;primaryKey" json:"id"`
	ModulID      int64           `gorm:"column:modul_id" json:"modul_id"`
	Aksi         AksiPermission  `gorm:"column:aksi" json:"aksi"`
	Kode         string          `gorm:"column:kode;uniqueIndex" json:"kode"`
	Nama         string          `gorm:"column:nama" json:"nama"`
	Keterangan   string          `gorm:"column:keterangan" json:"keterangan,omitempty"`
	IsBerbahaya  bool            `gorm:"column:is_berbahaya" json:"is_berbahaya"`
}

func (Permission) TableName() string { return "mst_permission" }

// Role maps to mst_role — with soft delete and hierarchy level.
type Role struct {
	model.Audit
	ID         int64      `gorm:"column:id;primaryKey" json:"id"`
	Kode       string     `gorm:"column:kode;uniqueIndex" json:"kode"`
	Nama       string     `gorm:"column:nama" json:"nama"`
	Level      int        `gorm:"column:level" json:"level"`
	Keterangan string     `gorm:"column:keterangan" json:"keterangan,omitempty"`
	IsSistem   bool       `gorm:"column:is_sistem" json:"is_sistem"`
	IsSuper    bool       `gorm:"column:is_super" json:"is_super"`
	IsAktif    bool       `gorm:"column:is_aktif" json:"is_aktif"`
}

func (Role) TableName() string { return "mst_role" }

// RolePermission maps to role_permission — the matrix row.
type RolePermission struct {
	ID           int64            `gorm:"column:id;primaryKey" json:"id"`
	RoleID       int64            `gorm:"column:role_id" json:"role_id"`
	PermissionID int64            `gorm:"column:permission_id" json:"permission_id"`
	Cakupan      CakupanPermission `gorm:"column:cakupan" json:"cakupan"`
	CreatedAt    time.Time        `gorm:"column:created_at" json:"created_at"`
	CreatedBy    *int64           `gorm:"column:created_by" json:"created_by,omitempty"`
}

func (RolePermission) TableName() string { return "role_permission" }

// UserRole maps to user_role — user ↔ role mapping.
type UserRole struct {
	ID            int64      `gorm:"column:id;primaryKey" json:"id"`
	UserID        int64      `gorm:"column:user_id" json:"user_id"`
	RoleID        int64      `gorm:"column:role_id" json:"role_id"`
	IsUtama       bool       `gorm:"column:is_utama" json:"is_utama"`
	BerlakuSampai *time.Time `gorm:"column:berlaku_sampai" json:"berlaku_sampai,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at" json:"created_at"`
	CreatedBy     *int64     `gorm:"column:created_by" json:"created_by,omitempty"`
}

func (UserRole) TableName() string { return "user_role" }
