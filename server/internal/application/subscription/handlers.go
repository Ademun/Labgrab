package subscription

import (
	"encoding/json"
	"labgrab/internal/auth"
	"labgrab/internal/shared/apperr"
	"labgrab/internal/subscription"
	"net/http"

	"labgrab/internal/application/subscription/dto"
	"labgrab/internal/application/subscription/usecase"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Handler struct {
	getSubscriptions      *usecase.GetSubscriptionsUsecase
	newSubscription       *usecase.NewSubscriptionUsecase
	editSubscription      *usecase.EditSubscriptionUsecase
	getTimePreferences    *usecase.GetTimePreferencesUsecase
	setTimePreferences    *usecase.SetTimePreferencesUsecase
	getTeacherPreferences *usecase.GetTeacherPreferencesUsecase
	setTeacherPreferences *usecase.SetTeacherPreferencesUsecase
	logger                *zap.SugaredLogger
}

func NewHandler(authSvc *auth.Service, subscriptionSvc *subscription.Service, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		getSubscriptions:      &usecase.GetSubscriptionsUsecase{AuthSvc: authSvc, SubscriptionSvc: subscriptionSvc},
		newSubscription:       &usecase.NewSubscriptionUsecase{AuthSvc: authSvc, SubscriptionSvc: subscriptionSvc},
		editSubscription:      &usecase.EditSubscriptionUsecase{AuthSvc: authSvc, SubscriptionSvc: subscriptionSvc},
		getTimePreferences:    &usecase.GetTimePreferencesUsecase{AuthSvc: authSvc, SubscriptionSvc: subscriptionSvc},
		setTimePreferences:    &usecase.SetTimePreferencesUsecase{AuthSvc: authSvc, SubscriptionSvc: subscriptionSvc},
		getTeacherPreferences: &usecase.GetTeacherPreferencesUsecase{AuthSvc: authSvc, SubscriptionSvc: subscriptionSvc},
		setTeacherPreferences: &usecase.SetTeacherPreferencesUsecase{AuthSvc: authSvc, SubscriptionSvc: subscriptionSvc},
		logger:                logger,
	}
}

func (h *Handler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "subscription_handler.get_subscriptions")
	defer span.End()

	vars := mux.Vars(r)
	subscriptionUUID := vars["id"]

	var subscriptionUUIDPtr *string
	if subscriptionUUID != "" {
		subscriptionUUIDPtr = &subscriptionUUID
		span.SetAttributes(attribute.String("subscription.uuid", subscriptionUUID))
	}

	session, ok := h.sessionFromCookie(w, r, span, "subscription handler: get subscriptions")
	if !ok {
		return
	}

	resp, err := h.getSubscriptions.Exec(ctx, &dto.GetSubscriptionsReqDTO{
		Session:          session,
		SubscriptionUUID: subscriptionUUIDPtr,
	})
	if err != nil {
		h.logger.Errorf("subscription handler: get subscriptions: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	span.SetAttributes(attribute.Int("response.count", len(resp)))
	span.SetStatus(codes.Ok, "")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("subscription handler: get subscriptions: encode response: %v", err)
	}
}

func (h *Handler) NewSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "subscription_handler.new_subscription")
	defer span.End()

	var req dto.NewSubscriptionReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warnf("subscription handler: new subscription: failed to decode body: %v", err)
		span.SetStatus(codes.Error, "invalid request payload")
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("lab.type", req.LabType),
		attribute.String("lab.topic", req.LabTopic),
		attribute.Int("lab.number", req.LabNumber),
	)

	session, ok := h.sessionFromCookie(w, r, span, "subscription handler: new subscription")
	if !ok {
		return
	}

	req.Session = session

	resp, err := h.newSubscription.Exec(ctx, &req)
	if err != nil {
		h.logger.Errorf("subscription handler: new subscription: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	span.SetAttributes(attribute.String("subscription.uuid", resp.UUID))
	span.SetStatus(codes.Ok, "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("subscription handler: new subscription: encode response: %v", err)
	}
}

func (h *Handler) EditSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "subscription_handler.edit_subscription")
	defer span.End()

	vars := mux.Vars(r)
	subscriptionUUID := vars["id"]
	span.SetAttributes(attribute.String("subscription.uuid", subscriptionUUID))

	var req dto.EditSubscriptionReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warnf("subscription handler: edit subscription: failed to decode body: %v", err)
		span.SetStatus(codes.Error, "invalid request payload")
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	session, ok := h.sessionFromCookie(w, r, span, "subscription handler: edit subscription")
	if !ok {
		return
	}

	req.Session = session
	req.SubscriptionUUID = subscriptionUUID

	resp, err := h.editSubscription.Exec(ctx, &req)
	if err != nil {
		h.logger.Errorf("subscription handler: edit subscription: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	span.SetStatus(codes.Ok, "")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("subscription handler: edit subscription: encode response: %v", err)
	}
}

