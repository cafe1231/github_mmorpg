package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"combat/internal/config"
	"combat/internal/models"
	"combat/internal/repository"
)

// EffectServiceInterface définit les méthodes du service d'effets
type EffectServiceInterface interface {
	ApplyEffect(effect *models.StatusEffect) error
	RemoveEffect(effectID uuid.UUID) error
	RemoveEffectsByName(characterID uuid.UUID, effectName string) error
	StackEffect(characterID uuid.UUID, effect *models.StatusEffect) error
	GetActiveEffects(characterID uuid.UUID) ([]*models.StatusEffect, error)
	DispelEffects(characterID uuid.UUID, dispelType string, maxCount int) ([]*models.StatusEffect, error)
	CleanupExpiredEffects() error
	CleanupSessionEffects(sessionID uuid.UUID) error
	CalculateEffectiveStats(characterID uuid.UUID, baseStats map[string]interface{}) (map[string]interface{}, error)
	ProcessPeriodicEffects(characterID uuid.UUID) error
	CheckEffectInteractions(characterID uuid.UUID, newEffect *models.StatusEffect) error
}

// EffectService implémente l'interface EffectServiceInterface
type EffectService struct {
	config     *config.Config
	effectRepo repository.EffectRepositoryInterface
	logRepo    repository.CombatLogRepositoryInterface
}

// NewEffectService crée une nouvelle instance du service d'effets
func NewEffectService(
	cfg *config.Config,
	effectRepo repository.EffectRepositoryInterface,
	logRepo repository.CombatLogRepositoryInterface,
) EffectServiceInterface {
	return &EffectService{
		config:     cfg,
		effectRepo: effectRepo,
		logRepo:    logRepo,
	}
}

// ApplyEffect applique un effet sur un personnage
func (s *EffectService) ApplyEffect(effect *models.StatusEffect) error {
	// Vérifier si l'effet peut être stacké
	if effect.MaxStacks > 1 {
		return s.StackEffect(effect.CharacterID, effect)
	}
	
	// Vérifier les interactions avec d'autres effets
	if err := s.CheckEffectInteractions(effect.CharacterID, effect); err != nil {
		return fmt.Errorf("effect interaction failed: %w", err)
	}
	
	// Chercher un effet existant du même nom
	existingEffects, err := s.effectRepo.GetByCharacterAndName(effect.CharacterID, effect.Name)
	if err != nil {
		return fmt.Errorf("failed to check existing effects: %w", err)
	}
	
	// Si un effet du même nom existe déjà
	if len(existingEffects) > 0 {
		existingEffect := existingEffects[0]
		
		// Stratégie de remplacement selon le type d'effet
		switch s.getRefreshStrategy(effect) {
		case "refresh":
			// Rafraîchir la durée
			existingEffect.ExpiresAt = time.Now().Add(effect.Duration)
			existingEffect.UpdatedAt = time.Now()
			
			if err := s.effectRepo.Update(existingEffect); err != nil {
				return fmt.Errorf("failed to refresh effect: %w", err)
			}
			
			s.logEffectRefreshed(existingEffect)
			return nil
			
		case "replace":
			// Remplacer l'ancien effet
			if err := s.effectRepo.Delete(existingEffect.ID); err != nil {
				logrus.WithError(err).Error("Failed to remove old effect")
			}
			
		case "stack":
			// Stacké (géré plus haut)
			return s.StackEffect(effect.CharacterID, effect)
			
		case "resist":
			// Résister au nouvel effet
			s.logEffectResisted(effect)
			return fmt.Errorf("effect resisted due to existing effect")
		}
	}
	
	// Appliquer le nouvel effet
	effect.IsActive = true
	effect.AppliedAt = time.Now()
	effect.UpdatedAt = time.Now()
	
	if err := s.effectRepo.Create(effect); err != nil {
		return fmt.Errorf("failed to create effect: %w", err)
	}
	
	// Logger l'application de l'effet
	s.logEffectApplied(effect)
	
	logrus.WithFields(logrus.Fields{
		"character_id": effect.CharacterID,
		"effect_name":  effect.Name,
		"effect_type":  effect.Type,
		"duration":     effect.Duration,
	}).Info("Effect applied successfully")
	
	return nil
}

