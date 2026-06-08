package apperr

import "fmt"

const (
	TypeBadRequest   = "bad_request"
	TypeUnauthorized = "unauthorized"
	TypeForbidden    = "forbidden"
	TypeNotFound     = "not_found"
	TypeConflict     = "conflict"
	TypeValidation   = "validation"
	TypeInternal     = "internal"
	TypeUnavailable  = "service_unavailable"
)

type AppError struct {
	Type    string
	Code    string
	Message string
	Status  int
	Err     error
	Errors  map[string]string
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func BadRequest(code, message string) *AppError {
	return &AppError{Type: TypeBadRequest, Code: code, Message: message, Status: 400}
}

func Unauthorized(code, message string) *AppError {
	return &AppError{Type: TypeUnauthorized, Code: code, Message: message, Status: 401}
}

func Forbidden(code, message string) *AppError {
	return &AppError{Type: TypeForbidden, Code: code, Message: message, Status: 403}
}

func NotFound(code, message string) *AppError {
	return &AppError{Type: TypeNotFound, Code: code, Message: message, Status: 404}
}

func Conflict(code, message string) *AppError {
	return &AppError{Type: TypeConflict, Code: code, Message: message, Status: 409}
}

func Validation(errors map[string]string) *AppError {
	return &AppError{
		Type:    TypeValidation,
		Code:    "validation.failed",
		Message: "Validation failed",
		Status:  422,
		Errors:  errors,
	}
}

func Internal(err error) *AppError {
	return &AppError{
		Type:    TypeInternal,
		Code:    "internal.error",
		Message: "Internal server error",
		Status:  500,
		Err:     err,
	}
}

func ServiceUnavailable(code, message string) *AppError {
	return &AppError{
		Type:    TypeUnavailable,
		Code:    code,
		Message: message,
		Status:  503,
	}
}

func NotImplemented() *AppError {
	return &AppError{
		Type:    TypeBadRequest,
		Code:    "feature.not_implemented",
		Message: "Feature not implemented yet",
		Status:  501,
	}
}
