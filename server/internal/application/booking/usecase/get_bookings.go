package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/application/booking/dto"
	"labgrab/internal/auth"
	"labgrab/internal/booking"
)

type GetBookingsUsecase struct {
	BookingSvc *booking.Service
	AuthSvc    *auth.Service
}

func (uc *GetBookingsUsecase) Exec(ctx context.Context, session string) ([]dto.GetBookingsResDTO, error) {
	if err := uc.AuthSvc.ValidateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("bookings usecase: get bookings: validate session: %w", err)
	}

	userUUID, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("bookings usecase: get bookings: get session data: %w", err)
	}

	creds, err := uc.AuthSvc.GetUserCredentials(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("booking usecase: get bookings: get credentials: %w", err)
	}

	if creds.Session == nil || creds.Cookies == nil {
		return nil, fmt.Errorf("booking usecase: get bookings: get credentials: no api session")
	}

	if err := uc.BookingSvc.LoadClientBookings(ctx, &booking.LoadClientBookingsReq{
		UserUUID: userUUID,
		Session:  *creds.Session,
		Cookies:  *creds.Cookies,
	}); err != nil {
		return nil, fmt.Errorf("booking usecase: get bookings: load bookings: %w", err)
	}

	bookings, err := uc.BookingSvc.GetBookings(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("booking usecase: get bookings: get bookings: %w", err)
	}

	dtoBookings := make([]dto.GetBookingsResDTO, len(bookings))
	for i, b := range bookings {
		dtoBookings[i] = dto.GetBookingsResDTO{
			Type:       string(b.Data.Type),
			Topic:      string(b.Data.Topic),
			Number:     b.Data.Number,
			Auditorium: b.Data.Auditorium,
			Spot:       b.Data.Spot,
			Lesson:     int(b.Data.Lesson),
			Start:      b.Data.Start,
			End:        b.Data.End,
			Status:     string(b.Status),
		}
	}

	return dtoBookings, nil
}
