// Package service holds the file layer's business rules: content-based type
// detection, hashing and dedup, the EXIF-aware resize pipeline, the disk
// layout, and who may read or delete a stored file.
//
// It is the only place in the system that writes uploaded bytes. Every other
// module calls it, stores the returned mst_file.id in its own `*_file_id`
// column, and never learns a path.
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"slam-team-api/internal/modules/core/file/domain"
	"slam-team-api/internal/modules/core/file/dto"
	"slam-team-api/internal/modules/core/file/repository"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Quality per variant tier (rancangan Bab 6.2). Unlike the pixel sizes these
// are not settings: they are the encoder's trade-off, not an operator's choice.
const (
	kualitasOriginal = 90
	kualitasMedium   = 80
	kualitasLow      = 70
)

// The mst_pengaturan keys the variant sizes come from. Reading them (rather
// than hardcoding 4000/1200/400) is what lets an operator change the ceiling
// from the settings page and have the next upload follow.
const (
	kunciOriginalPx = "file.varian_original_maks_px"
	kunciMediumPx   = "file.varian_medium_px"
	kunciLowPx      = "file.varian_low_px"
)

// Fallbacks used only when the setting row is missing or unreadable — an
// upload must not fail because a seeder has not run.
const (
	defaultOriginalPx = 4000
	defaultMediumPx   = 1200
	defaultLowPx      = 400
)

// levelAdmin mirrors middleware.LevelAdmin (mst_role.level for Admin). It is
// duplicated rather than imported so the service stays free of HTTP-layer
// packages; roles are ordered by privilege with 0 the highest.
const levelAdmin = 10

// SettingsReader is the slice of mst_pengaturan this module needs: the variant
// pixel sizes. It is an interface so the file service does not depend on the
// pengaturan module's concrete repository — main.file.go supplies the adapter.
type SettingsReader interface {
	// Int returns the setting as an integer, or def when the key is missing,
	// empty or unparsable. It never fails an upload.
	Int(ctx context.Context, kunci string, def int) int
}

// Actor is who is performing the action, plus the request fingerprint the audit
// row records. UserID is nil only for system/CLI callers; every HTTP route in
// this module that writes runs behind JWTAuth.
type Actor struct {
	UserID    *int64
	IsSuper   bool
	RoleLevel int
	IPAddress string
	UserAgent string
}

// admin reports whether the actor may act on any file rather than only their
// own — the `semua` cakupan of file.delete, and the equivalent read reach.
//
// Level 0 alone is not "better than Admin": a token carrying no role claims
// decodes to RoleLevel 0, and that must be denied, not mistaken for the super
// admin (who is identified by IsSuper). Same rule as middleware.RequireAdmin.
func (a Actor) admin() bool {
	return a.IsSuper || (a.RoleLevel > 0 && a.RoleLevel <= levelAdmin)
}

// authenticated reports whether the request carried verified claims.
func (a Actor) authenticated() bool { return a.UserID != nil }

// owns reports whether the actor created the file.
func (a Actor) owns(f domain.File) bool {
	return a.UserID != nil && f.CreatedBy != nil && *a.UserID == *f.CreatedBy
}

// FileService is the upload/serve/delete pipeline.
type FileService struct {
	repo     *repository.FileRepository
	store    *Storage
	settings SettingsReader
	audit    *audit.Writer
	// maxBytes is the accepted upload size, from FILE_MAX_UPLOAD_MB. The
	// handler caps the request body with the same number, and the service
	// re-checks the assembled part — a non-HTTP caller has no handler in front.
	maxBytes int64
}

// NewFileService composes the service. audit may be nil (tests); the writer
// tolerates it.
func NewFileService(repo *repository.FileRepository, store *Storage, settings SettingsReader, w *audit.Writer, maxUploadMB int) *FileService {
	if maxUploadMB <= 0 {
		maxUploadMB = 15
	}
	return &FileService{
		repo:     repo,
		store:    store,
		settings: settings,
		audit:    w,
		maxBytes: int64(maxUploadMB) << 20,
	}
}

// MaxBytes reports the accepted upload size in bytes, so the handler can cap
// the request body with the same number the service validates against.
func (s *FileService) MaxBytes() int64 { return s.maxBytes }

