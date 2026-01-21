package auth

import "fmt"

type ErrHashIntegrity struct {
	ExpectedHash string
	ActualHash   string
}

func (e ErrHashIntegrity) Error() string {
	return fmt.Sprintf("Failed to verify hash integrity. Expected hash %s. Actual hash %s.", e.ExpectedHash, e.ActualHash)
}
