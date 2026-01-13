package lab_polling

import (
	"errors"
	"fmt"
)

type ErrSlotParsing struct {
	errors []error
}

func (e *ErrSlotParsing) Error() string {
	return fmt.Sprintf("Encountered %d erros when parsing slot: %s", len(e.errors), errors.Join(e.errors...))
}
