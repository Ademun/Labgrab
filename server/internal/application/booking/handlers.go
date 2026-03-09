package booking

import (
	"encoding/json"
	"errors"
	"fmt"
	"labgrab/internal/application/booking/usecase"
	"labgrab/internal/auth"
	"labgrab/internal/booking"
	"net/http"

	"go.uber.org/zap"
)

type Handler struct {
	getBookings *usecase.GetBookingsUsecase
	logger      *zap.SugaredLogger
}

func NewHandler(bookingSvc *booking.Service, authSvc *auth.Service, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		getBookings: &usecase.GetBookingsUsecase{
			BookingSvc: bookingSvc,
			AuthSvc:    authSvc,
		},
		logger: logger,
	}
}

func (h *Handler) GetBookings(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		h.logger.Error(err)
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Failed to read cookie", http.StatusInternalServerError)
		return
	}

	bookings, err := h.getBookings.Exec(r.Context(), cookie.Value)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, "Failed to read bookings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(bookings); err != nil {
		err := fmt.Errorf("failed to write response: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error(err)
		return
	}
}
