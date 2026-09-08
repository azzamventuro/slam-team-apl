# Modul 03 — Pengaturan Sistem (`mst_pengaturan`) + Audit (`log_aktivitas`)

> nn=`03` · key=`pengaturan-log` · **Fase 0 — Fondasi** · Grup: Sistem
> Sumber otoritatif: `slamteam_db.dbml` (skema) > rancangan Bab 8.5 / Bab 9 / Bab 10.2–10.3 > `_shared/conventions-api.md` §7,§12 > file ini. Bila prosa dan `.dbml` berbeda, **`.dbml` menang**.

---

## 1. Ringkasan & tujuan modul

Modul infrastruktur Fase 0 yang menyediakan **dua kemampuan dasar** yang dipakai hampir semua modul lain:

1. **`mst_pengaturan`** — key-value **bertipe** dengan metadata yang cukup untuk **membangkitkan halaman pengaturannya sendiri**. Menambah setelan baru = satu baris `INSERT`, tanpa menyentuh kode frontend. Semua nilai ambang di sistem (radius absen, akurasi GPS maksimum, dimensi KTA, piksel varian file, masa berlaku KTA, kode wilayah NRA) dibaca dari sini — **tidak pernah di-hardcode**. Ini yang membuat KTA yang tercetak, tanggal kadaluarsa, dan kalimat peraturan di balik kartu berasal dari satu sumber (rancangan Bab 7.8, 8.5).

2. **`log_aktivitas`** — tabel **audit generik** (menggantikan `log_hak_akses` Rev.2). Satu tabel untuk semua tindakan yang mengubah status: perubahan hak akses, perubahan pengaturan, override absensi, pencabutan/pencetakan KTA, login, dan CRUD umum. Menyimpan `nilai_lama`/`nilai_baru` (jsonb), pelaku, dan `ip_address`. Bukan modul CRUD — ia adalah **penulis audit bersama** (shared writer) yang dipanggil modul lain.

Bukan bagian dari 19 `mst_modul` (tidak ada `kode`=`pengaturan` di menu RBAC). Pengaturan adalah **halaman sistem yang di-gate berdasarkan peran** (SA/Admin), bukan lewat izin `modul.aksi` yang di-seed. `log_aktivitas` tidak punya endpoint di Bab 9 sama sekali — ia infrastruktur tulis.

Dependensi: hanya **fondasi** (skeleton Go/Angular, koneksi DB, auth). Dibangun di urutan skema paling awal (rancangan Bab 10.2: "Tabel sistem: `mst_pengaturan`, `log_aktivitas`").

---

## 2. Tabel & kolom

### 2.1 `mst_pengaturan` (Kelompok 2 — Sistem)

| Kolom | Tipe | Aturan |
|---|---|---|
| `id` | `bigint` PK identity | |
| `kunci` | `varchar(100)` | **UNIQUE, NOT NULL**. Format `grup.nama` → `"kta.masa_berlaku_pelajar_bulan"`. Kunci alami untuk seeder idempoten & untuk `PUT`. |
| `grup` | `varchar(50)` | NOT NULL. Salah satu: `umum` / `nra` / `kta` / `absensi` / `file` / `notifikasi`. Dipakai memecah tab halaman pengaturan. |
| `label` | `varchar(150)` | NOT NULL. Teks yang tampil sebagai label kontrol di halaman pengaturan. |
| `nilai` | `text` | **Selalu disimpan sebagai teks**, ditafsirkan menurut `tipe_nilai`. Nullable. |
| `tipe_nilai` | enum `tipe_nilai_pengaturan` | default `string`. Nilai: `string` / `integer` / `boolean` / `json` / `date`. Menentukan kontrol form & cara casting nilai di API. |
| `nilai_bawaan` | `text` | Untuk tombol "kembalikan ke bawaan". |
| `opsi` | `jsonb` | Daftar pilihan bila kontrolnya dropdown (mis. `["png"]`, `["global","per_anggota"]`). Nullable. |
| `satuan` | `varchar(20)` | `meter` / `menit` / `bulan` / `px` / `mm` / `dpi`. Ditampilkan sebagai suffix input. |
| `keterangan` | `text` | Teks bantu di bawah kontrol. |
| `urutan` | `int` | default 0. Urutan tampil dalam grup. |
| `is_publik` | `boolean` | default `false`. `true` = boleh dibaca frontend **tanpa autentikasi** (`GET /public/pengaturan`), mis. nama klub. |
| `is_terkunci` | `boolean` | default `false`. `true` = **hanya super admin** yang boleh mengubah (mis. `nra.mode_penomoran`, ambang akurasi GPS). |
| `modified_at` | `timestamptz` | UTC. |
| `modified_by` | `bigint` | FK → `users.id`. |

