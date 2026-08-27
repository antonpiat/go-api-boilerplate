package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func BindJSON(c *gin.Context, dst any) error {
	if err := c.ShouldBindJSON(dst); err != nil {
		return mapBindError(err)
	}
	return nil
}

func mapBindError(err error) *AppError {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		fields := make([]FieldError, 0, len(ve))
		for _, fe := range ve {
			fields = append(fields, FieldError{
				Field:   jsonFieldName(fe),
				Message: validationMessage(fe),
			})
		}
		return Validation("invalid request body", fields...)
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return Validation("invalid request body", FieldError{
			Field:   typeErr.Field,
			Message: fmt.Sprintf("must be a %s", typeErr.Type.String()),
		})
	}

	if errors.Is(err, io.EOF) {
		return Validation("request body is required")
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return Validation("malformed JSON")
	}

	return Validation(err.Error())
}

func jsonFieldName(fe validator.FieldError) string {
	name := fe.Field()
	if name == "" {
		return fe.StructField()
	}
	return strings.ToLower(name[:1]) + name[1:]
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "uuid":
		return "must be a valid UUID"
	default:
		return "is invalid"
	}
}
