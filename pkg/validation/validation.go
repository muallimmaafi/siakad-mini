package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	// Pakai nama field dari tag json (misal "nim", bukan "NIM")
	validate.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// Struct mengembalikan nil kalau valid, atau map field -> daftar pesan error.
func Struct(s interface{}) map[string][]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var ves validator.ValidationErrors
	if !errors.As(err, &ves) {
		return map[string][]string{"body": {"Data tidak valid"}}
	}

	out := make(map[string][]string)
	for _, fe := range ves {
		out[fe.Field()] = append(out[fe.Field()], message(fe))
	}
	return out
}

func message(fe validator.FieldError) string {
	field := fe.Field()
	isString := fe.Kind() == reflect.String

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s wajib diisi", field)
	case "email":
		return "Format email tidak valid"
	case "numeric":
		return fmt.Sprintf("%s harus berupa angka", field)
	case "len":
		return fmt.Sprintf("%s harus %s karakter", field, fe.Param())
	case "min", "gte":
		if isString {
			return fmt.Sprintf("%s minimal %s karakter", field, fe.Param())
		}
		return fmt.Sprintf("%s minimal %s", field, fe.Param())
	case "max", "lte":
		if isString {
			return fmt.Sprintf("%s maksimal %s karakter", field, fe.Param())
		}
		return fmt.Sprintf("%s maksimal %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s tidak valid", field)
	}
}