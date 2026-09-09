package apperr

import "fmt"

// FieldsError is a validation failure that can name the offending fields, so a
// 422 carries {"errors": {"kta.dpi_cetak": "harus berupa bilangan bulat"}}
// instead of a bare message. It unwraps to ErrValidation, so callers that only
// switch on the sentinel keep working.
type FieldsError struct {
	Fields map[string]string
}

// Error satisfies error. The map itself is what a client reads; this is the
// envelope's "message".
func (e *FieldsError) Error() string { return ErrValidation.Error() }

// Unwrap makes errors.Is(err, ErrValidation) true for a FieldsError.
func (e *FieldsError) Unwrap() error { return ErrValidation }

// Invalid builds a FieldsError. Returns nil when there is nothing to report,
// so a service can `if err := apperr.Invalid(errs); err != nil { return err }`.
func Invalid(fields map[string]string) error {
	if len(fields) == 0 {
		return nil
	}
	return &FieldsError{Fields: fields}
}

// InvalidField is the single-field shorthand.
func InvalidField(field, msg string) error {
	return &FieldsError{Fields: map[string]string{field: msg}}
}

// Validationf builds a plain (field-less) validation error with a formatted
// message, for failures that do not belong to one input field.
func Validationf(format string, args ...any) error {
	return fmt.Errorf(format+": %w", append(args, ErrValidation)...)
}