**Indeks / constraint** (dari `.dbml`): `kunci` UNIQUE; index `(grup, urutan)`. Relasi: `mst_pengaturan.modified_by > users.id`.
**Catatan:** tabel ini TIDAK punya kolom soft-delete/`created_at` — setelan tidak dihapus, hanya diubah. Ia bukan tabel transaksional.

### 2.2 `log_aktivitas` (Kelompok 2 — Sistem)

| Kolom | Tipe | Aturan |
|---|---|---|
| `id` | `bigint` PK identity | |
| `aktor_user_id` | `bigint` | Nullable — **kosong untuk tindakan sistem/job terjadwal**. FK → `users.id`. |
| `modul` | `varchar(50)` | NOT NULL. `hak_akses` / `pengaturan` / `absensi` / `kta` / `anggota` / … |
| `aksi` | `varchar(50)` | NOT NULL. `buat` / `ubah` / `hapus` / `override` / `cabut` / `cetak` / `login` / … |
| `reff_type` | `varchar(50)` | Entitas sumber (polymorphic), mis. `mst_pengaturan`, `mst_role`. Nullable. |
| `reff_id` | `bigint` | Id baris sumber. Nullable. |
| `ringkasan` | `varchar(255)` | Kalimat siap tampil: `"Mengubah 3 setelan grup kta"`. |
| `nilai_lama` | `jsonb` | Snapshot sebelum perubahan. Nullable. |
| `nilai_baru` | `jsonb` | Snapshot sesudah perubahan. Nullable. |
| `ip_address` | `inet` | Alamat IP pelaku (tipe Postgres `inet`). |
| `user_agent` | `text` | |
| `created_at` | `timestamptz` | default `now()` (UTC). |

**Indeks** (dari `.dbml`): `(modul, created_at)`, `(reff_type, reff_id)`, `(aktor_user_id, created_at)`. Relasi: `log_aktivitas.aktor_user_id > users.id`. Tabel **append-only** — tidak ada update/delete, tidak ada soft-delete.

### 2.3 Enum baru yang dibuat modul ini

`tipe_nilai_pengaturan { string, integer, boolean, json, date }` (`.dbml` baris 149–155). Dibuat di migrasi (atau `0001_enums` bila sudah ada). `inet` & `jsonb` adalah tipe bawaan Postgres.

---

## 3. Endpoint (Bab 9 — grup "Pengaturan")

Semua di bawah `/api/v1`. `log_aktivitas` **tidak punya endpoint** — ia ditulis di dalam service modul lain.

| Method | Path | Gate | Auth | Ringkas |
|---|---|---|---|---|
| GET | `/pengaturan` | SA/Admin | bearer | Semua setelan + metadata (untuk halaman self-generating). |
| GET | `/pengaturan/{grup}` | SA/Admin | bearer | Setelan satu grup (`umum`/`nra`/`kta`/`absensi`/`file`/`notifikasi`). |
| PUT | `/pengaturan` | SA/Admin; baris `is_terkunci` → **SA only** | bearer | Update massal nilai. Menulis `log_aktivitas`. |
| GET | `/public/pengaturan` | — | **public** | Hanya baris `is_publik = true`, nilai sudah di-cast. |

