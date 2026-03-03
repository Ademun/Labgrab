package event

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"labgrab/internal/shared/api/dikidi"
	"labgrab/internal/shared/errors"
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
		return nil, err
	}

	kekCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	kekGCM, err := cipher.NewGCMWithRandomNonce(kekCipher)
	if err != nil {
		return nil, err
	}

	return &Service{
		repo:   repo,
		client: client,
		cfg:    cfg,
		kekGCM: kekGCM,
	}, nil
}

func (s *Service) CreateUserData(ctx context.Context, req *CreateUserDataReq) error {
	ctx, span := tracer.Start(ctx, "lab_enrollment_service.create_user_data")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUUID.String()),
	)

	pass, dek, err := s.EncryptPassword(req.DikidiPassword, req.UserUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "CreateUserData",
			Step:      "EncryptPassword",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to encrypt password")
		return err
	}

	dbUserData := &DBUserData{
		UserUUID:          req.UserUUID,
		DikidiPhoneNumber: req.DikidiPhoneNumber,
		DikidiPassword:    pass,
		DEK:               dek,
	}

	if err := s.repo.CreateUserData(ctx, dbUserData, req.Tx); err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "CreateUserData",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create user data in repository")
		return err
	}

	return nil
}

func (s *Service) AuthUser(ctx context.Context, userUUID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "lab_enrollment_service.auth_user")
	defer span.End()
	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	data, err := s.repo.GetUserData(ctx, userUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "AuthUser", Step: "GetUserData", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user data")
		return err
	}

	rawDEK, err := s.DecryptDEK(data.DEK, data.UserUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "AuthUser", Step: "DecryptDEK", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt DEK")
		return err
	}

	password, err := decryptWithDEK(data.DikidiPassword, rawDEK)
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "AuthUser", Step: "DecryptPassword", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt password")
		return err
	}

	client := mask.CreateRandomHTTPClient()

	telegramCSRF, err := s.client.AcquireTelegramCSRFToken(ctx, client)
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "AuthUser", Step: "AcquireTelegramCSRFToken", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire telegram csrf")
		return err
	}

	mask.Jitter(2000, 3000)
	csrf, err := s.client.AcquireCSRFToken(ctx, client, dikidi.CSRFTokenRequest{
		PhoneNumber:       sanitizePhoneNumber(data.DikidiPhoneNumber),
		TelegramCSRFToken: telegramCSRF,
	})
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "AuthUser", Step: "AcquireCSRFToken", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire CSRF token")
		return err
	}

	mask.Jitter(5000, 8000)
	if err = s.client.SendAuthRequest(ctx, client, dikidi.AuthRequest{
		PhoneNumber:       sanitizePhoneNumber(data.DikidiPhoneNumber),
		Password:          password,
		CSRFToken:         csrf,
		TelegramCSRFToken: telegramCSRF,
	}); err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "AuthUser", Step: "SendAuthRequest", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to send auth request")
		return err
	}

	cookies, err := s.client.AcquireClientCookies(client)
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "AuthUser", Step: "AcquireClientCookies", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire client cookies")
		return err
	}

	if cookies.CookieName == nil || cookies.Token == nil {
		err = &errors.ErrServiceProcedure{Procedure: "AuthUser", Step: "AcquireClientCookies", Err: fmt.Errorf("no cookie_name or token found")}
		span.RecordError(err)
		span.SetStatus(codes.Error, "bad client cookies")
		return err
	}

	session, err := s.client.AcquireSessionID(*cookies.CookieName)
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "AuthUser", Step: "AcquireSessionID", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire session ID")
		return err
	}

	if err = s.EncryptAndSaveCookies(ctx, data.UserUUID, rawDEK, cookies, session); err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "AuthUser", Step: "encryptAndSaveCookies", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to encrypt and save cookies")
		return err
	}

	return nil
}

