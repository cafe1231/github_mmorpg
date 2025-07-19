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
	// Gestion des effets
	ApplyEffect(req *models.ApplyEffectRequest) (*models.EffectResult, error)
	RemoveEffect(req *models.RemoveEffectRequest) error
	GetActiveEffects(targetID uuid.UUID) ([]*models.CombatEffect, error)

	// Traitement des effets
	ProcessEffects(participant *models.CombatParticipant) error
	ProcessEffectTurn(effect *models.CombatEffect) (*models.EffectProcessResult, error)

	// Effets par combat
	GetCombatEffects(combatID uuid.UUID) ([]*models.CombatEffect, error)
	CleanupExpiredEffects(combatID uuid.UUID) error

	// Utilitaires
	StackEffect(existingEffect *models.CombatEffect, newEffect *models.CombatEffect) (*models.CombatEffect, error)
	CalculateEffectModifiers(participant *models.CombatParticipant) (map[string]float64, error)
	DispelEffects(targetID uuid.UUID, effectType *models.EffectType, onlyDebuffs bool) error
}

// EffectService implémente l'interface EffectServiceInterface
type EffectService struct {
	effectRepo repository.EffectRepositoryInterface
	combatRepo repository.CombatRepositoryInterface
	config     *config.Config
}

// NewEffectService crée un nouveau service d'effets
func NewEffectService(
	effectRepo repository.EffectRepositoryInterface,
	combatRepo repository.CombatRepositoryInterface,
	config *config.Config,
) EffectServiceInterface {
	return &EffectService{
		effectRepo: effectRepo,
		combatRepo: combatRepo,
		config:     config,
	}
}

// ApplyEffect applique un effet sur une cible
func (s *EffectService) ApplyEffect(req *models.ApplyEffectRequest) (*models.EffectResult, error) {
	// Validation de la demande
	if err := req.Validate(); err != nil {
		return &models.EffectResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Récupérer le modèle d'effet
	effectTemplates := models.GetEffectTemplates()
	template, exists := effectTemplates[req.EffectID]
	if !exists {
		return &models.EffectResult{
			Success: false,
			Error:   fmt.Sprintf("Unknown effect: %s", req.EffectID),
		}, nil
	}

	// Créer l'application d'effet
	duration := 0
	if req.Duration != nil {
		duration = *req.Duration
	}
	stacks := 1
	if req.Stacks != nil {
		stacks = *req.Stacks
	}
	application := &models.EffectApplication{
		EffectTemplate: template,
		TargetID:       req.TargetID,
		CasterID:       req.CasterID,
		Duration:       duration,
		Stacks:         stacks,
		CustomModifier: req.CustomValue,
		Metadata:       req.Metadata,
	}

	// Vérifier si un effet similaire existe déjà
	existingEffects, err := s.effectRepo.GetActiveByTarget(req.TargetID)
	if err != nil {
		return &models.EffectResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to check existing effects: %v", err),
		}, nil
	}

	var existingEffect *models.CombatEffect
	for _, effect := range existingEffects {
		if effect.EffectName == template.Name && effect.IsActive {
			existingEffect = effect
			break
		}
	}

	// Si l'effet existe déjà
	if existingEffect != nil {
		if existingEffect.CanStack() {
			// Empiler l'effet
			stackedEffect, err := s.StackEffect(existingEffect, models.CreateEffectFromTemplate(template, application))
			if err != nil {
				return &models.EffectResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to stack effect: %v", err),
				}, nil
			}

			if err := s.effectRepo.Update(stackedEffect); err != nil {
				return &models.EffectResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to update stacked effect: %v", err),
				}, nil
			}

			return &models.EffectResult{
				Success:        true,
				Effect:         stackedEffect,
				ExistingEffect: existingEffect,
				Action:         "stacked",
				Message:        fmt.Sprintf("Effect %s stacked (x%d)", template.Name, stackedEffect.CurrentStacks),
			}, nil
		} else {
			// Rafraîchir l'effet existing
			existingEffect.RemainingTurns = template.BaseDuration
			if application.Duration > 0 {
				existingEffect.RemainingTurns = application.Duration
			}
			existingEffect.UpdatedAt = time.Now()

			if err := s.effectRepo.Update(existingEffect); err != nil {
				return &models.EffectResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to refresh effect: %v", err),
				}, nil
			}

			return &models.EffectResult{
				Success:        true,
				Effect:         existingEffect,
				ExistingEffect: existingEffect,
				Action:         "refreshed",
				Message:        fmt.Sprintf("Effect %s refreshed", template.Name),
			}, nil
		}
	}

	// Créer un nouvel effet
	newEffect := models.CreateEffectFromTemplate(template, application)

	// TODO: Récupérer le combat ID depuis le participant
	// Pour l'instant, on va devoir le passer différemment ou le récupérer
	// newEffect.CombatID = ...

	if err := s.effectRepo.Create(newEffect); err != nil {
		return &models.EffectResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to create effect: %v", err),
		}, nil
	}

	logrus.WithFields(logrus.Fields{
		"effect_id":   newEffect.ID,
		"effect_name": newEffect.EffectName,
		"target_id":   newEffect.TargetID,
		"caster_id":   newEffect.CasterID,
		"duration":    newEffect.DurationTurns,
	}).Info("Effect applied")

	return &models.EffectResult{
		Success: true,
		Effect:  newEffect,
		Action:  "applied",
		Message: fmt.Sprintf("Effect %s applied", template.Name),
	}, nil
}

