// Package validator hooks custom validation rules into Gin's default
// go-playground/validator engine. Gin already validates `binding` tags; this
// is the seam for registering project-specific rules.
package validator

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Register installs custom validators. Add new rules here, e.g.:
//
//	v.RegisterValidation("notblank", notBlank)
func Register() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v // no custom rules yet
	}
}