> **Gate Fase 0.** `PermGuard` (RBAC dinamis) baru ada di Fase 1 dan **tidak ada** izin `pengaturan.*` yang di-seed. Karena itu tiga route ber-auth di atas di-gate berbasis **peran**: `is_super` ATAU `role.level <= 10` (Admin). Baris `is_terkunci` hanya bisa diubah `is_super`. Ini konsisten dengan catatan modul: "settings SA/Admin; is_terkunci => SA only; is_publik readable unauth".

### Bentuk data

**`GET /pengaturan`** → `data` = array objek setelan **ter-cast**:

```json
{
  "success": true,
  "message": "OK",
  "data": [
    {
      "id": 8, "kunci": "kta.masa_berlaku_pelajar_bulan", "grup": "kta",
      "label": "Masa berlaku KTA pelajar", "nilai": 12, "tipe_nilai": "integer",
      "nilai_bawaan": 12, "opsi": null, "satuan": "bulan",
      "keterangan": "Dihitung dari tanggal terbit kartu",
      "urutan": 1, "is_publik": false, "is_terkunci": false,
      "modified_at": "2026-09-08T02:11:00Z", "modified_by": 1
    }
  ]
}
```

`nilai` & `nilai_bawaan` dikembalikan **ter-cast** menurut `tipe_nilai` (`integer`→number, `boolean`→bool, `json`→object/array, `date`/`string`→string). Frontend mengelompokkan array ini per `grup` untuk membangun tab & kontrol otomatis.

**`GET /pengaturan/{grup}`** → sama, difilter satu grup (404 bila grup tidak dikenal / kosong).

**`PUT /pengaturan`** — request update massal:

```json
{ "items": [
  { "kunci": "absensi.radius_bawaan_meter", "nilai": 120 },
  { "kunci": "umum.nama_klub", "nilai": "Scouting Legion Airsofter Malang" }
]}
```

`nilai` boleh dikirim sebagai tipe JSON asalnya (number/bool/string/object); backend menormalkan ke `text` untuk disimpan. Response `data` = array setelan terbaru (bentuk sama dengan `GET`). Menolak `kunci` tak dikenal (422) dan menolak baris `is_terkunci` bila pelaku bukan `is_super` (403).

**`GET /public/pengaturan`** → map ringkas `kunci → nilai ter-cast` (hanya `is_publik=true`):

```json
{ "success": true, "message": "OK",
  "data": { "umum.nama_klub": "Scouting Legion Airsofter Malang",
            "umum.singkatan": "SLAM",
            "umum.alamat_sekretariat": "Jalan Laksda Adi Sucipto Nomor 32 Blimbing, Kota Malang",
            "umum.timezone_bawaan": "Asia/Jakarta" } }
```

---

## 4. Hak akses

Modul ini **tidak muncul** di matriks izin `modul.aksi` (Bab 3.5) — tak ada `kode` modul `pengaturan`, jadi tidak ada baris `role_permission` untuk-nya. Aturannya hardcoded sebagai **kebijakan peran halaman sistem**:

| Aksi | Super Admin (level 0, `is_super`) | Admin (level 10) | Moderator (20) | User (30) | Guest (99 / publik) |
|---|---|---|---|---|---|
| Baca semua pengaturan (`GET /pengaturan`, `/{grup}`) | ✔ | ✔ | - | - | - |
| Ubah pengaturan biasa (`PUT`, baris `is_terkunci=false`) | ✔ | ✔ | - | - | - |
| Ubah pengaturan terkunci (`PUT`, baris `is_terkunci=true`) | ✔ | - | - | - | - |
| Baca pengaturan publik (`GET /public/pengaturan`, hanya `is_publik=true`) | ✔ | ✔ | ✔ | ✔ | ✔ (tanpa login) |

Tidak ada dimensi `cakupan` (setelan bersifat global, bukan milik instansi/anggota).

---

## 5. Kebutuhan BACKEND (Go)

Ikuti pola modul `internal/modules/core/<mod>/` (`conventions-api.md` §1). Direktori Go **`pengaturan`** (nama paket tanpa tanda hubung). Penulis audit ditempatkan di **`internal/shared/audit`** karena dipakai lintas modul (dibangun sekali di sini, dikonsumsi semua modul lain — `conventions-api.md` §12).

