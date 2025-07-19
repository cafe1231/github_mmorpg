package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"

	"gateway/internal/config"
	"gateway/internal/proxy"
)

// Server reprÃ©sente le serveur Gateway
type Server struct {
	config        *config.Config
	proxy         *proxy.ServiceProxy
	natsConn      *nats.Conn
	upgrader      websocket.Upgrader
	wsClients     map[*websocket.Conn]bool
	wsMutex       sync.RWMutex
	serviceHealth map[string]ServiceHealth
	healthMutex   sync.RWMutex
}

// ServiceHealth reprÃ©sente l'Ã©tat de santÃ© d'un service
type ServiceHealth struct {
	Name         string    `json:"name"`
	URL          string    `json:"url"`
	Status       string    `json:"status"` // "healthy", "unhealthy", "unknown"
	LastCheck    time.Time `json:"last_check"`
	ResponseTime int64     `json:"response_time_ms"`
	ErrorCount   int       `json:"error_count"`
}

// NewServer crÃ©e une nouvelle instance du serveur Gateway
func NewServer(cfg *config.Config) (*Server, error) {
	// Initialisation du proxy de services
	serviceProxy, err := proxy.NewServiceProxy(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create service proxy: %w", err)
	}

	// Connexion Ã  NATS
	natsConn, err := connectToNATS(cfg.NATS)
	if err != nil {
		logrus.Warn("Failed to connect to NATS, continuing without messaging: ", err)
	}

	// Configuration du WebSocket upgrader
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// En production, vÃ©rifier l'origine
			if cfg.Server.Environment == "production" {
				// Ajouter ici la logique de vÃ©rification d'origine
				return true // Temporaire
			}
			return true
		},
	}

	server := &Server{
		config:        cfg,
		proxy:         serviceProxy,
		natsConn:      natsConn,
		upgrader:      upgrader,
		wsClients:     make(map[*websocket.Conn]bool),
		serviceHealth: make(map[string]ServiceHealth),
	}

	// Initialiser le monitoring des services
	server.initServiceHealthMonitoring()

	logrus.Info("Gateway server initialized successfully")
	return server, nil
}

// ProxyTo retourne un handler Gin qui proxie vers un service spÃ©cifique
func (s *Server) ProxyTo(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// RÃ©cupÃ©rer l'endpoint du service
		endpoint, exists := s.getServiceEndpoint(serviceName)
		if !exists {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": fmt.Sprintf("Service %s not available", serviceName),
			})
			return
		}

		// VÃ©rifier la santÃ© du service
		if !s.isServiceHealthy(serviceName) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": fmt.Sprintf("Service %s is unhealthy", serviceName),
			})
			return
		}

		// Proxier la requÃªte
		if err := s.proxy.Forward(c, endpoint); err != nil {
			logrus.WithError(err).WithField("service", serviceName).Error("Proxy request failed")

			// Marquer le service comme potentiellement dÃ©faillant
			s.incrementServiceError(serviceName)

			c.JSON(http.StatusBadGateway, gin.H{
				"error": "Service request failed",
			})
			return
		}
	}
}

// HealthCheck endpoint de santÃ© du gateway
func (s *Server) HealthCheck(c *gin.Context) {
	healthStatus := s.getOverallHealth()

	// VÃ©rifier le statut depuis la map
	if status, ok := healthStatus["status"].(string); ok && status == "healthy" {
		c.JSON(http.StatusOK, healthStatus)
	} else {
		c.JSON(http.StatusServiceUnavailable, healthStatus)
	}
}

// HandleWebSocket gÃ¨re les connexions WebSocket
func (s *Server) HandleWebSocket(c *gin.Context) {
	// Upgrade de la connexion HTTP vers WebSocket
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to upgrade WebSocket connection")
		return
	}
	defer conn.Close()

	// Ajouter le client Ã  la liste
	s.wsMutex.Lock()
	s.wsClients[conn] = true
	clientCount := len(s.wsClients)
	s.wsMutex.Unlock()

	logrus.WithField("client_count", clientCount).Info("WebSocket client connected")

	// Envoyer un message de bienvenue
	welcomeMsg := map[string]interface{}{
		"type":    "welcome",
		"message": "Connected to MMORPG Gateway",
		"time":    time.Now().Unix(),
	}
	conn.WriteJSON(welcomeMsg)

	// GÃ©rer les messages du client
	for {
		var message map[string]interface{}
		err := conn.ReadJSON(&message)
		if err != nil {
			logrus.WithError(err).Debug("WebSocket read error")
			break
		}

		// Traiter le message
		s.handleWebSocketMessage(conn, message)
	}

	// Nettoyer lors de la dÃ©connexion
	s.wsMutex.Lock()
	delete(s.wsClients, conn)
	clientCount = len(s.wsClients)
	s.wsMutex.Unlock()

	logrus.WithField("client_count", clientCount).Info("WebSocket client disconnected")
}

