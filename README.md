# slam-team-apl

Monorepo aplikasi **SLAM Team** (Scouting Legion Airsofter Malang) — sistem
keanggotaan, penjadwalan, absensi, dan konten publik klub.

| Folder | Isi | Stack |
| --- | --- | --- |
| [`slam-team-api/`](slam-team-api/) | REST API | Go 1.24 · Gin · GORM/PostgreSQL · JWT HS256 · Zap |
| [`slam-team-app/`](slam-team-app/) | Frontend PWA | Angular 21 · standalone + signals · zoneless · Bootstrap 5 · ngx-translate |
| [`workflow/`](workflow/) | Build playbook — dokumen rencana & prompt per modul | Markdown |

## Dokumen

| Berkas | Fungsi |
| --- | --- |
| [`ARCHITECTURE.md`](ARCHITECTURE.md) | Standar arsitektur lintas kedua project — pola modul backend, pola fitur frontend, cara menambah fitur end-to-end. |
| [`slamteam_db.dbml`](slamteam_db.dbml) | **Sumber kebenaran skema database.** Kalau prosa manapun bertentangan dengan berkas ini, `.dbml` yang menang. |
| `Rancangan_SLAM_Team.docx` | Dokumen rancangan asli (routes, payload, matriks permission). |
| [`workflow/README.md`](workflow/README.md) | Cara menjalankan playbook: urutan fase, cara eksekusi satu modul. |
| [`workflow/00-general-tasks.md`](workflow/00-general-tasks.md) | Konvensi global (UTC/timestamptz, soft-delete, file layer, RBAC, response envelope). Baca sebelum modul manapun. |
| [`workflow/99-alignment-report.md`](workflow/99-alignment-report.md) | Catatan drift antara rencana dan realita build. |

Struktur `workflow/`: `modules/<n>.md` adalah lapisan **pemahaman** (apa modul itu,
tabel, route, permission, edge case), `prompts/<n>-api.md` dan `prompts/<n>-app.md`
adalah prompt **eksekusi** yang dijalankan per modul, `_shared/` berisi referensi yang
ditunjuk semua prompt (design tokens, matriks permission, peta skema, glossary).

## Cara menghubungkan keduanya

```
slam-team-app  ──HTTP──▶  slam-team-api
 (Angular :4200)          (Gin :8080, route di bawah /api/v1)
```

`POST /api/v1/auth/login` mengembalikan JWT. App menyimpannya di `localStorage`
(`slam_token`) dan mengirim `Authorization: Bearer <jwt>` lewat `authInterceptor`;
API memvalidasinya di `middleware.JWTAuth`.

## Menjalankan secara lokal

Prasyarat: Go 1.24+, Node 24, PostgreSQL 16 (atau Docker), Redis (opsional).

```bash
# API — butuh PostgreSQL, atau pakai docker-compose
cd slam-team-api
cp .env.example .env      # isi DB_PASSWORD dan JWT_SECRET
make run                  # atau: make dev (hot reload via Air)

# APP
cd slam-team-app
npm install --legacy-peer-deps   # flag ini WAJIB, lihat slam-team-app/CLAUDE.md
npm run serve:local              # http://localhost:4200 → API di :8080
```

Alternatif menjalankan API + PostgreSQL + Redis sekaligus:

```bash
cd slam-team-api && docker compose up
```

> `docker-compose.yml` hanya untuk pengembangan lokal. Kredensial di dalamnya
> bukan untuk produksi — set ulang lewat environment variable sebelum deploy.

## Catatan

- **Tidak ada `AutoMigrate`.** Skema dibuat manual mengikuti `slamteam_db.dbml`.
- `JWT_SECRET` harus stabil antar restart, atau token yang sudah terbit berhenti valid.
- Redis opsional — `REDIS_ADDR` kosong akan menonaktifkannya tanpa error.