func (s *Service) RefreshUserCookies(ctx context.Context, userUUID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "lab_enrollment_service.refresh_user_cookies")
	defer span.End()
	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	eData, err := s.repo.GetUserData(ctx, userUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "RefreshUserCookies", Step: "GetUserData", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user data")
		return err
	}

	rawDEK, err := s.DecryptDEK(eData.DEK, eData.UserUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "RefreshUserCookies", Step: "DecryptDEK", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt DEK")
		return err
	}

	existingCookies, err := decryptPtrWithDEK(eData.Cookies, rawDEK)
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "RefreshUserCookies", Step: "DecryptCookies", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt existing cookies")
		return err
	}

	client, err := mask.CreateClientWithCookies(existingCookies)
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "RefreshUserCookies", Step: "CreateClientWithCookies", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create http client with cookies")
		return err
	}

	if err = s.client.RenewCookies(ctx, client); err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "RefreshUserCookies", Step: "RenewCookies", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to renew cookies")
		return err
	}

	newCookies, err := s.client.AcquireClientCookies(client)
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "RefreshUserCookies", Step: "AcquireClientCookies", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire client cookies")
		return err
	}

	if newCookies.CookieName == nil || newCookies.Token == nil {
		err = &errors.ErrServiceProcedure{Procedure: "RefreshUserCookies", Step: "AcquireClientCookies", Err: fmt.Errorf("no cookie_name or token found")}
		span.RecordError(err)
		span.SetStatus(codes.Error, "bad client cookies")
		return err
	}

	session, err := s.client.AcquireSessionID(*newCookies.CookieName)
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "RefreshUserCookies", Step: "AcquireSessionID", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire session ID")
		return err
	}

	if err = s.EncryptAndSaveCookies(ctx, eData.UserUUID, rawDEK, newCookies, session); err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "RefreshUserCookies", Step: "encryptAndSaveCookies", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to encrypt and save cookies")
		return err
	}

	return nil
}

func (s *Service) Enroll(ctx context.Context, req *EnrollmentReq) (int, error) {
	ctx, span := tracer.Start(ctx, "lab_enrollment_service.Enroll")
	defer span.End()
	span.SetAttributes(attribute.String("user.uuid", req.UserUUID.String()))

	eData, err := s.repo.GetUserData(ctx, req.UserUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "Enroll", Step: "GetUserData", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user data")
		return 0, err
	}

	data, err := s.DecryptUserData(eData)
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "Enroll", Step: "DecryptUserData", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt user data")
		return 0, err
	}

	client, err := mask.CreateClientWithCookies(data.Cookies)
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "Enroll", Step: "CreateClientWithCookies", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create http client with cookies")
		return 0, err
	}

	timeStr := req.Time.Format("2006-01-02 15:04:05")
	reservation, err := s.client.AcquireTimeReservation(ctx, client, &dikidi.EventReservationRequest{
		EventID:    req.EventID,
		ServicesID: req.ServiceID,
		Time:       timeStr,
		Session:    *data.Session,
	})
	if err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "Enroll", Step: "AcquireTimeReservation", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire time reservation")
		return 0, err
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
		err = &errors.ErrServiceProcedure{Procedure: "Enroll", Step: "CheckEnrollment", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to check enrollment")
		return 0, err
	}

	mask.Jitter(500, 1500)

	if err = s.client.GetReservationInfo(ctx, client, &dikidi.ReservationInfoRequest{
		BookingID:  reservation.BookingID,
		MasterID:   req.EventID,
		ServicesID: req.ServiceID,
		Time:       refererTime,
		Session:    *data.Session,
	}); err != nil {
		err = &errors.ErrServiceProcedure{Procedure: "Enroll", Step: "GetReservationInfo", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get reservation info")
		return 0, err
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
		err = &errors.ErrServiceProcedure{Procedure: "Enroll", Step: "CreateRecord", Err: err}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create booking")
		return 0, err
	}

	fmt.Println("Succesfully enrolled")

	return recordID, nil
}

