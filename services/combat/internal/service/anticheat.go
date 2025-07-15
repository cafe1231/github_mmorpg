// internal/service/anticheat.go
package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"combat/internal/config"
	"combat/internal/models"
	"combat/internal/repository"
)

// PlayerStats représente les statistiques d'un joueur pour l'anti-cheat
type PlayerStats struct {
	ActionsPerSecond    float64               `json:"actions_per_second"`
	AverageDamage       float64               `json:"average_damage"`
	MaxDamageDealt      int                   `json:"max_damage_dealt"`
	SuspiciousActions   int                   `json:"suspicious_actions"`
	ViolationLevel      int                   `json:"violation_level"`
	LastViolationTime   time.Time             `json:"last_violation_time"`
	RecentPositions     []models.Position     `json:"recent_positions"`
	ActionHistory       []*models.ActionEvent `json:"action_history"`
}

// AntiCheatService gère la détection des triches
type AntiCheatService struct {
	config        *config.Config
	violationRepo repository.AntiCheatRepositoryInterface  // Corrigé le nom du type
	combatRepo    repository.CombatRepositoryInterface
	
	// Cache des statistiques des joueurs
	playerStats   map[uuid.UUID]*PlayerStats
	actionHistory map[uuid.UUID][]*models.ActionEvent
	mutex         sync.RWMutex
}

// NewAntiCheatService crée une nouvelle instance du service anti-cheat
func NewAntiCheatService(
	cfg *config.Config,
	violationRepo repository.AntiCheatRepositoryInterface,  // Corrigé le nom du type
	combatRepo repository.CombatRepositoryInterface,
) *AntiCheatService {
	return &AntiCheatService{
		config:        cfg,
		violationRepo: violationRepo,
		combatRepo:    combatRepo,
		playerStats:   make(map[uuid.UUID]*PlayerStats),
		actionHistory: make(map[uuid.UUID][]*models.ActionEvent),
	}
}

// ValidateCombatAction valide une action de combat
func (s *AntiCheatService) ValidateCombatAction(action *models.PerformActionRequest, participant *models.CombatParticipant) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Enregistrer l'action
	s.recordAction(action, participant)
	
	// Validations de base
	if err := s.validateBasicAction(action, participant); err != nil {
		s.recordViolation(participant.CharacterID, "invalid_action", err.Error())
		return err
	}
	
	// Vérifier le rate limiting
	if err := s.validateActionRate(participant.CharacterID); err != nil {
		s.recordViolation(participant.CharacterID, "action_spam", err.Error())
		return err
	}
	
	// Vérifier les dégâts
	if action.Type == "attack" || action.Type == "spell" {
		if err := s.validateDamage(action, participant); err != nil {
			s.recordViolation(participant.CharacterID, "impossible_damage", err.Error())
			return err
		}
	}
	
	// Vérifier les positions
	if action.Type == "move" {
		if err := s.validateMovement(action, participant); err != nil {
			s.recordViolation(participant.CharacterID, "impossible_movement", err.Error())
			return err
		}
	}
	
	// Vérifier les cooldowns
	if action.Type == "spell" || action.Type == "item" {
		if err := s.validateCooldowns(action, participant); err != nil {
			s.recordViolation(participant.CharacterID, "cooldown_violation", err.Error())
			return err
		}
	}
	
	// Mettre à jour les statistiques
	// Convertir ActionData en map si nécessaire
	var actionDataMap map[string]interface{}
	if action.ActionData != nil {
		json.Unmarshal(action.ActionData, &actionDataMap)
	}
	
	s.updatePlayerStats(participant.CharacterID, &models.ActionEvent{
		CharacterID: participant.CharacterID,
		ActionType:  action.Type,
		Timestamp:   time.Now(),
		Data:        actionDataMap, // Utiliser la map convertie au lieu de json.RawMessage
	})
	
	return nil
}

