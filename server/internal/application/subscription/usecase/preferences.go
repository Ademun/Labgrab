package usecase

import (
	"context"
	"labgrab/internal/application/subscription/dto"
	"labgrab/internal/auth"
	"labgrab/internal/subscription"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// GetTimePreferencesUseCase handles retrieval of time preferences
type GetTimePreferencesUseCase struct {
	authSvc         *auth.Service
	subscriptionSvc *subscription.Service
}

func NewGetTimePreferencesUseCase(authSvc *auth.Service, subscriptionSvc *subscription.Service) *GetTimePreferencesUseCase {
	return &GetTimePreferencesUseCase{
		authSvc:         authSvc,
		subscriptionSvc: subscriptionSvc,
	}
}

func (uc *GetTimePreferencesUseCase) Exec(ctx context.Context, session string) (*dto.GetTimePreferncesResDTO, error) {
	ctx, span := tracer.Start(ctx, "subscription_usecase.get_time_preferences")
	defer span.End()

	if err := uc.authSvc.ValidateSession(ctx, session); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "session validation failed")
		return nil, err
	}

	userUUID, err := uc.authSvc.GetSessionData(ctx, session)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to retrieve session data")
		return nil, err
	}

	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	preferences, err := uc.subscriptionSvc.GetTimePreferences(ctx, userUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get time preferences")
		return nil, err
	}

	span.SetAttributes(attribute.Int("preferences.weeks_count", len(preferences)))
	span.SetStatus(codes.Ok, "")

	return &dto.GetTimePreferncesResDTO{
		Preferences: preferences,
	}, nil
}

// SetTimePreferencesUseCase handles setting of time preferences
type SetTimePreferencesUseCase struct {
	authSvc         *auth.Service
	subscriptionSvc *subscription.Service
}

func NewSetTimePreferencesUseCase(authSvc *auth.Service, subscriptionSvc *subscription.Service) *SetTimePreferencesUseCase {
	return &SetTimePreferencesUseCase{
		authSvc:         authSvc,
		subscriptionSvc: subscriptionSvc,
	}
}

func (uc *SetTimePreferencesUseCase) Exec(ctx context.Context, session string, data *dto.SetTimePreferncesReqDTO) error {
	ctx, span := tracer.Start(ctx, "subscription_usecase.set_time_preferences")
	defer span.End()

	if err := uc.authSvc.ValidateSession(ctx, session); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "session validation failed")
		return err
	}

	userUUID, err := uc.authSvc.GetSessionData(ctx, session)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to retrieve session data")
		return err
	}

	span.SetAttributes(
		attribute.String("user.uuid", userUUID.String()),
		attribute.Int("preferences.weeks_count", len(data.Preferences)),
	)

	if err := uc.subscriptionSvc.SetTimePreferences(ctx, userUUID, data.Preferences); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to set time preferences")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// GetTeacherPreferencesUseCase handles retrieval of teacher preferences
type GetTeacherPreferencesUseCase struct {
	authSvc         *auth.Service
	subscriptionSvc *subscription.Service
}

func NewGetTeacherPreferencesUseCase(authSvc *auth.Service, subscriptionSvc *subscription.Service) *GetTeacherPreferencesUseCase {
	return &GetTeacherPreferencesUseCase{
		authSvc:         authSvc,
		subscriptionSvc: subscriptionSvc,
	}
}

func (uc *GetTeacherPreferencesUseCase) Exec(ctx context.Context, session string) (*dto.GetTeacherPreferencesResDTO, error) {
	ctx, span := tracer.Start(ctx, "subscription_usecase.get_teacher_preferences")
	defer span.End()

	if err := uc.authSvc.ValidateSession(ctx, session); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "session validation failed")
		return nil, err
	}

	userUUID, err := uc.authSvc.GetSessionData(ctx, session)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to retrieve session data")
		return nil, err
	}

	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	preferences, err := uc.subscriptionSvc.GetTeacherPreferences(ctx, userUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get teacher preferences")
		return nil, err
	}

	span.SetAttributes(attribute.Int("preferences.blacklisted_count", len(preferences)))
	span.SetStatus(codes.Ok, "")

	return &dto.GetTeacherPreferencesResDTO{
		Preferences: preferences,
	}, nil
}

// SetTeacherPreferencesUseCase handles setting of teacher preferences
type SetTeacherPreferencesUseCase struct {
	authSvc         *auth.Service
	subscriptionSvc *subscription.Service
}

func NewSetTeacherPreferencesUseCase(authSvc *auth.Service, subscriptionSvc *subscription.Service) *SetTeacherPreferencesUseCase {
	return &SetTeacherPreferencesUseCase{
		authSvc:         authSvc,
		subscriptionSvc: subscriptionSvc,
	}
}

func (uc *SetTeacherPreferencesUseCase) Exec(ctx context.Context, session string, data *dto.SetTeacherPreferencesReqDTO) error {
	ctx, span := tracer.Start(ctx, "subscription_usecase.set_teacher_preferences")
	defer span.End()

	if err := uc.authSvc.ValidateSession(ctx, session); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "session validation failed")
		return err
	}

	userUUID, err := uc.authSvc.GetSessionData(ctx, session)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to retrieve session data")
		return err
	}

	span.SetAttributes(
		attribute.String("user.uuid", userUUID.String()),
		attribute.Int("preferences.blacklisted_count", len(data.Preferences)),
	)

	if err := uc.subscriptionSvc.SetTeacherPreferences(ctx, userUUID, data.Preferences); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to set teacher preferences")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
