package main

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func init() {
	seedOrder = append(seedOrder, "instansi")
	seeders["instansi"] = seedInstansi
}

// seedInstansi inserts the initial SLAM Team instansi.
// Keyed on kode; safe to re-run (ON CONFLICT DO NOTHING).
func seedInstansi(_ context.Context, db *gorm.DB) error {
	type row struct {
		Kode string
		Nama string
	}

	seeds := []row{
		{Kode: "SLAM", Nama: "Scouting Legion Airsofter Malang"},
	}

	for _, r := range seeds {
		res := db.Exec(`
			INSERT INTO mst_instansi (kode, nama, status, created_at)
			VALUES (?, ?, 1, now())
			ON CONFLICT (kode) DO NOTHING
		`, r.Kode, r.Nama)
		if res.Error != nil {
			return fmt.Errorf("seed instansi %s: %w", r.Kode, res.Error)
		}
		if res.RowsAffected == 0 {
			fmt.Printf("  %s sudah ada, skip\n", r.Kode)
		}
	}
	return nil
}
