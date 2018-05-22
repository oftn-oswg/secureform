package secureform

import (
	"errors"
	"fmt"
)

// ErrExpectedStructPtr is produced when a type other than a struct pointer is passed to be populated.
var ErrExpectedStructPtr = errors.New("Invalid field interface, expected struct pointer")

// ErrInvalidKind is produced when a struct field has an unsupported type.
// Currently, only string, bool, numeric, secureform.Type and secureform.File types are supported.
var ErrInvalidKind = errors.New("Invalid field kind, expected string, bool, numeric, secureform.Type, or secureform.File")

// ErrValidMin is produced when a form value or string length is too small.
var ErrValidMin = errors.New("Result falls below minimum range")

// ErrValidMax is produced when a form value or string length is too large.
var ErrValidMax = errors.New("Result falls above maximum range")

// FieldError is an error that relates to a specific struct field.
type FieldError struct {
	Name string
	Err  error
}

func (err *FieldError) Error() string {
	return fmt.Sprintf("Error parsing %q form field: %s", err.Name, err.Err)
}
