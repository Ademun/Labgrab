package web

import (
	_ "embed"
	"encoding/json"
	"labgrab/internal/shared/apperr"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

//go:embed config.json
var jsonConf []byte

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
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(jsonConf); err != nil {
		h.logger.Errorf("web handler: get config: %v", err)
		http.Error(w, err.Error(), apperr.HTTPErrorCode(err))
		return
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/web/config", h.GetConfig).Methods(http.MethodGet)
}
