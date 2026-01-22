package subscription

import (
	"encoding/json"
	"net/http"

	"labgrab/internal/application/subscription/dto"
	"labgrab/internal/application/subscription/usecase"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("subscription-handler")

type Handler struct {
	getSubscriptions *usecase.GetSubscriptionsUseCase
	newSubscription  *usecase.NewSubscriptionUseCase
	editSubscription *usecase.EditSubscriptionUseCase
	logger           *zap.SugaredLogger
}

func NewHandler(
	getSubscriptions *usecase.GetSubscriptionsUseCase,
	newSubscription *usecase.NewSubscriptionUseCase,
	editSubscription *usecase.EditSubscriptionUseCase,
	logger *zap.SugaredLogger,
) *Handler {
	return &Handler{
		getSubscriptions: getSubscriptions,
		newSubscription:  newSubscription,
		editSubscription: editSubscription,
		logger:           logger,
	}
}

func (h *Handler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "subscription.handler.GetSubscriptions")
	defer span.End()

	vars := mux.Vars(r)
	userUUID := vars["user_uuid"]

	subscriptionUUID := r.URL.Query().Get("subscription_uuid")
	var subscriptionUUIDPtr *string
	if subscriptionUUID != "" {
		subscriptionUUIDPtr = &subscriptionUUID
	}

	req := &dto.GetSubscriptionsReqDTO{
		UserUUID:         userUUID,
		SubscriptionUUID: subscriptionUUIDPtr,
	}

	h.logger.Infow("handling get subscriptions request",
		"user_uuid", userUUID,
		"subscription_uuid", subscriptionUUID)

	resp, err := h.getSubscriptions.Exec(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get subscriptions")
		h.logger.Errorw("failed to get subscriptions",
			"user_uuid", userUUID,
			"error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to encode response")
		h.logger.Errorw("failed to encode response",
			"user_uuid", userUUID,
			"error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infow("successfully retrieved subscriptions",
		"user_uuid", userUUID,
		"count", len(resp))
}

func (h *Handler) NewSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "subscription.handler.NewSubscription")
	defer span.End()

	vars := mux.Vars(r)
	userUUID := vars["user_uuid"]

	var req dto.NewSubscriptionReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request payload")
		h.logger.Errorw("failed to decode request",
			"user_uuid", userUUID,
			"error", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	req.UserUUID = userUUID

	h.logger.Infow("handling new subscription request",
		"user_uuid", userUUID,
		"lab_type", req.LabType,
		"lab_topic", req.LabTopic)

	resp, err := h.newSubscription.Exec(ctx, &req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create subscription")
		h.logger.Errorw("failed to create subscription",
			"user_uuid", userUUID,
			"error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to encode response")
		h.logger.Errorw("failed to encode response",
			"user_uuid", userUUID,
			"error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infow("successfully created subscription",
		"user_uuid", userUUID,
		"subscription_uuid", resp.UUID)
}

func (h *Handler) EditSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "subscription.handler.EditSubscription")
	defer span.End()

	vars := mux.Vars(r)
	userUUID := vars["user_uuid"]
	subscriptionUUID := vars["id"]

	var req dto.EditSubscriptionReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request payload")
		h.logger.Errorw("failed to decode request",
			"user_uuid", userUUID,
			"subscription_uuid", subscriptionUUID,
			"error", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	req.UserUUID = userUUID
	req.SubscriptionUUID = subscriptionUUID

	h.logger.Infow("handling edit subscription request",
		"user_uuid", userUUID,
		"subscription_uuid", subscriptionUUID)

	resp, err := h.editSubscription.Exec(ctx, &req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to edit subscription")
		h.logger.Errorw("failed to edit subscription",
			"user_uuid", userUUID,
			"subscription_uuid", subscriptionUUID,
			"error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to encode response")
		h.logger.Errorw("failed to encode response",
			"user_uuid", userUUID,
			"subscription_uuid", subscriptionUUID,
			"error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infow("successfully edited subscription",
		"user_uuid", userUUID,
		"subscription_uuid", subscriptionUUID)
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/subscriptions/{user_uuid}", h.GetSubscriptions).Methods(http.MethodGet)
	r.HandleFunc("/api/subscriptions/{user_uuid}/new", h.NewSubscription).Methods(http.MethodPost)
	r.HandleFunc("/api/subscriptions/{user_uuid}/{id}", h.GetSubscriptions).Methods(http.MethodGet)
	r.HandleFunc("/api/subscriptions/{user_uuid}/{id}", h.EditSubscription).Methods(http.MethodPatch)
}
