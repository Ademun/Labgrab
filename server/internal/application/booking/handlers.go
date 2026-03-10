package booking

import (
	"encoding/json"
	"errors"
	"labgrab/internal/application/booking/usecase"
	"labgrab/internal/auth"
	"labgrab/internal/booking"
	"labgrab/internal/shared/apperr"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	getBookings   *usecase.GetBookingsUsecase
	loadBookings  *usecase.LoadBookingsUsecase
	cancelBooking *usecase.CancelBookingUsecase
	logger        *zap.SugaredLogger
}

func NewHandler(bookingSvc *booking.Service, authSvc *auth.Service, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		getBookings: &usecase.GetBookingsUsecase{
			BookingSvc: bookingSvc,
			AuthSvc:    authSvc,
		},
		loadBookings: &usecase.LoadBookingsUsecase{
			BookingSvc: bookingSvc,
			AuthSvc:    authSvc,
		},
		cancelBooking: &usecase.CancelBookingUsecase{
			BookingSvc: bookingSvc,
			AuthSvc:    authSvc,
		},
		logger: logger,
	}
}

func (h *Handler) GetBookings(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		h.logger.Errorf("booking handler: get bookings: read cookie: %v", err)
		http.Error(w, "Failed to read cookie", http.StatusInternalServerError)
		return
	}

	result, err := h.getBookings.Exec(r.Context(), cookie.Value)
	if err != nil {
		h.logger.Errorf("booking handler: get bookings: %v", err)
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(result); err != nil {
		h.logger.Errorf("booking handler: get bookings: encode response: %v", err)
	}
}

func (h *Handler) LoadBookings(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		h.logger.Errorf("booking handler: load bookings: read cookie: %v", err)
		http.Error(w, "Failed to read cookie", http.StatusInternalServerError)
		return
	}

	if err := h.loadBookings.Exec(r.Context(), cookie.Value); err != nil {
		h.logger.Errorf("booking handler: load bookings: %v", err)
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		h.logger.Errorf("booking handler: get bookings: read cookie: %v", err)
		http.Error(w, "Failed to read cookie", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	bookingID, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.logger.Errorf("booking handler: get bookings: parse id: %v", err)
		http.Error(w, "Failed to parse id", http.StatusBadRequest)
		return
	}

	if err := h.cancelBooking.Exec(r.Context(), cookie.Value, bookingID); err != nil {
		h.logger.Errorf("booking handler: cancel booking: %v", err)
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/bookings", h.GetBookings).Methods(http.MethodGet)
	r.HandleFunc("/api/bookings/load", h.LoadBookings).Methods(http.MethodPost)
	r.HandleFunc("/api/bookings/{id}", h.CancelBooking).Methods(http.MethodDelete)
}
