package subscription

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrCreateSubscription   = errors.New("failed to create subscription")
	ErrUpdateSubscription   = errors.New("failed to update subscription")
	ErrDeleteSubscription   = errors.New("failed to delete subscription")
	ErrCloseSubscription    = errors.New("failed to close subscription")
	ErrRestoreSubscription  = errors.New("failed to restore subscription")
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
