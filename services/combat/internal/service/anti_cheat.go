package service

import (
	"combat/internal/config"
	"combat/internal/models"
	"combat/internal/repository"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AntiCheatServiceInterface définit les méthodes du service anti-cheat
type AntiCheatServiceInterface interface {
	// Validation des actions
	ValidateAction(actor *models.CombatParticipant, req *models.ActionRequest) *models.ValidationResponse
	CheckActionFrequency(actorID uuid.UUID, timeWindow time.Duration) (bool, int, error)
	ValidateTimestamp(clientTime, serverTime time.Time) (bool, string)

	// Détection de patterns suspects
	DetectSuspiciousPatterns(actorID uuid.UUID) (*models.AntiCheatResult, error)
	CheckDamageIntegrity(action *models.CombatAction, actor, target *models.CombatParticipant) (bool, string)
	ValidateMovement(oldPos, newPos *models.Position, timeElapsed time.Duration) (bool, string)

	// Système de scoring
	CalculateSuspicionScore(actorID uuid.UUID) (float64, []string, error)
	RecordSuspiciousActivity(actorID uuid.UUID, activityType string, details map[string]interface{})

	// Actions correctives
	ApplyAntiCheatMeasures(actorID uuid.UUID, score float64, flags []string) string
	TemporaryBan(actorID uuid.UUID, duration time.Duration, reason string) error
}

// AntiCheatService implémente l'interface AntiCheatServiceInterface
type AntiCheatService struct {
	actionRepo     repository.ActionRepositoryInterface
	config         *config.Config
	suspiciousLogs map[uuid.UUID][]SuspiciousActivity
	playerStats    map[uuid.UUID]*PlayerStats
}

// SuspiciousActivity représente une action suspecte
type SuspiciousActivity struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Severity  int                    `json:"severity"`
	Details   map[string]interface{} `json:"details"`
	ActionID  *uuid.UUID             `json:"action_id,omitempty"`
}

// PlayerStats représente les statistiques d'un joueur pour l'anti-cheat
type PlayerStats struct {
	ActorID              uuid.UUID   `json:"actor_id"`
	LastActionTime       time.Time   `json:"last_action_time"`
	ActionsInLastMinute  []time.Time `json:"actions_in_last_minute"`
	AverageDamage        float64     `json:"average_damage"`
	MaxDamageRecorded    int         `json:"max_damage_recorded"`
	SuspicionScore       float64     `json:"suspicion_score"`
	LastPositionUpdate   time.Time   `json:"last_position_update"`
	ConsistentHighDamage int         `json:"consistent_high_damage"`
	ImpossibleActions    int         `json:"impossible_actions"`
}

// NewAntiCheatService crée un nouveau service anti-cheat
func NewAntiCheatService(
	actionRepo repository.ActionRepositoryInterface,
	config *config.Config,
) AntiCheatServiceInterface {
	return &AntiCheatService{
		actionRepo:     actionRepo,
		config:         config,
		suspiciousLogs: make(map[uuid.UUID][]SuspiciousActivity),
		playerStats:    make(map[uuid.UUID]*PlayerStats),
	}
}

