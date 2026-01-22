package user

import (
	"encoding/json"
	"labgrab/internal/application/user/dto"
	"labgrab/internal/application/user/usecase"
	"net/http"
)

type Handler struct {
	authUser *usecase.AuthUserUseCase
	newUser  *usecase.NewUserUseCase
}

func NewHandler(authUser *usecase.AuthUserUseCase, newUser *usecase.NewUserUseCase) *Handler {
	return &Handler{
		authUser: authUser,
		newUser:  newUser,
	}
}

func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	var req dto.AuthUserReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	w.Header().Set("Content-Type", "application/json")
}

func (h *Handler) NewUser(w http.ResponseWriter, r *http.Request) {
	var req dto.NewUserReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	resp, err := h.newUser.Exec(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
}
