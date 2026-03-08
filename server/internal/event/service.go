package event

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"labgrab/internal/shared/api/dikidi"
	"labgrab/internal/shared/mask"
	"labgrab/pkg/config"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("event-service")

type Service struct {
	repo   *Repo
	client *dikidi.Client
	cfg    *config.EncryptionConfig
	kekGCM cipher.AEAD
}

func NewService(repo *Repo, client *dikidi.Client, cfg *config.EncryptionConfig) (*Service, error) {
	key, err := hex.DecodeString(cfg.PasswordKEK)
	if err != nil {
		return nil, fmt.Errorf("event service: new service: decode kek key: %w", err)
	}

	kekCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("event service: new service: create aes cipher: %w", err)
	}

	kekGCM, err := cipher.NewGCMWithRandomNonce(kekCipher)
	if err != nil {
		return nil, fmt.Errorf("event service: new service: create gcm: %w", err)
	}

	return &Service{
		repo:   repo,
		client: client,
		cfg:    cfg,
		kekGCM: kekGCM,
	}, nil
}

func (s *Service) CreateUserData(ctx context.Context, req *CreateUserDataReq) error {
	ctx, span := tracer.Start(ctx, "event_service.create_user_data")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", req.UserUUID.String()))

	pass, dek, err := s.EncryptPassword(req.DikidiPassword, req.UserUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to encrypt password")
		return fmt.Errorf("event service: create user data: encrypt password: %w", err)
	}

	if err := s.repo.CreateUserData(ctx, &DBUserData{
		UserUUID:          req.UserUUID,
		DikidiPhoneNumber: req.DikidiPhoneNumber,
		DikidiPassword:    pass,
		DEK:               dek,
	}, req.Tx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create user data")
		return fmt.Errorf("event service: create user data: repository call: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *Service) AuthUser(ctx context.Context, userUUID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "event_service.auth_user")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	data, err := s.repo.GetUserData(ctx, userUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user data")
		return fmt.Errorf("event service: auth user: get user data: %w", err)
	}

	rawDEK, err := s.DecryptDEK(data.DEK, data.UserUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt dek")
		return fmt.Errorf("event service: auth user: decrypt dek: %w", err)
	}

	password, err := decryptWithDEK(data.DikidiPassword, rawDEK)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt password")
		return fmt.Errorf("event service: auth user: decrypt password: %w", err)
	}

	client := mask.CreateRandomHTTPClient()

	telegramCSRF, err := s.client.AcquireTelegramCSRFToken(ctx, client)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire telegram csrf token")
		return fmt.Errorf("event service: auth user: acquire telegram csrf token: %w", err)
	}

	mask.Jitter(2000, 3000)
	csrf, err := s.client.AcquireCSRFToken(ctx, client, dikidi.CSRFTokenRequest{
		PhoneNumber:       sanitizePhoneNumber(data.DikidiPhoneNumber),
		TelegramCSRFToken: telegramCSRF,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire csrf token")
		return fmt.Errorf("event service: auth user: acquire csrf token: %w", err)
	}

	mask.Jitter(5000, 8000)
	if err = s.client.SendAuthRequest(ctx, client, dikidi.AuthRequest{
		PhoneNumber:       sanitizePhoneNumber(data.DikidiPhoneNumber),
		Password:          password,
		CSRFToken:         csrf,
		TelegramCSRFToken: telegramCSRF,
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to send auth request")
		return fmt.Errorf("event service: auth user: send auth request: %w", err)
	}

	cookies, err := s.client.AcquireClientCookies(client)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire client cookies")
		return fmt.Errorf("event service: auth user: acquire client cookies: %w", err)
	}

	if cookies.CookieName == nil || cookies.Token == nil {
		err = fmt.Errorf("event service: auth user: acquire client cookies: cookie_name or token is missing")
		span.RecordError(err)
		span.SetStatus(codes.Error, "bad client cookies")
		return err
	}

	session, err := s.client.AcquireSessionID(*cookies.CookieName)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire session id")
		return fmt.Errorf("event service: auth user: acquire session id: %w", err)
	}

	if err = s.EncryptAndSaveCookies(ctx, data.UserUUID, rawDEK, cookies, session); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to encrypt and save cookies")
		return fmt.Errorf("event service: auth user: encrypt and save cookies: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *Service) RefreshUserCookies(ctx context.Context, userUUID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "event_service.refresh_user_cookies")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	eData, err := s.repo.GetUserData(ctx, userUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user data")
		return fmt.Errorf("event service: refresh user cookies: get user data: %w", err)
	}

	rawDEK, err := s.DecryptDEK(eData.DEK, eData.UserUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt dek")
		return fmt.Errorf("event service: refresh user cookies: decrypt dek: %w", err)
	}

	existingCookies, err := decryptPtrWithDEK(eData.Cookies, rawDEK)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt existing cookies")
		return fmt.Errorf("event service: refresh user cookies: decrypt cookies: %w", err)
	}

	client, err := mask.CreateClientWithCookies(existingCookies)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create http client with cookies")
		return fmt.Errorf("event service: refresh user cookies: create http client with cookies: %w", err)
	}

	if err = s.client.RenewCookies(ctx, client); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to renew cookies")
		return fmt.Errorf("event service: refresh user cookies: renew cookies: %w", err)
	}

	newCookies, err := s.client.AcquireClientCookies(client)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire client cookies")
		return fmt.Errorf("event service: refresh user cookies: acquire client cookies: %w", err)
	}

	if newCookies.CookieName == nil || newCookies.Token == nil {
		err = fmt.Errorf("event service: refresh user cookies: acquire client cookies: cookie_name or token is missing")
		span.RecordError(err)
		span.SetStatus(codes.Error, "bad client cookies")
		return err
	}

	session, err := s.client.AcquireSessionID(*newCookies.CookieName)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire session id")
		return fmt.Errorf("event service: refresh user cookies: acquire session id: %w", err)
	}

	if err = s.EncryptAndSaveCookies(ctx, eData.UserUUID, rawDEK, newCookies, session); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to encrypt and save cookies")
		return fmt.Errorf("event service: refresh user cookies: encrypt and save cookies: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *Service) Enroll(ctx context.Context, req *EnrollmentReq) (int, error) {
	ctx, span := tracer.Start(ctx, "event_service.enroll")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", req.UserUUID.String()))

	eData, err := s.repo.GetUserData(ctx, req.UserUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user data")
		return 0, fmt.Errorf("event service: enroll: get user data: %w", err)
	}

	data, err := s.DecryptUserData(eData)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt user data")
		return 0, fmt.Errorf("event service: enroll: decrypt user data: %w", err)
	}

	client, err := mask.CreateClientWithCookies(data.Cookies)
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
		Session:    *data.Session,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire time reservation")
		return 0, fmt.Errorf("event service: enroll: acquire time reservation: %w", err)
	}

	refererTime := req.Time.Format("200601021504")

	mask.Jitter(1000, 2000)
	if err = s.client.CheckEnrollment(ctx, client, &dikidi.EnrollmentCheckRequest{
		MasterID:   req.EventID,
		ServicesID: req.ServiceID,
		Time:       refererTime,
		RecordID:   reservation.BookingID,
		Session:    *data.Session,
		Phone:      sanitizePhoneNumber(data.DikidiPhoneNumber),
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
		Session:    *data.Session,
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
		Session:   *data.Session,
		Phone:     sanitizePhoneNumber(data.DikidiPhoneNumber),
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

func (s *Service) GetUserCredentials(ctx context.Context, userUUID uuid.UUID) (*GetUserCredentialsRes, error) {
	ctx, span := tracer.Start(ctx, "event_service.get_user_credentials")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	data, err := s.repo.GetUserData(ctx, userUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user data")
		return nil, fmt.Errorf("event service: get user credentials: get user data: %w", err)
	}

	rawDEK, err := s.DecryptDEK(data.DEK, data.UserUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt dek")
		return nil, fmt.Errorf("event service: get user credentials: decrypt dek: %w", err)
	}

	session, err := decryptPtrWithDEK(data.Session, rawDEK)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt session")
		return nil, fmt.Errorf("event service: get user credentials: decrypt session: %w", err)
	}

	token, err := decryptPtrWithDEK(data.Token, rawDEK)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt token")
		return nil, fmt.Errorf("event service: get user credentials: decrypt token: %w", err)
	}

	cookies, err := decryptPtrWithDEK(data.Cookies, rawDEK)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt cookies")
		return nil, fmt.Errorf("event service: get user credentials: decrypt cookies: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return &GetUserCredentialsRes{
		Session: session,
		Token:   token,
		Cookies: cookies,
	}, nil
}

func sanitizePhoneNumber(phoneNumber string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, phoneNumber)
}
