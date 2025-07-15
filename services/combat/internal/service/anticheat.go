package service

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"combat/internal/config"
	"combat/internal/models"
	"combat/internal/repository"
	"combat/internal/external"
)

// AntiCheatServiceInterface définit les méthodes du service anti-cheat
type AntiCheatServiceInterface interface {
	ValidateAction(action *models.CombatActionRequest, participant *models.CombatParticipant) error
	ValidateMovement(characterID uuid.UUID, oldPos, newPos map[string]interface{}, deltaTime time.Duration) error
	ValidateSpellCast(characterID uuid.UUID, spell *models.Spell, castTime time.Duration) error
	ValidateDamage(attacker, defender *models.CombatParticipant, damage int, damageType string) error
	CheckSuspiciousActivity(characterID uuid.UUID) (*models.SuspicionReport, error)
	ReportViolation(violation *models.AntiCheatViolation) error
	GetPlayerViolations(characterID uuid.UUID) ([]*models.AntiCheatViolation, error)
	IsPlayerBanned(characterID uuid.UUID) (bool, string, error)
}

// AntiCheatService implémente l'interface AntiCheatServiceInterface
type AntiCheatService struct {
	config        *config.Config
	violationRepo repository.AntiCheatRepositoryInterface
	playerClient  external.PlayerClientInterface
	
	// Caches pour la détection de patterns
	actionHistory map[uuid.UUID][]*models.ActionEvent
	playerStats   map[uuid.UUID]*models.PlayerBehaviorStats
}

// NewAntiCheatService crée une nouvelle instance du service anti-cheat
func NewAntiCheatService(
	cfg *config.Config,
	violationRepo repository.AntiCheatRepositoryInterface,
	playerClient external.PlayerClientInterface,
) AntiCheatServiceInterface {
	return &AntiCheatService{
		config:        cfg,
		violationRepo: violationRepo,
		playerClient:  playerClient,
		actionHistory: make(map[uuid.UUID][]*models.ActionEvent),
		playerStats:   make(map[uuid.UUID]*models.PlayerBehaviorStats),
	}
}

// ValidateAction valide une action de combat contre le cheating
func (s *AntiCheatService) ValidateAction(action *models.CombatActionRequest, participant *models.CombatParticipant) error {
	// Enregistrer l'action pour l'historique
	s.recordAction(action, participant)
	
	// Vérifications de base
	if err := s.validateBasicAction(action, participant); err != nil {
		return s.reportSuspiciousActivity(participant.CharacterID, "invalid_action", err.Error(), "medium")
	}
	
	// Vérifier le rate limiting
	if err := s.validateActionRate(participant.CharacterID, action.Type); err != nil {
		return s.reportSuspiciousActivity(participant.CharacterID, "action_spam", err.Error(), "high")
	}
	
	// Vérifier les patterns suspects
	if err := s.detectSuspiciousPatterns(participant.CharacterID); err != nil {
		return s.reportSuspiciousActivity(participant.CharacterID, "suspicious_pattern", err.Error(), "medium")
	}
	
	// Validations spécifiques par type d'action
	switch action.Type {
	case "attack":
		return s.validateAttackAction(action, participant)
	case "spell":
		return s.validateSpellAction(action, participant)
	case "move":
		return s.validateMoveAction(action, participant)
	case "item":
		return s.validateItemAction(action, participant)
	}
	
	return nil
}