// RemoveEffect supprime un effet
func (s *EffectService) RemoveEffect(req *models.RemoveEffectRequest) error {
	if req.EffectID != nil {
		// Supprimer un effet spécifique
		return s.effectRepo.Delete(*req.EffectID)
	}

	// Récupérer les effets de la cible
	effects, err := s.effectRepo.GetActiveByTarget(req.TargetID)
	if err != nil {
		return fmt.Errorf("failed to get target effects: %w", err)
	}

	for _, effect := range effects {
		shouldRemove := false

		if req.RemoveAll {
			shouldRemove = true
		} else if req.EffectType != nil && effect.EffectType == *req.EffectType {
			shouldRemove = true
		} else if req.OnlyDebuffs && effect.IsHarmful() {
			shouldRemove = true
		} else if req.OnlyBuffs && effect.IsBeneficial() {
			shouldRemove = true
		}

		if shouldRemove && effect.IsDispellable {
			if err := s.effectRepo.Delete(effect.ID); err != nil {
				logrus.WithError(err).WithField("effect_id", effect.ID).Error("Failed to remove effect")
			}
		}
	}

	return nil
}

// GetActiveEffects récupère les effets actifs d'une cible
func (s *EffectService) GetActiveEffects(targetID uuid.UUID) ([]*models.CombatEffect, error) {
	return s.effectRepo.GetActiveByTarget(targetID)
}

