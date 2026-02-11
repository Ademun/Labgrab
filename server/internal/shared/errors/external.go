package errors

import "fmt"

type ExternalAPIError struct {
	Procedure string
	Step      string
	Err       error
}

func (e ExternalAPIError) Error() string {
	return fmt.Sprintf("External API error. %s: %s: %s", e.Procedure, e.Step, e.Err)
}
