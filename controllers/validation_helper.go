package controllers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func formatBindError(err error) []string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return []string{strings.TrimSpace(err.Error())}
	}
	out := make([]string, 0, len(ve))
	for _, fe := range ve {
		out = append(out, fmt.Sprintf("%s: %s", fe.Field(), msgForTag(fe)))
	}
	return out
}

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		return "minimal " + fe.Param() + " karakter"
	case "max":
		return "maksimal " + fe.Param() + " karakter"
	default:
		return "tidak valid"
	}
}
