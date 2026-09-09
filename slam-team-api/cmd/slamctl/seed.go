package main

import (
	"context"
	"fmt"
	"sort"

	"gorm.io/gorm"
)

// seeder fills reference data. Every seeder MUST be idempotent — running it
// twice never duplicates rows (INSERT ... ON CONFLICT (kode) DO NOTHING/UPDATE
// keyed on the natural unique column).
type seeder func(ctx context.Context, db *gorm.DB) error

// seedOrder lists the seeders in dependency order; seeders holds the
// implementations. Reference data belongs to the module that owns its table
// (mst_pengaturan, RBAC matrix, mst_wilayah, mst_instansi, …), so each of those
// modules registers its seeder here as it lands.
var (
	seedOrder []string
	seeders   = map[string]seeder{}
)

func cmdSeed(args []string) error {
	if len(seeders) == 0 {
		return fmt.Errorf("belum ada seeder terdaftar — seeder data referensi " +
			"(mst_pengaturan, RBAC, mst_wilayah, mst_instansi) didaftarkan oleh " +
			"modul pemilik tabelnya di cmd/slamctl/seed.go")
	}

	names := seedOrder
	if len(args) > 0 {
		if _, ok := seeders[args[0]]; !ok {
			return fmt.Errorf("seeder %q tidak dikenal — tersedia: %v", args[0], available())
		}
		names = []string{args[0]}
	}

	_, db, err := openDB()
	if err != nil {
		return err
	}
	defer closeDB(db)

	ctx := context.Background()
	for _, name := range names {
		fmt.Printf("seed %s … ", name)
		if err := seeders[name](ctx, db); err != nil {
			fmt.Println("gagal")
			return fmt.Errorf("seeder %s: %w", name, err)
		}
		fmt.Println("ok")
	}
	return nil
}

func available() []string {
	names := make([]string, 0, len(seeders))
	for n := range seeders {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
