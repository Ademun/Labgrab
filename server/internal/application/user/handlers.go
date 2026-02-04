package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"labgrab/internal/application/user/dto"
	"labgrab/internal/application/user/usecase"
	"labgrab/internal/auth"
	"labgrab/internal/user"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("user-handler")

type Handler struct {
	authUser   *usecase.AuthUserUseCase
	getUser    *usecase.GetUserUseCase
	updateUser *usecase.UpdateUserUseCase
	logger     *zap.SugaredLogger
}

func NewHandler(authSvc *auth.Service, userSvc *user.Service, pool *pgxpool.Pool, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		authUser:   usecase.NewAuthUserUseCase(authSvc, userSvc, pool),
		getUser:    usecase.NewGetUserUseCase(authSvc, userSvc),
		updateUser: usecase.NewUpdateUserUseCase(authSvc, userSvc),
		logger:     logger,
	}
}

func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "user.handler.Auth")
	defer span.End()

	var req dto.AuthUserReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = fmt.Errorf("failed to decode request: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		h.logger.Error(err)
		return
	}

	resp, err := h.authUser.Exec(ctx, &req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error(err)
		return
	}

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    resp.Session,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   false,                //TODO: set to true in prod
		SameSite: http.SameSiteLaxMode, //TODO: set to other type in prod
	}

	http.SetCookie(w, cookie)
	if err = json.NewEncoder(w).Encode(resp.IsNew); err != nil {
		err := fmt.Errorf("failed to write response: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error(err)
		return
	}
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "user.handler.NewUser")
	defer span.End()

	cookie, err := r.Cookie("session_id")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			h.logger.Error(err)
			return
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Failed to read cookie", http.StatusInternalServerError)
		h.logger.Error(err)
		return
	}

	resp, err := h.getUser.Exec(ctx, cookie.Value)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error(err)
		return
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		err := fmt.Errorf("failed to write response: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error(err)
		return
	}
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "user.handler.UpdateUser")
	defer span.End()

	cookie, err := r.Cookie("session_id")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		h.logger.Error(err)
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Failed to read cookie", http.StatusInternalServerError)
		return
	}

	var req dto.UpdateUserReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = fmt.Errorf("failed to decode request: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		h.logger.Error(err)
		return
	}

	if err := h.updateUser.Exec(ctx, cookie.Value, &req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/users/auth", h.Auth).Methods(http.MethodPost)
	r.HandleFunc("/api/users", h.GetUser).Methods(http.MethodGet)
	r.HandleFunc("/api/users", h.UpdateUser).Methods(http.MethodPatch)
}
