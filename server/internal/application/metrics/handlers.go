package metrics

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(r *mux.Router) {
	r.Handle("/api/metrics", promhttp.Handler()).Methods(http.MethodGet)
}
