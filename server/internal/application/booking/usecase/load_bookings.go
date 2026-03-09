package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/auth"
	"labgrab/internal/booking"
)

type LoadBookingsUsecase struct {
	BookingSvc *booking.Service
	AuthSvc    *auth.Service
}

func (uc *LoadBookingsUsecase) Exec(ctx context.Context, session string) error {
	sessionData, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return fmt.Errorf("booking usecase: load bookings: get session data: %w", err)
	}

	creds, err := uc.AuthSvc.GetUserCredentials(ctx, sessionData)
	if err != nil {
		return fmt.Errorf("booking usecase: load bookings: get credentials: %w", err)
	}

	if creds.Session == nil || creds.Cookies == nil {
		return fmt.Errorf("booking usecase: load bookings: get credentials: no api session")
	}

	if err := uc.BookingSvc.LoadClientBookings(ctx, &booking.LoadClientBookingsReq{
		UserUUID: sessionData,
		Session:  *creds.Session,
		Cookies:  *creds.Cookies,
	}); err != nil {
		return fmt.Errorf("booking usecase: load bookings: load client bookings: %w", err)
	}

	return nil
}