func sanitizePhoneNumber(phoneNumber string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, phoneNumber)
}

//func (s *Service) GetBookings(ctx context.Context, userUUID uuid.UUID) ([]domain.Booking, error) {
//	ctx, span := tracer.Start(ctx, "lab_enrollment_service.GetRecords")
//	defer span.End()
//	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))
//
//	eData, err := s.repo.GetUserData(ctx, userUUID)
//	if err != nil {
//		err = &errors.ErrServiceProcedure{Procedure: "GetRecords", Step: "GetUserData", Err: err}
//		span.RecordError(err)
//		span.SetStatus(codes.Error, "failed to get user data")
//		return nil, err
//	}
//
//	data, err := s.DecryptUserData(eData)
//	if err != nil {
//		err = &errors.ErrServiceProcedure{Procedure: "GetRecords", Step: "DecryptUserData", Err: err}
//		span.RecordError(err)
//		span.SetStatus(codes.Error, "failed to decrypt user data")
//		return nil, err
//	}
//
//	client, err := s.CreateClientWithCookies(data.Cookies)
//	if err != nil {
//		err = &errors.ErrServiceProcedure{Procedure: "GetRecords", Step: "CreateClientWithCookies", Err: err}
//		span.RecordError(err)
//		span.SetStatus(codes.Error, "failed to create http client with cookies")
//		return nil, err
//	}
//
//	result, err := s.client.GetBookings(ctx, client, &dikidi.GetBookingsRequest{
//		Session: *data.Session,
//	})
//	if err != nil {
//		err = &errors.ErrServiceProcedure{Procedure: "GetRecords", Step: "GetRecords", Err: err}
//		span.RecordError(err)
//		span.SetStatus(codes.Error, "failed to get records")
//		return nil, err
//	}
//
//	return &GetRecordsRes{
//		New: mapToRecordItems(result.New),
//		Old: mapToRecordItems(result.Old),
//	}, nil
//}
//
//func (s *Service) RemoveBooking(ctx context.Context, req *RemoveRecordReq) error {
//	ctx, span := tracer.Start(ctx, "lab_enrollment_service.RemoveRecord")
//	defer span.End()
//	span.SetAttributes(
//		attribute.String("user.uuid", req.UserUUID.String()),
//		attribute.String("booking.id", req.RecordID),
//	)
//
//	eData, err := s.repo.GetUserData(ctx, req.UserUUID)
//	if err != nil {
//		err = &errors.ErrServiceProcedure{Procedure: "RemoveRecord", Step: "GetUserData", Err: err}
//		span.RecordError(err)
//		span.SetStatus(codes.Error, "failed to get user data")
//		return err
//	}
//
//	data, err := s.DecryptUserData(eData)
//	if err != nil {
//		err = &errors.ErrServiceProcedure{Procedure: "RemoveRecord", Step: "DecryptUserData", Err: err}
//		span.RecordError(err)
//		span.SetStatus(codes.Error, "failed to decrypt user data")
//		return err
//	}
//
//	client, err := s.CreateClientWithCookies(data.Cookies)
//	if err != nil {
//		err = &errors.ErrServiceProcedure{Procedure: "RemoveRecord", Step: "CreateClientWithCookies", Err: err}
//		span.RecordError(err)
//		span.SetStatus(codes.Error, "failed to create http client with cookies")
//		return err
//	}
//
//	if err = s.client.RemoveBooking(ctx, client, &dikidi.RemoveBookingRequest{
//		BookingID: req.RecordID,
//		Session:   *data.Session,
//	}); err != nil {
//		err = &errors.ErrServiceProcedure{Procedure: "RemoveRecord", Step: "RemoveRecord", Err: err}
//		span.RecordError(err)
//		span.SetStatus(codes.Error, "failed to remove booking")
//		return err
//	}
//
//	return nil
//}
