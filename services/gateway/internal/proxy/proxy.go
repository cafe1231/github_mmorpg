package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"gateway/internal/config"
)

// ServiceProxy gère le proxy vers les microservices
type ServiceProxy struct {
	config *config.Config
	client *http.Client
}

// NewServiceProxy crée une nouvelle instance du proxy
func NewServiceProxy(cfg *config.Config) (*ServiceProxy, error) {
	// Client HTTP optimisé pour les microservices
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &ServiceProxy{
		config: cfg,
		client: client,
	}, nil
}

// Forward proxie une requête vers un service backend
func (sp *ServiceProxy) Forward(c *gin.Context, endpoint config.ServiceEndpoint) error {
	// Construire l'URL de destination
	targetURL, err := url.Parse(endpoint.URL)
	if err != nil {
		return fmt.Errorf("invalid service URL: %w", err)
	}

	// Modifier le path pour enlever le préfixe du service
	originalPath := c.Request.URL.Path
	targetPath := sp.transformPath(originalPath)

	// Construire l'URL complète avec le path transformé
	targetURL.Path = targetPath
	targetURL.RawQuery = c.Request.URL.RawQuery

	// Lire le body de la requête
	var body []byte
	if c.Request.Body != nil {
		body, err = io.ReadAll(c.Request.Body)
		if err != nil {
			return fmt.Errorf("failed to read request body: %w", err)
		}
	}

	// Créer la nouvelle requête
	req, err := http.NewRequestWithContext(
		c.Request.Context(),
		c.Request.Method,
		targetURL.String(),
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Copier les headers importants
	sp.copyHeaders(c.Request, req)

	// Ajouter des headers spécifiques au proxy
	req.Header.Set("X-Forwarded-For", c.ClientIP())
	req.Header.Set("X-Forwarded-Proto", "http")
	req.Header.Set("X-Gateway-Version", "1.0.0")

	// Log de la requête proxy
	logrus.WithFields(logrus.Fields{
		"method":         req.Method,
		"original_path":  originalPath,
		"target_path":    targetPath,
		"target_url":     targetURL.String(),
		"client_ip":      c.ClientIP(),
		"request_id":     c.GetHeader("X-Request-ID"),
		"user_id":        c.GetHeader("X-User-ID"),
		"content_length": len(body),
	}).Debug("Proxying request to service")

	// Exécuter la requête avec retry
	resp, err := sp.executeWithRetry(req, endpoint.Retries, endpoint.Timeout)
	if err != nil {
		return fmt.Errorf("service request failed: %w", err)
	}
	defer resp.Body.Close()

	// Lire la réponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Copier les headers de réponse
	sp.copyResponseHeaders(resp, c)

	// Log de la réponse
	logrus.WithFields(logrus.Fields{
		"status_code":     resp.StatusCode,
		"response_length": len(responseBody),
		"target_url":      targetURL.String(),
		"request_id":      c.GetHeader("X-Request-ID"),
	}).Debug("Service response received")

	// Envoyer la réponse au client
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), responseBody)

	return nil
}

// transformPath transforme le path de la requête pour le service de destination
func (sp *ServiceProxy) transformPath(originalPath string) string {
	// Mapping des préfixes Gateway vers les paths des services
	pathMappings := map[string]string{
		"/auth/health":     "/health",                    // /auth/health -> /health
		"/auth/metrics":    "/metrics",                   // /auth/metrics -> /metrics
		"/auth/":           "/api/v1/auth/",             // /auth/register -> /api/v1/auth/register
		"/api/v1/auth/":    "/api/v1/auth/",             // /api/v1/auth/login -> /api/v1/auth/login  
		"/api/v1/user/":    "/api/v1/user/",             // /api/v1/user/profile -> /api/v1/user/profile
		"/api/v1/validate/": "/api/v1/validate/",        // /api/v1/validate/token -> /api/v1/validate/token
	}

	// Chercher le mapping correspondant (plus spécifique en premier)
	for prefix, replacement := range pathMappings {
		if strings.HasPrefix(originalPath, prefix) {
			// Pour les mappings exacts (health, metrics)
			if originalPath == prefix {
				newPath := replacement
				
				logrus.WithFields(logrus.Fields{
					"original_path": originalPath,
					"new_path":     newPath,
					"mapping_type": "exact",
				}).Debug("Path transformation")
				
				return newPath
			}
			
			// Pour les mappings de préfixe
			if strings.HasSuffix(prefix, "/") {
				newPath := replacement + strings.TrimPrefix(originalPath, prefix)
				
				logrus.WithFields(logrus.Fields{
					"original_path": originalPath,
					"new_path":     newPath,
					"prefix":       prefix,
					"replacement":  replacement,
					"mapping_type": "prefix",
				}).Debug("Path transformation")
				
				return newPath
			}
		}
	}

	// Si aucun mapping trouvé, garder le path original
	logrus.WithFields(logrus.Fields{
		"original_path": originalPath,
		"new_path":     originalPath,
		"mapping_type": "none",
	}).Debug("Path transformation")
	
	return originalPath
}