// Upload runs the whole pipeline for one multipart part:
//
//  1. read the bytes (bounded), detect the type from the CONTENT, hash them;
//  2. return an existing row when the same uploader already stored these exact
//     bytes in this kategori;
//  3. for an image: decode with EXIF auto-orientation, cap the longest side at
//     the configured maximum, write `original`;
//     for anything else: write the bytes as they arrived;
//  4. insert mst_file + the `original` variant in one transaction;
//  5. for an image: generate `medium` and `low` asynchronously, then move
//     status_proses to selesai (or gagal).
//
// The response therefore always carries a readable `original`; the two derived
// sizes follow within seconds.
func (s *FileService) Upload(ctx context.Context, actor Actor, req dto.UploadReq, fh *multipart.FileHeader) (dto.UploadResp, error) {
	kategori := req.Kategori
	if !domain.KategoriValid(kategori) {
		return dto.UploadResp{}, apperr.InvalidField("kategori", "kategori berkas tidak dikenal")
	}
	if fh == nil {
		return dto.UploadResp{}, apperr.InvalidField("file", "berkas wajib diunggah")
	}

	data, err := s.readPart(fh)
	if err != nil {
		return dto.UploadResp{}, err
	}

	// Type comes from the first bytes, never from the filename: an .exe renamed
	// to .jpg must not become a "photo".
	mimeType := detectMIME(data)
	if err := checkKategoriMIME(kategori, mimeType); err != nil {
		return dto.UploadResp{}, err
	}

	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])

	// Dedup: identical bytes, same kategori, same uploader. Scoped per uploader
	// on purpose — a shared row would let one person's delete take away
	// somebody else's file.
	// ponytail: cross-user dedup (shared bytes, refcounted rows) deferred.
	if existing, ok, err := s.repo.ExistsByHash(ctx, hash, kategori, actor.UserID); err != nil {
		return dto.UploadResp{}, err
	} else if ok {
		return s.respond(ctx, existing)
	}

	now := time.Now().UTC()
	fileUUID := uuid.NewString()
	relDir := s.store.BaseDir(kategori, fileUUID, now)

	f := domain.File{
		UUID:          fileUUID,
		NamaAsli:      truncate(fh.Filename, 255),
		NamaSlug:      slugify(fh.Filename),
		Ekstensi:      extensionOf(fh.Filename, mimeType),
		MimeType:      mimeType,
		UkuranByte:    int64(len(data)),
		HashSHA256:    hash,
		Kategori:      kategori,
		StorageDriver: "local",
		PathDasar:     relDir,
		IsPublik:      publikFor(req, kategori),
		StatusProses:  domain.StatusSelesai,
	}
	if owner := req.Owner(); owner != "" {
		f.ReffType = &owner
	}
	f.ReffID = req.ReffID
	f.CreatedBy = actor.UserID

	// Decode BEFORE creating any directory: a corrupt image must leave neither
	// a row nor an empty folder behind.
	var (
		img    image.Image
		format mimeFormat
	)
	isImage := IsImageMIME(mimeType)
	if isImage {
		var ok bool
		if format, ok = imageFormat(mimeType); !ok {
			return dto.UploadResp{}, apperr.InvalidField("file",
				fmt.Sprintf("tipe gambar %s belum didukung", mimeType))
		}
		if img, err = decodeOriented(data); err != nil {
			return dto.UploadResp{}, apperr.InvalidField("file", "gambar tidak bisa dibaca atau rusak")
		}
		f.MetadataEXIF = readEXIF(data)
		// Derived variants are still missing at this point; the async pass
		// moves the row to selesai (or gagal) when it is done.
		f.StatusProses = domain.StatusMenunggu
	}

	if _, err := s.store.MkdirBase(relDir); err != nil {
		return dto.UploadResp{}, err
	}

	original, err := s.writeOriginal(ctx, relDir, data, img, format, isImage, f)
	if err != nil {
		s.cleanup(relDir)
		return dto.UploadResp{}, err
	}
	if isImage {
		f.LebarPx = original.LebarPx
		f.TinggiPx = original.TinggiPx
	}

	if err := s.repo.Create(ctx, &f, []domain.FileVarian{original}); err != nil {
		// The bytes are already on disk; without a row nothing can ever reach
		// them, so take the folder back rather than leak it.
		s.cleanup(relDir)
		return dto.UploadResp{}, err
	}

	// The response is assembled BEFORE the variant pass is started, so what it
	// reports is always self-consistent: status_proses "menunggu" alongside the
	// one variant that exists. Reading it afterwards would race a fast resize
	// and could answer "menunggu" next to three finished variants.
	out, err := s.respond(ctx, f)
	if err != nil {
		return dto.UploadResp{}, err
	}

	if isImage {
		s.startVariants(ctx, f, img, format, relDir)
	}

	s.audit.Log(ctx, audit.Entry{
		AktorUserID: actor.UserID,
		Modul:       "file",
		Aksi:        "buat",
		ReffType:    "mst_file",
		ReffID:      &f.ID,
		Ringkasan:   fmt.Sprintf("Mengunggah berkas %s (%s)", f.NamaAsli, f.Kategori),
		NilaiBaru: map[string]any{
			"uuid": f.UUID, "kategori": f.Kategori, "mime_type": f.MimeType,
			"ukuran_byte": f.UkuranByte, "is_publik": f.IsPublik,
		},
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	})

	return out, nil
}

