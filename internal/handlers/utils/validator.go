package utils

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

const SpecialCharacters string = " !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

var Validate *validator.Validate

func init() {
	v := validator.New()

	if err := v.RegisterValidation("containsUppercase", containsUppercaseValidator); err != nil {
		panic(err)
	}

	if err := v.RegisterValidation("containsLowercase", containsLowercaseValidator); err != nil {
		panic(err)
	}

	if err := v.RegisterValidation("containsDigit", containsDigitValidator); err != nil {
		panic(err)
	}

	if err := v.RegisterValidation("containsSpecialCharacter", containsSpecialCharacterValidator); err != nil {
		panic(err)
	}

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

func containsUppercaseValidator(fl validator.FieldLevel) bool {
	fieldString := fl.Field().String()

	for _, char := range fieldString {
		if unicode.IsUpper(char) {
			return true
		}
	}

	return false
}

func containsLowercaseValidator(fl validator.FieldLevel) bool {
	fieldString := fl.Field().String()

	for _, char := range fieldString {
		if unicode.IsLower(char) {
			return true
		}
	}

	return false
}

func containsDigitValidator(fl validator.FieldLevel) bool {
	fieldString := fl.Field().String()

	for _, char := range fieldString {
		if unicode.IsDigit(char) {
			return true
		}
	}

	return false
}

func containsSpecialCharacterValidator(fl validator.FieldLevel) bool {
	return strings.ContainsAny(fl.Field().String(), SpecialCharacters)
}
