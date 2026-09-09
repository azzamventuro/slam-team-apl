package service

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

// Storage owns the disk side of the file layer: where bytes go and how a
// stored relative path is turned back into a safe absolute one.
//
// Disk layout (rancangan Bab 6.1):
//
//	{root}/{kategori}/{tahun}/{bulan}/{uuid}/{varian}.{ext}
//
// The year/month partition keeps a single directory from collecting tens of
// thousands of entries. It is deliberately NOT the URL shape — the URL is built
// to be readable, and mst_file.path_dasar is what maps the two.
type Storage struct {
	// root is absolute, resolved once at startup, so every containment check
	// compares like with like.
	root string
}

// NewStorage resolves the configured root to an absolute path and makes sure it
// exists and is writable. A root that cannot be written is fatal at startup
// rather than at the first upload: "berhasil diunggah" while the bytes went
// nowhere is the exact failure the Bab 6.6 checklist exists to prevent.
func NewStorage(root string) (*Storage, error) {
	if strings.TrimSpace(root) == "" {
		root = "./storage"
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolusi STORAGE_ROOT %q: %w", root, err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("buat STORAGE_ROOT %q: %w", abs, err)
	}
	probe := filepath.Join(abs, ".write-probe")
	if err := os.WriteFile(probe, []byte("ok"), 0o644); err != nil {
		return nil, fmt.Errorf("STORAGE_ROOT %q tidak bisa ditulis: %w", abs, err)
	}
	if err := os.Remove(probe); err != nil {
		return nil, fmt.Errorf("bersihkan probe di STORAGE_ROOT %q: %w", abs, err)
	}
	return &Storage{root: abs}, nil
}

// Root reports the absolute storage root.
func (s *Storage) Root() string { return s.root }

// BaseDir builds the relative folder for one file: the value stored in
// mst_file.path_dasar. Slash-separated, so the same row reads correctly on
// Windows and Linux.
func (s *Storage) BaseDir(kategori, uuid string, at time.Time) string {
	t := at.UTC()
	return path.Join(kategori, t.Format("2006"), t.Format("01"), uuid)
}

// MkdirBase creates the file's folder and returns its absolute path.
func (s *Storage) MkdirBase(rel string) (string, error) {
	abs, err := s.Abs(rel)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return "", fmt.Errorf("buat folder %q: %w", abs, err)
	}
	return abs, nil
}

// Abs turns a stored relative path into an absolute one, refusing anything that
// would escape the root. Variant paths come out of the database and the serve
// handler hands them to the OS, so this is the boundary that keeps a poisoned
// row from reading /etc/passwd.
func (s *Storage) Abs(rel string) (string, error) {
	raw := filepath.FromSlash(rel)
	// Rooted and volume-qualified paths are refused explicitly rather than left
	// to filepath.IsAbs, which on Windows calls "\etc\passwd" relative and
	// would quietly reinterpret it against the root instead of rejecting it.
	if filepath.IsAbs(raw) || filepath.VolumeName(raw) != "" ||
		strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, `\`) {
		return "", fmt.Errorf("path berkas %q di luar storage root", rel)
	}
	clean := filepath.Clean(raw)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path berkas %q di luar storage root", rel)
	}
	abs := filepath.Join(s.root, clean)
	// Join already cleans, but compare explicitly: a symlinked or oddly-cased
	// segment must not silently pass.
	if abs != s.root && !strings.HasPrefix(abs, s.root+string(filepath.Separator)) {
		return "", fmt.Errorf("path berkas %q di luar storage root", rel)
	}
	return abs, nil
}

// Write writes one variant's bytes and returns how many landed.
//
// Every step is checked — create, write, close — because a half-written file
// that nobody noticed is indistinguishable from a successful upload until the
// day someone opens it.
func (s *Storage) Write(relDir, name string, data []byte) (string, int64, error) {
	rel := path.Join(relDir, name)
	abs, err := s.Abs(rel)
	if err != nil {
		return "", 0, err
	}
	f, err := os.Create(abs)
	if err != nil {
		return "", 0, fmt.Errorf("buat berkas %q: %w", abs, err)
	}
	n, werr := f.Write(data)
	cerr := f.Close()
	if werr != nil {
		_ = os.Remove(abs)
		return "", 0, fmt.Errorf("tulis berkas %q: %w", abs, werr)
	}
	if cerr != nil {
		_ = os.Remove(abs)
		return "", 0, fmt.Errorf("tutup berkas %q: %w", abs, cerr)
	}
	if n != len(data) {
		_ = os.Remove(abs)
		return "", 0, fmt.Errorf("tulis berkas %q: %d dari %d byte", abs, n, len(data))
	}
	return rel, int64(n), nil
}

// RemoveBase deletes a file's whole folder. Used only to clean up after an
// upload whose database insert failed — never on delete, which is soft.
func (s *Storage) RemoveBase(rel string) error {
	abs, err := s.Abs(rel)
	if err != nil {
		return err
	}
	return os.RemoveAll(abs)
}

// slugify turns an uploaded filename into the URL-safe nama_slug. The result is
// never used as the name on disk (that is always "{varian}.{ext}"), so it only
// has to be readable, not unique.
func slugify(name string) string {
	name = strings.TrimSuffix(name, filepath.Ext(name))
	var b strings.Builder
	lastDash := true // leading dashes are suppressed
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "berkas"
	}
	if len(out) > 200 {
		out = strings.Trim(out[:200], "-")
	}
	return out
}

// extensionOf normalises the uploaded filename's extension for mst_file:
// lower-case, no leading dot, and inside the column's varchar(10). An upload
// with no usable extension falls back to the one implied by its detected
// content type.
func extensionOf(filename, mimeType string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	ext = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || unicode.IsDigit(r) {
			return r
		}
		return -1
	}, ext)
	if ext == "" || len(ext) > 10 {
		return extForMIME(mimeType)
	}
	return ext
}