// readPart streams the multipart part into memory, refusing anything over the
// configured limit. The limit reader stops one byte past it, so a part header
// that under-reported its size is caught too.
//
// Buffering is a deliberate ceiling: the bytes are needed three times over
// (hash, type detection, decode), and FILE_MAX_UPLOAD_MB keeps the cost bounded.
// ponytail: spool to a temp file if that limit is ever raised past tens of MB.
func (s *FileService) readPart(fh *multipart.FileHeader) ([]byte, error) {
	tooBig := apperr.InvalidField("file",
		fmt.Sprintf("ukuran berkas melebihi batas %d MB", s.maxBytes>>20))

	if fh.Size > s.maxBytes {
		return nil, tooBig
	}
	src, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("buka bagian multipart: %w", err)
	}
	defer src.Close()

	data, err := io.ReadAll(io.LimitReader(src, s.maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("baca bagian multipart: %w", err)
	}
	if int64(len(data)) > s.maxBytes {
		return nil, tooBig
	}
	if len(data) == 0 {
		return nil, apperr.InvalidField("file", "berkas kosong")
	}
	return data, nil
}

// writeOriginal renders and writes the `original` variant and returns its row.
// Non-images are written byte-for-byte: a PDF is evidence, not something to
// re-encode.
func (s *FileService) writeOriginal(ctx context.Context, relDir string, data []byte, img image.Image, format mimeFormat, isImage bool, f domain.File) (domain.FileVarian, error) {
	if !isImage {
		// The name on disk comes from the DETECTED type, never from the
		// uploaded filename: mst_file.ekstensi keeps what the user called it,
		// but nothing lets a client choose the suffix of a file we write.
		rel, n, err := s.store.Write(relDir, "original."+extForMIME(f.MimeType), data)
		if err != nil {
			return domain.FileVarian{}, err
		}
		return domain.FileVarian{
			Varian:     domain.VarianOriginal,
			Path:       rel,
			URLPublik:  publikURL(f, domain.VarianOriginal),
			UkuranByte: n,
			MimeType:   f.MimeType,
		}, nil
	}

	maxPx := s.settings.Int(ctx, kunciOriginalPx, defaultOriginalPx)
	enc, err := encodeOriginal(img, format, maxPx, kualitasOriginal)
	if err != nil {
		return domain.FileVarian{}, err
	}
	rel, n, err := s.store.Write(relDir, "original."+enc.ext, enc.data)
	if err != nil {
		return domain.FileVarian{}, err
	}
	q := kualitasOriginal
	return domain.FileVarian{
		Varian:     domain.VarianOriginal,
		Path:       rel,
		URLPublik:  publikURL(f, domain.VarianOriginal),
		LebarPx:    &enc.lebarPx,
		TinggiPx:   &enc.tinggiPx,
		UkuranByte: n,
		MimeType:   enc.mime,
		Kualitas:   &q,
	}, nil
}

