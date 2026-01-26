package errors

import "fmt"

type ErrDBProcedure struct {
	Procedure string
	Step      string
	Err       error
}

func (e *ErrDBProcedure) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.Procedure, e.Step, e.Err)
}