// StackEffect gère les effets stackables
func (s *EffectService) StackEffect(characterID uuid.UUID, newEffect *models.StatusEffect) error {
	// Chercher les effets existants du même nom
	existingEffects, err := s.effectRepo.GetByCharacterAndName(characterID, newEffect.Name)
	if err != nil {
		return fmt.Errorf("failed to check existing effects: %w", err)
	}
	
	if len(existingEffects) == 0 {
		// Aucun effet existant, créer le premier stack
		newEffect.Stacks = 1
		newEffect.IsActive = true
		newEffect.AppliedAt = time.Now()
		newEffect.UpdatedAt = time.Now()
		
		if err := s.effectRepo.Create(newEffect); err != nil {
			return fmt.Errorf("failed to create stacked effect: %w", err)
		}
		
		s.logEffectApplied(newEffect)
		return nil
	}
	
	// Effet existant trouvé
	existingEffect := existingEffects[0]
	
	// Vérifier si on peut ajouter un stack
	if existingEffect.Stacks >= newEffect.MaxStacks {
		// Déjà au maximum, rafraîchir la durée
		existingEffect.ExpiresAt = time.Now().Add(newEffect.Duration)
		existingEffect.UpdatedAt = time.Now()
		
		if err := s.effectRepo.Update(existingEffect); err != nil {
			return fmt.Errorf("failed to refresh max stack effect: %w", err)
		}
		
		s.logEffectRefreshed(existingEffect)
		return nil
	}
	
	// Ajouter un stack
	existingEffect.Stacks++
	existingEffect.ExpiresAt = time.Now().Add(newEffect.Duration)
	existingEffect.UpdatedAt = time.Now()
	
	// Recalculer les modificateurs selon le nombre de stacks
	existingEffect.StatModifiers = s.calculateStackedModifiers(newEffect.StatModifiers, existingEffect.Stacks)
	
	if err := s.effectRepo.Update(existingEffect); err != nil {
		return fmt.Errorf("failed to update stacked effect: %w", err)
	}
	
	s.logEffectStacked(existingEffect)
	
	logrus.WithFields(logrus.Fields{
		"character_id": characterID,
		"effect_name":  newEffect.Name,
		"stacks":       existingEffect.Stacks,
		"max_stacks":   newEffect.MaxStacks,
	}).Info("Effect stacked successfully")
	
	return nil
}

// RemoveEffect supprime un effet spécifique
func (s *EffectService) RemoveEffect(effectID uuid.UUID) error {
	effect, err := s.effectRepo.GetByID(effectID)
	if err != nil {
		return fmt.Errorf("effect not found: %w", err)
	}
	
	if err := s.effectRepo.Delete(effectID); err != nil {
		return fmt.Errorf("failed to delete effect: %w", err)
	}
	
	s.logEffectRemoved(effect)
	
	logrus.WithFields(logrus.Fields{
		"effect_id":    effectID,
		"character_id": effect.CharacterID,
		"effect_name":  effect.Name,
	}).Info("Effect removed successfully")
	
	return nil
}

// RemoveEffectsByName supprime tous les effets d'un nom donné d'un personnage
func (s *EffectService) RemoveEffectsByName(characterID uuid.UUID, effectName string) error {
	effects, err := s.effectRepo.GetByCharacterAndName(characterID, effectName)
	if err != nil {
		return fmt.Errorf("failed to get effects by name: %w", err)
	}
	
	for _, effect := range effects {
		if err := s.effectRepo.Delete(effect.ID); err != nil {
			logrus.WithError(err).WithField("effect_id", effect.ID).Error("Failed to delete effect")
			continue
		}
		
		s.logEffectRemoved(effect)
	}
	
	return nil
}

// GetActiveEffects récupère tous les effets actifs d'un personnage
func (s *EffectService) GetActiveEffects(characterID uuid.UUID) ([]*models.StatusEffect, error) {
	effects, err := s.effectRepo.GetActiveEffects(characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active effects: %w", err)
	}
	
	// Filtrer les effets expirés
	activeEffects := make([]*models.StatusEffect, 0)
	now := time.Now()
	
	for _, effect := range effects {
		if effect.ExpiresAt.After(now) {
			activeEffects = append(activeEffects, effect)
		} else {
			// Supprimer les effets expirés
			if err := s.effectRepo.Delete(effect.ID); err != nil {
				logrus.WithError(err).Error("Failed to delete expired effect")
			}
			s.logEffectExpired(effect)
		}
	}
	
	return activeEffects, nil
}

