package user

import (
	"encoding/json"
	"fmt"
	"labgrab/internal/application/user/dto"
	"labgrab/internal/application/user/usecase"
	"labgrab/internal/auth"
	"labgrab/internal/subscription"
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
	authUser *usecase.AuthUserUseCase
	newUser  *usecase.NewUserUseCase
	logger   *zap.SugaredLogger
}

func NewHandler(authSvc *auth.Service, userSvc *user.Service, subscriptionSvc *subscription.Service, pool *pgxpool.Pool, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		authUser: usecase.NewAuthUserUseCase(authSvc, userSvc),
		newUser:  usecase.NewNewUserUseCase(userSvc, subscriptionSvc, pool),
		logger:   logger,
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
		return
	}

	resp, err := h.authUser.Exec(ctx, &req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		err := fmt.Errorf("failed to write response: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) NewUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "user.handler.NewUser")
	defer span.End()

	var req dto.NewUserReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = fmt.Errorf("failed to decode request: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	resp, err := h.newUser.Exec(ctx, &req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		err := fmt.Errorf("failed to write response: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/users/auth", h.Auth).Methods(http.MethodPost)
	r.HandleFunc("/api/users", h.NewUser).Methods(http.MethodPost)
}
