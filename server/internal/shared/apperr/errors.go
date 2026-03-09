package apperr

import (
	"errors"
	"fmt"
	"strings"
)

var ErrUnauthorized = errors.New("unauthorized")

var ErrNotFound = errors.New("not found")

var ErrInternal = errors.New("internal server error")

var ErrForbidden = errors.New("forbidden")

type ValidationError struct {
	Details map[string]error
}

func NewValidationError() *ValidationError {
	return &ValidationError{
		Details: make(map[string]error),
	}
}

func (e ValidationError) Error() string {
	var sb strings.Builder
	for k, v := range e.Details {
		sb.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}
	return sb.String()
}

func (e ValidationError) AddErr(field string, err error) {
	e.Details[field] = err
}

func (e ValidationError) IsEmpty() bool {
	return len(e.Details) == 0
}
