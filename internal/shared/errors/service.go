package errors

import "fmt"

type ErrServiceProcedure struct {
	Procedure string
	Step      string
	Err       error
}

func (e ErrServiceProcedure) Error() string {
	return fmt.Sprintf("Service error. %s: %s: %s", e.Procedure, e.Step, e.Err)
}
