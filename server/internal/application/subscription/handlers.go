package subscription

import (
	"encoding/json"
	"errors"
	"fmt"
	"labgrab/internal/auth"
	"labgrab/internal/subscription"
	"net/http"

	"labgrab/internal/application/subscription/dto"
	"labgrab/internal/application/subscription/usecase"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type Handler struct {
	getSubscriptions *usecase.GetSubscriptionsUseCase
	newSubscription  *usecase.NewSubscriptionUseCase
	editSubscription *usecase.EditSubscriptionUseCase
	logger           *zap.SugaredLogger
}

func NewHandler(authSvc *auth.Service, subscriptionSvc *subscription.Service,
	logger *zap.SugaredLogger,
) *Handler {
	return &Handler{
		getSubscriptions: usecase.NewGetSubscriptionsUseCase(authSvc, subscriptionSvc, logger),
		newSubscription:  usecase.NewNewSubscriptionUseCase(authSvc, subscriptionSvc, logger),
		editSubscription: usecase.NewEditSubscriptionUseCase(authSvc, subscriptionSvc, logger),
		logger:           logger,
	}
}

func (h *Handler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "subscription_handler.get_subscriptions")
	defer span.End()

	subscriptionUUID := r.URL.Query().Get("subscription_uuid")
	var subscriptionUUIDPtr *string
	if subscriptionUUID != "" {
		subscriptionUUIDPtr = &subscriptionUUID
		span.SetAttributes(attribute.String("subscription.uuid", subscriptionUUID))
	}

	req := &dto.GetSubscriptionsReqDTO{
		SubscriptionUUID: subscriptionUUIDPtr,
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			span.RecordError(err)
			span.SetStatus(codes.Error, "unauthorized: missing session cookie")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			h.logger.Error(err)
			return
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read session cookie")
		http.Error(w, "Failed to read cookie", http.StatusInternalServerError)
		h.logger.Error(err)
		return
	}

	resp, err := h.getSubscriptions.Exec(ctx, cookie.Value, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error(err)
		return
	}

	span.SetAttributes(attribute.Int("response.count", len(resp)))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		err = fmt.Errorf("failed to write response: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error(err)
		return
	}

	span.SetStatus(codes.Ok, "")
}

func (h *Handler) NewSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "subscription_handler.new_subscription")
	defer span.End()

	var req dto.NewSubscriptionReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = fmt.Errorf("failed to decode request: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		h.logger.Error(err)
		return
	}

	span.SetAttributes(
		attribute.String("lab.type", req.LabType),
		attribute.String("lab.topic", req.LabTopic),
		attribute.Int("lab.number", req.LabNumber),
	)

	cookie, err := r.Cookie("session_id")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			span.RecordError(err)
			span.SetStatus(codes.Error, "unauthorized: missing session cookie")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			h.logger.Error(err)
			return
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read session cookie")
		http.Error(w, "Failed to read cookie", http.StatusInternalServerError)
		h.logger.Error(err)
		return
	}

	resp, err := h.newSubscription.Exec(ctx, cookie.Value, &req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error(err)
		return
	}

	span.SetAttributes(attribute.String("subscription.uuid", resp.String()))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		err = fmt.Errorf("failed to write response: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error(err)
		return
	}

	span.SetStatus(codes.Ok, "")
}

func (h *Handler) EditSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "subscription_handler.edit_subscription")
	defer span.End()

	vars := mux.Vars(r)
	subscriptionUUID := vars["id"]

	cookie, err := r.Cookie("session_id")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			span.RecordError(err)
			span.SetStatus(codes.Error, "unauthorized: missing session cookie")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			h.logger.Error(err)
			return
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read session cookie")
		http.Error(w, "Failed to read cookie", http.StatusInternalServerError)
		h.logger.Error(err)
		return
	}

	span.SetAttributes(
		attribute.String("subscription.uuid", subscriptionUUID),
	)

	var req dto.EditSubscriptionReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = fmt.Errorf("failed to decode request: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	req.SubscriptionUUID = subscriptionUUID

	resp, err := h.editSubscription.Exec(ctx, cookie.Value, &req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		err = fmt.Errorf("failed to write response: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "")
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/subscriptions", h.GetSubscriptions).Methods(http.MethodGet)
	r.HandleFunc("/api/subscriptions", h.NewSubscription).Methods(http.MethodPost)
	r.HandleFunc("/api/subscriptions/{id}", h.GetSubscriptions).Methods(http.MethodGet)
	r.HandleFunc("/api/subscriptions/{id}", h.EditSubscription).Methods(http.MethodPatch)
}
