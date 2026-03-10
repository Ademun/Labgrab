package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/auth"
	"labgrab/internal/booking"
)

type CancelBookingUsecase struct {
	AuthSvc    *auth.Service
	BookingSvc *booking.Service
}

func (uc *CancelBookingUsecase) Exec(ctx context.Context, session string, bookingID int) error {
	if err := uc.AuthSvc.ValidateSession(ctx, session); err != nil {
		return fmt.Errorf("bookings usecase: cancel booking: validate session: %w", err)
	}

	userUUID, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return fmt.Errorf("bookings usecase: cancel bboking: get session data: %w", err)
	}

	creds, err := uc.AuthSvc.GetUserCredentials(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("booking usecase: cancel booking: get credentials: %w", err)
	}

	if creds.Session == nil || creds.Cookies == nil {
		return fmt.Errorf("booking usecase: cancel booking: get credentials: no api session")
	}

	if err := uc.BookingSvc.CancelClientBooking(ctx, &booking.CancelClientBookingReq{
		UserUUID:  userUUID,
		BookingID: bookingID,
		Session:   *creds.Session,
		Cookies:   *creds.Cookies,
	}); err != nil {
		return fmt.Errorf("bookings usecase: cancel booking: cancel booking: %w", err)
	}

	return nil
}