// startVariants kicks off the medium/low pass.
//
// It runs after the response so a client is not made to wait for two resizes
// (rancangan Bab 6.2). The context is detached from the request — cancelling
// the HTTP call must not abandon a half-written variant set — and the goroutine
// recovers, because a panic inside an image decoder would otherwise take the
// whole process down long after the upload "succeeded".
//
// ponytail: one goroutine per upload; move to a worker pool or a queue if
// upload throughput ever makes that a problem.
func (s *FileService) startVariants(ctx context.Context, f domain.File, img image.Image, format mimeFormat, relDir string) {
	bg := context.WithoutCancel(ctx)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("file: panic saat membuat varian",
					zap.String("uuid", f.UUID), zap.Any("panic", r))
				s.markGagal(bg, f)
			}
		}()
		if err := s.generateVariants(bg, f, img, format, relDir); err != nil {
			logger.Error("file: gagal membuat varian",
				zap.String("uuid", f.UUID), zap.Error(err))
			s.markGagal(bg, f)
			return
		}
		if err := s.repo.UpdateStatus(bg, f.ID, domain.StatusSelesai); err != nil {
			logger.Error("file: gagal menandai status selesai",
				zap.String("uuid", f.UUID), zap.Error(err))
		}
	}()
}

// generateVariants writes `medium` and `low`. Both read their pixel ceiling
// from mst_pengaturan and neither ever upscales, so an 800 px original yields
// an 800 px medium and a 400 px low.
func (s *FileService) generateVariants(ctx context.Context, f domain.File, img image.Image, format mimeFormat, relDir string) error {
	tiers := []struct {
		varian   domain.Varian
		maxPx    int
		kualitas int
	}{
		{domain.VarianMedium, s.settings.Int(ctx, kunciMediumPx, defaultMediumPx), kualitasMedium},
		{domain.VarianLow, s.settings.Int(ctx, kunciLowPx, defaultLowPx), kualitasLow},
	}

	for _, t := range tiers {
		enc, err := encodeDerived(img, format, t.maxPx, t.kualitas)
		if err != nil {
			return fmt.Errorf("varian %s: %w", t.varian, err)
		}
		rel, n, err := s.store.Write(relDir, string(t.varian)+"."+enc.ext, enc.data)
		if err != nil {
			return fmt.Errorf("varian %s: %w", t.varian, err)
		}
		q := t.kualitas
		row := domain.FileVarian{
			FileID:     f.ID,
			Varian:     t.varian,
			Path:       rel,
			URLPublik:  publikURL(f, t.varian),
			LebarPx:    &enc.lebarPx,
			TinggiPx:   &enc.tinggiPx,
			UkuranByte: n,
			MimeType:   enc.mime,
			Kualitas:   &q,
		}
		if err := s.repo.CreateVarian(ctx, &row); err != nil {
			return fmt.Errorf("varian %s: %w", t.varian, err)
		}
	}
	return nil
}

// markGagal records a failed variant pass. A file stuck at 'menunggu' would
// look like work still in flight; 'gagal' is a state somebody can act on.
func (s *FileService) markGagal(ctx context.Context, f domain.File) {
	if err := s.repo.UpdateStatus(ctx, f.ID, domain.StatusGagal); err != nil {
		logger.Error("file: gagal menandai status gagal",
			zap.String("uuid", f.UUID), zap.Error(err))
	}
}

// cleanup drops a half-built folder, reporting rather than masking a failure —
// leftover bytes nothing points at are a disk leak worth seeing in the log.
func (s *FileService) cleanup(relDir string) {
	if err := s.store.RemoveBase(relDir); err != nil {
		logger.Error("file: gagal membersihkan folder unggahan",
			zap.String("path", relDir), zap.Error(err))
	}
}

// Detail returns one file's metadata and variants, subject to the same read
// rule as the byte stream.
func (s *FileService) Detail(ctx context.Context, actor Actor, fileUUID string) (dto.UploadResp, error) {
	f, err := s.repo.FindByUUID(ctx, fileUUID)
	if err != nil {
		return dto.UploadResp{}, err
	}
	if err := s.authorizeRead(actor, f); err != nil {
		return dto.UploadResp{}, err
	}
	return s.respond(ctx, f)
}

