package errors

import (
	"errors"
	"fmt"
)

type ErrParsing struct {
	Errors []error
}

func (e *ErrParsing) Error() string {
	return fmt.Sprintf("Encountered %d errors when parsing slot: %s", len(e.Errors), errors.Join(e.Errors...))
}