// ValidateMovement valide un mouvement de joueur
func (s *AntiCheatService) ValidateMovement(characterID uuid.UUID, oldPos, newPos map[string]interface{}, deltaTime time.Duration) error {
	// Calculer la distance parcourue
	distance := s.calculateDistance(oldPos, newPos)
	
	// Calculer la vitesse
	speed := distance / deltaTime.Seconds()
	
	// Vérifier la vitesse maximale autorisée
	maxSpeed := s.getMaxMovementSpeed(characterID)
	if speed > maxSpeed*1.2 { // 20% de tolérance pour la latence
		violation := &models.AntiCheatViolation{
			CharacterID:  characterID,
			Type:         "speed_hack",
			Description:  fmt.Sprintf("Movement speed %.2f exceeds maximum %.2f", speed, maxSpeed),
			Severity:     "high",
			Data: map[string]interface{}{
				"calculated_speed": speed,
				"max_speed":       maxSpeed,
				"distance":        distance,
				"delta_time":      deltaTime.Seconds(),
				"old_position":    oldPos,
				"new_position":    newPos,
			},
			Timestamp: time.Now(),
		}
		
		return s.ReportViolation(violation)
	}
	
	// Vérifier la téléportation
	if speed > maxSpeed*3 { // Téléportation détectée
		violation := &models.AntiCheatViolation{
			CharacterID:  characterID,
			Type:         "teleport_hack",
			Description:  fmt.Sprintf("Possible teleportation detected: speed %.2f", speed),
			Severity:     "critical",
			Data: map[string]interface{}{
				"calculated_speed": speed,
				"distance":        distance,
				"delta_time":      deltaTime.Seconds(),
			},
			Timestamp: time.Now(),
		}
		
		return s.ReportViolation(violation)
	}
	
	// Vérifier les collisions avec le terrain
	if err := s.validateTerrainCollision(newPos); err != nil {
		violation := &models.AntiCheatViolation{
			CharacterID:  characterID,
			Type:         "no_clip",
			Description:  "Player moved through solid terrain",
			Severity:     "high",
			Data: map[string]interface{}{
				"position": newPos,
				"error":    err.Error(),
			},
			Timestamp: time.Now(),
		}
		
		return s.ReportViolation(violation)
	}
	
	return nil
}

// ValidateSpellCast valide le lancement d'un sort
func (s *AntiCheatService) ValidateSpellCast(characterID uuid.UUID, spell *models.Spell, castTime time.Duration) error {
	// Vérifier le temps d'incantation
	expectedCastTime := time.Duration(spell.CastTime) * time.Millisecond
	tolerance := 200 * time.Millisecond // 200ms de tolérance
	
	if castTime < (expectedCastTime - tolerance) {
		violation := &models.AntiCheatViolation{
			CharacterID:  characterID,
			Type:         "instant_cast",
			Description:  fmt.Sprintf("Spell cast too fast: %v (expected: %v)", castTime, expectedCastTime),
			Severity:     "high",
			Data: map[string]interface{}{
				"spell_id":          spell.ID,
				"spell_name":        spell.Name,
				"actual_cast_time":  castTime.Milliseconds(),
				"expected_cast_time": expectedCastTime.Milliseconds(),
			},
			Timestamp: time.Now(),
		}
		
		return s.ReportViolation(violation)
	}
	
	// Vérifier la fréquence de lancement
	if err := s.validateSpellFrequency(characterID, spell.ID); err != nil {
		return err
	}
	
	return nil
}

// ValidateDamage valide les dégâts infligés
func (s *AntiCheatService) ValidateDamage(attacker, defender *models.CombatParticipant, damage int, damageType string) error {
	// Calculer les dégâts théoriques maximaux
	maxDamage := s.calculateMaxPossibleDamage(attacker, damageType)
	
	// Vérifier si les dégâts sont impossibles
	maxDamageFloat := float64(maxDamage) * 1.5 // 50% de tolérance pour les critiques et buffs
	if damage > int(maxDamageFloat) {
		violation := &models.AntiCheatViolation{
			CharacterID:  attacker.CharacterID,
			Type:         "damage_hack",
			Description:  fmt.Sprintf("Damage %d exceeds maximum possible %d", damage, maxDamage),
			Severity:     "critical",
			Data: map[string]interface{}{
				"actual_damage":  damage,
				"max_damage":     maxDamage,
				"damage_type":    damageType,
				"attacker_stats": attacker.Stats,
				"defender_id":    defender.CharacterID,
			},
			Timestamp: time.Now(),
		}
		
		return s.ReportViolation(violation)
	}
	
	// Vérifier les patterns de dégâts suspects
	if err := s.validateDamagePattern(attacker.CharacterID, damage); err != nil {
		return err
	}
	
	return nil
}

