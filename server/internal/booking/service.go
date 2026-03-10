package booking

import (
	"context"
	"fmt"
	"labgrab/internal/shared/api/dikidi"
	"labgrab/internal/shared/domain"
	"labgrab/internal/shared/mask"
	"strconv"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("booking-service")

type Service struct {
	repo   *Repo
	client *dikidi.Client
}

func NewService(repo *Repo, client *dikidi.Client) *Service {
	return &Service{repo: repo, client: client}
}

func (s *Service) GetBookings(ctx context.Context, userUUID uuid.UUID) ([]GetBookingsRes, error) {
	bookings, err := s.repo.GetBookings(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("booking service: get booking: repository call: %w", err)
	}

	result := make([]GetBookingsRes, len(bookings))
	for i, b := range bookings {
		result[i] = GetBookingsRes{
			UserUUID: b.UserUUID,
			Status:   b.Status,
			Data: &domain.Booking{
				ID:         b.BookingID,
				Type:       b.Type,
				Topic:      b.Topic,
				Auditorium: b.Auditorium,
				Spot:       b.Spot,
				Lesson:     b.Lesson,
				Start:      b.Start,
				End:        b.End,
			},
		}
	}

	return result, nil
}

func (s *Service) LoadClientBookings(ctx context.Context, req *LoadClientBookingsReq) error {
	client, err := mask.CreateClientWithCookies(&req.Cookies)
	if err != nil {
		return fmt.Errorf("booking service: load client booking: client initialization: %w", err)
	}

	bookings, err := s.client.GetBookings(ctx, client, &dikidi.GetBookingsRequest{
		Session: req.Session,
	})
	if err != nil {
		return fmt.Errorf("booking service: load client booking: get booking: %w", err)
	}

	dbBookings := make([]DBBooking, 0)
	for _, b := range bookings.Active {
		dbBookings = append(dbBookings, DBBooking{
			BookingID:  b.ID,
			Type:       b.Type,
			Topic:      b.Topic,
			Number:     b.Number,
			Auditorium: b.Auditorium,
			Spot:       b.Spot,
			Lesson:     b.Lesson,
			Start:      b.Start,
			End:        b.End,
			Status:     StatusOpen,
			UserUUID:   req.UserUUID,
		})
	}

	for _, b := range bookings.Closed {
		dbBookings = append(dbBookings, DBBooking{
			BookingID:  b.ID,
			Type:       b.Type,
			Topic:      b.Topic,
			Number:     b.Number,
			Auditorium: b.Auditorium,
			Spot:       b.Spot,
			Lesson:     b.Lesson,
			Start:      b.Start,
			End:        b.End,
			Status:     StatusClosed,
			UserUUID:   req.UserUUID,
		})
	}

	if err = s.repo.LoadBookings(ctx, req.UserUUID, dbBookings); err != nil {
		return fmt.Errorf("booking service: load client booking: load booking: %w", err)
	}

	return nil
}

func (s *Service) CancelClientBooking(ctx context.Context, req *CancelClientBookingReq) error {
	client, err := mask.CreateClientWithCookies(&req.Cookies)
	if err != nil {
		return fmt.Errorf("booking service: cancel client booking: client initialization: %w", err)
	}

	if err = s.client.RemoveBooking(ctx, client, &dikidi.RemoveBookingRequest{
		BookingID: strconv.Itoa(req.BookingID),
		Session:   req.Session,
	}); err != nil {
		return fmt.Errorf("booking service: cancel client booking: remove booking: %w", err)
	}

	mask.Jitter(1000, 2000)
	if err = s.LoadClientBookings(ctx, &LoadClientBookingsReq{
		UserUUID: req.UserUUID,
		Session:  req.Session,
		Cookies:  req.Cookies,
	}); err != nil {
		return fmt.Errorf("booking service: cancel client booking: load client booking: %w", err)
	}

	return nil
}

func (s *Service) DeleteBookings(ctx context.Context, req *DeleteBookingsReq) error {
	if err := s.repo.DeleteBookings(ctx, req.UserUUID, req.Tx); err != nil {
		return fmt.Errorf("booking service: delete booking: repository call: %w", err)
	}
	return nil
}

func (s *Service) FilterSchedule(ctx context.Context, req *FilterScheduleReq) (domain.Schedule, error) {
	schedule, err := s.repo.FilterSchedule(ctx, &DBSlotFilter{
		UserUUID: req.UserUUID,
		Schedule: req.Schedule,
		Type:     req.Type,
		Topic:    req.Topic,
		Number:   req.Number,
	})
	if err != nil {
		return nil, fmt.Errorf("booking service: filter schedule: repository call: %w", err)
	}
	return schedule, nil
}
