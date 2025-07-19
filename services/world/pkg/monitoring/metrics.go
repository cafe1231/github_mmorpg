package monitoring

import (
	"net/http"
)

// Handler retourne un handler stub pour les métriques
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Metrics endpoint - not implemented yet\n"))
	})
}
