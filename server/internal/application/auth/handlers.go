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
	authUser       *usecase.AuthUserUsecase
	createUserData *usecase.CreateUserDataUsecase
	dikidiAuth     *usecase.DikidiAuthUsecase
	getUserInfo    *usecase.GetUserInfoUsecase
	logger         *zap.SugaredLogger
}

func NewHandler(authSvc *auth.Service, userSvc *user.Service, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		authUser: &usecase.AuthUserUsecase{
			AuthSvc: authSvc,
			UserSvc: userSvc,
		},
		createUserData: &usecase.CreateUserDataUsecase{
			AuthSvc: authSvc,
		},
		dikidiAuth: &usecase.DikidiAuthUsecase{
			AuthSvc: authSvc,
		},
		getUserInfo: &usecase.GetUserInfoUsecase{
			AuthSvc: authSvc,
		},
		logger: logger,
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

func (h *Handler) CreateUserData(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserDataReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warnf("auth handler: create user data: failed to decode body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.createUserData.Exec(r.Context(), cookie.Value, &req); err != nil {
		h.logger.Errorf("auth handler: create user data: failed to create user data: %v", err)
		code := apperr.HTTPErrorCode(err)
		http.Error(w, err.Error(), code)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) DikidiAuth(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.dikidiAuth.Exec(r.Context(), cookie.Value); err != nil {
		h.logger.Errorf("auth handler: dikidi auth: failed to create user data: %v", err)
		code := apperr.HTTPErrorCode(err)
		http.Error(w, err.Error(), code)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	result, err := h.getUserInfo.Exec(r.Context(), cookie.Value)
	if err != nil {
		h.logger.Errorf("auth handler: get user info: %v", err)
		code := apperr.HTTPErrorCode(err)
		http.Error(w, err.Error(), code)
		return
	}

	if result == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(result); err != nil {
		h.logger.Errorf("auth handler: get user info: encode response: %v", err)
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/auth/user", h.AuthUser).Methods(http.MethodPost)
	r.HandleFunc("/api/auth/data", h.CreateUserData).Methods(http.MethodPost)
	r.HandleFunc("/api/auth/dikidi", h.DikidiAuth).Methods(http.MethodGet)
	r.HandleFunc("/api/auth/data", h.GetUserInfo).Methods(http.MethodGet)
}