// ListRoutes affiche toutes les routes disponibles (debug)
func (s *Server) ListRoutes(c *gin.Context) {
	if s.config.Server.Environment == "production" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Debug endpoints disabled in production"})
		return
	}

	routes := []map[string]string{
		{"method": "POST", "path": "/api/v1/auth/register", "service": "auth"},
		{"method": "POST", "path": "/api/v1/auth/login", "service": "auth"},
		{"method": "GET", "path": "/api/v1/player/profile", "service": "player"},
		{"method": "GET", "path": "/api/v1/world/zones", "service": "world"},
		{"method": "POST", "path": "/api/v1/combat/action", "service": "combat"},
		{"method": "GET", "path": "/api/v1/inventory/:characterId", "service": "inventory"},
		{"method": "GET", "path": "/api/v1/guild/", "service": "guild"},
		{"method": "GET", "path": "/api/v1/chat/channels", "service": "chat"},
		{"method": "GET", "path": "/api/v1/analytics/dashboard", "service": "analytics"},
		{"method": "GET", "path": "/ws", "service": "gateway-websocket"},
	}

	c.JSON(http.StatusOK, gin.H{
		"routes": routes,
		"total":  len(routes),
	})
}

// ShowConfig affiche la configuration (debug, sans secrets)
func (s *Server) ShowConfig(c *gin.Context) {
	if s.config.Server.Environment == "production" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Debug endpoints disabled in production"})
		return
	}

	safeConfig := map[string]interface{}{
		"server": map[string]interface{}{
			"port":        s.config.Server.Port,
			"host":        s.config.Server.Host,
			"environment": s.config.Server.Environment,
		},
		"services":   s.config.Services,
		"rate_limit": s.config.RateLimit,
		"monitoring": s.config.Monitoring,
	}

	c.JSON(http.StatusOK, safeConfig)
}

// ServiceStatus affiche l'Ã©tat de tous les services
func (s *Server) ServiceStatus(c *gin.Context) {
	s.healthMutex.RLock()
	services := make([]ServiceHealth, 0, len(s.serviceHealth))
	for _, health := range s.serviceHealth {
		services = append(services, health)
	}
	s.healthMutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"services": services,
		"total":    len(services),
	})
}

// Close ferme proprement le serveur Gateway
func (s *Server) Close() error {
	var errors []error

	// Fermer les connexions WebSocket
	s.wsMutex.Lock()
	for conn := range s.wsClients {
		conn.Close()
	}
	s.wsClients = make(map[*websocket.Conn]bool)
	s.wsMutex.Unlock()

	// Fermer la connexion NATS
	if s.natsConn != nil {
		s.natsConn.Close()
	}

	// Fermer le proxy
	if s.proxy != nil {
		if err := s.proxy.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during gateway shutdown: %v", errors)
	}

	logrus.Info("Gateway server closed successfully")
	return nil
}

// MÃ©thodes privÃ©es

// getServiceEndpoint rÃ©cupÃ¨re l'endpoint d'un service
func (s *Server) getServiceEndpoint(serviceName string) (config.ServiceEndpoint, bool) {
	switch serviceName {
	case "auth":
		return s.config.Services.Auth, true
	case "player":
		return s.config.Services.Player, true
	case "world":
		return s.config.Services.World, true
	case "combat":
		return s.config.Services.Combat, true
	case "inventory":
		return s.config.Services.Inventory, true
	case "guild":
		return s.config.Services.Guild, true
	case "chat":
		return s.config.Services.Chat, true
	case "analytics":
		return s.config.Services.Analytics, true
	default:
		return config.ServiceEndpoint{}, false
	}
}

// isServiceHealthy vÃ©rifie si un service est en bonne santÃ©
func (s *Server) isServiceHealthy(serviceName string) bool {
	s.healthMutex.RLock()
	defer s.healthMutex.RUnlock()

	health, exists := s.serviceHealth[serviceName]
	if !exists {
		return true // Si pas de donnÃ©es, on assume que c'est sain
	}

	return health.Status == "healthy"
}

// incrementServiceError incrÃ©mente le compteur d'erreurs d'un service
func (s *Server) incrementServiceError(serviceName string) {
	s.healthMutex.Lock()
	defer s.healthMutex.Unlock()

	if health, exists := s.serviceHealth[serviceName]; exists {
		health.ErrorCount++
		if health.ErrorCount > 5 {
			health.Status = "unhealthy"
		}
		s.serviceHealth[serviceName] = health
	}
}

// getOverallHealth calcule l'Ã©tat de santÃ© global
func (s *Server) getOverallHealth() map[string]interface{} {
	s.healthMutex.RLock()
	defer s.healthMutex.RUnlock()

	totalServices := len(s.serviceHealth)
	healthyServices := 0

	for _, health := range s.serviceHealth {
		if health.Status == "healthy" {
			healthyServices++
		}
	}

	status := "healthy"
	if totalServices > 0 && float64(healthyServices)/float64(totalServices) < 0.8 {
		status = "degraded"
	}
	if healthyServices == 0 && totalServices > 0 {
		status = "unhealthy"
	}

	return map[string]interface{}{
		"status":            status,
		"timestamp":         time.Now().Unix(),
		"services_total":    totalServices,
		"services_healthy":  healthyServices,
		"websocket_clients": len(s.wsClients),
		"version":           "1.0.0",
	}
}