// DispelEffects dissipe des effets selon leur type
func (s *EffectService) DispelEffects(characterID uuid.UUID, dispelType string, maxCount int) ([]*models.StatusEffect, error) {
	allEffects, err := s.effectRepo.GetActiveEffects(characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get effects for dispel: %w", err)
	}
	
	// Filtrer les effets selon le type de dispel
	dispelableEffects := make([]*models.StatusEffect, 0)
	
	for _, effect := range allEffects {
		if s.canDispel(effect, dispelType) {
			dispelableEffects = append(dispelableEffects, effect)
		}
	}
	
	// Limiter le nombre d'effets à dissiper
	if len(dispelableEffects) > maxCount {
		dispelableEffects = dispelableEffects[:maxCount]
	}
	
	// Supprimer les effets
	removedEffects := make([]*models.StatusEffect, 0)
	for _, effect := range dispelableEffects {
		if err := s.effectRepo.Delete(effect.ID); err != nil {
			logrus.WithError(err).Error("Failed to dispel effect")
			continue
		}
		
		removedEffects = append(removedEffects, effect)
		s.logEffectDispelled(effect)
	}
	
	return removedEffects, nil
}

// CleanupExpiredEffects nettoie les effets expirés
func (s *EffectService) CleanupExpiredEffects() error {
	count, err := s.effectRepo.CleanupExpiredEffects()
	if err != nil {
		return fmt.Errorf("failed to cleanup expired effects: %w", err)
	}
	
	if count > 0 {
		logrus.WithField("count", count).Info("Cleaned up expired effects")
	}
	
	return nil
}

// CleanupSessionEffects nettoie les effets d'une session
func (s *EffectService) CleanupSessionEffects(sessionID uuid.UUID) error {
	count, err := s.effectRepo.RemoveSessionEffects(sessionID)
	if err != nil {
		return fmt.Errorf("failed to cleanup session effects: %w", err)
	}
	
	logrus.WithFields(logrus.Fields{
		"session_id": sessionID,
		"count":      count,
	}).Info("Cleaned up session effects")
	
	return nil
}

// CalculateEffectiveStats calcule les stats effectives avec les effets appliqués
func (s *EffectService) CalculateEffectiveStats(characterID uuid.UUID, baseStats map[string]interface{}) (map[string]interface{}, error) {
	effects, err := s.GetActiveEffects(characterID)
	if err != nil {
		return baseStats, fmt.Errorf("failed to get active effects: %w", err)
	}
	
	// Copier les stats de base
	effectiveStats := make(map[string]interface{})
	for key, value := range baseStats {
		effectiveStats[key] = value
	}
	
	// Appliquer les modificateurs de chaque effet
	for _, effect := range effects {
		if effect.StatModifiers != nil {
			s.applyStatModifiers(effectiveStats, effect.StatModifiers)
		}
	}
	
	return effectiveStats, nil
}

// ProcessPeriodicEffects traite les effets périodiques pour un personnage
func (s *EffectService) ProcessPeriodicEffects(characterID uuid.UUID) error {
	effects, err := s.GetActiveEffects(characterID)
	if err != nil {
		return fmt.Errorf("failed to get active effects: %w", err)
	}
	
	now := time.Now()
	
	for _, effect := range effects {
		if s.isPeriodicEffect(effect) {
			// Vérifier si le tick doit être traité
			tickInterval := s.getTickInterval(effect)
			
			// Si pas de dernier tick ou si assez de temps s'est écoulé
			if effect.LastTick == nil || now.Sub(*effect.LastTick) >= tickInterval {
				if err := s.processPeriodicTick(effect); err != nil {
					logrus.WithError(err).WithField("effect_id", effect.ID).Error("Failed to process periodic tick")
					continue
				}
				
				// Mettre à jour le timestamp du dernier tick
				effect.LastTick = &now
				if err := s.effectRepo.Update(effect); err != nil {
					logrus.WithError(err).Error("Failed to update effect last tick")
				}
			}
		}
	}
	
	return nil
}

