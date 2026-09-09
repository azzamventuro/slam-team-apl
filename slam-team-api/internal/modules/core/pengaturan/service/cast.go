package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"slam-team-api/internal/modules/core/pengaturan/domain"
)

// dateLayout is the storage format for TipeDate values.
const dateLayout = "2006-01-02"

// castOut turns the stored text into the JSON type the client expects, so the
// settings page binds a number input to a number and a switch to a bool.
//
// A value that does not parse is returned as its raw text rather than as an
// error: a malformed row is a data problem to surface in the UI, not a reason
// to fail the whole GET.
func castOut(tipe domain.TipeNilai, raw *string) any {
	if raw == nil {
		return nil
	}
	s := *raw
	switch tipe {
	case domain.TipeInteger:
		if n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
			return n
		}
		return s
	case domain.TipeBoolean:
		if b, err := strconv.ParseBool(strings.TrimSpace(s)); err == nil {
			return b
		}
		return s
	case domain.TipeJSON:
		if json.Valid([]byte(s)) {
			return json.RawMessage(s)
		}
		return s
	default: // TipeString, TipeDate, and any unknown enum value
		return s
	}
}

// normalizeIn parses what the client sent against the setting's declared type
// and returns the canonical text to store. It accepts both the native JSON type
// (12, true) and its string form ("12", "true"), because HTML inputs hand back
// strings; the stored form is identical either way.
//
// The returned error message is client-facing (it lands in the 422 errors map),
// so it names the expectation, never the parse failure.
func normalizeIn(tipe domain.TipeNilai, raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "", fmt.Errorf("nilai wajib diisi")
	}

	switch tipe {
	case domain.TipeInteger:
		return normalizeInteger(raw)
	case domain.TipeBoolean:
		return normalizeBoolean(raw)
	case domain.TipeJSON:
		if !json.Valid(raw) {
			return "", fmt.Errorf("harus berupa JSON yang valid")
		}
		return compactJSON(raw)
	case domain.TipeDate:
		s, ok := asScalarText(raw)
		if !ok {
			return "", fmt.Errorf("harus berupa tanggal YYYY-MM-DD")
		}
		if _, err := time.Parse(dateLayout, strings.TrimSpace(s)); err != nil {
			return "", fmt.Errorf("harus berupa tanggal YYYY-MM-DD")
		}
		return strings.TrimSpace(s), nil
	case domain.TipeString:
		s, ok := asScalarText(raw)
		if !ok {
			return "", fmt.Errorf("harus berupa teks")
		}
		return s, nil
	default:
		// An enum value the code does not know about: refuse rather than store
		// something the reader will misinterpret.
		return "", fmt.Errorf("tipe nilai %q tidak dikenal", tipe)
	}
}

func normalizeInteger(raw json.RawMessage) (string, error) {
	const msg = "harus berupa bilangan bulat"
	s, ok := asScalarText(raw)
	if !ok {
		return "", fmt.Errorf(msg)
	}
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return "", fmt.Errorf(msg)
	}
	return strconv.FormatInt(n, 10), nil
}

func normalizeBoolean(raw json.RawMessage) (string, error) {
	const msg = "harus berupa true atau false"
	s, ok := asScalarText(raw)
	if !ok {
		return "", fmt.Errorf(msg)
	}
	b, err := strconv.ParseBool(strings.TrimSpace(s))
	if err != nil {
		return "", fmt.Errorf(msg)
	}
	return strconv.FormatBool(b), nil
}

// asScalarText unwraps a JSON scalar to its text form: a string yields its
// contents, a number or boolean its literal. Objects, arrays and null are not
// scalars and report false.
func asScalarText(raw json.RawMessage) (string, bool) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return "", false
	}
	switch trimmed[0] {
	case '{', '[':
		return "", false
	case '"':
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return "", false
		}
		return s, true
	default: // number, true, false
		return trimmed, true
	}
}

// compactJSON strips insignificant whitespace so the stored text is stable and
// two writes of the same value compare equal.
func compactJSON(raw json.RawMessage) (string, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return "", fmt.Errorf("harus berupa JSON yang valid")
	}
	return buf.String(), nil
}

// checkOpsi enforces a dropdown's allowed set. opsi is a JSON array; each entry
// is compared in its stored text form, so ["png"] matches the string "png" and
// [1,2] matches the integer 1.
func checkOpsi(opsi json.RawMessage, nilai string) error {
	if len(opsi) == 0 || string(opsi) == "null" {
		return nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(opsi, &items); err != nil {
		// A malformed opsi column must not block an otherwise valid edit.
		return nil
	}
	allowed := make([]string, 0, len(items))
	for _, it := range items {
		s, ok := asScalarText(it)
		if !ok {
			continue
		}
		if s == nilai {
			return nil
		}
		allowed = append(allowed, s)
	}
	if len(allowed) == 0 {
		return nil
	}
	return fmt.Errorf("harus salah satu dari: %s", strings.Join(allowed, ", "))
}
