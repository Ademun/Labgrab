package user

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"labgrab/internal/application/user/dto"
	"labgrab/internal/application/user/usecase"
	"labgrab/internal/auth"
	"labgrab/internal/shared/apperr"
	"labgrab/internal/user"
)

type Handler struct {
	getUser    *usecase.GetUserUseCase
	updateUser *usecase.UpdateUserUseCase
	logger     *zap.SugaredLogger
}

func NewHandler(authSvc *auth.Service, userSvc *user.Service, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		getUser: &usecase.GetUserUseCase{
			AuthSvc: authSvc,
			UserSvc: userSvc,
		},
		updateUser: &usecase.UpdateUserUseCase{
			AuthSvc: authSvc,
			UserSvc: userSvc,
		},
		logger: logger,
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/user", h.GetUser).Methods(http.MethodGet)
	r.HandleFunc("/api/user", h.UpdateUser).Methods(http.MethodPatch)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	result, err := h.getUser.Exec(r.Context(), cookie.Value)
	if err != nil {
		h.logger.Errorf("user handler: get user: %v", err)
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.logger.Errorf("user handler: get user: encode response: %v", err)
	}
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req dto.UpdateUserReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warnf("user handler: update user: failed to decode body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.updateUser.Exec(r.Context(), cookie.Value, &req); err != nil {
		h.logger.Errorf("user handler: update user: %v", err)
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
