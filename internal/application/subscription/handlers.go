package subscription

import (
	"encoding/json"
	"fmt"
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

	resp, err := h.getSubscriptions.Exec(ctx, req)
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
}

func (h *Handler) NewSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "subscription.handler.NewSubscription")
	defer span.End()

	vars := mux.Vars(r)
	userUUID := vars["user_uuid"]

	var req dto.NewSubscriptionReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = fmt.Errorf("failed to decode request: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	req.UserUUID = userUUID

	resp, err := h.newSubscription.Exec(ctx, &req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		err = fmt.Errorf("failed to write response: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) EditSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "subscription.handler.EditSubscription")
	defer span.End()

	vars := mux.Vars(r)
	userUUID := vars["user_uuid"]
	subscriptionUUID := vars["id"]

	var req dto.EditSubscriptionReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = fmt.Errorf("failed to decode request: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	req.UserUUID = userUUID
	req.SubscriptionUUID = subscriptionUUID

	resp, err := h.editSubscription.Exec(ctx, &req)
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
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/subscriptions/{user_uuid}", h.GetSubscriptions).Methods(http.MethodGet)
	r.HandleFunc("/api/subscriptions/{user_uuid}", h.NewSubscription).Methods(http.MethodPost)
	r.HandleFunc("/api/subscriptions/{user_uuid}/{id}", h.GetSubscriptions).Methods(http.MethodGet)
	r.HandleFunc("/api/subscriptions/{user_uuid}/{id}", h.EditSubscription).Methods(http.MethodPatch)
}
