package auth

import (
	"encoding/json"
	"labgrab/internal/application/auth/dto"
	"labgrab/internal/application/auth/usecase"
	"labgrab/internal/auth"
	"labgrab/internal/shared/apperr"
	"labgrab/internal/user"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	authUser *usecase.AuthUserUsecase
	logger   *zap.SugaredLogger
}

func NewHandler(authSvc *auth.Service, userSvc *user.Service, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		authUser: &usecase.AuthUserUsecase{
			AuthSvc: authSvc,
			UserSvc: userSvc,
		},
	}
}

func (h *Handler) AuthUser(w http.ResponseWriter, r *http.Request) {
	var req dto.AuthUserReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warnf("auth handler: auth user: failed to decode body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	session, err := h.authUser.Exec(r.Context(), &req)
	if err != nil {
		h.logger.Errorf("auth handler: auth user: failed to auth user: %v", err)
		code := apperr.HTTPErrorCode(err)
		http.Error(w, err.Error(), code)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session,
		Path:     "/",
		MaxAge:   864000,
		HttpOnly: true,
		Secure:   false,                //TODO: set to true in prod
		SameSite: http.SameSiteLaxMode, //TODO: set to other type in prod
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/auth/user", h.AuthUser).Methods(http.MethodPost)
}
