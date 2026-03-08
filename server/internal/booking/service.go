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
	"go.opentelemetry.io/otel/codes"
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
	ctx, span := tracer.Start(ctx, "booking_service.get_bookings")
	defer span.End()

	bookings, err := s.repo.GetBookings(ctx, userUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get bookings")
		return nil, fmt.Errorf("booking service: get bookings: repository call: %w", err)
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

	span.SetStatus(codes.Ok, "")
	return result, nil
}

func (s *Service) LoadClientBookings(ctx context.Context, req *LoadClientBookingsReq) error {
	ctx, span := tracer.Start(ctx, "booking_service.load_client_bookings")
	defer span.End()

	client, err := mask.CreateClientWithCookies(&req.Cookies)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to load client bookings")
		return fmt.Errorf("booking service: load client bookings: client initialization: %w", err)
	}

	bookings, err := s.client.GetBookings(ctx, client, &dikidi.GetBookingsRequest{
		Session: req.Session,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to load client bookings")
		return fmt.Errorf("booking service: load client bookings: get bookings: %w", err)
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
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to load client bookings")
		return fmt.Errorf("booking service: load client bookings: load bookings: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *Service) CancelClientBooking(ctx context.Context, req *CancelClientBookingReq) error {
	ctx, span := tracer.Start(ctx, "booking_service.cancel_client_booking")
	defer span.End()

	client, err := mask.CreateClientWithCookies(&req.Cookies)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to cancel client booking")
		return fmt.Errorf("booking service: cancel client booking: client initialization: %w", err)
	}

	if err = s.client.RemoveBooking(ctx, client, &dikidi.RemoveBookingRequest{
		BookingID: strconv.Itoa(req.BookingID),
		Session:   req.Session,
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to cancel client booking")
		return fmt.Errorf("booking service: cancel client booking: remove booking: %w", err)
	}

	mask.Jitter(1000, 2000)
	if err = s.LoadClientBookings(ctx, &LoadClientBookingsReq{
		UserUUID: req.UserUUID,
		Session:  req.Session,
		Cookies:  req.Cookies,
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to cancel client booking")
		return fmt.Errorf("booking service: cancel client booking: load client bookings: %w", err)
	}

	return nil
}

func (s *Service) FilterSchedule(ctx context.Context, req *FilterScheduleReq) (domain.Schedule, error) {
	ctx, span := tracer.Start(ctx, "booking_service.filter_available_slots")
	defer span.End()

	schedule, err := s.repo.FilterSchedule(ctx, &DBSlotFilter{
		UserUUID: req.UserUUID,
		Schedule: req.Schedule,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to filter schedule")
		return nil, fmt.Errorf("booking service: filter schedule: repository call: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return schedule, nil
}