// CheckSuspiciousActivity vérifie l'activité suspecte d'un joueur
func (s *AntiCheatService) CheckSuspiciousActivity(characterID uuid.UUID) (*models.SuspicionReport, error) {
	report := &models.SuspicionReport{
		CharacterID: characterID,
		Timestamp:   time.Now(),
		Factors:     make([]models.SuspicionFactor, 0),
	}
	
	// Vérifier l'historique des violations
	violations, err := s.GetPlayerViolations(characterID)
	if err != nil {
		return nil, err
	}
	
	// Analyser les violations récentes
	recentViolations := s.getRecentViolations(violations, 24*time.Hour)
	if len(recentViolations) > 3 {
		report.Factors = append(report.Factors, models.SuspicionFactor{
			Type:        "frequent_violations",
			Severity:    "high",
			Description: fmt.Sprintf("%d violations in last 24 hours", len(recentViolations)),
			Weight:      0.8,
		})
	}
	
	// Analyser les statistiques de comportement
	stats := s.getPlayerStats(characterID)
	if stats != nil {
		// Taux de coups critiques anormal
		if stats.CriticalHitRate > 0.5 { // Plus de 50% de critiques
			report.Factors = append(report.Factors, models.SuspicionFactor{
				Type:        "abnormal_crit_rate",
				Severity:    "medium",
				Description: fmt.Sprintf("Critical hit rate: %.2f%%", stats.CriticalHitRate*100),
				Weight:      0.6,
			})
		}
		
		// Précision anormale
		if stats.AccuracyRate > 0.95 { // Plus de 95% de précision
			report.Factors = append(report.Factors, models.SuspicionFactor{
				Type:        "abnormal_accuracy",
				Severity:    "medium",
				Description: fmt.Sprintf("Accuracy rate: %.2f%%", stats.AccuracyRate*100),
				Weight:      0.5,
			})
		}
		
		// Actions par minute anormales
		if stats.ActionsPerMinute > 120 { // Plus de 2 actions par seconde
			report.Factors = append(report.Factors, models.SuspicionFactor{
				Type:        "abnormal_apm",
				Severity:    "high",
				Description: fmt.Sprintf("Actions per minute: %.1f", stats.ActionsPerMinute),
				Weight:      0.7,
			})
		}
	}
	
	// Calculer le score de suspicion total
	report.SuspicionScore = s.calculateSuspicionScore(report.Factors)
	
	// Déterminer le niveau de risque
	if report.SuspicionScore >= 0.8 {
		report.RiskLevel = "critical"
	} else if report.SuspicionScore >= 0.6 {
		report.RiskLevel = "high"
	} else if report.SuspicionScore >= 0.4 {
		report.RiskLevel = "medium"
	} else {
		report.RiskLevel = "low"
	}
	
	return report, nil
}

// ReportViolation signale une violation anti-cheat
func (s *AntiCheatService) ReportViolation(violation *models.AntiCheatViolation) error {
	// Sauvegarder la violation
	if err := s.violationRepo.CreateViolation(violation); err != nil {
		return fmt.Errorf("failed to save violation: %w", err)
	}
	
	// Logger la violation
	logrus.WithFields(logrus.Fields{
		"character_id": violation.CharacterID,
		"type":         violation.Type,
		"severity":     violation.Severity,
		"description":  violation.Description,
	}).Warn("Anti-cheat violation detected")
	
	// Actions automatiques selon la gravité
	switch violation.Severity {
	case "critical":
		// Bannissement temporaire immédiat
		return s.applyTemporaryBan(violation.CharacterID, 24*time.Hour, "Automatic ban: "+violation.Type)
		
	case "high":
		// Vérifier si c'est une récidive
		violations, _ := s.GetPlayerViolations(violation.CharacterID)
		recentHighViolations := 0
		for _, v := range violations {
			if v.Severity == "high" && time.Since(v.Timestamp) < 7*24*time.Hour {
				recentHighViolations++
			}
		}
		
		if recentHighViolations >= 3 {
			return s.applyTemporaryBan(violation.CharacterID, 12*time.Hour, "Multiple high severity violations")
		}
	}
	
	return nil
}

