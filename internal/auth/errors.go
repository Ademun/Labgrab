package auth

import (
	"fmt"
	"time"
)

type ErrHashIntegrity struct {
	ExpectedHash string
	ActualHash   string
}

func (e ErrHashIntegrity) Error() string {
	return fmt.Sprintf("failed to verify hash integrity. Expected hash %s. Actual hash %s.", e.ExpectedHash, e.ActualHash)
}

type ErrAuthDateExpired struct {
	AuthDate    time.Time
	CurrentDate time.Time
}

func (e ErrAuthDateExpired) Error() string {
	return fmt.Sprintf("auth date has expired. Expected time diff < 24 hours. Actual time diff %f hours.", e.CurrentDate.Sub(e.AuthDate).Hours())
}
