package monitoring

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// Init initialise le serveur de métriques Prometheus
func Init(port int) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logrus.WithField("port", port).Info("Prometheus metrics server starting")

		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			logrus.WithError(err).Error("Failed to start metrics server")
		}
	}()
}

// Handler retourne le handler Prometheus
func Handler() http.Handler {
	return promhttp.Handler()
}
