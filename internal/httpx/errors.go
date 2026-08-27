package httpx

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	CodeValidation   ErrorCode = "VALIDATION_ERROR"
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	CodeForbidden    ErrorCode = "FORBIDDEN"
	CodeNotFound     ErrorCode = "NOT_FOUND"
	CodeConflict     ErrorCode = "CONFLICT"
	CodeInternal     ErrorCode = "INTERNAL_ERROR"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type AppError struct {
	Code    ErrorCode
	Message string
	Status  int
	Fields  []FieldError
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewError(status int, code ErrorCode, message string) *AppError {
	return &AppError{Code: code, Message: message, Status: status}
}

func Validation(message string, fields ...FieldError) *AppError {
	return &AppError{
		Code:    CodeValidation,
		Message: message,
		Status:  http.StatusBadRequest,
		Fields:  fields,
	}
}

func Unauthorized(message string) *AppError {
	return NewError(http.StatusUnauthorized, CodeUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return NewError(http.StatusForbidden, CodeForbidden, message)
}

func NotFound(message string) *AppError {
	return NewError(http.StatusNotFound, CodeNotFound, message)
}

func Conflict(message string) *AppError {
	return NewError(http.StatusConflict, CodeConflict, message)
}

func Internal(message string, err error) *AppError {
	return &AppError{
		Code:    CodeInternal,
		Message: message,
		Status:  http.StatusInternalServerError,
		Err:     err,
	}
}

func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