func (h *Handler) GetTimePreferences(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "subscription_handler.get_time_preferences")
	defer span.End()

	session, ok := h.sessionFromCookie(w, r, span, "subscription handler: get time preferences")
	if !ok {
		return
	}

	resp, err := h.getTimePreferences.Exec(ctx, &dto.GetTimePreferencesReqDTO{Session: session})
	if err != nil {
		h.logger.Errorf("subscription handler: get time preferences: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	span.SetStatus(codes.Ok, "")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("subscription handler: get time preferences: encode response: %v", err)
	}
}

func (h *Handler) SetTimePreferences(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "subscription_handler.set_time_preferences")
	defer span.End()

	var req dto.SetTimePreferncesReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warnf("subscription handler: set time preferences: failed to decode body: %v", err)
		span.SetStatus(codes.Error, "invalid request payload")
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	session, ok := h.sessionFromCookie(w, r, span, "subscription handler: set time preferences")
	if !ok {
		return
	}

	req.Session = session

	if err := h.setTimePreferences.Exec(ctx, &req); err != nil {
		h.logger.Errorf("subscription handler: set time preferences: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	span.SetStatus(codes.Ok, "")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetTeacherPreferences(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "subscription_handler.get_teacher_preferences")
	defer span.End()

	session, ok := h.sessionFromCookie(w, r, span, "subscription handler: get teacher preferences")
	if !ok {
		return
	}

	resp, err := h.getTeacherPreferences.Exec(ctx, &dto.GetTeacherPreferencesReqDTO{Session: session})
	if err != nil {
		h.logger.Errorf("subscription handler: get teacher preferences: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	span.SetStatus(codes.Ok, "")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("subscription handler: get teacher preferences: encode response: %v", err)
	}
}

func (h *Handler) SetTeacherPreferences(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "subscription_handler.set_teacher_preferences")
	defer span.End()

	var req dto.SetTeacherPreferencesReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warnf("subscription handler: set teacher preferences: failed to decode body: %v", err)
		span.SetStatus(codes.Error, "invalid request payload")
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	session, ok := h.sessionFromCookie(w, r, span, "subscription handler: set teacher preferences")
	if !ok {
		return
	}

	req.Session = session

	if err := h.setTeacherPreferences.Exec(ctx, &req); err != nil {
		h.logger.Errorf("subscription handler: set teacher preferences: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	span.SetStatus(codes.Ok, "")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/subscriptions", h.GetSubscriptions).Methods(http.MethodGet)
	r.HandleFunc("/api/subscriptions", h.NewSubscription).Methods(http.MethodPost)
	r.HandleFunc("/api/subscriptions/{id}", h.GetSubscriptions).Methods(http.MethodGet)
	r.HandleFunc("/api/subscriptions/{id}", h.EditSubscription).Methods(http.MethodPatch)
	r.HandleFunc("/api/subscriptions/preferences/time", h.GetTimePreferences).Methods(http.MethodGet)
	r.HandleFunc("/api/subscriptions/preferences/time", h.SetTimePreferences).Methods(http.MethodPost)
	r.HandleFunc("/api/subscription/preferences/teachers", h.GetTeacherPreferences).Methods(http.MethodGet)
	r.HandleFunc("/api/subscription/preferences/teachers", h.SetTeacherPreferences).Methods(http.MethodPost)
}

func (h *Handler) sessionFromCookie(w http.ResponseWriter, r *http.Request, span trace.Span, logPrefix string) (string, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		h.logger.Warnf("%s: missing session cookie: %v", logPrefix, err)
		span.SetStatus(codes.Error, "unauthorized: missing session cookie")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return "", false
	}
	return cookie.Value, true
}
