package handler

import (
"net/http"
"github.com/gin-gonic/gin"
"github.com/gorilla/websocket"
"github.com/sirupsen/logrus"
"combat/internal/service"
)

// WebSocketHandler gère les connexions WebSocket
type WebSocketHandler struct {
upgrader        websocket.Upgrader
realtimeService service.RealtimeServiceInterface
}

// NewWebSocketHandler crée une nouvelle instance du handler WebSocket
func NewWebSocketHandler(realtimeService service.RealtimeServiceInterface) *WebSocketHandler {
return &WebSocketHandler{
upgrader: websocket.Upgrader{
ReadBufferSize:  1024,
WriteBufferSize: 1024,
CheckOrigin: func(r *http.Request) bool {
return true // En production, vérifier l'origine
},
},
realtimeService: realtimeService,
}
}

// HandleWebSocket gère les connexions WebSocket
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
if err != nil {
logrus.WithError(err).Error("Failed to upgrade WebSocket connection")
return
}
defer conn.Close()

// Récupérer l'ID utilisateur (si authentifié)
userID := "anonymous"
if uid, exists := c.Get("user_id"); exists {
userID = uid.(string)
}

// Ajouter la connexion au service temps réel
if err := h.realtimeService.AddConnection(conn, userID); err != nil {
logrus.WithError(err).Error("Failed to add WebSocket connection")
return
}
defer h.realtimeService.RemoveConnection(conn)

// Envoyer un message de bienvenue
welcomeMsg := map[string]interface{}{
"type":    "welcome",
"message": "Connected to Combat Service",
"user_id": userID,
}
if err := conn.WriteJSON(welcomeMsg); err != nil {
logrus.WithError(err).Error("Failed to send welcome message")
return
}

// Boucle de lecture des messages
for {
var message map[string]interface{}
err := conn.ReadJSON(&message)
if err != nil {
if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
logrus.WithError(err).Error("WebSocket unexpected close error")
}
break
}

// Traiter le message reçu
h.handleMessage(conn, userID, message)
}

logrus.WithField("user_id", userID).Info("WebSocket connection closed")
}

// handleMessage traite un message WebSocket reçu
func (h *WebSocketHandler) handleMessage(conn *websocket.Conn, userID string, message map[string]interface{}) {
messageType, ok := message["type"].(string)
if !ok {
return
}

switch messageType {
case "ping":
// Répondre au ping
response := map[string]interface{}{
"type":    "pong",
"message": "pong",
}
conn.WriteJSON(response)

case "join_combat":
// Rejoindre un combat
sessionID, ok := message["session_id"].(string)
if ok {
response := map[string]interface{}{
"type":       "combat_joined",
"session_id": sessionID,
"message":    "Joined combat session",
}
conn.WriteJSON(response)
}

case "leave_combat":
// Quitter un combat
response := map[string]interface{}{
"type":    "combat_left",
"message": "Left combat session",
}
conn.WriteJSON(response)

default:
// Message non reconnu
response := map[string]interface{}{
"type":    "error",
"message": "Unknown message type",
}
conn.WriteJSON(response)
}
}