// Served is everything the handler needs to stream one variant.
type Served struct {
	AbsPath  string
	MimeType string
	// NamaAsli drives the download filename; the name on disk is meaningless.
	NamaAsli string
	IsPublik bool
	// Varian is what is actually being served, which may not be what was asked
	// for — see the fallback in Serve.
	Varian domain.Varian
}

// Serve resolves one variant to a file on disk, after deciding whether this
// caller may read it.
//
// A public file is served to anyone. A private one requires an authenticated
// caller who either reaches every file (super admin / Admin) or owns this one;
// this is the Fase 0 stand-in for the full owner-permission check — see
// authorizeRead.
//
// When a derived variant has not been written yet, the original is served
// instead: a picture that is momentarily larger than asked for beats a broken
// image in the page.
func (s *FileService) Serve(ctx context.Context, actor Actor, fileUUID string, varian domain.Varian) (Served, error) {
	if !varian.Valid() {
		return Served{}, apperr.Validationf("varian %q tidak dikenal", varian)
	}
	f, err := s.repo.FindByUUID(ctx, fileUUID)
	if err != nil {
		return Served{}, err
	}
	if err := s.authorizeRead(actor, f); err != nil {
		return Served{}, err
	}

	v, err := s.repo.FindVarian(ctx, fileUUID, varian)
	if err != nil {
		if !errors.Is(err, apperr.ErrNotFound) || varian == domain.VarianOriginal {
			return Served{}, err
		}
		if v, err = s.repo.FindVarian(ctx, fileUUID, domain.VarianOriginal); err != nil {
			return Served{}, err
		}
	}

	abs, err := s.store.Abs(v.Path)
	if err != nil {
		return Served{}, err
	}
	if st, statErr := os.Stat(abs); statErr != nil || st.IsDir() {
		// The row exists but the bytes do not: report it, and answer 404
		// rather than 500 — the client can do nothing about either.
		logger.Error("file: varian tercatat tapi tidak ada di disk",
			zap.String("uuid", fileUUID), zap.String("path", v.Path), zap.Error(statErr))
		return Served{}, fmt.Errorf("berkas %s varian %s: %w", fileUUID, v.Varian, apperr.ErrNotFound)
	}

	return Served{
		AbsPath:  abs,
		MimeType: v.MimeType,
		NamaAsli: f.NamaAsli,
		IsPublik: f.IsPublik,
		Varian:   v.Varian,
	}, nil
}

// Delete soft-deletes a file. Bytes stay on disk for a separate sweeper, so a
// mistaken delete is still recoverable.
//
// Ownership is enforced here, not in middleware: SA/Admin hold file.delete with
// cakupan `semua`, Moderator and User with `milik_sendiri`, and only this layer
// knows whose file the uuid names.
func (s *FileService) Delete(ctx context.Context, actor Actor, fileUUID string) error {
	f, err := s.repo.FindByUUID(ctx, fileUUID)
	if err != nil {
		return err
	}
	if !actor.admin() && !actor.owns(f) {
		return fmt.Errorf("berkas ini bukan milik Anda: %w", apperr.ErrForbidden)
	}
	if err := s.repo.SoftDelete(ctx, f.ID, actor.UserID); err != nil {
		return err
	}

	s.audit.Log(ctx, audit.Entry{
		AktorUserID: actor.UserID,
		Modul:       "file",
		Aksi:        "hapus",
		ReffType:    "mst_file",
		ReffID:      &f.ID,
		Ringkasan:   fmt.Sprintf("Menghapus berkas %s (%s)", f.NamaAsli, f.Kategori),
		NilaiLama: map[string]any{
			"uuid": f.UUID, "kategori": f.Kategori, "nama_asli": f.NamaAsli,
			"reff_type": f.ReffType, "reff_id": f.ReffID,
		},
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	})
	return nil
}

