package main

import (
	"fmt"
	"strconv"

	"slam-team-api/internal/migrator"

	"github.com/golang-migrate/migrate/v4"
)

func cmdMigrate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("penggunaan: slamctl migrate up|down [n|all]|version")
	}

	_, db, err := openDB()
	if err != nil {
		return err
	}
	defer closeDB(db)

	raw, err := sqlDB(db)
	if err != nil {
		return err
	}
	m, err := migrator.New(raw)
	if err != nil {
		return err
	}

	switch args[0] {
	case "up":
		if err := migrator.Up(m); err != nil {
			return fmt.Errorf("migrate up: %w", err)
		}
	case "down":
		// Default to a single step: rolling everything back must be explicit.
		n := 1
		if len(args) > 1 {
			if args[1] == "all" {
				n = 0
			} else {
				n, err = strconv.Atoi(args[1])
				if err != nil || n < 1 {
					return fmt.Errorf("jumlah langkah tidak valid: %q (pakai angka ≥ 1 atau \"all\")", args[1])
				}
			}
		}
		if err := migrator.Down(m, n); err != nil {
			return fmt.Errorf("migrate down: %w", err)
		}
	case "version":
		// handled below by the shared report
	default:
		return fmt.Errorf("sub-perintah migrate tidak dikenal: %q (up|down|version)", args[0])
	}

	return reportVersion(m)
}

func reportVersion(m *migrate.Migrate) error {
	version, dirty, applied, err := migrator.Version(m)
	if err != nil {
		return fmt.Errorf("baca versi skema: %w", err)
	}
	if !applied {
		fmt.Println("skema: belum ada migrasi yang diterapkan")
		return nil
	}
	if dirty {
		return fmt.Errorf("skema versi %d berstatus DIRTY — migrasi sebelumnya gagal "+
			"di tengah jalan; perbaiki manual lalu set ulang versinya", version)
	}
	fmt.Printf("skema: versi %d (bersih)\n", version)
	return nil
}