// ValidateAction valide une action avec l'anti-cheat
func (s *AntiCheatService) ValidateAction(actor *models.CombatParticipant, req *models.ActionRequest) *models.ValidationResponse {
	response := &models.ValidationResponse{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
		AntiCheat: &models.AntiCheatResult{
			Suspicious: false,
			Score:      0,
			Flags:      []string{},
			Action:     "allow",
		},
	}

	// Vérifier la fréquence d'actions
	actions, err := s.actionRepo.GetRecentActionsByActor(actor.CharacterID, config.DefaultTimeWindow*time.Minute)
	if err == nil && len(actions) > 0 {
		if suspicious, count, err := s.CheckActionFrequency(actor.CharacterID, 1*time.Minute); err == nil && suspicious {
			response.AntiCheat.Suspicious = true
			response.AntiCheat.Score += 25
			response.AntiCheat.Flags = append(response.AntiCheat.Flags, "high_action_frequency")
			response.Warnings = append(response.Warnings, fmt.Sprintf("High action frequency: %d actions/minute", count))

			s.RecordSuspiciousActivity(actor.CharacterID, "high_frequency", map[string]interface{}{
				"actions_per_minute": count,
				"threshold":          s.config.AntiCheat.MaxActionsPerSecond * config.DefaultPercentageMultiplier,
			})
		}
	}

	// Valider le timestamp
	if !req.ClientTimestamp.IsZero() {
		if valid, reason := s.ValidateTimestamp(req.ClientTimestamp, time.Now()); !valid {
			response.AntiCheat.Score += 15
			response.AntiCheat.Flags = append(response.AntiCheat.Flags, "timestamp_anomaly")
			response.Warnings = append(response.Warnings, reason)

			s.RecordSuspiciousActivity(actor.CharacterID, "timestamp_anomaly", map[string]interface{}{
				"client_timestamp": req.ClientTimestamp,
				"server_timestamp": time.Now(),
				"reason":           reason,
			})
		}
	}

	// Vérifier les patterns de comportement
	if patterns, err := s.DetectSuspiciousPatterns(actor.CharacterID); err == nil {
		if patterns.Suspicious {
			response.AntiCheat.Score += patterns.Score
			response.AntiCheat.Flags = append(response.AntiCheat.Flags, patterns.Flags...)
		}
	}

	// Calculer le score de suspicion final
	finalScore, flags, _ := s.CalculateSuspicionScore(actor.CharacterID)
	response.AntiCheat.Score = math.Max(response.AntiCheat.Score, finalScore)
	response.AntiCheat.Flags = append(response.AntiCheat.Flags, flags...)

	// Déterminer l'action à prendre
	response.AntiCheat.Action = s.ApplyAntiCheatMeasures(actor.CharacterID, response.AntiCheat.Score, response.AntiCheat.Flags)

	// Si le score est trop élevé, bloquer l'action
	if response.AntiCheat.Score > config.DefaultMinScore3 {
		response.AntiCheat.Action = AntiCheatActionBlock
		response.Valid = false
	} else if response.AntiCheat.Score > config.DefaultMinScore2 {
		response.AntiCheat.Action = "warn"
	}

	if response.AntiCheat.Score > config.DefaultMinSuspicionScore3 {
		response.Valid = false
		response.Errors = append(response.Errors, "Action blocked by anti-cheat system")
		response.AntiCheat.Suspicious = true
	} else if response.AntiCheat.Score > config.DefaultMinSuspicionScore2 {
		response.AntiCheat.Suspicious = true
	}

	return response
}

// CheckActionFrequency vérifie la fréquence d'actions d'un joueur
func (s *AntiCheatService) CheckActionFrequency(actorID uuid.UUID, timeWindow time.Duration) (isValid bool, actionCount int, err error) {
	actions, err := s.actionRepo.GetRecentActionsByActor(actorID, timeWindow)
	if err != nil {
		return false, 0, err
	}

	actionCount = len(actions)
	maxAllowed := int(timeWindow.Seconds()) * s.config.AntiCheat.MaxActionsPerSecond

	return actionCount > maxAllowed, actionCount, nil
}

// ValidateTimestamp valide un timestamp client
func (s *AntiCheatService) ValidateTimestamp(clientTime, serverTime time.Time) (isValid bool, reason string) {
	diff := serverTime.Sub(clientTime).Abs()

	// Tolérance de 5 secondes
	if diff > 5*time.Second {
		return false, fmt.Sprintf("Timestamp too far from server time: %v difference", diff)
	}

	// Vérifier si le timestamp est dans le futur
	if clientTime.After(serverTime.Add(1 * time.Second)) {
		return false, "Timestamp is in the future"
	}

	// Vérifier si le timestamp est trop ancien
	if clientTime.Before(serverTime.Add(-30 * time.Second)) {
		return false, "Timestamp is too old"
	}

	return true, ""
}

