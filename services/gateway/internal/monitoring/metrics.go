package monitoring

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// Constantes pour le serveur de métriques
const (
	MetricsReadTimeout  = 10 // secondes
	MetricsWriteTimeout = 10 // secondes
	MetricsIdleTimeout  = 60 // secondes
)

// Init initialize le serveur de métriques Prometheus
func Init(port int) {
	go func() {
		// Créer un serveur HTTP avec timeouts pour la sécurité
		server := &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      nil,
			ReadTimeout:  MetricsReadTimeout * time.Second,
			WriteTimeout: MetricsWriteTimeout * time.Second,
			IdleTimeout:  MetricsIdleTimeout * time.Second,
		}

		http.Handle("/metrics", promhttp.Handler())
		logrus.WithField("port", port).Info("Prometheus metrics server starting")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Error("Failed to start metrics server")
		}
	}()
}

// Handler retourne le handler Prometheus
func Handler() http.Handler {
	return promhttp.Handler()
}