### 5.1 `domain/`
- `Pengaturan` → `TableName() = "mst_pengaturan"`. Field memetakan semua kolom. `Opsi datatypes.JSON` (gorm datatypes) atau `json.RawMessage`. Tipe `TipeNilai string` + konstanta (`TipeString`, `TipeInteger`, `TipeBoolean`, `TipeJSON`, `TipeDate`).
- `internal/shared/audit`: `LogAktivitas` → `TableName() = "log_aktivitas"`. `NilaiLama`/`NilaiBaru datatypes.JSON`, `IPAddress string` (kolom `inet`; simpan sebagai string, GORM menulis apa adanya), `AktorUserID *int64`.

### 5.2 `dto/`
- `SettingItem` (response) — semua metadata + `Nilai any` & `NilaiBawaan any` (ter-cast).
- `UpdatePengaturanReq` → `{ Items []SettingKV }`, `SettingKV{ Kunci string binding:"required", Nilai any binding:"required" }` (izinkan `false`/`0` — pakai pointer atau validasi manual di service, bukan `required` yang menolak zero value; lihat edge case).
- `PublicSetting` = `map[string]any`.

### 5.3 `repository/`
- `FindAll() []Pengaturan` (order by `grup, urutan`).
- `FindByGrup(grup string) []Pengaturan`.
- `FindPublik() []Pengaturan` (`WHERE is_publik = true`).
- `FindByKunci(kunci string) (Pengaturan, error)` — map `gorm.ErrRecordNotFound` → `ErrNotFound`.
- `UpdateNilai(kunci, nilai string, by int64)` — set `nilai`, `modified_at=now()`, `modified_by`. Lakukan seluruh update `PUT` dalam **satu transaksi**.
- `audit.Repository.Insert(LogAktivitas)` — append.

### 5.4 `service/`
- **Casting**: `castOut(tipe, nilaiText) any` untuk response; `normalizeIn(tipe, any) (string, error)` untuk simpan (integer→`strconv`, boolean→`"true"/"false"`, json→`json.Marshal`, date→validasi `YYYY-MM-DD`, string→apa adanya). Casting gagal → `ErrValidation` dengan pesan field.
- **`Update(ctx, actor, req)`**: untuk tiap item → cari `Pengaturan` by `kunci` (tak ada → 422 `{kunci: "tidak dikenal"}`); bila `is_terkunci` dan `!actor.IsSuper` → `ErrForbidden`; validasi nilai vs `tipe_nilai` (dan vs `opsi` bila ada — nilai harus salah satu opsi); kumpulkan `nilai_lama`. Terapkan semua dalam **satu transaksi**; setelah commit tulis **satu** baris `log_aktivitas` (`modul="pengaturan"`, `aksi="ubah"`, `reff_type="mst_pengaturan"`, `nilai_lama`/`nilai_baru` = map `kunci→nilai`, `ringkasan="Mengubah N setelan"`, `ip_address`+`user_agent` dari request, `aktor_user_id`=actor).
- **`List`/`ListByGrup`/`Publik`**: baca + cast.
- `internal/shared/audit`: `Writer.Log(ctx, entry)` — helper yang dipanggil modul lain; menerima `aktor_user_id` (nil untuk sistem), modul, aksi, reff, nilai lama/baru, dan `*gin.Context`/`http.Request` opsional untuk mengekstrak `ip_address`+`user_agent`. Jangan menggagalkan aksi bisnis hanya karena penulisan audit gagal — log error via Zap dan lanjut.

### 5.5 `handler/`
- Bind req; validasi via `validator.Explain`; panggil service; balas dengan `response.OK/Created`; error via `response.FromError` (tambahkan `Forbidden`/`NotFound` ke paket `response` bila belum ada — `conventions-api.md` §2). Handler membaca `middleware.Claims(c)` untuk `IsSuper`+`RoleLevel`, dan `c.ClientIP()` + `c.Request.UserAgent()` untuk audit.