// validateDamage vérifie que les dégâts sont réalistes
func (s *AntiCheatService) validateDamage(action *models.PerformActionRequest, attacker *models.CombatParticipant) error {
	// Convertir ActionData depuis json.RawMessage
	var actionData map[string]interface{}
	if err := json.Unmarshal(action.ActionData, &actionData); err != nil {
		return fmt.Errorf("failed to unmarshal action data: %w", err)
	}
	
	// Extraire les dégâts de l'action
	damageData, ok := actionData["damage"]
	if !ok {
		return nil // Pas de dégâts spécifiés
	}
	
	damage, ok := damageData.(float64)
	if !ok {
		return fmt.Errorf("invalid damage data type")
	}
	
	// Calculer les dégâts maximaux possibles basés sur les stats réelles du participant
	maxPossibleDamage := s.calculateMaxPossibleDamage(attacker, action.Type)
	
	if int(damage) > maxPossibleDamage {
		return fmt.Errorf("damage too high: %d (max: %d)", int(damage), maxPossibleDamage)
	}
	
	return nil
}

// validateMovement vérifie la validité d'un mouvement
func (s *AntiCheatService) validateMovement(action *models.PerformActionRequest, participant *models.CombatParticipant) error {
	// Convertir ActionData depuis json.RawMessage
	var actionData map[string]interface{}
	if err := json.Unmarshal(action.ActionData, &actionData); err != nil {
		return fmt.Errorf("failed to unmarshal action data: %w", err)
	}
	
	// Extraire la nouvelle position
	positionData, ok := actionData["position"]
	if !ok {
		return fmt.Errorf("missing position data")
	}
	
	positionMap, ok := positionData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid position data format")
	}
	
	x, ok := positionMap["x"].(float64)
	if !ok {
		return fmt.Errorf("invalid x coordinate")
	}
	
	y, ok := positionMap["y"].(float64)
	if !ok {
		return fmt.Errorf("invalid y coordinate")
	}
	
	z, ok := positionMap["z"].(float64)
	if !ok {
		return fmt.Errorf("invalid z coordinate")
	}
	
	newPosition := models.Position{X: x, Y: y, Z: z}
	
	// Calculer la distance parcourue
	distance := s.calculateDistance(participant.Position, newPosition)
	
	// Calculer la vitesse maximale autorisée (basée sur les stats du participant)
	maxSpeed := s.calculateMaxSpeed(participant)
	
	// Temps écoulé depuis la dernière action
	timeElapsed := time.Since(*participant.LastActionAt).Seconds()
	maxDistance := maxSpeed * timeElapsed
	
	if distance > maxDistance {
		return fmt.Errorf("movement too fast: %.2f units in %.2f seconds (max: %.2f)", 
			distance, timeElapsed, maxDistance)
	}
	
	// Vérifier la position Y (anti-fly)
	if err := s.validateYPosition(y); err != nil {
		return err
	}
	
	return nil
}

// validateYPosition vérifie la validité de la position Y
func (s *AntiCheatService) validateYPosition(y float64) error {
	// Vérifier que le joueur n'est pas sous le sol (arbitraire)
	if y < -100 {
		return fmt.Errorf("position below ground level: y=%.2f", y)
	}
	
	// Vérifier que le joueur n'est pas trop haut (anti-fly hack)
	if y > 1000 {
		return fmt.Errorf("position too high: y=%.2f", y)
	}
	
	return nil
}

// calculateMaxPossibleDamage calcule les dégâts maximaux théoriques
func (s *AntiCheatService) calculateMaxPossibleDamage(attacker *models.CombatParticipant, damageType string) int {
	// Utiliser les stats réelles du participant au lieu de Stats inexistant
	baseDamage := attacker.Damage
	
	// Facteur de critique maximum
	criticalMultiplier := 1.0 + (attacker.CritChance * 2.5) // Crit fait 2.5x les dégâts
	
	// Facteur de buffs maximum estimé
	buffMultiplier := 2.0
	
	return int(float64(baseDamage) * criticalMultiplier * buffMultiplier)
}