// DetectSuspiciousPatterns détecte des patterns suspects dans le comportement
func (s *AntiCheatService) DetectSuspiciousPatterns(actorID uuid.UUID) (*models.AntiCheatResult, error) {
	result := &models.AntiCheatResult{
		Suspicious: false,
		Score:      0,
		Flags:      []string{},
		Action:     "allow",
	}

	// Récupérer les statistiques du joueur
	stats := s.getOrCreatePlayerStats(actorID)

	// Pattern 1: Actions trop régulières (bot-like behavior)
	actions, err := s.actionRepo.GetRecentActionsByActor(actorID, config.DefaultTimeWindow*time.Minute)
	if err != nil {
		return result, err
	}

	if len(actions) > config.DefaultMaxActions {
		result.Flags = append(result.Flags, "high_frequency")
		result.Score += config.DefaultMaxActionsPerSecond
	}

	// Pattern 2: Dégâts inconsistants avec les stats
	highDamageActions := 0
	for _, action := range actions {
		if action.DamageDealt > 0 {
			// Si les dégâts sont beaucoup plus élevés que prévu
			expectedMaxDamage := stats.MaxDamageRecorded
			if expectedMaxDamage == 0 {
				expectedMaxDamage = 100 // Valeur par défaut
			}

			if float64(action.DamageDealt) > float64(expectedMaxDamage)*s.config.AntiCheat.MaxDamageMultiplier {
				highDamageActions++
				result.Score += 10
			}
		}
	}

	if highDamageActions > config.DefaultHighDamageThreshold {
		result.Flags = append(result.Flags, "excessive_damage")
		s.RecordSuspiciousActivity(actorID, "excessive_damage", map[string]interface{}{
			"high_damage_actions":  highDamageActions,
			"max_damage_threshold": float64(stats.MaxDamageRecorded) * s.config.AntiCheat.MaxDamageMultiplier,
		})
	}

	// Pattern 3: Critique rate anormalement élevé
	criticalActions := 0
	totalDamageActions := 0
	for _, action := range actions {
		if action.DamageDealt > 0 {
			totalDamageActions++
			if action.IsCritical {
				criticalActions++
			}
		}
	}

	if totalDamageActions > config.DefaultTotalDamageThreshold {
		critRate := float64(criticalActions) / float64(totalDamageActions)
		if critRate > config.DefaultCritRateThreshold { // Plus de 50% de critiques est suspect
			result.Score += 15
			result.Flags = append(result.Flags, "high_critical_rate")

			s.RecordSuspiciousActivity(actorID, "high_critical_rate", map[string]interface{}{
				"critical_rate":        critRate,
				"critical_actions":     criticalActions,
				"total_damage_actions": totalDamageActions,
			})
		}
	}

	// Déterminer si c'est suspect
	if result.Score > config.DefaultScoreThreshold {
		result.Suspicious = true
	}

	return result, nil
}

// CheckDamageIntegrity vérifie l'intégrité des calculations de dégâts
func (s *AntiCheatService) CheckDamageIntegrity(
	action *models.CombatAction,
	actor, target *models.CombatParticipant,
) (isValid bool, reason string) {
	if action.DamageDealt <= 0 {
		return true, "" // Pas de dégâts à vérifier
	}

	// Calculer les dégâts attendus
	expectedDamage := action.CalculateDamage(actor, target, nil)

	// Tolérance de ±20% pour la variance
	tolerance := 0.20
	minExpected := float64(expectedDamage) * (1 - tolerance)
	maxExpected := float64(expectedDamage) * (1 + tolerance)

	actualDamage := float64(action.DamageDealt)

	if actualDamage < minExpected || actualDamage > maxExpected {
		return false, fmt.Sprintf("Damage integrity check failed: expected %d (±20%%), got %d", expectedDamage, action.DamageDealt)
	}

	return true, ""
}

// ValidateMovement valide un mouvement de joueur
func (s *AntiCheatService) ValidateMovement(oldPos, newPos *models.Position, timeElapsed time.Duration) (isValid bool, reason string) {
	if oldPos == nil || newPos == nil {
		return true, "" // Pas de mouvement à valider
	}

	// Calculer la distance
	dx := newPos.X - oldPos.X
	dy := newPos.Y - oldPos.Y
	dz := newPos.Z - oldPos.Z
	distance := math.Sqrt(dx*dx + dy*dy + dz*dz)

	// Calculer la vitesse (unités par seconde)
	speed := distance / timeElapsed.Seconds()

	// Vitesse maximale autorisée (configurable)
	maxSpeed := 20.0 // unités par seconde

	if speed > maxSpeed {
		return false, fmt.Sprintf("Movement too fast: %.2f units/s (max: %.2f)", speed, maxSpeed)
	}

	// Vérifier les téléportations impossibles
	if distance > 100 && timeElapsed < 1*time.Second {
		return false, fmt.Sprintf("Impossible teleportation: %.2f units in %.2f seconds", distance, timeElapsed.Seconds())
	}

	return true, ""
}