### 5.6 `main.pengaturan.go` + router
```go
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
    // publik — tanpa JWTAuth
    rg.GET("/public/pengaturan", m.h.Public)
    // ber-auth — SA/Admin
    g := rg.Group("/pengaturan", middleware.JWTAuth(m.jwtMgr))
    g.GET("",        middleware.RequireAdmin(), m.h.List)
    g.GET("/:grup",  middleware.RequireAdmin(), m.h.ListByGrup)
    g.PUT("",        middleware.RequireAdmin(), m.h.Update) // is_terkunci → SA-only dicek di service
}
```
Daftarkan di `internal/router/router.go`: `pengaturan.Initialize(db, jwtMgr, auditWriter).SetupRoutes(apiV1)`. `RequireAdmin()` = helper middleware kecil (is_super OR level<=10) — bila belum ada, buat di `internal/middleware`. Catatan: `PermGuard` Fase 1 tidak dipakai di sini (tak ada izin `pengaturan.*`).

### 5.7 Migrasi & seeder
- **Migrasi** `migrations/0002_pengaturan_log.up.sql` / `.down.sql`: `CREATE TYPE tipe_nilai_pengaturan AS ENUM (...)` (bila `0001_enums` belum membuatnya — pakai guard `DO $$ ... IF NOT EXISTS`), `CREATE TABLE mst_pengaturan`, `CREATE TABLE log_aktivitas`, semua indeks & FK dari `.dbml`. `.down.sql` = `DROP TABLE` (urutan terbalik) + `DROP TYPE`.
- **Seeder** `cmd/slamctl seed pengaturan` — **idempoten**: `INSERT ... ON CONFLICT (kunci) DO UPDATE SET label=EXCLUDED.label, tipe_nilai=EXCLUDED.tipe_nilai, satuan=EXCLUDED.satuan, opsi=EXCLUDED.opsi, keterangan=EXCLUDED.keterangan, urutan=EXCLUDED.urutan, is_publik=EXCLUDED.is_publik, is_terkunci=EXCLUDED.is_terkunci, nilai_bawaan=EXCLUDED.nilai_bawaan` — **JANGAN timpa `nilai`** (pelaku mungkin sudah menyuntingnya); `nilai` hanya diisi saat baris pertama kali dibuat. Jalankan ulang aman.

### 5.8 Seed 24 baris (Bab 8.5 / `.dbml` "SEED AWAL")

> `.dbml` mencantumkan **24 kunci** di blok "SEED AWAL". Prosa Bab 10.3 menyebut "Dua puluh baris" — **`.dbml` menang → 24 baris**. `is_terkunci`/`is_publik` di bawah adalah keputusan seed (bisa diubah runtime oleh SA).

| # | kunci | grup | nilai | tipe_nilai | satuan | is_publik | is_terkunci |
|--:|---|---|---|---|---|:--:|:--:|
| 1 | `umum.nama_klub` | umum | Scouting Legion Airsofter Malang | string | | ✔ | |
| 2 | `umum.singkatan` | umum | SLAM | string | | ✔ | |
| 3 | `umum.alamat_sekretariat` | umum | Jalan Laksda Adi Sucipto Nomor 32 Blimbing, Kota Malang | string | | ✔ | |
| 4 | `umum.timezone_bawaan` | umum | Asia/Jakarta | string | | ✔ | |
| 5 | `umum.kode_wilayah_paten` | umum | 3573 | string | | | ✔ |
| 6 | `nra.mode_penomoran` | nra | global | string (opsi `["global","per_anggota"]`) | | | ✔ |
| 7 | `nra.panjang_urut_minimal` | nra | 3 | integer | digit | | ✔ |
| 8 | `kta.masa_berlaku_pelajar_bulan` | kta | 12 | integer | bulan | | |
| 9 | `kta.masa_berlaku_dewasa_bulan` | kta | 24 | integer | bulan | | |
| 10 | `kta.format_keluaran` | kta | png | string (opsi `["png"]`) | | | |
| 11 | `kta.dpi_cetak` | kta | 300 | integer | dpi | | |
| 12 | `kta.lebar_mm` | kta | 85.6 | string¹ | mm | | |
| 13 | `kta.tinggi_mm` | kta | 54 | integer | mm | | |
| 14 | `kta.bleed_mm` | kta | 2 | integer | mm | | |
| 15 | `kta.qr_ukuran_mm` | kta | 24 | integer | mm | | |
| 16 | `kta.kartu_per_lembar` | kta | 8 | integer | | | |
| 17 | `absensi.akurasi_gps_maks_meter` | absensi | 100 | integer | meter | | ✔ |
| 18 | `absensi.radius_bawaan_meter` | absensi | 100 | integer | meter | | |
| 19 | `absensi.toleransi_telat_menit` | absensi | 15 | integer | menit | | |
| 20 | `absensi.buka_absen_menit` | absensi | 30 | integer | menit | | |
| 21 | `absensi.tutup_absen_menit` | absensi | 60 | integer | menit | | |
| 22 | `file.varian_original_maks_px` | file | 4000 | integer | px | | ✔ |
| 23 | `file.varian_medium_px` | file | 1200 | integer | px | | ✔ |
| 24 | `file.varian_low_px` | file | 400 | integer | px | | ✔ |

