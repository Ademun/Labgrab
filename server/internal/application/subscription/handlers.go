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
	"go.uber.org/zap"
)

type Handler struct {
	getSubscriptions      *usecase.GetSubscriptionsUsecase
	newSubscription       *usecase.NewSubscriptionUsecase
	editSubscription      *usecase.EditSubscriptionUsecase
	getTimeRestrictions   *usecase.GetTimeRestrictionsUsecase
	setTimeRestrictions   *usecase.SetTimeRestrictionsUsecase
	getTeacherPreferences *usecase.GetTeacherPreferencesUsecase
	setTeacherPreferences *usecase.SetTeacherPreferencesUsecase
	logger                *zap.SugaredLogger
}

func NewHandler(authSvc *auth.Service, subscriptionSvc *subscription.Service, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		getSubscriptions:      &usecase.GetSubscriptionsUsecase{AuthSvc: authSvc, SubscriptionSvc: subscriptionSvc},
		newSubscription:       &usecase.NewSubscriptionUsecase{AuthSvc: authSvc, SubscriptionSvc: subscriptionSvc},
		editSubscription:      &usecase.EditSubscriptionUsecase{AuthSvc: authSvc, SubscriptionSvc: subscriptionSvc},
		getTimeRestrictions:   &usecase.GetTimeRestrictionsUsecase{AuthSvc: authSvc, SubscriptionSvc: subscriptionSvc},
		setTimeRestrictions:   &usecase.SetTimeRestrictionsUsecase{AuthSvc: authSvc, SubscriptionSvc: subscriptionSvc},
		getTeacherPreferences: &usecase.GetTeacherPreferencesUsecase{AuthSvc: authSvc, SubscriptionSvc: subscriptionSvc},
		setTeacherPreferences: &usecase.SetTeacherPreferencesUsecase{AuthSvc: authSvc, SubscriptionSvc: subscriptionSvc},
		logger:                logger,
	}
}

func (h *Handler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscriptionUUID := vars["id"]

	var subscriptionUUIDPtr *string
	if subscriptionUUID != "" {
		subscriptionUUIDPtr = &subscriptionUUID
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.getSubscriptions.Exec(r.Context(), cookie.Value, subscriptionUUIDPtr)
	if err != nil {
		h.logger.Errorf("subscription handler: get subscriptions: %v", err)
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("subscription handler: get subscriptions: encode response: %v", err)
	}
}

func (h *Handler) NewSubscription(w http.ResponseWriter, r *http.Request) {
	var req dto.NewSubscriptionReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warnf("subscription handler: new subscription: failed to decode body: %v", err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.newSubscription.Exec(r.Context(), cookie.Value, &req)
	if err != nil {
		h.logger.Errorf("subscription handler: new subscription: %v", err)
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("subscription handler: new subscription: encode response: %v", err)
	}
}

func (h *Handler) EditSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscriptionUUID := vars["id"]

	var req dto.EditSubscriptionReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warnf("subscription handler: edit subscription: failed to decode body: %v", err)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.editSubscription.Exec(r.Context(), cookie.Value, subscriptionUUID, &req)
	if err != nil {
		h.logger.Errorf("subscription handler: edit subscription: %v", err)
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("subscription handler: edit subscription: encode response: %v", err)
	}
}

func (h *Handler) GetTimeRestrictions(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.getTimeRestrictions.Exec(r.Context(), cookie.Value)
	if err != nil {
		h.logger.Errorf("subscription handler: get time preferences: %v", err)
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("subscription handler: get time preferences: encode response: %v", err)
	}
}

func (h *Handler) SetTimeRestrictions(w http.ResponseWriter, r *http.Request) {
	var req dto.SetTimeRestrictionsReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warnf("subscription handler: set time preferences: failed to decode body: %v", err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.setTimeRestrictions.Exec(r.Context(), cookie.Value, &req); err != nil {
		h.logger.Errorf("subscription handler: set time preferences: %v", err)
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetTeacherPreferences(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.getTeacherPreferences.Exec(r.Context(), cookie.Value)
	if err != nil {
		h.logger.Errorf("subscription handler: get teacher preferences: %v", err)
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("subscription handler: get teacher preferences: encode response: %v", err)
	}
}

func (h *Handler) SetTeacherPreferences(w http.ResponseWriter, r *http.Request) {
	var req dto.SetTeacherPreferencesReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warnf("subscription handler: set teacher preferences: failed to decode body: %v", err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.setTeacherPreferences.Exec(r.Context(), cookie.Value, &req); err != nil {
		h.logger.Errorf("subscription handler: set teacher preferences: %v", err)
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/subscriptions", h.GetSubscriptions).Methods(http.MethodGet)
	r.HandleFunc("/api/subscriptions", h.NewSubscription).Methods(http.MethodPost)
	r.HandleFunc("/api/subscriptions/{id}", h.GetSubscriptions).Methods(http.MethodGet)
	r.HandleFunc("/api/subscriptions/{id}", h.EditSubscription).Methods(http.MethodPatch)
	r.HandleFunc("/api/subscriptions/restrictions/time", h.GetTimeRestrictions).Methods(http.MethodGet)
	r.HandleFunc("/api/subscriptions/restrictions/time", h.SetTimeRestrictions).Methods(http.MethodPost)
	r.HandleFunc("/api/subscription/preferences/teachers", h.GetTeacherPreferences).Methods(http.MethodGet)
	r.HandleFunc("/api/subscription/preferences/teachers", h.SetTeacherPreferences).Methods(http.MethodPost)
}
