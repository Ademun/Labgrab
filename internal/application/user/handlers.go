package user

import (
	"encoding/json"
	"labgrab/internal/application/user/dto"
	"labgrab/internal/application/user/usecase"
	"log/slog"
	"net/http"
)

type Handler struct {
	authUser *usecase.AuthUserUseCase
}

func NewHandler(authUser *usecase.AuthUserUseCase) *Handler {
	return &Handler{authUser: authUser}
}

func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	var req dto.AuthUserReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	resp, err := h.authUser.Exec(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
}