¹ `tipe_nilai_pengaturan` tak punya `decimal`; `85.6` disimpan `string` dan di-parse `float64` oleh konsumen KTA. `nilai_bawaan` = sama dengan `nilai` di atas untuk setiap baris.

### 5.9 Edge cases / aturan
- **Zero-value bukan kosong**: `nilai=0`, `nilai=false`, `nilai=""` adalah nilai sah. Jangan pakai `binding:"required"` pada `SettingKV.Nilai` (menolak `0`/`false`). Validasi "hadir" via pointer/`map` presence.
- **`opsi` non-null** → nilai wajib salah satu opsi, else 422.
- **`is_terkunci`** dicek di **service** (bukan hanya middleware) — pelaku Admin yang mengubah baris terkunci → 403, walau route lolos `RequireAdmin`.
- **`PUT` atomik**: semua item lolos validasi atau tidak ada yang tersimpan (satu transaksi). Bila satu `kunci` gagal, batalkan seluruhnya.
- **Audit tak boleh menggagalkan bisnis**: kegagalan tulis `log_aktivitas` di-log Zap, bukan 500.
- **`inet`**: `c.ClientIP()` bisa berupa string kosong di test — simpan `NULL` bila kosong (jangan kirim string kosong ke kolom `inet`).
- **`GET /public/pengaturan`** tak pernah membocorkan baris non-publik walau tanpa filter permission — filter `is_publik=true` di query.

---

## 6. Kebutuhan FRONTEND (Angular)

> **Awali dengan Taste Skill** (umumkan "Using design-taste-frontend"), jalankan pre-flight/audit, petakan ke Bootstrap 5, dan **tegakkan `workflow/_shared/design-tokens.md`** (dark default, SLAM red accent, heading UPPERCASE wide-tracking).
> **Catatan blog-fe:** "Jika sumber `blog-fe` dipulihkan, tiru layout/menu-nya untuk layar ini; jika tidak, ikuti design tokens."

Cakupan frontend fase ini = **satu halaman Pengaturan** (`log_aktivitas` tidak punya endpoint/viewer di Bab 9 → di luar cakupan; ia hanya jalur tulis backend).

### 6.1 Halaman & komponen
- `pages/pengaturan/pengaturan.ts` + `.html` + `.scss` (standalone, signals, zoneless). Rute lazy `loadComponent`, `canActivate: [authGuard]` + guard peran (SA/Admin) — lihat 6.4.
- **Self-generating**: ambil `GET /pengaturan`, kelompokkan per `grup` (computed signal), render **tab per grup** (`umum`/`nra`/`kta`/`absensi`/`file`/`notifikasi`). Untuk tiap setelan render kontrol menurut `tipe_nilai`:
  - `integer` → `<input type="number">` (+ suffix `satuan`)
  - `boolean` → `<input type="checkbox">`/switch
  - `string` dengan `opsi` → `<select>`; tanpa `opsi` → `<input type="text">`
  - `date` → `<input type="date">` (native, bukan lib picker)
  - `json` → `<textarea>` (validasi JSON sebelum submit)
