package errors

import (
	"fmt"
	"strings"
)

type Validator interface {
	Validate() error
}

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
	var messages []string
	for field, message := range e.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", field, message))
	}
	return fmt.Sprintf("validation failed: %s", strings.Join(messages, "; "))
}
