package web

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	logger *zap.SugaredLogger
}

func NewHandler(logger *zap.SugaredLogger) *Handler {
	return &Handler{
		logger: logger,
	}
}

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.ServeFile(w, r, "internal/application/web/config.json")
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/web/config", h.GetConfig).Methods(http.MethodGet)
}
