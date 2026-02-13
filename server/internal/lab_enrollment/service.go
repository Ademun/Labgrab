package lab_enrollment

import (
	"golang.org/x/time/rate"
)

type Service struct {
	repo    *Repo
	limiter *rate.Limiter
}