// ProcessEffects traite tous les effets d'un participant pour un tour
func (s *EffectService) ProcessEffects(participant *models.CombatParticipant) error {
	effects, err := s.effectRepo.GetActiveByTarget(participant.CharacterID)
	if err != nil {
		return fmt.Errorf("failed to get participant effects: %w", err)
	}

	var totalDamage, totalHealing int
	var expiredEffects []uuid.UUID

	for _, effect := range effects {
		if !effect.IsActive {
			continue
		}

		result := effect.ProcessTurn()

		// Appliquer les résultats
		if result.DamageDealt > 0 {
			totalDamage += result.DamageDealt
			participant.Health -= result.DamageDealt
			if participant.Health < 0 {
				participant.Health = 0
				participant.IsAlive = false
			}
		}

		if result.HealingDone > 0 {
			totalHealing += result.HealingDone
			participant.Health += result.HealingDone
			if participant.Health > participant.MaxHealth {
				participant.Health = participant.MaxHealth
			}
		}

		// Marquer les effets expirés pour suppression
		if result.Expired {
			expiredEffects = append(expiredEffects, effect.ID)
		} else {
			// Mettre à jour l'effet
			if err := s.effectRepo.Update(effect); err != nil {
				logrus.WithError(err).WithField("effect_id", effect.ID).Error("Failed to update effect")
			}
		}

		// Log des effets
		if result.Message != "" {
			logrus.WithFields(logrus.Fields{
				"participant_id": participant.CharacterID,
				"effect_id":      effect.ID,
				"effect_name":    effect.EffectName,
				"damage":         result.DamageDealt,
				"healing":        result.HealingDone,
			}).Debug(result.Message)
		}
	}

	// Supprimer les effets expirés
	for _, effectID := range expiredEffects {
		if err := s.effectRepo.Delete(effectID); err != nil {
			logrus.WithError(err).WithField("effect_id", effectID).Error("Failed to delete expired effect")
		}
	}

	// Mettre à jour le participant si nécessaire
	if totalDamage > 0 || totalHealing > 0 {
		if err := s.combatRepo.UpdateParticipant(participant); err != nil {
			return fmt.Errorf("failed to update participant: %w", err)
		}

		logrus.WithFields(logrus.Fields{
			"participant_id":  participant.CharacterID,
			"total_damage":    totalDamage,
			"total_healing":   totalHealing,
			"expired_effects": len(expiredEffects),
		}).Debug("Processed participant effects")
	}

	return nil
}

// ProcessEffectTurn traite un effet pour un tour
func (s *EffectService) ProcessEffectTurn(effect *models.CombatEffect) (*models.EffectProcessResult, error) {
	if !effect.IsActive {
		return &models.EffectProcessResult{
			EffectID: effect.ID,
			Expired:  true,
		}, nil
	}

	return effect.ProcessTurn(), nil
}

// GetCombatEffects récupère tous les effets d'un combat
func (s *EffectService) GetCombatEffects(combatID uuid.UUID) ([]*models.CombatEffect, error) {
	return s.effectRepo.GetActiveByCombat(combatID)
}

// CleanupExpiredEffects nettoie les effets expirés d'un combat
func (s *EffectService) CleanupExpiredEffects(combatID uuid.UUID) error {
	effects, err := s.effectRepo.GetActiveByCombat(combatID)
	if err != nil {
		return fmt.Errorf("failed to get combat effects: %w", err)
	}

	var expiredCount int
	for _, effect := range effects {
		if effect.IsExpired() {
			if err := s.effectRepo.Delete(effect.ID); err != nil {
				logrus.WithError(err).WithField("effect_id", effect.ID).Error("Failed to delete expired effect")
			} else {
				expiredCount++
			}
		}
	}

	if expiredCount > 0 {
		logrus.WithFields(logrus.Fields{
			"combat_id":       combatID,
			"expired_effects": expiredCount,
		}).Debug("Cleaned up expired effects")
	}

	return nil
}

// StackEffect empile un effet existing avec un nouveau
func (s *EffectService) StackEffect(existingEffect *models.CombatEffect, newEffect *models.CombatEffect) (*models.CombatEffect, error) {
	if !existingEffect.CanStack() {
		return nil, fmt.Errorf("effect cannot be stacked")
	}

	// Augmenter le nombre de stacks
	existingEffect.CurrentStacks++
	if existingEffect.CurrentStacks > existingEffect.MaxStacks {
		existingEffect.CurrentStacks = existingEffect.MaxStacks
	}

	// Rafraîchir la durée
	existingEffect.RemainingTurns = newEffect.DurationTurns
	existingEffect.UpdatedAt = time.Now()

	// Calculer les nouvelles valeurs avec les stacks
	if existingEffect.StatAffected != nil {
		stat, value, modType := existingEffect.GetStatModifier()
		logrus.WithFields(logrus.Fields{
			"effect_id": existingEffect.ID,
			"stacks":    existingEffect.CurrentStacks,
			"stat":      stat,
			"value":     value,
			"mod_type":  modType,
		}).Debug("Effect stacked")
	}

	return existingEffect, nil
}