// executeWithRetry exécute une requête avec retry automatique
func (sp *ServiceProxy) executeWithRetry(req *http.Request, maxRetries int, timeout time.Duration) (*http.Response, error) {
	// Sauvegarder le body original pour les retries
	var originalBody []byte
	if req.Body != nil {
		var err error
		originalBody, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body for retry: %w", err)
		}
		req.Body.Close()
	}

	var lastErr error

	// Créer un client avec timeout spécifique
	client := &http.Client{
		Timeout:   timeout,
		Transport: sp.client.Transport,
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Restaurer le body pour chaque tentative
		if originalBody != nil {
			req.Body = io.NopCloser(bytes.NewReader(originalBody))
		}

		// Log de la tentative
		if attempt > 0 {
			logrus.WithFields(logrus.Fields{
				"attempt":     attempt + 1,
				"max_retries": maxRetries + 1,
				"url":         req.URL.String(),
				"request_id":  req.Header.Get("X-Request-ID"),
			}).Warn("Retrying service request")
		}

		// Exécuter la requête
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err

			// Attendre avant de retry (backoff exponentiel)
			if attempt < maxRetries {
				waitTime := time.Duration(attempt+1) * 500 * time.Millisecond
				time.Sleep(waitTime)
			}
			continue
		}

		// Vérifier si la réponse indique une erreur de service
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("service returned status %d", resp.StatusCode)

			if attempt < maxRetries {
				waitTime := time.Duration(attempt+1) * 500 * time.Millisecond
				time.Sleep(waitTime)
			}
			continue
		}

		// Succès
		return resp, nil
	}

	return nil, fmt.Errorf("all retry attempts failed, last error: %w", lastErr)
}

// copyHeaders copie les headers importants de la requête originale
func (sp *ServiceProxy) copyHeaders(originalReq *http.Request, newReq *http.Request) {
	// Headers à copier
	headersToProxy := []string{
		"Content-Type",
		"Content-Length",
		"Accept",
		"Accept-Encoding",
		"Accept-Language",
		"User-Agent",
		"X-Request-ID",
		"X-User-ID",
		"X-Username",
		"X-User-Role",
		"X-Client-Version",
		"X-Game-Session",
		"Authorization",
	}

	for _, header := range headersToProxy {
		if value := originalReq.Header.Get(header); value != "" {
			newReq.Header.Set(header, value)
		}
	}

	// Copier tous les headers personnalisés commençant par X-
	for name, values := range originalReq.Header {
		if len(name) > 2 && name[:2] == "X-" {
			for _, value := range values {
				newReq.Header.Add(name, value)
			}
		}
	}
}

// copyResponseHeaders copie les headers de réponse appropriés
func (sp *ServiceProxy) copyResponseHeaders(resp *http.Response, c *gin.Context) {
	// Headers à copier dans la réponse
	headersToProxy := []string{
		"Content-Type",
		"Cache-Control",
		"Expires",
		"Last-Modified",
		"ETag",
		"X-Rate-Limit-Remaining",
		"X-Request-ID",
	}

	for _, header := range headersToProxy {
		if value := resp.Header.Get(header); value != "" {
			c.Header(header, value)
		}
	}

	// Ajouter des headers spécifiques au gateway
	c.Header("X-Gateway-Service", extractServiceFromURL(resp.Request.URL.String()))
	c.Header("X-Response-Time", fmt.Sprintf("%dms", time.Now().UnixMilli()))
}

// Close ferme le proxy proprement
func (sp *ServiceProxy) Close() error {
	// Fermer les connexions idle
	if transport, ok := sp.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}

	logrus.Info("Service proxy closed")
	return nil
}

// extractServiceFromURL extrait le nom du service depuis l'URL
func extractServiceFromURL(urlStr string) string {
	// http://localhost:8081/health -> localhost:8081
	if parsedURL, err := url.Parse(urlStr); err == nil {
		return parsedURL.Host
	}
	return "unknown"
}