// CheckEffectInteractions vérifie les interactions entre effets
func (s *EffectService) CheckEffectInteractions(characterID uuid.UUID, newEffect *models.StatusEffect) error {
	existingEffects, err := s.GetActiveEffects(characterID)
	if err != nil {
		return fmt.Errorf("failed to get existing effects: %w", err)
	}
	
	for _, existing := range existingEffects {
		// Vérifier les incompatibilités
		if s.areEffectsIncompatible(existing, newEffect) {
			return fmt.Errorf("effect %s is incompatible with existing effect %s", newEffect.Name, existing.Name)
		}
		
		// Vérifier les suppressions mutuelles
		if s.shouldSuppressExisting(existing, newEffect) {
			if err := s.RemoveEffect(existing.ID); err != nil {
				logrus.WithError(err).Error("Failed to remove suppressed effect")
			}
		}
	}
	
	return nil
}

// Méthodes utilitaires

// getRefreshStrategy détermine la stratégie de rafraîchissement d'un effet
func (s *EffectService) getRefreshStrategy(effect *models.StatusEffect) string {
	// Logique basée sur le type d'effet et ses propriétés
	if effect.MaxStacks > 1 {
		return "stack"
	}
	
	switch effect.Type {
	case "buff":
		return "refresh" // Les buffs se rafraîchissent généralement
	case "debuff":
		return "refresh" // Les debuffs aussi
	case "poison", "burn", "bleed":
		return "replace" // Les DoT se remplacent
	case "immunity":
		return "resist" // Les immunités résistent aux nouveaux effets
	default:
		return "refresh"
	}
}

// calculateStackedModifiers calcule les modificateurs pour un effet stacké
func (s *EffectService) calculateStackedModifiers(baseModifiers map[string]interface{}, stacks int) map[string]interface{} {
	stackedModifiers := make(map[string]interface{})
	
	for key, value := range baseModifiers {
		switch v := value.(type) {
		case int:
			stackedModifiers[key] = v * stacks
		case float64:
			stackedModifiers[key] = v * float64(stacks)
		default:
			stackedModifiers[key] = value // Valeurs non-stackables
		}
	}
	
	return stackedModifiers
}

// canDispel vérifie si un effet peut être dissipé
func (s *EffectService) canDispel(effect *models.StatusEffect, dispelType string) bool {
	if !effect.IsDispellable {
		return false
	}
	
	switch dispelType {
	case "magic":
		return effect.Type == "buff" || effect.Type == "debuff"
	case "poison":
		return effect.Type == "poison"
	case "disease":
		return effect.Type == "disease"
	case "curse":
		return effect.Type == "curse"
	case "all":
		return true
	default:
		return false
	}
}

// applyStatModifiers applique les modificateurs de stats
func (s *EffectService) applyStatModifiers(stats map[string]interface{}, modifiers map[string]interface{}) {
	for key, modifier := range modifiers {
		switch key {
		case "health_bonus", "mana_bonus", "attack_power_bonus", "spell_power_bonus":
			// Bonus additifs
			s.addToStat(stats, key, modifier)
		case "damage_multiplier", "healing_multiplier", "speed_multiplier":
			// Multiplicateurs
			s.multiplyStat(stats, key, modifier)
		default:
			// Autres modificateurs
			s.addToStat(stats, key, modifier)
		}
	}
}

// addToStat ajoute une valeur à une stat
func (s *EffectService) addToStat(stats map[string]interface{}, key string, value interface{}) {
	baseKey := key
	if key == "health_bonus" {
		baseKey = "health"
	} else if key == "mana_bonus" {
		baseKey = "mana"
	}
	
	currentValue := s.getStatValue(stats, baseKey, 0)
	bonusValue := s.getIntValue(value, 0)
	
	stats[baseKey] = currentValue + bonusValue
}

// multiplyStat multiplie une stat
func (s *EffectService) multiplyStat(stats map[string]interface{}, key string, value interface{}) {
	baseKey := key
	if key == "damage_multiplier" {
		baseKey = "damage"
	}
	
	currentValue := s.getStatValue(stats, baseKey, 1)
	multiplier := s.getFloatValue(value, 1.0)
	
	stats[baseKey] = int(float64(currentValue) * multiplier)
}

