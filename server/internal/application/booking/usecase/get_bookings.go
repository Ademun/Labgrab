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

func (uc *GetBookingsUsecase) Exec(ctx context.Context, session string) ([]dto.GetBookingsRespDTO, error) {
	sessionData, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("booking: usacase: get session data: %w", err)
	}

	creds, err := uc.AuthSvc.GetUserCredentials(ctx, sessionData)
	if err != nil {
		return nil, fmt.Errorf("booking: usecase: get credentials: %w", err)
	}

	if creds.Session == nil || creds.Cookies == nil {
		return nil, fmt.Errorf("booking: usecase: get credentials: no api session")
	}

	if err := uc.BookingSvc.LoadClientBookings(ctx, &booking.LoadClientBookingsReq{
		UserUUID: sessionData,
		Session:  *creds.Session,
		Cookies:  *creds.Cookies,
	}); err != nil {
		return nil, fmt.Errorf("booking: usecase: load bookings: %w", err)
	}

	bookings, err := uc.BookingSvc.GetBookings(ctx, sessionData)

	dtoBookings := make([]dto.GetBookingsRespDTO, len(bookings))
	for i, b := range bookings {
		dtoBookings[i] = dto.GetBookingsRespDTO{
			Type:       b.Data.Type,
			Topic:      b.Data.Topic,
			Number:     b.Data.Number,
			Auditorium: b.Data.Auditorium,
			Spot:       b.Data.Spot,
			Lesson:     b.Data.Lesson,
			Start:      b.Data.Start,
			End:        b.Data.End,
			Status:     b.Status,
		}
	}

	return dtoBookings, nil
}