// CalculateSuspicionScore calcule le score de suspicion global d'un joueur
func (s *AntiCheatService) CalculateSuspicionScore(actorID uuid.UUID) (score float64, flags []string, err error) {
	stats := s.getOrCreatePlayerStats(actorID)
	score = 0.0

	// Facteur 1: Actions suspicious récentes
	suspiciousActivities := s.suspiciousLogs[actorID]
	recentActivities := 0
	for _, activity := range suspiciousActivities {
		if time.Since(activity.Timestamp) < config.DefaultTimeWindow*time.Minute {
			recentActivities++
			score += float64(activity.Severity)
		}
	}

	if recentActivities > config.DefaultRecentActivitiesThreshold {
		flags = append(flags, "multiple_recent_infractions")
		score += 15
	}

	// Facteur 2: Consistency des performances
	if stats.ConsistentHighDamage > config.DefaultConsistentDamageThreshold {
		flags = append(flags, "consistent_high_performance")
		score += 10
	}

	// Facteur 3: Actions impossibles
	if stats.ImpossibleActions > config.DefaultImpossibleActionsThreshold {
		flags = append(flags, "impossible_actions")
		score += 20
	}

	// Facteur 4: Timing patterns
	actions, err := s.actionRepo.GetRecentActionsByActor(actorID, config.DefaultTimeWindow*time.Minute)
	if err == nil && len(actions) > 0 {
		avgProcessingTime := s.calculateAverageProcessingTime(actions)
		if avgProcessingTime < config.DefaultMinProcessingTime { // Moins de 50ms est suspect pour un humain
			flags = append(flags, "superhuman_reflexes")
			score += 15
		}
	}

	// Limiter le score entre 0 et 100
	if score > config.DefaultMaxScore2 {
		score = config.DefaultMaxScore2
	}

	stats.SuspicionScore = score
	return score, flags, nil
}

// RecordSuspiciousActivity enregistre une action suspecte
func (s *AntiCheatService) RecordSuspiciousActivity(actorID uuid.UUID, activityType string, details map[string]interface{}) {
	activity := SuspiciousActivity{
		Type:      activityType,
		Timestamp: time.Now(),
		Severity:  s.getSeverityForActivity(activityType),
		Details:   details,
	}

	if s.suspiciousLogs[actorID] == nil {
		s.suspiciousLogs[actorID] = []SuspiciousActivity{}
	}

	s.suspiciousLogs[actorID] = append(s.suspiciousLogs[actorID], activity)

	// Garder seulement les 50 dernières actions
	if len(s.suspiciousLogs[actorID]) > config.DefaultSuspiciousLogsLimit {
		s.suspiciousLogs[actorID] = s.suspiciousLogs[actorID][1:]
	}

	logrus.WithFields(logrus.Fields{
		"actor_id":      actorID,
		"activity_type": activityType,
		"severity":      activity.Severity,
		"details":       details,
	}).Warn("Suspicious activity recorded")
}

// ApplyAntiCheatMeasures applique des mesures correctives
func (s *AntiCheatService) ApplyAntiCheatMeasures(actorID uuid.UUID, score float64, flags []string) string {
	if score < config.DefaultScoreThreshold {
		return "allow"
	}

	if score < config.DefaultScoreThreshold*1.67 { // 50
		// Avertissement léger
		s.RecordSuspiciousActivity(actorID, "warning_issued", map[string]interface{}{
			"score": score,
			"flags": flags,
		})
		return "warn"
	}

	if score < config.DefaultScoreThreshold*2.67 { // 80
		// Surveillance renforcée
		s.RecordSuspiciousActivity(actorID, "enhanced_monitoring", map[string]interface{}{
			"score": score,
			"flags": flags,
		})
		return "monitor"
	}

	// Score élevé - action drastique
	s.RecordSuspiciousActivity(actorID, "high_suspicion_score", map[string]interface{}{
		"score": score,
		"flags": flags,
	})

	// Appliquer un ban temporaire selon la gravité
	duration := s.calculateBanDuration(score, flags)
	if err := s.TemporaryBan(actorID, duration, "High suspicion score detected"); err != nil {
		logrus.WithError(err).Error("Failed to apply temporary ban")
	}

	return "block"
}

