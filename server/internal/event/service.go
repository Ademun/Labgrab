package event

import (
	"context"
	"fmt"
	"labgrab/internal/shared/api/dikidi"
	"labgrab/internal/shared/mask"
	"strings"
	"unicode"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("event-service")

type Service struct {
	client *dikidi.Client
}

func NewService(client *dikidi.Client) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) Enroll(ctx context.Context, req *EnrollmentReq) (int, error) {
	ctx, span := tracer.Start(ctx, "event_service.enroll")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", req.UserUUID.String()))

	client, err := mask.CreateClientWithCookies(&req.Cookies)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create http client with cookies")
		return 0, fmt.Errorf("event service: enroll: create http client with cookies: %w", err)
	}

	timeStr := req.Time.Format("2006-01-02 15:04:05")
	reservation, err := s.client.AcquireTimeReservation(ctx, client, &dikidi.EventReservationRequest{
		EventID:    req.EventID,
		ServicesID: req.ServiceID,
		Time:       timeStr,
		Session:    req.Session,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire time reservation")
		return 0, fmt.Errorf("event service: enroll: acquire time reservation: %w", err)
	}

	refererTime := req.Time.Format("200601021504")
	phone := sanitizePhoneNumber(req.PhoneNumber)

	mask.Jitter(1000, 2000)

	if err = s.client.CheckEnrollment(ctx, client, &dikidi.EnrollmentCheckRequest{
		MasterID:   req.EventID,
		ServicesID: req.ServiceID,
		Time:       refererTime,
		RecordID:   reservation.BookingID,
		Session:    req.Session,
		Phone:      phone,
		FirstName:  fmt.Sprintf("%s %s", req.Name, req.Patronymic),
		LastName:   req.Surname,
		Comments:   req.Group,
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to check enrollment")
		return 0, fmt.Errorf("event service: enroll: check enrollment: %w", err)
	}

	mask.Jitter(500, 1000)

	if err = s.client.GetReservationInfo(ctx, client, &dikidi.ReservationInfoRequest{
		BookingID:  reservation.BookingID,
		MasterID:   req.EventID,
		ServicesID: req.ServiceID,
		Time:       refererTime,
		Session:    req.Session,
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get reservation info")
		return 0, fmt.Errorf("event service: enroll: get reservation info: %w", err)
	}

	mask.Jitter(1000, 2000)

	recordID, err := s.client.CreateBooking(ctx, client, &dikidi.CreateBookingRequest{
		EventID:   req.EventID,
		ServiceID: req.ServiceID,
		Time:      refererTime,
		BookingID: reservation.BookingID,
		Session:   req.Session,
		Phone:     phone,
		FirstName: fmt.Sprintf("%s %s", req.Name, req.Patronymic),
		LastName:  req.Surname,
		Comments:  req.Group,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create booking")
		return 0, fmt.Errorf("event service: enroll: create booking: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return recordID, nil
}

func (s *Service) GetCurrentEvents(ctx context.Context, clientCookies *string) (chan *GetEventsRes, error) {
	ctx, span := tracer.Start(ctx, "event_service.get_current_events")

	client, err := mask.CreateClientWithCookies(clientCookies)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create http client with cookies")
		return nil, fmt.Errorf("event service: get current events: create http client: %w", err)
	}

	ch := make(chan *GetEventsRes)
	go func() {
		defer span.End()
		for event := range s.client.GetEventStream(ctx, client) {
			ch <- &GetEventsRes{
				Data: event.Event,
				Err:  event.Error,
			}
		}
		close(ch)
	}()

	span.SetStatus(codes.Ok, "")
	return ch, nil
}

func (s *Service) UpdateServiceIDs(ctx context.Context, clientCookies *string) error {
	ctx, span := tracer.Start(ctx, "event_service.update_service_ids")
	defer span.End()

	client, err := mask.CreateClientWithCookies(clientCookies)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create http client with cookies")
		return fmt.Errorf("event service: update service ids: create http client: %w", err)
	}

	if err := s.client.UpdateServiceIDs(ctx, client); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update service ids")
		return fmt.Errorf("event service: update service ids: client call: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func sanitizePhoneNumber(phoneNumber string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, phoneNumber)
}