// calculateMaxSpeed calcule la vitesse maximale d'un participant
func (s *AntiCheatService) calculateMaxSpeed(participant *models.CombatParticipant) float64 {
	// Vitesse de base
	baseSpeed := 10.0 // unités par seconde
	
	// Modificateur basé sur l'attack speed (approximation)
	speedModifier := participant.AttackSpeed
	
	return baseSpeed * speedModifier
}

// calculateDistance calcule la distance entre deux positions
func (s *AntiCheatService) calculateDistance(pos1, pos2 models.Position) float64 {
	dx := pos2.X - pos1.X
	dy := pos2.Y - pos1.Y
	dz := pos2.Z - pos1.Z
	
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// validateActionRate vérifie le taux d'actions
func (s *AntiCheatService) validateActionRate(characterID uuid.UUID) error {
	history := s.actionHistory[characterID]
	if len(history) == 0 {
		return nil
	}
	
	// Compter les actions dans la dernière seconde
	now := time.Now()
	recentActions := 0
	
	for _, event := range history {
		if now.Sub(event.Timestamp) <= time.Second {
			recentActions++
		}
	}
	
	// Limite d'actions par seconde
	maxActionsPerSecond := 5
	
	if recentActions > maxActionsPerSecond {
		return fmt.Errorf("too many actions: %d in 1 second (max: %d)", 
			recentActions, maxActionsPerSecond)
	}
	
	return nil
}

// validateCooldowns vérifie les cooldowns des sorts et objets
func (s *AntiCheatService) validateCooldowns(action *models.PerformActionRequest, participant *models.CombatParticipant) error {
	// Pour l'instant, validation basique
	// TODO: Intégrer avec le service de sorts pour vérifier les vrais cooldowns
	
	if action.Type == "spell" {
		// Cooldown minimum entre les sorts
		if participant.LastActionAt != nil {
			timeSinceLastAction := time.Since(*participant.LastActionAt)
			minCooldown := time.Second * 2 // 2 secondes minimum
			
			if timeSinceLastAction < minCooldown {
				return fmt.Errorf("spell on cooldown: %v remaining", 
					minCooldown - timeSinceLastAction)
			}
		}
	}
	
	return nil
}

// recordViolation enregistre une violation
func (s *AntiCheatService) recordViolation(characterID uuid.UUID, violationType, description string) {
	violation := &models.AntiCheatViolation{
		ID:          uuid.New(),
		CharacterID: characterID,
		Type:        violationType,
		Description: description,  // Corrigé: utilise Description au lieu de Details
		Severity:    s.calculateSeverityString(violationType),  // Corrigé: retourne une string
		Data:        make(map[string]interface{}),
		Timestamp:   time.Now(),
		CreatedAt:   time.Now(),
	}
	
	if err := s.violationRepo.CreateViolation(violation); err != nil {
		logrus.WithError(err).Error("Failed to record violation")
	}
	
	logrus.WithFields(logrus.Fields{
		"character_id": characterID,
		"type":         violationType,
		"description":  description,
		"severity":     violation.Severity,
	}).Warn("Anti-cheat violation detected")
}

// calculateSeverityString calcule la sévérité d'une violation (retourne une string)
func (s *AntiCheatService) calculateSeverityString(violationType string) string {
	severityMap := map[string]string{
		"action_spam":         "low",
		"cooldown_violation":  "medium",
		"impossible_movement": "high",
		"impossible_damage":   "high",
		"invalid_action":      "medium",
	}
	
	if severity, exists := severityMap[violationType]; exists {
		return severity
	}
	
	return "low" // Sévérité par défaut
}

// updatePlayerStats met à jour les statistiques d'un joueur
func (s *AntiCheatService) updatePlayerStats(characterID uuid.UUID, event *models.ActionEvent) {
	if _, exists := s.playerStats[characterID]; !exists {
		s.playerStats[characterID] = &PlayerStats{
			ActionHistory: make([]*models.ActionEvent, 0),
		}
	}
	
	stats := s.playerStats[characterID]
	stats.ActionHistory = append(stats.ActionHistory, event)
	
	// Garder seulement les 50 derniers événements
	if len(stats.ActionHistory) > 50 {
		stats.ActionHistory = stats.ActionHistory[len(stats.ActionHistory)-50:]
	}
	
	// Calculer les actions par seconde sur la dernière minute
	now := time.Now()
	recentActions := 0
	
	for _, action := range stats.ActionHistory {
		if now.Sub(action.Timestamp) <= time.Minute {
			recentActions++
		}
	}
	
	stats.ActionsPerSecond = float64(recentActions) / 60.0
}

// GetPlayerStats retourne les statistiques d'un joueur
func (s *AntiCheatService) GetPlayerStats(characterID uuid.UUID) *PlayerStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if stats, exists := s.playerStats[characterID]; exists {
		return stats
	}
	
	return &PlayerStats{}
}