- Baris `is_terkunci` → gembok + `[disabled]` bila user bukan super admin (mirror backend; server tetap penentu).
- `label` sebagai label kontrol; `keterangan` sebagai help text; `nilai_bawaan` → tombol "Kembalikan ke bawaan" per baris.

### 6.2 Form
- Bangun `FormGroup` dinamis dari metadata (`FormBuilder`, satu control per `kunci`). Validator cermin backend: number untuk `integer`, `oneOf(opsi)` untuk dropdown, `Validators.required`? → **jangan** untuk boolean/0. JSON control divalidasi `JSON.parse`.
- Submit → kumpulkan hanya control yang **berubah** (dirty) → `PUT /pengaturan` `{ items:[{kunci,nilai}] }`. Pada `422`/`403`, petakan `res.errors` kembali ke control terkait dan tampilkan toast.

### 6.3 Service API (envelope via `ApiService`)
```ts
@Injectable({ providedIn: 'root' })
export class PengaturanService {
  private api = inject(ApiService);
  list()                        { return this.api.get<Setting[]>('/pengaturan'); }
  byGrup(grup: string)          { return this.api.get<Setting[]>(`/pengaturan/${grup}`); }
  update(items: SettingKV[])    { return this.api.put<Setting[]>('/pengaturan', { items }); }
  publik()                      { return this.api.get<Record<string, unknown>>('/public/pengaturan'); }
}
```
`Setting` (model): `{ id, kunci, grup, label, nilai, tipe_nilai, nilai_bawaan, opsi, satuan, keterangan, urutan, is_publik, is_terkunci, modified_at, modified_by }`. `GET /public/pengaturan` dipakai app shell (nama klub, timezone bawaan) **tanpa token** — panggil sebelum login (bootstrap public config).

### 6.4 Guard & permission-gating
- Tidak ada izin `pengaturan.*`. Gate berbasis peran: `permissionGuard` tidak cocok → buat guard kecil `adminGuard` (izinkan bila `AuthService.user().role.is_super || role.level <= 10`). Sembunyikan menu Pengaturan untuk non-Admin. Menyembunyikan bukan keamanan — backend tetap menolak.
- Tombol Simpan hanya aktif untuk Admin; kontrol `is_terkunci` disabled untuk non-super.

### 6.5 i18n (namespace `PENGATURAN.*`, sinkron IND/ENG)
`PENGATURAN.TITLE`, `PENGATURAN.TAB.UMUM|NRA|KTA|ABSENSI|FILE|NOTIFIKASI`, `PENGATURAN.LOCKED_HINT`, `PENGATURAN.RESET_DEFAULT`, `COMMON.SAVE`, `COMMON.CANCEL`, `VALIDATION.REQUIRED`, `VALIDATION.NUMBER`, `VALIDATION.JSON`. Jangan hardcode string; `label`/`keterangan` datang dari DB (bukan i18n) — tampilkan apa adanya.

### 6.6 File handling
Tidak ada unggah berkas di modul ini.

---

## 7. ALUR form → API → DATABASE (kontrak koherensi)

**Pemetaan field (halaman Pengaturan, satu baris = satu setelan):**

| Kontrol form | Sumber metadata (kolom) | Payload key (`PUT`) | Kolom DB tujuan |
|---|---|---|---|
| kontrol per setelan | `mst_pengaturan.tipe_nilai` (menentukan jenis kontrol) | `items[].kunci` = `mst_pengaturan.kunci` | (identifier baris) |
| nilai kontrol | nilai awal = `mst_pengaturan.nilai` (ter-cast) | `items[].nilai` | `mst_pengaturan.nilai` (dinormalkan ke text) |
| label kontrol | `mst_pengaturan.label` | — | — |
| help text | `mst_pengaturan.keterangan` | — | — |
| suffix satuan | `mst_pengaturan.satuan` | — | — |
| opsi dropdown | `mst_pengaturan.opsi` | — | — |
| gembok/disabled | `mst_pengaturan.is_terkunci` | — | — |
| tombol reset | `mst_pengaturan.nilai_bawaan` | `items[].nilai` = nilai_bawaan | `mst_pengaturan.nilai` |

