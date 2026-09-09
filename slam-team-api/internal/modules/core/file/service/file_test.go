package service

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
	"time"

	"slam-team-api/internal/modules/core/file/domain"
	"slam-team-api/internal/modules/core/file/dto"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/model"
)

func TestSlugify(t *testing.T) {
	tests := map[string]string{
		"Foto Profil KTP.JPG":   "foto-profil-ktp",
		"SK  Pengurus_2026.pdf": "sk-pengurus-2026",
		"---.png":               "berkas",
		"":                      "berkas",
		"Ünïcødé Fïlé.jpeg":     "n-c-d-f-l",
	}
	for in, want := range tests {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestExtensionOf(t *testing.T) {
	tests := []struct {
		filename, mime, want string
	}{
		{"foto.JPG", "image/jpeg", "jpg"},
		{"dok.pdf", "application/pdf", "pdf"},
		// No usable extension: fall back to what the content says it is.
		{"tanpa-ekstensi", "image/png", "png"},
		{"aneh.thisiswaytoolong", "image/jpeg", "jpg"},
	}
	for _, tc := range tests {
		if got := extensionOf(tc.filename, tc.mime); got != tc.want {
			t.Errorf("extensionOf(%q, %q) = %q, want %q", tc.filename, tc.mime, got, tc.want)
		}
	}
}

// A stored path is handed straight to the OS by the serve handler, so anything
// that climbs out of the storage root must be refused rather than opened.
func TestStorageAbsRefusesEscape(t *testing.T) {
	store, err := NewStorage(t.TempDir())
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}

	for _, bad := range []string{
		"../../etc/passwd",
		"..",
		"foto_profil/../../../secret.txt",
		filepath.Join(string(filepath.Separator), "etc", "passwd"),
		"/etc/passwd",
		`\windows\system32\config`,
		`C:\Windows\System32`,
	} {
		if _, err := store.Abs(bad); err == nil {
			t.Errorf("Abs(%q) diterima, seharusnya ditolak", bad)
		}
	}

	good := "foto_profil/2026/09/abc/original.jpg"
	abs, err := store.Abs(good)
	if err != nil {
		t.Fatalf("Abs(%q): %v", good, err)
	}
	if !bytes.HasPrefix([]byte(abs), []byte(store.Root())) {
		t.Errorf("Abs(%q) = %q, di luar root %q", good, abs, store.Root())
	}
}

// A write must either land completely or report why not — never both "success"
// and a missing file.
func TestStorageWrite(t *testing.T) {
	store, err := NewStorage(t.TempDir())
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}
	relDir := store.BaseDir("banner", "uuid-1", nowFixed())
	if _, err := store.MkdirBase(relDir); err != nil {
		t.Fatalf("MkdirBase: %v", err)
	}

	data := []byte("halo")
	rel, n, err := store.Write(relDir, "original.jpg", data)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != int64(len(data)) {
		t.Errorf("n = %d, want %d", n, len(data))
	}
	if want := "banner/2026/09/uuid-1/original.jpg"; rel != want {
		t.Errorf("rel = %q, want %q", rel, want)
	}
	abs, _ := store.Abs(rel)
	got, err := os.ReadFile(abs)
	if err != nil || !bytes.Equal(got, data) {
		t.Errorf("isi di disk = %q (%v), want %q", got, err, data)
	}
}

func TestDetectMIMEIgnoresFilename(t *testing.T) {
	if got := detectMIME([]byte("%PDF-1.4\nbukan gambar")); got != "application/pdf" {
		t.Errorf("detectMIME(pdf) = %q, want application/pdf", got)
	}
	if got := detectMIME(jpegFixture(t, 8, 8)); got != "image/jpeg" {
		t.Errorf("detectMIME(jpeg) = %q, want image/jpeg", got)
	}
	// The charset parameter must not reach the varchar(100) column.
	if got := detectMIME([]byte("halo dunia, ini teks biasa")); got != "text/plain" {
		t.Errorf("detectMIME(text) = %q, want text/plain (tanpa charset)", got)
	}
}

func TestCheckKategoriMIME(t *testing.T) {
	tests := []struct {
		kategori, mime string
		wantErr        bool
	}{
		{domain.KategoriFotoProfil, "image/jpeg", false},
		{domain.KategoriFotoProfil, "application/pdf", true},
		{domain.KategoriDokumen, "application/pdf", false},
		{domain.KategoriDokumen, "image/png", false},
		{domain.KategoriDokumen, "text/plain", true},
		{domain.KategoriBanner, "application/octet-stream", true},
	}
	for _, tc := range tests {
		err := checkKategoriMIME(tc.kategori, tc.mime)
		if (err != nil) != tc.wantErr {
			t.Errorf("checkKategoriMIME(%q, %q) = %v, wantErr=%v", tc.kategori, tc.mime, err, tc.wantErr)
		}
		if err != nil && !errors.Is(err, apperr.ErrValidation) {
			t.Errorf("checkKategoriMIME(%q, %q) bukan ErrValidation: %v", tc.kategori, tc.mime, err)
		}
	}
}