// TemporaryBan applique un ban temporaire
func (s *AntiCheatService) TemporaryBan(actorID uuid.UUID, duration time.Duration, reason string) error {
	// TODO: Implémenter le système de ban
	// Pour l'instant, juste enregistrer l'action
	s.RecordSuspiciousActivity(actorID, "temporary_ban", map[string]interface{}{
		"duration": duration.String(),
		"reason":   reason,
	})

	logrus.WithFields(logrus.Fields{
		"actor_id": actorID,
		"duration": duration,
		"reason":   reason,
	}).Warn("Temporary ban applied")

	return nil
}

// Helper methods

func (s *AntiCheatService) getOrCreatePlayerStats(actorID uuid.UUID) *PlayerStats {
	if stats, exists := s.playerStats[actorID]; exists {
		return stats
	}

	stats := &PlayerStats{
		ActorID:             actorID,
		ActionsInLastMinute: []time.Time{},
		SuspicionScore:      0,
	}

	s.playerStats[actorID] = stats
	return stats
}

func (s *AntiCheatService) calculateAverageProcessingTime(actions []*models.CombatAction) float64 {
	if len(actions) == 0 {
		return 0
	}

	total := 0.0
	count := 0

	for _, action := range actions {
		if action.ProcessingTimeMs != nil && *action.ProcessingTimeMs > 0 {
			total += float64(*action.ProcessingTimeMs)
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return total / float64(count)
}

func (s *AntiCheatService) getSeverityForActivity(activityType string) int {
	severityMap := map[string]int{
		"high_frequency":       config.DefaultMinScore,
		"timestamp_anomaly":    config.DefaultMinScore2,
		"regular_pattern":      config.DefaultMinScore3,
		"excessive_damage":     config.DefaultMaxScore,
		"high_critical_rate":   config.DefaultMinScore2,
		"impossible_movement":  config.DefaultMaxScore2,
		"damage_integrity":     config.DefaultMaxScore,
		"superhuman_reflexes":  config.DefaultMinScore3,
		"multiple_infractions": config.DefaultMaxSuspiciousLogs,
	}

	if severity, exists := severityMap[activityType]; exists {
		return severity
	}

	return config.DefaultMinScore // Sévérité par défaut
}

func (s *AntiCheatService) calculateBanDuration(score float64, flags []string) time.Duration {
	baseDuration := time.Duration(config.DefaultBaseDurationAntiCheat) * time.Minute

	// Augmenter selon le score
	multiplier := score / config.DefaultScoreDivisor
	duration := time.Duration(float64(baseDuration) * multiplier)

	// Augmenter selon les flags critiques
	criticalFlags := []string{"impossible_actions", "excessive_damage", "superhuman_reflexes"}
	for _, flag := range flags {
		for _, critical := range criticalFlags {
			if flag == critical {
				duration *= 2
				break
			}
		}
	}

	// Limiter la durée maximale
	maxDuration := time.Duration(config.DefaultMaxDurationAntiCheat) * time.Minute
	if duration > maxDuration {
		duration = maxDuration
	}

	return duration
}

// UpdatePlayerStats met à jour les statistiques d'un joueur après une action
func (s *AntiCheatService) UpdatePlayerStats(actorID uuid.UUID, action *models.CombatAction) {
	stats := s.getOrCreatePlayerStats(actorID)

	// Mettre à jour le temps de dernière action
	stats.LastActionTime = action.ServerTimestamp

	// Ajouter l'action à la liste des actions récentes
	now := time.Now()
	stats.ActionsInLastMinute = append(stats.ActionsInLastMinute, now)

	// Nettoyer les actions trop anciennes
	cutoff := now.Add(-1 * time.Minute)
	newActions := []time.Time{}
	for _, actionTime := range stats.ActionsInLastMinute {
		if actionTime.After(cutoff) {
			newActions = append(newActions, actionTime)
		}
	}
	stats.ActionsInLastMinute = newActions

	// Mettre à jour les statistiques de dégâts
	if action.DamageDealt > 0 {
		if action.DamageDealt > stats.MaxDamageRecorded {
			stats.MaxDamageRecorded = action.DamageDealt
		}

		// Calculer la moyenne mobile des dégâts
		if stats.AverageDamage == 0 {
			stats.AverageDamage = float64(action.DamageDealt)
		} else {
			damageMultiplier := config.DefaultAverageDamageMultiplier
			damageMultiplier2 := config.DefaultAverageDamageMultiplier2
			stats.AverageDamage = (stats.AverageDamage * damageMultiplier) + (float64(action.DamageDealt) * damageMultiplier2)
		}

		// Vérifier si les dégâts sont anormalement élevés
		if float64(action.DamageDealt) > stats.AverageDamage*2.0 {
			stats.ConsistentHighDamage++
		} else {
			stats.ConsistentHighDamage = 0 // Reset si les dégâts redeviennent normaux
		}
	}
}

// GetPlayerSuspicionReport génère un rapport de suspicion pour un joueur
func (s *AntiCheatService) GetPlayerSuspicionReport(actorID uuid.UUID) *SuspicionReport {
	stats := s.getOrCreatePlayerStats(actorID)
	activities := s.suspiciousLogs[actorID]

	report := &SuspicionReport{
		ActorID:        actorID,
		SuspicionScore: stats.SuspicionScore,
		LastActivity:   stats.LastActionTime,
		RecentFlags:    []string{},
		Summary:        "",
	}

	// Analyser les actions récentes
	recentActivities := 0
	flagCounts := make(map[string]int)

	for _, activity := range activities {
		if time.Since(activity.Timestamp) < 24*time.Hour {
			recentActivities++
			flagCounts[activity.Type]++
		}
	}

	// Générer les flags récents
	for flag, count := range flagCounts {
		if count > 1 {
			report.RecentFlags = append(report.RecentFlags, fmt.Sprintf("%s (x%d)", flag, count))
		} else {
			report.RecentFlags = append(report.RecentFlags, flag)
		}
	}

	// Générer le résumé
	if stats.SuspicionScore < config.DefaultMinSuspicionScore {
		report.Summary = "Player behavior appears normal"
	} else if stats.SuspicionScore < config.DefaultMinSuspicionScore2 {
		report.Summary = "Some suspicious patterns detected, monitoring recommended"
	} else if stats.SuspicionScore < config.DefaultMinSuspicionScore3 {
		report.Summary = "Multiple suspicious activities, enhanced monitoring active"
	} else {
		report.Summary = "High suspicion score, immediate review recommended"
	}

	return report
}

// SuspicionReport représente un rapport de suspicion
type SuspicionReport struct {
	ActorID        uuid.UUID `json:"actor_id"`
	SuspicionScore float64   `json:"suspicion_score"`
	LastActivity   time.Time `json:"last_activity"`
	RecentFlags    []string  `json:"recent_flags"`
	Summary        string    `json:"summary"`
}

// CleanupOldData nettoie les anciennes données d'anti-cheat
func (s *AntiCheatService) CleanupOldData() {
	cutoff := time.Now().Add(-24 * time.Hour)

	// Nettoyer les logs suspects anciens
	for actorID, activities := range s.suspiciousLogs {
		newActivities := []SuspiciousActivity{}
		for _, activity := range activities {
			if activity.Timestamp.After(cutoff) {
				newActivities = append(newActivities, activity)
			}
		}

		if len(newActivities) == 0 {
			delete(s.suspiciousLogs, actorID)
		} else {
			s.suspiciousLogs[actorID] = newActivities
		}
	}

	// Nettoyer les stats des joueurs inactifs
	inactiveCutoff := time.Now().Add(-7 * 24 * time.Hour)
	for actorID, stats := range s.playerStats {
		if stats.LastActionTime.Before(inactiveCutoff) {
			delete(s.playerStats, actorID)
		}
	}

	logrus.Debug("Anti-cheat data cleanup completed")
}

// StartCleanupRoutine démarre la routine de nettoyage
func (s *AntiCheatService) StartCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			s.CleanupOldData()
		}
	}()
}