// IsPlayerBanned vérifie si un joueur est banni
func (s *AntiCheatService) IsPlayerBanned(characterID uuid.UUID) (bool, string, error) {
	ban, err := s.violationRepo.GetActiveBan(characterID)
	if err != nil {
		return false, "", err
	}
	
	if ban == nil {
		return false, "", nil
	}
	
	// Vérifier si le ban est encore actif
	if ban.ExpiresAt != nil && time.Now().After(*ban.ExpiresAt) {
		return false, "", nil
	}
	
	return true, ban.Reason, nil
}

// Méthodes privées

// recordAction enregistre une action dans l'historique
func (s *AntiCheatService) recordAction(action *models.PerformActionRequest, participant *models.CombatParticipant) {
	// Convertir ActionData en map pour les événements
	var actionDataMap map[string]interface{}
	if action.ActionData != nil {
		json.Unmarshal(action.ActionData, &actionDataMap)
	}
	
	event := &models.ActionEvent{
		CharacterID: participant.CharacterID,
		ActionType:  action.Type,
		Timestamp:   time.Now(),
		Data:        actionDataMap,
	}
	
	// Ajouter à l'historique
	if _, exists := s.actionHistory[participant.CharacterID]; !exists {
		s.actionHistory[participant.CharacterID] = make([]*models.ActionEvent, 0)
	}
	
	s.actionHistory[participant.CharacterID] = append(s.actionHistory[participant.CharacterID], event)
	
	// Garder seulement les 100 dernières actions
	history := s.actionHistory[participant.CharacterID]
	if len(history) > 100 {
		s.actionHistory[participant.CharacterID] = history[len(history)-100:]
	}
	
	// Mettre à jour les stats de comportement
	s.updatePlayerStats(participant.CharacterID, event)
}

// validateBasicAction effectue les validations de base
func (s *AntiCheatService) validateBasicAction(action *models.PerformActionRequest, participant *models.CombatParticipant) error {
	// Vérifier que le joueur est vivant
	if participant.Status != "alive" {
		return fmt.Errorf("dead player cannot perform actions")
	}
	
	// Vérifier les données d'action
	if action.ActionData == nil || len(action.ActionData) == 0 {
		return fmt.Errorf("missing action data")
	}
	
	// Vérifier les cibles
	if len(action.Targets) == 0 && action.Type != "move" && action.Type != "defend" {
		return fmt.Errorf("action requires targets")
	}
	
	// Vérifier le nombre de cibles
	maxTargets := s.getMaxTargetsForAction(action.Type)
	if len(action.Targets) > maxTargets {
		return fmt.Errorf("too many targets: %d (max: %d)", len(action.Targets), maxTargets)
	}
	
	return nil
}

// getMaxTargetsForAction retourne le nombre maximum de cibles pour une action
func (s *AntiCheatService) getMaxTargetsForAction(actionType string) int {
	limits := map[string]int{
		"attack": 1,  // Une cible pour les attaques normales
		"spell":  5,  // Jusqu'à 5 cibles pour les sorts de zone
		"item":   1,  // Une cible pour les objets
		"move":   0,  // Pas de cible pour les mouvements
		"defend": 0,  // Pas de cible pour la défense
	}
	
	if limit, exists := limits[actionType]; exists {
		return limit
	}
	
	return 1 // Par défaut
}