**Aksi → endpoint → tulisan DB:**

| Aksi user | Endpoint | Tulisan DB |
|---|---|---|
| Buka halaman | `GET /pengaturan` | (baca) `SELECT * FROM mst_pengaturan ORDER BY grup, urutan` |
| Pilih tab grup | `GET /pengaturan/{grup}` | (baca) `WHERE grup = ?` |
| Simpan perubahan | `PUT /pengaturan` | Per item: `UPDATE mst_pengaturan SET nilai=?, modified_at=now(), modified_by=<aktor> WHERE kunci=?` (1 transaksi) **+** 1× `INSERT INTO log_aktivitas (aktor_user_id, modul='pengaturan', aksi='ubah', reff_type='mst_pengaturan', nilai_lama, nilai_baru, ringkasan, ip_address, user_agent, created_at)` |
| App shell / publik | `GET /public/pengaturan` | (baca) `WHERE is_publik = true` |

**Audit lintas modul (dipakai modul lain, disediakan di sini):** setiap aksi state-changing modul lain memanggil `audit.Writer.Log(...)` → `INSERT INTO log_aktivitas (...)` dengan `modul`/`aksi`/`reff_*`/`nilai_lama`/`nilai_baru`/`ip_address`/`user_agent`; `aktor_user_id` NULL untuk job sistem.

---

## 8. Dependencies & Acceptance criteria

**Prasyarat:** fondasi Fase 0 (skeleton Go/Angular, koneksi Postgres `slamteam_db`, runner migrasi, `cmd/slamctl`, modul `auth` yang mengeluarkan claims `role_id`/`is_super`/`level`). Enum `tipe_nilai_pengaturan` (dibuat di sini bila `0001_enums` belum ada). Tidak bergantung pada file layer maupun RBAC Fase 1.

**Acceptance (checkable):**
- [ ] Migrasi `0002_pengaturan_log.up/.down.sql` membuat & merollback `mst_pengaturan` + `log_aktivitas` + enum `tipe_nilai_pengaturan`, dengan seluruh indeks/FK sesuai `.dbml`.
- [ ] `slamctl seed pengaturan` mengisi **24** baris; dijalankan dua kali tidak menggandakan baris dan **tidak menimpa `nilai`** yang sudah disunting.
- [ ] `GET /pengaturan` (SA/Admin) mengembalikan 24 setelan dengan `nilai` **ter-cast** menurut `tipe_nilai`; non-Admin → 403.
- [ ] `GET /pengaturan/{grup}` memfilter per grup; grup tak dikenal → 404.
- [ ] `PUT /pengaturan` memperbarui nilai (atomik), menolak `kunci` tak dikenal (422), menolak nilai di luar `opsi` (422), dan menolak baris `is_terkunci` bila pelaku bukan `is_super` (403).
- [ ] Setiap `PUT` sukses menulis satu baris `log_aktivitas` (`modul=pengaturan`, `aksi=ubah`) berisi `nilai_lama`/`nilai_baru`/`ip_address`/`user_agent`.
- [ ] `GET /public/pengaturan` **tanpa token** hanya mengembalikan baris `is_publik=true` (nama klub, singkatan, alamat, timezone bawaan) dan tak pernah baris lain.
- [ ] Zero-value (`0`/`false`/`""`) tersimpan benar (tidak ditolak sebagai kosong).
- [ ] Halaman Angular Pengaturan membangkitkan kontrol otomatis dari metadata per grup, kontrol `is_terkunci` disabled untuk non-super, submit → `PUT` → baris DB berubah; tokens & taste-skill diterapkan.
- [ ] `audit.Writer` tersedia dan dipanggil dari modul lain (kontrak; kegagalan audit tidak menggagalkan aksi bisnis).