// GetPlayerViolations récupère les violations d'un joueur
func (s *AntiCheatService) GetPlayerViolations(characterID uuid.UUID) ([]*models.AntiCheatViolation, error) {
	return s.violationRepo.GetPlayerViolations(characterID)
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
func (s *AntiCheatService) recordAction(action *models.CombatActionRequest, participant *models.CombatParticipant) {
	event := &models.ActionEvent{
		CharacterID: participant.CharacterID,
		ActionType:  action.Type,
		Timestamp:   time.Now(),
		Data:        action.ActionData,
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
func (s *AntiCheatService) validateBasicAction(action *models.CombatActionRequest, participant *models.CombatParticipant) error {
	// Vérifier que le joueur est vivant
	if participant.Status != "alive" {
		return fmt.Errorf("dead player cannot perform actions")
	}
	
	// Vérifier les données d'action
	if action.ActionData == nil {
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

// validateActionRate vérifie le taux d'actions
func (s *AntiCheatService) validateActionRate(characterID uuid.UUID, actionType string) error {
	history := s.actionHistory[characterID]
	if len(history) == 0 {
		return nil
	}
	
	// Compter les actions du même type dans la dernière seconde
	now := time.Now()
	recentActions := 0
	
	for i := len(history) - 1; i >= 0; i-- {
		if now.Sub(history[i].Timestamp) > time.Second {
			break
		}
		if history[i].ActionType == actionType {
			recentActions++
		}
	}
	
	// Limites par type d'action
	limits := map[string]int{
		"attack": 2,  // 2 attaques par seconde max
		"spell":  1,  // 1 sort par seconde max
		"move":   10, // 10 mouvements par seconde max
		"item":   1,  // 1 objet par seconde max
	}
	
	limit, exists := limits[actionType]
	if !exists {
		limit = 3 // Limite par défaut
	}
	
	if recentActions > limit {
		return fmt.Errorf("action rate limit exceeded: %d %s actions in 1 second (limit: %d)", 
			recentActions, actionType, limit)
	}
	
	return nil
}

// detectSuspiciousPatterns détecte les patterns suspects
func (s *AntiCheatService) detectSuspiciousPatterns(characterID uuid.UUID) error {
	history := s.actionHistory[characterID]
	if len(history) < 10 {
		return nil // Pas assez de données
	}
	
	// Détecter les patterns trop réguliers (bot-like)
	if s.detectBotPattern(history) {
		return fmt.Errorf("suspicious bot-like pattern detected")
	}
	
	// Détecter les séquences impossibles
	if s.detectImpossibleSequence(history) {
		return fmt.Errorf("impossible action sequence detected")
	}
	
	return nil
}

// validateAttackAction valide une action d'attaque
func (s *AntiCheatService) validateAttackAction(action *models.CombatActionRequest, participant *models.CombatParticipant) error {
	// Vérifier que le joueur a une arme équipée (si nécessaire)
	// TODO: Intégration avec le service d'inventaire
	
	// Vérifier la portée des cibles
	for _, targetID := range action.Targets {
		if err := s.validateTargetRange(participant, targetID, 5.0); err != nil {
			return fmt.Errorf("target out of range: %w", err)
		}
	}
	
	return nil
}

// validateSpellAction valide une action de sort
func (s *AntiCheatService) validateSpellAction(action *models.CombatActionRequest, participant *models.CombatParticipant) error {
	spellIDStr, ok := action.ActionData["spell_id"].(string)
	if !ok {
		return fmt.Errorf("spell_id missing from action data")
	}
	
	_, err := uuid.Parse(spellIDStr)
	if err != nil {
		return fmt.Errorf("invalid spell_id format")
	}
	
	// Vérifier que le joueur connaît ce sort
	// TODO: Vérification via le service de sorts
	
	// Vérifier le coût en mana
	// TODO: Récupérer la mana depuis le service Player
	// Pour l'instant, on suppose que c'est OK
	
	// Vérifier la portée du sort
	for _, targetID := range action.Targets {
		if err := s.validateTargetRange(participant, targetID, 30.0); err != nil {
			return fmt.Errorf("spell target out of range: %w", err)
		}
	}
	
	return nil
}

// validateMoveAction valide une action de mouvement
func (s *AntiCheatService) validateMoveAction(action *models.CombatActionRequest, participant *models.CombatParticipant) error {
	if _, okX := action.ActionData["x"].(float64); !okX {
		return fmt.Errorf("invalid x coordinate")
	}
	if _, okY := action.ActionData["y"].(float64); !okY {
		return fmt.Errorf("invalid y coordinate")
	}
	if _, okZ := action.ActionData["z"].(float64); !okZ {
		return fmt.Errorf("invalid z coordinate")
	}
	return nil
}

// validateItemAction valide une action d'objet
func (s *AntiCheatService) validateItemAction(action *models.CombatActionRequest, participant *models.CombatParticipant) error {
	itemType, ok := action.ActionData["item_type"].(string)
	if !ok {
		return fmt.Errorf("item_type missing from action data")
	}
	
	// Vérifier que le joueur possède cet objet
	// TODO: Intégration avec le service d'inventaire
	
	// Vérifier les cooldowns d'objets
	if err := s.validateItemCooldown(participant.CharacterID, itemType); err != nil {
		return fmt.Errorf("item on cooldown: %w", err)
	}
	
	return nil
}

// validateSpellFrequency vérifie la fréquence de lancement des sorts
func (s *AntiCheatService) validateSpellFrequency(characterID, spellID uuid.UUID) error {
	history := s.actionHistory[characterID]
	if len(history) == 0 {
		return nil
	}
	
	// Compter les lancements de ce sort dans la dernière minute
	now := time.Now()
	spellCasts := 0
	
	for i := len(history) - 1; i >= 0; i-- {
		if now.Sub(history[i].Timestamp) > time.Minute {
			break
		}
		
		if history[i].ActionType == "spell" {
			if sid, exists := history[i].Data["spell_id"]; exists {
				if sidStr, ok := sid.(string); ok && sidStr == spellID.String() {
					spellCasts++
				}
			}
		}
	}
	
	// Limite arbitraire: max 10 fois le même sort par minute
	if spellCasts > 10 {
		violation := &models.AntiCheatViolation{
			CharacterID:  characterID,
			Type:         "spell_spam",
			Description:  fmt.Sprintf("Spell cast %d times in 1 minute", spellCasts),
			Severity:     "medium",
			Data: map[string]interface{}{
				"spell_id":    spellID,
				"cast_count":  spellCasts,
				"time_window": "1 minute",
			},
			Timestamp: time.Now(),
		}
		
		return s.ReportViolation(violation)
	}
	
	return nil
}

// validateDamagePattern vérifie les patterns de dégâts suspects
func (s *AntiCheatService) validateDamagePattern(characterID uuid.UUID, damage int) error {
	stats := s.getPlayerStats(characterID)
	if stats == nil {
		return nil
	}
	
	// Ajouter ce dégât aux statistiques
	stats.TotalDamageDealt += damage
	stats.DamageEvents++
	stats.AverageDamage = float64(stats.TotalDamageDealt) / float64(stats.DamageEvents)
	
	// Vérifier si les dégâts sont anormalement constants
	if stats.DamageEvents > 20 {
		variance := s.calculateDamageVariance(characterID)
		if variance < 0.05 { // Variance trop faible = dégâts trop constants
			violation := &models.AntiCheatViolation{
				CharacterID:  characterID,
				Type:         "constant_damage",
				Description:  fmt.Sprintf("Damage variance too low: %.4f", variance),
				Severity:     "medium",
				Data: map[string]interface{}{
					"damage_variance": variance,
					"average_damage":  stats.AverageDamage,
					"damage_events":   stats.DamageEvents,
				},
				Timestamp: time.Now(),
			}
			
			return s.ReportViolation(violation)
		}
	}
	
	return nil
}

// calculateDistance calcule la distance entre deux positions
func (s *AntiCheatService) calculateDistance(pos1, pos2 map[string]interface{}) float64 {
	if pos1 == nil || pos2 == nil {
		return 0
	}
	
	x1, _ := pos1["x"].(float64)
	y1, _ := pos1["y"].(float64)
	z1, _ := pos1["z"].(float64)
	
	x2, _ := pos2["x"].(float64)
	y2, _ := pos2["y"].(float64)
	z2, _ := pos2["z"].(float64)
	
	return math.Sqrt(math.Pow(x2-x1, 2) + math.Pow(y2-y1, 2) + math.Pow(z2-z1, 2))
}

// getMaxMovementSpeed récupère la vitesse maximale d'un joueur
func (s *AntiCheatService) getMaxMovementSpeed(characterID uuid.UUID) float64 {
	// TODO: Récupérer la vitesse réelle depuis les stats du joueur
	baseSpeed := 7.0 // unités par seconde
	
	// Appliquer les modificateurs de vitesse
	// TODO: Intégration avec le service d'effets pour les buffs de vitesse
	
	return baseSpeed
}

// validateTerrainCollision vérifie les collisions avec le terrain
func (s *AntiCheatService) validateTerrainCollision(position map[string]interface{}) error {
	// TODO: Intégration avec le service World pour vérifier les collisions
	// Pour l'instant, vérifications basiques
	
	y, ok := position["y"].(float64)
	if !ok {
		return fmt.Errorf("invalid y coordinate")
	}
	
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
	// TODO: Calcul basé sur les vraies stats du joueur
	baseDamage := 100 // Dégâts de base arbitraires
	
	// Pour l'instant, on utilise des stats par défaut car Stats n'existe pas dans CombatParticipant
	// TODO: Intégrer avec le service Player pour récupérer les vraies stats
	
	// Facteur de critique maximum
	criticalMultiplier := 2.5
	
	// Facteur de buffs maximum
	buffMultiplier := 2.0
	
	return int(float64(baseDamage) * criticalMultiplier * buffMultiplier)
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

// validateTargetRange vérifie la portée d'une cible
func (s *AntiCheatService) validateTargetRange(attacker *models.CombatParticipant, targetID uuid.UUID, maxRange float64) error {
	// TODO: Récupérer la position de la cible depuis le service de combat
	// Pour l'instant, on suppose que la validation est OK
	return nil
}

// validateItemCooldown vérifie le cooldown d'un objet
func (s *AntiCheatService) validateItemCooldown(characterID uuid.UUID, itemType string) error {
	// TODO: Intégration avec le système de cooldowns d'objets
	// Pour l'instant, on suppose que c'est OK
	return nil
}

// detectBotPattern détecte les patterns de bot
func (s *AntiCheatService) detectBotPattern(history []*models.ActionEvent) bool {
	if len(history) < 20 {
		return false
	}
	
	// Vérifier les intervalles trop réguliers entre les actions
	intervals := make([]time.Duration, 0)
	for i := 1; i < len(history); i++ {
		interval := history[i].Timestamp.Sub(history[i-1].Timestamp)
		intervals = append(intervals, interval)
	}
	
	// Calculer la variance des intervalles
	variance := s.calculateTimeVariance(intervals)
	
	// Si la variance est trop faible, c'est suspect (actions trop régulières)
	return variance < 50*time.Millisecond
}

// detectImpossibleSequence détecte les séquences d'actions impossibles
func (s *AntiCheatService) detectImpossibleSequence(history []*models.ActionEvent) bool {
	if len(history) < 2 {
		return false
	}
	
	// Vérifier les actions simultanées impossibles
	for i := 1; i < len(history); i++ {
		prev := history[i-1]
		curr := history[i]
		
		timeDiff := curr.Timestamp.Sub(prev.Timestamp)
		
		// Deux actions de combat en moins de 50ms = suspect
		if timeDiff < 50*time.Millisecond {
			if (prev.ActionType == "attack" || prev.ActionType == "spell") &&
			   (curr.ActionType == "attack" || curr.ActionType == "spell") {
				return true
			}
		}
	}
	
	return false
}

// getPlayerStats récupère les statistiques de comportement d'un joueur
func (s *AntiCheatService) getPlayerStats(characterID uuid.UUID) *models.PlayerBehaviorStats {
	if stats, exists := s.playerStats[characterID]; exists {
		return stats
	}
	
	// Créer de nouvelles stats
	stats := &models.PlayerBehaviorStats{
		CharacterID: characterID,
		StartTime:   time.Now(),
	}
	
	s.playerStats[characterID] = stats
	return stats
}

// updatePlayerStats met à jour les statistiques de comportement
func (s *AntiCheatService) updatePlayerStats(characterID uuid.UUID, event *models.ActionEvent) {
	stats := s.getPlayerStats(characterID)
	
	stats.TotalActions++
	
	// Calculer les actions par minute
	duration := time.Since(stats.StartTime)
	if duration > 0 {
		stats.ActionsPerMinute = float64(stats.TotalActions) / duration.Minutes()
	}
	
	// Mettre à jour selon le type d'action
	switch event.ActionType {
	case "attack":
		stats.AttackActions++
	case "spell":
		stats.SpellActions++
	case "move":
		stats.MoveActions++
	}
}

// getRecentViolations filtre les violations récentes
func (s *AntiCheatService) getRecentViolations(violations []*models.AntiCheatViolation, duration time.Duration) []*models.AntiCheatViolation {
	recent := make([]*models.AntiCheatViolation, 0)
	cutoff := time.Now().Add(-duration)
	
	for _, violation := range violations {
		if violation.Timestamp.After(cutoff) {
			recent = append(recent, violation)
		}
	}
	
	return recent
}

// calculateSuspicionScore calcule le score de suspicion total
func (s *AntiCheatService) calculateSuspicionScore(factors []models.SuspicionFactor) float64 {
	if len(factors) == 0 {
		return 0.0
	}
	
	totalWeight := 0.0
	weightedSum := 0.0
	
	for _, factor := range factors {
		totalWeight += factor.Weight
		weightedSum += factor.Weight
	}
	
	if totalWeight == 0 {
		return 0.0
	}
	
	return weightedSum / totalWeight
}

// calculateDamageVariance calcule la variance des dégâts
func (s *AntiCheatService) calculateDamageVariance(characterID uuid.UUID) float64 {
	history := s.actionHistory[characterID]
	if len(history) < 10 {
		return 1.0 // Variance par défaut si pas assez de données
	}
	
	damages := make([]float64, 0)
	
	// Extraire les dégâts des actions d'attaque
	for _, event := range history {
		if event.ActionType == "attack" {
			if dmg, exists := event.Data["damage"]; exists {
				if damage, ok := dmg.(int); ok {
					damages = append(damages, float64(damage))
				}
			}
		}
	}
	
	if len(damages) < 5 {
		return 1.0
	}
	
	// Calculer la moyenne
	mean := 0.0
	for _, dmg := range damages {
		mean += dmg
	}
	mean /= float64(len(damages))
	
	// Calculer la variance
	variance := 0.0
	for _, dmg := range damages {
		variance += math.Pow(dmg-mean, 2)
	}
	variance /= float64(len(damages))
	
	// Retourner la variance normalisée
	if mean > 0 {
		return variance / (mean * mean) // Coefficient de variation au carré
	}
	
	return 1.0
}

// calculateTimeVariance calcule la variance des intervalles de temps
func (s *AntiCheatService) calculateTimeVariance(intervals []time.Duration) time.Duration {
	if len(intervals) == 0 {
		return 0
	}
	
	// Calculer la moyenne
	sum := time.Duration(0)
	for _, interval := range intervals {
		sum += interval
	}
	mean := sum / time.Duration(len(intervals))
	
	// Calculer la variance
	variance := time.Duration(0)
	for _, interval := range intervals {
		diff := interval - mean
		variance += time.Duration(int64(diff) * int64(diff) / int64(len(intervals)))
	}
	
	return variance
}

// reportSuspiciousActivity signale une activité suspecte
func (s *AntiCheatService) reportSuspiciousActivity(characterID uuid.UUID, violationType, description, severity string) error {
	violation := &models.AntiCheatViolation{
		CharacterID:  characterID,
		Type:         violationType,
		Description:  description,
		Severity:     severity,
		Timestamp:    time.Now(),
	}
	
	return s.ReportViolation(violation)
}

// applyTemporaryBan applique un bannissement temporaire
func (s *AntiCheatService) applyTemporaryBan(characterID uuid.UUID, duration time.Duration, reason string) error {
	expiresAt := time.Now().Add(duration)
	
	ban := &models.AntiCheatBan{
		ID:          uuid.New(),
		CharacterID: characterID,
		Reason:      reason,
		Duration:    duration,
		ExpiresAt:   &expiresAt,
		CreatedAt:   time.Now(),
		IsActive:    true,
	}
	
	if err := s.violationRepo.CreateBan(ban); err != nil {
		return fmt.Errorf("failed to create ban: %w", err)
	}
	
	logrus.WithFields(logrus.Fields{
		"character_id": characterID,
		"duration":     duration,
		"reason":       reason,
		"expires_at":   expiresAt,
	}).Warn("Temporary ban applied")
	
	return nil
}