// handleWebSocketMessage traite les messages WebSocket
func (s *Server) handleWebSocketMessage(conn *websocket.Conn, message map[string]interface{}) {
	messageType, ok := message["type"].(string)
	if !ok {
		conn.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": "Message type required",
		})
		return
	}

	switch messageType {
	case "ping":
		conn.WriteJSON(map[string]interface{}{
			"type": "pong",
			"time": time.Now().Unix(),
		})
	case "join_channel":
		// Logique pour rejoindre un channel de chat
		s.handleJoinChannel(conn, message)
	case "chat_message":
		// Relayer vers le service de chat via NATS
		s.handleChatMessage(message)
	default:
		conn.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": "Unknown message type",
		})
	}
}

// handleJoinChannel gÃ¨re l'adhÃ©sion Ã  un channel
func (s *Server) handleJoinChannel(conn *websocket.Conn, message map[string]interface{}) {
	channel, ok := message["channel"].(string)
	if !ok {
		conn.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": "Channel name required",
		})
		return
	}

	// Envoyer confirmation
	conn.WriteJSON(map[string]interface{}{
		"type":    "channel_joined",
		"channel": channel,
		"time":    time.Now().Unix(),
	})

	logrus.WithField("channel", channel).Debug("Client joined channel")
}

// handleChatMessage gÃ¨re les messages de chat
func (s *Server) handleChatMessage(message map[string]interface{}) {
	if s.natsConn == nil {
		return
	}

	// Publier le message sur NATS pour le service de chat
	data, err := json.Marshal(message)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal chat message")
		return
	}

	err = s.natsConn.Publish("chat.message", data)
	if err != nil {
		logrus.WithError(err).Error("Failed to publish chat message")
	}
}

// connectToNATS Ã©tablit la connexion NATS
func connectToNATS(cfg config.NATSConfig) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.Name(cfg.ClientID),
		nats.Timeout(cfg.ConnectTimeout),
		nats.ReconnectWait(cfg.ReconnectDelay),
		nats.MaxReconnects(cfg.MaxReconnectAttempts),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logrus.WithError(err).Warn("NATS disconnected")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logrus.Info("NATS reconnected")
		}),
	}

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	logrus.WithField("url", cfg.URL).Info("Connected to NATS")
	return nc, nil
}

// initServiceHealthMonitoring initialise le monitoring des services
func (s *Server) initServiceHealthMonitoring() {
	services := map[string]config.ServiceEndpoint{
		"auth":      s.config.Services.Auth,
		"player":    s.config.Services.Player,
		"world":     s.config.Services.World,
		"combat":    s.config.Services.Combat,
		"inventory": s.config.Services.Inventory,
		"guild":     s.config.Services.Guild,
		"chat":      s.config.Services.Chat,
		"analytics": s.config.Services.Analytics,
	}

	// Initialiser l'Ã©tat de santÃ©
	s.healthMutex.Lock()
	for name, endpoint := range services {
		s.serviceHealth[name] = ServiceHealth{
			Name:         name,
			URL:          endpoint.URL,
			Status:       "unknown",
			LastCheck:    time.Now(),
			ResponseTime: 0,
			ErrorCount:   0,
		}
	}
	s.healthMutex.Unlock()

	// DÃ©marrer le monitoring pÃ©riodique
	go s.monitorServicesHealth()
}

// monitorServicesHealth surveille pÃ©riodiquement la santÃ© des services
func (s *Server) monitorServicesHealth() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkAllServicesHealth()
		}
	}
}

// checkAllServicesHealth vÃ©rifie la santÃ© de tous les services
func (s *Server) checkAllServicesHealth() {
	s.healthMutex.RLock()
	services := make(map[string]ServiceHealth)
	for k, v := range s.serviceHealth {
		services[k] = v
	}
	s.healthMutex.RUnlock()

	for name, health := range services {
		go s.checkServiceHealth(name, health)
	}
}

// checkServiceHealth vÃ©rifie la santÃ© d'un service spÃ©cifique
func (s *Server) checkServiceHealth(name string, health ServiceHealth) {
	start := time.Now()

	// Faire un ping simple au service
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(health.URL + "/health")

	responseTime := time.Since(start).Milliseconds()

	s.healthMutex.Lock()
	defer s.healthMutex.Unlock()

	updatedHealth := health
	updatedHealth.LastCheck = time.Now()
	updatedHealth.ResponseTime = responseTime

	if err != nil || resp.StatusCode != http.StatusOK {
		updatedHealth.ErrorCount++
		if updatedHealth.ErrorCount > 3 {
			updatedHealth.Status = "unhealthy"
		}
	} else {
		updatedHealth.ErrorCount = 0
		updatedHealth.Status = "healthy"
	}

	if resp != nil {
		resp.Body.Close()
	}

	s.serviceHealth[name] = updatedHealth
}