// isPeriodicEffect vérifie si un effet est périodique
func (s *EffectService) isPeriodicEffect(effect *models.StatusEffect) bool {
	periodicTypes := []string{"poison", "burn", "bleed", "regeneration", "mana_regen"}
	
	for _, pType := range periodicTypes {
		if effect.Type == pType {
			return true
		}
	}
	
	// Vérifier si l'effet a des propriétés périodiques dans ses modificateurs
	if effect.StatModifiers != nil {
		if _, hasDoT := effect.StatModifiers["damage_per_tick"]; hasDoT {
			return true
		}
		if _, hasHoT := effect.StatModifiers["healing_per_tick"]; hasHoT {
			return true
		}
	}
	
	return false
}

// getTickInterval retourne l'intervalle entre les ticks d'un effet
func (s *EffectService) getTickInterval(effect *models.StatusEffect) time.Duration {
	// Intervalle par défaut de 3 secondes
	defaultInterval := 3 * time.Second
	
	if effect.StatModifiers != nil {
		if interval, exists := effect.StatModifiers["tick_interval"]; exists {
			if seconds, ok := interval.(int); ok {
				return time.Duration(seconds) * time.Second
			}
			if seconds, ok := interval.(float64); ok {
				return time.Duration(seconds) * time.Second
			}
		}
	}
	
	return defaultInterval
}

// processPeriodicTick traite un tick périodique
func (s *EffectService) processPeriodicTick(effect *models.StatusEffect) error {
	// TODO: Implémenter la logique de traitement des ticks
	// Par exemple, appliquer des dégâts ou des soins périodiques
	
	logrus.WithFields(logrus.Fields{
		"effect_id":    effect.ID,
		"character_id": effect.CharacterID,
		"effect_name":  effect.Name,
	}).Debug("Processing periodic tick")
	
	return nil
}

// areEffectsIncompatible vérifie si deux effets sont incompatibles
func (s *EffectService) areEffectsIncompatible(existing, newEffect *models.StatusEffect) bool {
	// Logique d'incompatibilité basique
	incompatiblePairs := map[string][]string{
		"haste":     {"slow"},
		"slow":      {"haste"},
		"strength":  {"weakness"},
		"weakness":  {"strength"},
	}
	
	if incompatible, exists := incompatiblePairs[existing.Name]; exists {
		for _, name := range incompatible {
			if newEffect.Name == name {
				return true
			}
		}
	}
	
	return false
}

// shouldSuppressExisting vérifie si un nouvel effet doit supprimer un existant
func (s *EffectService) shouldSuppressExisting(existing, newEffect *models.StatusEffect) bool {
	// Les nouveaux effets du même type mais plus puissants suppriment les anciens
	if existing.Name == newEffect.Name && existing.Priority < newEffect.Priority {
		return true
	}
	
	return false
}

// getStatValue récupère une valeur de stat avec une valeur par défaut
func (s *EffectService) getStatValue(stats map[string]interface{}, key string, defaultValue int) int {
	if value, exists := stats[key]; exists {
		if intVal, ok := value.(int); ok {
			return intVal
		}
		if floatVal, ok := value.(float64); ok {
			return int(floatVal)
		}
	}
	return defaultValue
}

func (s *EffectService) getIntValue(value interface{}, defaultValue int) int {
	if intVal, ok := value.(int); ok {
		return intVal
	}
	if floatVal, ok := value.(float64); ok {
		return int(floatVal)
	}
	return defaultValue
}

func (s *EffectService) getFloatValue(value interface{}, defaultValue float64) float64 {
	if floatVal, ok := value.(float64); ok {
		return floatVal
	}
	if intVal, ok := value.(int); ok {
		return float64(intVal)
	}
	return defaultValue
}

// Méthodes de logging