// CalculateEffectModifiers calcule tous les modificateurs d'effets d'un participant
func (s *EffectService) CalculateEffectModifiers(participant *models.CombatParticipant) (map[string]float64, error) {
	modifiers := make(map[string]float64)

	effects, err := s.effectRepo.GetActiveByTarget(participant.CharacterID)
	if err != nil {
		return modifiers, fmt.Errorf("failed to get participant effects: %w", err)
	}

	for _, effect := range effects {
		if !effect.IsActive || effect.StatAffected == nil {
			continue
		}

		stat, value, modType := effect.GetStatModifier()

		switch modType {
		case models.ModifierTypeFlat:
			// Modificateur fixe
			modifiers[stat] += float64(value)
		case models.ModifierTypePercentage:
			// Modificateur en pourcentage
			key := stat + "_percent"
			modifiers[key] += float64(value)
		}
	}

	return modifiers, nil
}

// DispelEffects dissipe des effets selon les critères
func (s *EffectService) DispelEffects(targetID uuid.UUID, effectType *models.EffectType, onlyDebuffs bool) error {
	effects, err := s.effectRepo.GetActiveByTarget(targetID)
	if err != nil {
		return fmt.Errorf("failed to get target effects: %w", err)
	}

	var dispelledCount int
	for _, effect := range effects {
		if !effect.IsDispellable {
			continue
		}

		shouldDispel := false

		if effectType != nil && effect.EffectType == *effectType {
			shouldDispel = true
		} else if onlyDebuffs && effect.IsHarmful() {
			shouldDispel = true
		} else if !onlyDebuffs && effect.IsBeneficial() {
			shouldDispel = true
		}

		if shouldDispel {
			if err := s.effectRepo.Delete(effect.ID); err != nil {
				logrus.WithError(err).WithField("effect_id", effect.ID).Error("Failed to dispel effect")
			} else {
				dispelledCount++
				logrus.WithFields(logrus.Fields{
					"effect_id":   effect.ID,
					"effect_name": effect.EffectName,
					"target_id":   targetID,
				}).Debug("Effect dispelled")
			}
		}
	}

	if dispelledCount > 0 {
		logrus.WithFields(logrus.Fields{
			"target_id":         targetID,
			"dispelled_effects": dispelledCount,
		}).Info("Effects dispelled")
	}

	return nil
}

// ApplyTemporaryEffect applique un effet temporaire (pour une action)
func (s *EffectService) ApplyTemporaryEffect(
	targetID, casterID uuid.UUID,
	effectType models.EffectType,
	duration int,
	value int,
) (*models.CombatEffect, error) {
	template := &models.EffectTemplate{
		ID:            fmt.Sprintf("temp_%s", effectType),
		Name:          string(effectType),
		Description:   fmt.Sprintf("Temporary %s effect", effectType),
		EffectType:    effectType,
		ModifierValue: value,
		ModifierType:  models.ModifierTypeFlat,
		BaseDuration:  duration,
		MaxStacks:     1,
		IsDispellable: true,
		IsBeneficial:  effectType == models.EffectTypeBuff || effectType == models.EffectTypeHot,
	}

	application := &models.EffectApplication{
		EffectTemplate: template,
		TargetID:       targetID,
		CasterID:       &casterID,
		Duration:       duration,
	}

	effect := models.CreateEffectFromTemplate(template, application)

	if err := s.effectRepo.Create(effect); err != nil {
		return nil, fmt.Errorf("failed to create temporary effect: %w", err)
	}

	return effect, nil
}

