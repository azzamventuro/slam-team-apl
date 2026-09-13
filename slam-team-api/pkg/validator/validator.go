// Package validator hooks custom validation rules into Gin's default
// go-playground/validator engine. Gin already validates `binding` tags; this
// is the seam for registering project-specific rules.
package validator

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Register installs custom validators. Add new rules here, e.g.:
//
//	v.RegisterValidation("notblank", notBlank)
func Register() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Report fields by their wire name (json, else form) so Explain's keys
		// match what the client sent: "alasan_batal", not "AlasanBatal".
		v.RegisterTagNameFunc(wireName)
	}
}

// wireName returns the json (or form) tag name of a struct field, without
// options; "" lets validator fall back to the Go field name.
func wireName(f reflect.StructField) string {
	for _, tag := range []string{"json", "form"} {
		name := strings.SplitN(f.Tag.Get(tag), ",", 2)[0]
		if name == "-" {
			return ""
		}
		if name != "" {
			return name
		}
	}
	return ""
}
