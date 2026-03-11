package event

import (
	"context"
	"fmt"
	"labgrab/internal/shared/api/dikidi"
	"labgrab/internal/shared/mask"
	"labgrab/internal/shared/sanitizing"
)

type Service struct {
	client *dikidi.Client
}

func NewService(client *dikidi.Client) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) Enroll(ctx context.Context, req *EnrollmentReq) (int, error) {
	client, err := mask.CreateClientWithCookies(&req.Cookies)
	if err != nil {
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
		return 0, fmt.Errorf("event service: enroll: acquire time reservation: %w", err)
	}

	refererTime := req.Time.Format("200601021504")
	phone := sanitizing.SanitizePhoneNumber(req.PhoneNumber)

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
		return 0, fmt.Errorf("event service: enroll: create booking: %w", err)
	}

	return recordID, nil
}

func (s *Service) GetCurrentEvents(ctx context.Context, clientCookies *string) (chan *GetEventsRes, error) {
	client, err := mask.CreateClientWithCookies(clientCookies)
	if err != nil {
		return nil, fmt.Errorf("event service: get current events: create http client: %w", err)
	}

	ch := make(chan *GetEventsRes)
	go func() {
		for event := range s.client.GetEventStream(ctx, client) {
			ch <- &GetEventsRes{
				Data: event.Event,
				Err:  event.Error,
			}
		}
		close(ch)
	}()

	return ch, nil
}

func (s *Service) UpdateServiceIDs(ctx context.Context, clientCookies *string) error {
	client, err := mask.CreateClientWithCookies(clientCookies)
	if err != nil {
		return fmt.Errorf("event service: update service ids: create http client: %w", err)
	}

	if err := s.client.UpdateServiceIDs(ctx, client); err != nil {
		return fmt.Errorf("event service: update service ids: client call: %w", err)
	}
	return nil
}
