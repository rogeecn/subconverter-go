package errors

import (
	"net/http"

)

type Error struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Status  int                    `json:"status"`
}

func (e *Error) Error() string {
	return e.Message
}

func New(code string, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Status:  http.StatusInternalServerError,
	}
}

func NewWithStatus(code string, message string, status int) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

func Wrap(err error, message string) *Error {
	return &Error{
		Code:    "INTERNAL_ERROR",
		Message: message,
		Status:  http.StatusInternalServerError,
		Details: map[string]interface{}{
			"original_error": err.Error(),
		},
	}
}

func BadRequest(code string, message string) *Error {
	return NewWithStatus(code, message, http.StatusBadRequest)
}

func NotFound(code string, message string) *Error {
	return NewWithStatus(code, message, http.StatusNotFound)
}

func Unauthorized(code string, message string) *Error {
	return NewWithStatus(code, message, http.StatusUnauthorized)
}

func Forbidden(code string, message string) *Error {
	return NewWithStatus(code, message, http.StatusForbidden)
}

func InternalError(code string, message string) *Error {
	return NewWithStatus(code, message, http.StatusInternalServerError)
}

func ValidationError(field string, message string) *Error {
	return &Error{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Status:  http.StatusBadRequest,
		Details: map[string]interface{}{
			"field": field,
		},
	}
}

// Common errors
var (
	ErrInvalidURL       = BadRequest("INVALID_URL", "invalid subscription URL")
	ErrInvalidFormat    = BadRequest("INVALID_FORMAT", "invalid subscription format")
	ErrEmptyResponse    = BadRequest("EMPTY_RESPONSE", "empty subscription content")
	ErrParseFailed      = BadRequest("PARSE_FAILED", "failed to parse subscription")
	ErrGenerationFailed = InternalError("GENERATION_FAILED", "failed to generate configuration")
	ErrCacheError       = InternalError("CACHE_ERROR", "cache operation failed")
	ErrNotFound         = NotFound("NOT_FOUND", "resource not found")
	ErrTimeout          = InternalError("TIMEOUT", "operation timeout")
	ErrRateLimit        = NewWithStatus("RATE_LIMIT", "too many requests", http.StatusTooManyRequests)
)