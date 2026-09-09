package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

// Image encoding choices for the file layer.
//
// The pipeline is pure Go (github.com/disintegration/imaging) so the API stays
// a single binary with no cgo and no system image libraries to install.
//
// ponytail: medium/low are encoded JPEG (or PNG when the source carries
// transparency) — the "cadangan JPEG" of rancangan Bab 6.2. imaging cannot
// encode WebP without cgo; adding WebP later changes this file and the variant
// row's mime_type, not the schema.

// mimeFormat pairs a detected content type with what imaging can do about it.
type mimeFormat struct {
	format imaging.Format
	ext    string
	mime   string
	// alpha marks formats that can carry transparency: flattening those into
	// JPEG would put a black box behind every transparent logo, so their
	// variants stay in the same lossless format.
	alpha bool
}

// decodableImages is the set of image types this layer accepts. Anything else
// (WebP, HEIC, SVG, …) is refused at upload with a 422 rather than stored as
// something no variant pass can read.
var decodableImages = map[string]mimeFormat{
	"image/jpeg": {imaging.JPEG, "jpg", "image/jpeg", false},
	"image/png":  {imaging.PNG, "png", "image/png", true},
	"image/gif":  {imaging.GIF, "gif", "image/gif", true},
	"image/bmp":  {imaging.BMP, "bmp", "image/bmp", false},
	"image/tiff": {imaging.TIFF, "tiff", "image/tiff", false},
}

// IsImageMIME reports whether the detected content type is an image at all,
// including ones this layer cannot decode.
func IsImageMIME(mime string) bool { return strings.HasPrefix(mime, "image/") }

// imageFormat looks the detected content type up, reporting false for images
// the encoder cannot round-trip.
func imageFormat(mime string) (mimeFormat, bool) {
	f, ok := decodableImages[mime]
	return f, ok
}

// extForMIME is the fallback extension for an upload whose filename carries
// none.
func extForMIME(mime string) string {
	if f, ok := decodableImages[mime]; ok {
		return f.ext
	}
	switch mime {
	case "application/pdf":
		return "pdf"
	default:
		return "bin"
	}
}

// decodeOriented decodes an upload and bakes in its EXIF Orientation.
//
// Auto-orient is mandatory, not a nicety: phone cameras record a portrait shot
// as landscape pixels plus an Orientation tag, so a pipeline that ignores it
// stores every selfie rotated 90°. Re-encoding then strips EXIF from what is
// written, which is why the tags are kept in mst_file.metadata_exif instead.
func decodeOriented(data []byte) (image.Image, error) {
	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("gambar tidak bisa dibaca: %w", err)
	}
	return img, nil
}

// fitDown scales an image so its longest side is at most maxPx, and returns it
// untouched when it already fits.
//
// Images are NEVER upscaled (rancangan Bab 6.2): an 800 px original stays 800
// px as its "medium", because inventing pixels only costs bytes.
func fitDown(img image.Image, maxPx int) image.Image {
	if maxPx <= 0 {
		return img
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxPx && h <= maxPx {
		return img
	}
	return imaging.Fit(img, maxPx, maxPx, imaging.Lanczos)
}

// encoded is one rendered variant, still in memory.
type encoded struct {
	data     []byte
	ext      string
	mime     string
	lebarPx  int
	tinggiPx int
}

// encodeOriginal renders the `original` variant: the source format, capped at
// the configured longest side, at the original tier's quality.
func encodeOriginal(img image.Image, src mimeFormat, maxPx, quality int) (encoded, error) {
	return encodeAs(fitDown(img, maxPx), src.format, src.ext, src.mime, quality)
}

// encodeDerived renders `medium` or `low`. Sources that can carry transparency
// keep their lossless format; everything else becomes JPEG at the tier's
// quality.
func encodeDerived(img image.Image, src mimeFormat, maxPx, quality int) (encoded, error) {
	fitted := fitDown(img, maxPx)
	if src.alpha {
		return encodeAs(fitted, imaging.PNG, "png", "image/png", quality)
	}
	return encodeAs(fitted, imaging.JPEG, "jpg", "image/jpeg", quality)
}

// encodeAs runs the encoder and reports the dimensions actually written, so the
// variant row records the file rather than the intent.
func encodeAs(img image.Image, format imaging.Format, ext, mime string, quality int) (encoded, error) {
	var buf bytes.Buffer
	opts := []imaging.EncodeOption{}
	if format == imaging.JPEG {
		opts = append(opts, imaging.JPEGQuality(quality))
	}
	if err := imaging.Encode(&buf, img, format, opts...); err != nil {
		return encoded{}, fmt.Errorf("encode %s: %w", ext, err)
	}
	b := img.Bounds()
	return encoded{
		data:     buf.Bytes(),
		ext:      ext,
		mime:     mime,
		lebarPx:  b.Dx(),
		tinggiPx: b.Dy(),
	}, nil
}

// readEXIF pulls the original EXIF block out as JSON for mst_file.metadata_exif.
//
// Best effort by design: most uploads carry no EXIF, and a photo whose tags
// cannot be parsed is still a perfectly good photo. It returns nil rather than
// an error so a malformed tag block never fails an upload — the value is
// evidence, not a requirement.
func readEXIF(data []byte) *string {
	x, err := exif.Decode(bytes.NewReader(data))
	if err != nil || x == nil {
		return nil
	}
	raw, err := x.MarshalJSON()
	if err != nil || len(raw) == 0 || !json.Valid(raw) {
		return nil
	}
	s := string(raw)
	return &s
}