func (s *EffectService) logEffectApplied(effect *models.StatusEffect) {
	if effect.SessionID != nil {
		logEntry := &models.CombatLog{
			ID:        uuid.New(),
			SessionID: *effect.SessionID,
			TargetID:  &effect.CharacterID,
			EventType: "effect_applied",
			Message:   fmt.Sprintf("gained %s", effect.Name),
			Color:     s.getEffectColor(effect.Type),
			Timestamp: time.Now(),
		}
		s.logRepo.CreateLog(logEntry)
	}
}

func (s *EffectService) logEffectRemoved(effect *models.StatusEffect) {
	if effect.SessionID != nil {
		logEntry := &models.CombatLog{
			ID:        uuid.New(),
			SessionID: *effect.SessionID,
			TargetID:  &effect.CharacterID,
			EventType: "effect_removed",
			Message:   fmt.Sprintf("lost %s", effect.Name),
			Color:     "#CCCCCC",
			Timestamp: time.Now(),
		}
		s.logRepo.CreateLog(logEntry)
	}
}

func (s *EffectService) logEffectStacked(effect *models.StatusEffect) {
	if effect.SessionID != nil {
		logEntry := &models.CombatLog{
			ID:        uuid.New(),
			SessionID: *effect.SessionID,
			TargetID:  &effect.CharacterID,
			EventType: "effect_stacked",
			Message:   fmt.Sprintf("%s (x%d)", effect.Name, effect.Stacks),
			Color:     s.getEffectColor(effect.Type),
			Timestamp: time.Now(),
		}
		s.logRepo.CreateLog(logEntry)
	}
}

func (s *EffectService) logEffectRefreshed(effect *models.StatusEffect) {
	if effect.SessionID != nil {
		logEntry := &models.CombatLog{
			ID:        uuid.New(),
			SessionID: *effect.SessionID,
			TargetID:  &effect.CharacterID,
			EventType: "effect_refreshed",
			Message:   fmt.Sprintf("%s refreshed", effect.Name),
			Color:     s.getEffectColor(effect.Type),
			Timestamp: time.Now(),
		}
		s.logRepo.CreateLog(logEntry)
	}
}

func (s *EffectService) logEffectExpired(effect *models.StatusEffect) {
	if effect.SessionID != nil {
		logEntry := &models.CombatLog{
			ID:        uuid.New(),
			SessionID: *effect.SessionID,
			TargetID:  &effect.CharacterID,
			EventType: "effect_expired",
			Message:   fmt.Sprintf("%s expired", effect.Name),
			Color:     "#999999",
			Timestamp: time.Now(),
		}
		s.logRepo.CreateLog(logEntry)
	}
}

func (s *EffectService) logEffectDispelled(effect *models.StatusEffect) {
	if effect.SessionID != nil {
		logEntry := &models.CombatLog{
			ID:        uuid.New(),
			SessionID: *effect.SessionID,
			TargetID:  &effect.CharacterID,
			EventType: "effect_dispelled",
			Message:   fmt.Sprintf("%s dispelled", effect.Name),
			Color:     "#FF6B6B",
			Timestamp: time.Now(),
		}
		s.logRepo.CreateLog(logEntry)
	}
}

func (s *EffectService) logEffectResisted(effect *models.StatusEffect) {
	if effect.SessionID != nil {
		logEntry := &models.CombatLog{
			ID:        uuid.New(),
			SessionID: *effect.SessionID,
			TargetID:  &effect.CharacterID,
			EventType: "effect_resisted",
			Message:   fmt.Sprintf("%s resisted", effect.Name),
			Color:     "#FFA500",
			Timestamp: time.Now(),
		}
		s.logRepo.CreateLog(logEntry)
	}
}

func (s *EffectService) getEffectColor(effectType string) string {
	colors := map[string]string{
		"buff":         "#4CAF50", // Vert
		"debuff":       "#F44336", // Rouge
		"poison":       "#9C27B0", // Violet
		"burn":         "#FF5722", // Orange rouge
		"freeze":       "#2196F3", // Bleu
		"stun":         "#FF9800", // Orange
		"slow":         "#607D8B", // Gris bleu
		"haste":        "#FFEB3B", // Jaune
		"shield":       "#00BCD4", // Cyan
		"regeneration": "#8BC34A", // Vert clair
	}
	
	if color, exists := colors[effectType]; exists {
		return color
	}
	
	return "#FFFFFF" // Blanc par défaut
}