// GetEffectImpact calcule l'impact total des effets sur les stats d'un participant
func (s *EffectService) GetEffectImpact(participant *models.CombatParticipant) (*EffectImpact, error) {
	effects, err := s.effectRepo.GetActiveByTarget(participant.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participant effects: %w", err)
	}

	impact := &EffectImpact{
		StatModifiers:    make(map[string]int),
		PercentModifiers: make(map[string]float64),
		SpecialEffects:   []string{},
	}

	for _, effect := range effects {
		if !effect.IsActive {
			continue
		}

		// Modificateurs de stats
		if effect.StatAffected != nil {
			stat, value, modType := effect.GetStatModifier()

			switch modType {
			case models.ModifierTypeFlat:
				impact.StatModifiers[stat] += value
			case models.ModifierTypePercentage:
				impact.PercentModifiers[stat] += float64(value) / float64(config.DefaultVarianceDivisor)
			}
		}

		// Effets spéciaux
		switch effect.EffectType {
		case models.EffectTypeStun:
			impact.SpecialEffects = append(impact.SpecialEffects, "stunned")
			impact.CanAct = false
		case models.EffectTypeSilence:
			impact.SpecialEffects = append(impact.SpecialEffects, "silenced")
			impact.CanUseSkills = false
		case models.EffectTypeShield:
			impact.ShieldAmount += effect.ModifierValue * effect.CurrentStacks
		}
	}

	return impact, nil
}

// EffectImpact représente l'impact total des effets sur un participant
type EffectImpact struct {
	StatModifiers    map[string]int     `json:"stat_modifiers"`
	PercentModifiers map[string]float64 `json:"percent_modifiers"`
	SpecialEffects   []string           `json:"special_effects"`
	CanAct           bool               `json:"can_act"`
	CanUseSkills     bool               `json:"can_use_skills"`
	ShieldAmount     int                `json:"shield_amount"`
}

// ApplyDamageReduction applique la réduction de dégâts des effets
func (s *EffectService) ApplyDamageReduction(participant *models.CombatParticipant, incomingDamage int) int {
	effects, err := s.effectRepo.GetActiveByTarget(participant.CharacterID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get effects for damage reduction")
		return incomingDamage
	}

	finalDamage := float64(incomingDamage)
	var shieldAbsorption int

effectLoop:
	for _, effect := range effects {
		if !effect.IsActive {
			continue
		}

		switch effect.EffectType {
		case models.EffectTypeShield:
			// Les boucliers absorbent les dégâts en premier
			absorption := effect.ModifierValue * effect.CurrentStacks
			if absorption >= int(finalDamage) {
				// Le bouclier absorbe tous les dégâts
				shieldAbsorption = int(finalDamage)
				finalDamage = 0

				// Réduire la valeur du bouclier
				effect.ModifierValue -= int(finalDamage) / effect.CurrentStacks
				if effect.ModifierValue <= 0 {
					// Bouclier détruit
					if err := s.effectRepo.Delete(effect.ID); err != nil {
						logrus.WithError(err).Error("Failed to delete destroyed shield effect")
					}
				} else {
					if err := s.effectRepo.Update(effect); err != nil {
						logrus.WithError(err).Error("Failed to update shield effect")
					}
				}
				break effectLoop
			} else {
				// Le bouclier absorbe partiellement
				finalDamage -= float64(absorption)
				shieldAbsorption = absorption

				// Bouclier détruit
				if err := s.effectRepo.Delete(effect.ID); err != nil {
					logrus.WithError(err).Error("Failed to delete shield effect")
				}
			}

		case models.EffectTypeBuff:
			// Vérifier si c'est un effet de réduction de dégâts
			if effect.StatAffected != nil && *effect.StatAffected == "damage_reduction" {
				reduction := float64(effect.ModifierValue*effect.CurrentStacks) / float64(config.DefaultVarianceDivisor)
				finalDamage *= (1.0 - reduction)
			}
		}
	}

	if shieldAbsorption > 0 {
		logrus.WithFields(logrus.Fields{
			"participant_id":  participant.CharacterID,
			"original_damage": incomingDamage,
			"shield_absorbed": shieldAbsorption,
			"final_damage":    int(finalDamage),
		}).Debug("Shield absorbed damage")
	}

	return int(finalDamage)
}

