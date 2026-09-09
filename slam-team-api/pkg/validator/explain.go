package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Explain turns a binding failure into a field → message map for the envelope's
// "errors" object:
//
//	if err := c.ShouldBindJSON(&req); err != nil {
//	    response.Unprocess(c, "validasi gagal", validator.Explain(err))
//	    return
//	}
//
// Keys are the JSON field paths the client sent (items[0].kunci), never the Go
// struct names. A non-validator error (malformed JSON, wrong type) has no
// fields to report and yields a single "_" entry.
func Explain(err error) map[string]string {
	if err == nil {
		return nil
	}

	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return map[string]string{"_": "format permintaan tidak valid"}
	}

	out := make(map[string]string, len(verrs))
	for _, fe := range verrs {
		out[fieldPath(fe)] = message(fe)
	}
	return out
}

// fieldPath drops the top-level struct name from validator's namespace
// ("UpdatePengaturanReq.Items[0].Kunci" → "items[0].kunci"), leaving the path a
// client can match against the JSON body it sent.
func fieldPath(fe validator.FieldError) string {
	ns := fe.Namespace()
	if i := strings.Index(ns, "."); i >= 0 {
		ns = ns[i+1:]
	}
	return strings.ToLower(ns)
}

func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		return fmt.Sprintf("minimal %s", fe.Param())
	case "max":
		return fmt.Sprintf("maksimal %s", fe.Param())
	case "len":
		return fmt.Sprintf("panjangnya harus %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("harus salah satu dari: %s", strings.ReplaceAll(fe.Param(), " ", ", "))
	case "gte":
		return fmt.Sprintf("minimal %s", fe.Param())
	case "lte":
		return fmt.Sprintf("maksimal %s", fe.Param())
	case "latitude", "longitude":
		return "koordinat tidak valid"
	default:
		return fmt.Sprintf("tidak memenuhi aturan %q", fe.Tag())
	}
}
