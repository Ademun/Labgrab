package apperr

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var ErrUnauthorized = errors.New("unauthorized")

var ErrNotFound = errors.New("not found")

var ErrForbidden = errors.New("forbidden")

var ErrConflict = errors.New("conflict")

type ValidationError struct {
	Details map[string]error
}

func NewValidationError() *ValidationError {
	return &ValidationError{
		Details: make(map[string]error),
	}
}

func (e *ValidationError) Error() string {
	var sb strings.Builder
	for k, v := range e.Details {
		sb.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}
	return sb.String()
}

func (e *ValidationError) AddErr(field string, err error) {
	e.Details[field] = err
}

func (e *ValidationError) IsEmpty() bool {
	return len(e.Details) == 0
}

func HTTPErrorCode(err error) int {
	if errors.Is(err, ErrUnauthorized) {
		return http.StatusUnauthorized
	}
	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, ErrForbidden) {
		return http.StatusForbidden
	}
	if errors.Is(err, ErrConflict) {
		return http.StatusConflict
	}
	if _, ok := errors.AsType[*ValidationError](err); ok {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
