package utils

import "github.com/go-playground/validator/v10"

var validate = validator.New()

// ValidateStruct validates a struct using its `validate` struct tags.
// Returns an error describing the first validation failure, or nil if valid.
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}
