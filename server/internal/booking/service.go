package booking

import (
	"context"
	"labgrab/internal/shared/api/dikidi"
	"labgrab/internal/shared/domain"
	"labgrab/internal/shared/errors"
	"labgrab/internal/shared/mask"
	"strconv"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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

	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	bookings, err := s.repo.GetBookings(ctx, userUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "GetBookings",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to retrieve bookings from repository")
		return nil, err
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

	span.SetAttributes(attribute.Int("bookings.count", len(result)))
	span.SetStatus(codes.Ok, "")
	return result, nil
}

func (s *Service) LoadClientBookings(ctx context.Context, req *LoadClientBookingsReq) error {
	ctx, span := tracer.Start(ctx, "booking_service.load_client_bookings")
	defer span.End()

	client, err := mask.CreateClientWithCookies(&req.Cookies)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "LoadClientBookings",
			Step:      "Client initialization",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create http client")
		return err
	}

	bookings, err := s.client.GetBookings(ctx, client, &dikidi.GetBookingsRequest{
		Session: req.Session,
	})
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "LoadClientBookings",
			Step:      "Load",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to load bookings")
		return err
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

	err = s.repo.LoadBookings(ctx, dbBookings)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "LoadClientBookings",
			Step:      "Load",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to load bookings")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *Service) CancelClientBooking(ctx context.Context, req *CancelClientBookingReq) error {
	ctx, span := tracer.Start(ctx, "booking_service.cancel_client_booking")
	defer span.End()

	span.SetAttributes(attribute.Int("booking.id", req.BookingID))

	client, err := mask.CreateClientWithCookies(&req.Cookies)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "CancelClientBooking",
			Step:      "Client initialization",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create http client")
		return err
	}

	err = s.client.RemoveBooking(ctx, client, &dikidi.RemoveBookingRequest{
		BookingID: strconv.Itoa(req.BookingID),
		Session:   req.Session,
	})

	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "CancelClientBooking",
			Step:      "API call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to cancel booking in repository")
		return err
	}

	err = s.LoadClientBookings(ctx, &LoadClientBookingsReq{
		UserUUID: req.UserUUID,
		Session:  req.Session,
		Cookies:  req.Cookies,
	})

	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "CancelClientBooking",
			Step:      "Service call",
			Err:       err,
		}
	}

	return nil
}

func (s *Service) FilterSchedule(ctx context.Context, req *FilterScheduleReq) (domain.Schedule, error) {
	ctx, span := tracer.Start(ctx, "booking_service.filter_available_slots")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUUID.String()),
		attribute.Int("schedule.input_count", len(req.Schedule)),
	)

	schedule, err := s.repo.FilterSchedule(ctx, &DBSlotFilter{
		UserUUID: req.UserUUID,
		Schedule: req.Schedule,
	})
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "FilterAvailableSlots",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to filter available slots in repository")
		return nil, err
	}

	span.SetAttributes(attribute.Int("slots.output_count", len(schedule)))
	span.SetStatus(codes.Ok, "")
	return schedule, nil
}
