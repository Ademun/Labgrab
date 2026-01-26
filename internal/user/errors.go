package user

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Errors map[string]string
}

func NewValidationError() *ValidationError {
	return &ValidationError{
		Errors: make(map[string]string),
	}
}

func (e *ValidationError) Add(field, message string) {
	e.Errors[field] = message
}

func (e *ValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}

func (e *ValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "validation error"
	}

	var msgs []string
	for field, msg := range e.Errors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", field, msg))
	}
	return fmt.Sprintf("validation failed: %s", strings.Join(msgs, "; "))
}