func TestPublikDefaultsPerKategori(t *testing.T) {
	// Anything carrying a face, an identity document or attendance evidence
	// stays private unless the client says otherwise.
	for _, k := range []string{domain.KategoriFotoProfil, domain.KategoriFotoFormal,
		domain.KategoriAbsensi, domain.KategoriDokumen} {
		if publikFor(dto.UploadReq{}, k) {
			t.Errorf("kategori %q default publik, seharusnya privat", k)
		}
	}
	for _, k := range []string{domain.KategoriBanner, domain.KategoriLogo,
		domain.KategoriFlyer, domain.KategoriKTA} {
		if !publikFor(dto.UploadReq{}, k) {
			t.Errorf("kategori %q default privat, seharusnya publik", k)
		}
	}

	yes, no := true, false
	if !publikFor(dto.UploadReq{IsPublik: &yes}, domain.KategoriAbsensi) {
		t.Error("is_publik eksplisit true diabaikan")
	}
	if publikFor(dto.UploadReq{IsPublik: &no}, domain.KategoriLogo) {
		t.Error("is_publik eksplisit false diabaikan")
	}
}

// Images are never upscaled: an 800 px original stays 800 px as its "medium".
func TestFitDownNeverUpscales(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 800, 400))

	same := fitDown(img, 1200)
	if b := same.Bounds(); b.Dx() != 800 || b.Dy() != 400 {
		t.Errorf("fitDown(800x400, 1200) = %dx%d, want 800x400", b.Dx(), b.Dy())
	}

	smaller := fitDown(img, 400)
	if b := smaller.Bounds(); b.Dx() != 400 || b.Dy() != 200 {
		t.Errorf("fitDown(800x400, 400) = %dx%d, want 400x200", b.Dx(), b.Dy())
	}
}

// A PNG source keeps a lossless variant format: flattening it into JPEG would
// put a black box behind every transparent logo.
func TestEncodeDerivedKeepsTransparency(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	png, err := encodeDerived(img, decodableImages["image/png"], 50, kualitasMedium)
	if err != nil {
		t.Fatalf("encodeDerived(png): %v", err)
	}
	if png.mime != "image/png" {
		t.Errorf("varian dari PNG = %q, want image/png", png.mime)
	}

	jpg, err := encodeDerived(img, decodableImages["image/jpeg"], 50, kualitasMedium)
	if err != nil {
		t.Fatalf("encodeDerived(jpeg): %v", err)
	}
	if jpg.mime != "image/jpeg" {
		t.Errorf("varian dari JPEG = %q, want image/jpeg", jpg.mime)
	}
	if jpg.lebarPx != 50 || jpg.tinggiPx != 50 {
		t.Errorf("dimensi varian = %dx%d, want 50x50", jpg.lebarPx, jpg.tinggiPx)
	}
}

func TestActorReach(t *testing.T) {
	id := int64(7)
	other := int64(8)
	file := domain.File{Audit: model.Audit{CreatedBy: &id}}

	super := Actor{UserID: &other, IsSuper: true}
	admin := Actor{UserID: &other, RoleLevel: 10}
	owner := Actor{UserID: &id, RoleLevel: 30}
	stranger := Actor{UserID: &other, RoleLevel: 30}
	// A token with no role claims decodes to RoleLevel 0; that is "no role",
	// not "super admin".
	roleless := Actor{UserID: &other}

	for name, tc := range map[string]struct {
		a          Actor
		admin, own bool
	}{
		"super":    {super, true, false},
		"admin":    {admin, true, false},
		"owner":    {owner, false, true},
		"stranger": {stranger, false, false},
		"roleless": {roleless, false, false},
	} {
		if got := tc.a.admin(); got != tc.admin {
			t.Errorf("%s.admin() = %v, want %v", name, got, tc.admin)
		}
		if got := tc.a.owns(file); got != tc.own {
			t.Errorf("%s.owns() = %v, want %v", name, got, tc.own)
		}
	}
}

func TestAuthorizeRead(t *testing.T) {
	id := int64(7)
	other := int64(8)
	svc := &FileService{}

	privat := domain.File{Audit: model.Audit{CreatedBy: &id}}
	publik := domain.File{IsPublik: true, Audit: model.Audit{CreatedBy: &id}}

	if err := svc.authorizeRead(Actor{}, publik); err != nil {
		t.Errorf("berkas publik ditolak untuk anonim: %v", err)
	}
	if err := svc.authorizeRead(Actor{}, privat); !errors.Is(err, apperr.ErrUnauthorized) {
		t.Errorf("berkas privat + anonim = %v, want ErrUnauthorized", err)
	}
	if err := svc.authorizeRead(Actor{UserID: &other, RoleLevel: 30}, privat); !errors.Is(err, apperr.ErrForbidden) {
		t.Errorf("berkas privat + pengguna lain = %v, want ErrForbidden", err)
	}
	if err := svc.authorizeRead(Actor{UserID: &id, RoleLevel: 30}, privat); err != nil {
		t.Errorf("pemilik ditolak: %v", err)
	}
	if err := svc.authorizeRead(Actor{UserID: &other, RoleLevel: 10}, privat); err != nil {
		t.Errorf("admin ditolak: %v", err)
	}
}

func TestVarianValid(t *testing.T) {
	for _, v := range []domain.Varian{domain.VarianOriginal, domain.VarianMedium, domain.VarianLow} {
		if !v.Valid() {
			t.Errorf("%q seharusnya valid", v)
		}
	}
	for _, v := range []domain.Varian{"", "tiny", "../original", "ORIGINAL"} {
		if domain.Varian(v).Valid() {
			t.Errorf("%q seharusnya ditolak", v)
		}
	}
}

func TestServeURL(t *testing.T) {
	if got := dto.ServeURL("abc", domain.VarianMedium); got != "/api/v1/files/abc/medium" {
		t.Errorf("ServeURL = %q", got)
	}
}

// --- helpers ---------------------------------------------------------------

func jpegFixture(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 0, 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	return buf.Bytes()
}

// nowFixed pins the year/month partition so the expected path is stable.
func nowFixed() time.Time {
	return time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
}
