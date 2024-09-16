package utils

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	v := validator.New()

	// Use the names which have been specified for JSON or Query Params
	// representations of structs, rather than normal Go field names
	v.RegisterTagNameFunc(
		func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("params"), ",", 2)[0]

			if name == "" {
				name = strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			}

			if name == "" {
				name = strings.SplitN(fld.Tag.Get("query"), ",", 2)[0]
			}

			if name == "-" {
				return ""
			}

			return name
		},
	)

	Validate = v
}
