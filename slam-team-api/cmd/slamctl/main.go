// Command slamctl runs the operational tasks that must never be reachable over
// HTTP: schema migrations, idempotent reference-data seeders, and the one-time
// bootstrap of the first super admin.
//
//	slamctl migrate up|down [n]|version
//	slamctl seed [name]
//	slamctl create-superadmin --nama ... --username ... --email ... --tanggal-lahir ...
package main

import (
	"fmt"
	"os"

	// Same reason as cmd/api: IANA zones must resolve without OS tz files.
	_ "time/tzdata"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "migrate":
		err = cmdMigrate(os.Args[2:])
	case "seed":
		err = cmdSeed(os.Args[2:])
	case "create-superadmin":
		err = cmdCreateSuperadmin(os.Args[2:])
	case "help", "-h", "--help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "perintah tidak dikenal: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "slamctl: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `slamctl — perkakas operasional slam-team-api

Penggunaan:
  slamctl migrate up               Jalankan semua migrasi yang tertunda
  slamctl migrate down [n|all]     Rollback n migrasi (bawaan 1), atau semua
  slamctl migrate version          Tampilkan versi skema saat ini
  slamctl seed [nama]              Jalankan seeder referensi (idempoten)
  slamctl create-superadmin        Buat super admin pertama (sekali saja)

Konfigurasi dibaca dari environment (+ .env bila ada), sama seperti cmd/api.
Jalankan sebuah perintah dengan -h untuk melihat flag-nya.
`)
}