// authorizeRead decides whether actor may read f.
//
// Public files are open. For a private one the rule is the reach the caller has
// over the OWNING row: reading an absensi selfie is really absensi.read with
// its cakupan, which is why the file layer has no `file.read` permission of its
// own (rancangan Bab 3.5).
//
// Fase 0 approximates that with the two limbs the JWT already carries — super
// admin / Admin reach every file, everyone else reaches what they uploaded.
// SEAM: when the dynamic RBAC guard lands (05-hak-akses), resolve
// f.ReffType → the owning module's read permission and evaluate its cakupan
// here. The route wrapper and this function are the only two places that change.
func (s *FileService) authorizeRead(actor Actor, f domain.File) error {
	if f.IsPublik {
		return nil
	}
	if !actor.authenticated() {
		return fmt.Errorf("berkas privat: %w", apperr.ErrUnauthorized)
	}
	if actor.admin() || actor.owns(f) {
		return nil
	}
	return fmt.Errorf("tidak berhak membaca berkas ini: %w", apperr.ErrForbidden)
}

// respond assembles the wire shape, reading the variants that exist right now.
func (s *FileService) respond(ctx context.Context, f domain.File) (dto.UploadResp, error) {
	rows, err := s.repo.VariansOf(ctx, f.ID)
	if err != nil {
		return dto.UploadResp{}, err
	}

	variants := make([]dto.VarianResp, 0, len(rows))
	for _, v := range rows {
		variants = append(variants, dto.VarianResp{
			Varian:     string(v.Varian),
			URL:        dto.ServeURL(f.UUID, v.Varian),
			URLPublik:  v.URLPublik,
			LebarPx:    v.LebarPx,
			TinggiPx:   v.TinggiPx,
			UkuranByte: v.UkuranByte,
			MimeType:   v.MimeType,
			Kualitas:   v.Kualitas,
		})
	}

	// The URL an owning form should render: medium for an image (the size the
	// UI actually wants), original for anything else.
	display := domain.VarianOriginal
	if IsImageMIME(f.MimeType) {
		display = domain.VarianMedium
	}

	return dto.UploadResp{
		UUID:         f.UUID,
		NamaAsli:     f.NamaAsli,
		NamaSlug:     f.NamaSlug,
		Ekstensi:     f.Ekstensi,
		MimeType:     f.MimeType,
		UkuranByte:   f.UkuranByte,
		Kategori:     f.Kategori,
		ReffType:     f.ReffType,
		ReffID:       f.ReffID,
		IsPublik:     f.IsPublik,
		StatusProses: string(f.StatusProses),
		HashSHA256:   f.HashSHA256,
		LebarPx:      f.LebarPx,
		TinggiPx:     f.TinggiPx,
		Variants:     variants,
		URL:          dto.ServeURL(f.UUID, display),
	}, nil
}

// detectMIME reads the content type out of the first bytes and drops the
// charset parameter, so the column stores "text/plain", not
// "text/plain; charset=utf-8".
func detectMIME(data []byte) string {
	head := data
	if len(head) > 512 {
		head = head[:512]
	}
	mime := http.DetectContentType(head)
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		mime = strings.TrimSpace(mime[:i])
	}
	return mime
}

// checkKategoriMIME enforces what a kategori accepts, from the DETECTED type.
// Only `dokumen` takes non-images, and only a PDF: every other kategori exists
// to hold a picture.
func checkKategoriMIME(kategori, mime string) error {
	if IsImageMIME(mime) {
		return nil
	}
	if domain.DokumenMIME(kategori) {
		if mime == "application/pdf" {
			return nil
		}
		return apperr.InvalidField("file", "kategori dokumen hanya menerima PDF atau gambar")
	}
	return apperr.InvalidField("file",
		fmt.Sprintf("kategori %s hanya menerima gambar, bukan %s", kategori, mime))
}

// publikFor resolves is_publik: what the client sent, otherwise the kategori's
// default (faces, identity documents and attendance evidence are private; club
// artwork is not).
func publikFor(req dto.UploadReq, kategori string) bool {
	if req.IsPublik != nil {
		return *req.IsPublik
	}
	return domain.PublikDefault(kategori)
}

// publikURL fills mst_file_varian.url_publik, and only for public files: a
// private variant has no URL that skips the permission check.
//
// It is the same route either way — nothing is ever served straight out of a
// static folder — the column simply records which ones may be cached and shared.
func publikURL(f domain.File, v domain.Varian) *string {
	if !f.IsPublik {
		return nil
	}
	u := dto.ServeURL(f.UUID, v)
	return &u
}

// truncate keeps a value inside its column width.
func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}