// CreatePredefinedEffect crée un effet prédéfini
func (s *EffectService) CreatePredefinedEffect(
	effectID string,
	targetID, casterID uuid.UUID,
	combatID uuid.UUID,
) (*models.CombatEffect, error) {
	templates := models.GetEffectTemplates()
	template, exists := templates[effectID]
	if !exists {
		return nil, fmt.Errorf("unknown predefined effect: %s", effectID)
	}

	application := &models.EffectApplication{
		EffectTemplate: template,
		TargetID:       targetID,
		CasterID:       &casterID,
	}

	effect := models.CreateEffectFromTemplate(template, application)
	effect.CombatID = combatID

	if err := s.effectRepo.Create(effect); err != nil {
		return nil, fmt.Errorf("failed to create predefined effect: %w", err)
	}

	return effect, nil
}

// GetEffectDuration calcule la durée restante d'un effet
func (s *EffectService) GetEffectDuration(effectID uuid.UUID) (time.Duration, error) {
	effect, err := s.effectRepo.GetByID(effectID)
	if err != nil {
		return 0, err
	}

	if effect.ExpiresAt != nil {
		return time.Until(*effect.ExpiresAt), nil
	}

	// Si pas d'expiration absolute, estimer basé sur les tours restants
	// Assume 30 secondes par tour (devrait être configuré)
	estimatedDuration := time.Duration(effect.RemainingTurns) * time.Duration(config.DefaultCombatTurnTimeout) * time.Second
	return estimatedDuration, nil
}

// BulkApplyEffects applique plusieurs effets en une fois
func (s *EffectService) BulkApplyEffects(requests []*models.ApplyEffectRequest) ([]*models.EffectResult, error) {
	results := make([]*models.EffectResult, len(requests))

	for i, req := range requests {
		result, err := s.ApplyEffect(req)
		if err != nil {
			result = &models.EffectResult{
				Success: false,
				Error:   err.Error(),
			}
		}
		results[i] = result
	}

	return results, nil
}

// GetEffectSummary génère un résumé des effets actifs
func (s *EffectService) GetEffectSummary(targetID uuid.UUID) (*EffectSummary, error) {
	effects, err := s.effectRepo.GetActiveByTarget(targetID)
	if err != nil {
		return nil, err
	}

	summary := &EffectSummary{
		TotalEffects: len(effects),
		Buffs:        0,
		Debuffs:      0,
		DoTs:         0,
		HoTs:         0,
		Effects:      make(map[string]int),
	}

	for _, effect := range effects {
		if !effect.IsActive {
			continue
		}

		switch effect.EffectType {
		case models.EffectTypeBuff:
			summary.Buffs++
		case models.EffectTypeDebuff:
			summary.Debuffs++
		case models.EffectTypeDot:
			summary.DoTs++
		case models.EffectTypeHot:
			summary.HoTs++
		}

		summary.Effects[effect.EffectName] = effect.CurrentStacks
	}

	return summary, nil
}

// EffectSummary représente un résumé des effets
type EffectSummary struct {
	TotalEffects int            `json:"total_effects"`
	Buffs        int            `json:"buffs"`
	Debuffs      int            `json:"debuffs"`
	DoTs         int            `json:"dots"`
	HoTs         int            `json:"hots"`
	Effects      map[string]int `json:"effects"`
}

// StartEffectCleanupRoutine démarre une routine de nettoyage des effets
func (s *EffectService) StartEffectCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			// Nettoyer les effets expirés de tous les combats actifs
			// TODO: Récupérer la liste des combats actifs et nettoyer leurs effets
			logrus.Debug("Running effect cleanup routine")
		}
	}()
}
