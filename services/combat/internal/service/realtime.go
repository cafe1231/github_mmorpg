package service

import (
"context"
"github.com/gorilla/websocket"
"github.com/sirupsen/logrus"
"combat/internal/config"
)

// RealtimeServiceInterface définit les méthodes du service temps réel
type RealtimeServiceInterface interface {
Start() error
Stop() error
BroadcastToSession(sessionID string, message interface{}) error
AddConnection(conn *websocket.Conn, userID string) error
RemoveConnection(conn *websocket.Conn) error
}

// RealtimeService implémente l'interface RealtimeServiceInterface
type RealtimeService struct {
config      *config.Config
connections map[*websocket.Conn]string
ctx         context.Context
cancel      context.CancelFunc
}

// NewRealtimeService crée une nouvelle instance du service temps réel
func NewRealtimeService(cfg *config.Config) RealtimeServiceInterface {
ctx, cancel := context.WithCancel(context.Background())

return &RealtimeService{
config:      cfg,
connections: make(map[*websocket.Conn]string),
ctx:         ctx,
cancel:      cancel,
}
}

// Start démarre le service temps réel
func (s *RealtimeService) Start() error {
logrus.Info("Realtime service started")
return nil
}

// Stop arrête le service temps réel
func (s *RealtimeService) Stop() error {
s.cancel()

// Fermer toutes les connexions
for conn := range s.connections {
conn.Close()
}

logrus.Info("Realtime service stopped")
return nil
}

// BroadcastToSession diffuse un message à une session
func (s *RealtimeService) BroadcastToSession(sessionID string, message interface{}) error {
// TODO: Implement broadcast logic
return nil
}

// AddConnection ajoute une connexion WebSocket
func (s *RealtimeService) AddConnection(conn *websocket.Conn, userID string) error {
s.connections[conn] = userID
logrus.WithField("user_id", userID).Info("WebSocket connection added")
return nil
}

// RemoveConnection supprime une connexion WebSocket
func (s *RealtimeService) RemoveConnection(conn *websocket.Conn) error {
if userID, exists := s.connections[conn]; exists {
delete(s.connections, conn)
logrus.WithField("user_id", userID).Info("WebSocket connection removed")
}
return nil
}
