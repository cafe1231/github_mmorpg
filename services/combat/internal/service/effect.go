// internal/service/effect.go - Service de gestion des effets
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
	ProcessPeriodicEffects() error
	CleanupExpiredEffects() error
	GetActiveEffects(characterID uuid.UUID) ([]*models.StatusEffect, error)
	DispelEffects(characterID uuid.UUID, dispelType string, maxCount int) ([]*models.StatusEffect, error)
}

// EffectService implémente l'interface EffectServiceInterface
type EffectService struct {
	config     *config.Config
	effectRepo repository.EffectRepositoryInterface
	combatRepo repository.CombatRepositoryInterface
	logRepo    repository.CombatLogRepositoryInterface
}

// NewEffectService crée une nouvelle instance du service d'effets
func NewEffectService(
	cfg *config.Config,
	effectRepo repository.EffectRepositoryInterface,
	combatRepo repository.CombatRepositoryInterface,
	logRepo repository.CombatLogRepositoryInterface,
) EffectServiceInterface {
	return &EffectService{
		config:     cfg,
		effectRepo: effectRepo,
		combatRepo: combatRepo,
		logRepo:    logRepo,
	}
}

// ApplyEffect applique un effet de statut
func (s *EffectService) ApplyEffect(effect *models.StatusEffect) error {
	return s.effectRepo.ApplyEffect(effect)
}

// RemoveEffect supprime un effet
func (s *EffectService) RemoveEffect(effectID uuid.UUID) error {
	// Récupérer l'effet avant de le supprimer pour les logs
	effect, err := s.effectRepo.GetEffectByID(effectID)
	if err != nil {
		return err
	}

	// Supprimer l'effet
	if err := s.effectRepo.RemoveEffect(effectID); err != nil {
		return err
	}

	// Log de suppression
	if effect.SessionID != nil {
		logEntry := &models.CombatLog{
			ID:        uuid.New(),
			SessionID: *effect.SessionID,
			TargetID:  &effect.CharacterID,
			EventType: "effect_removed",
			Message:   fmt.Sprintf("removed %s", effect.Name),
			Color:     "#FFFF00",
			Timestamp: time.Now(),
		}
		s.logRepo.CreateLog(logEntry)
	}

	return nil
}

// ProcessPeriodicEffects traite tous les effets périodiques
func (s *EffectService) ProcessPeriodicEffects() error {
	count, err := s.effectRepo.ProcessPeriodicEffects()
	if err != nil {
		return err
	}

	if count > 0 {
		logrus.WithField("count", count).Debug("Processed periodic effects")
	}

	return nil
}

// CleanupExpiredEffects nettoie les effets expirés
func (s *EffectService) CleanupExpiredEffects() error {
	count, err := s.effectRepo.CleanupExpiredEffects()
	if err != nil {
		return err
	}

	if count > 0 {
		logrus.WithField("count", count).Info("Cleaned up expired effects")
	}

	return nil
}

// GetActiveEffects récupère les effets actifs d'un personnage
func (s *EffectService) GetActiveEffects(characterID uuid.UUID) ([]*models.StatusEffect, error) {
	return s.effectRepo.GetActiveEffects(characterID)
}

// DispelEffects dissipe des effets selon leur type
func (s *EffectService) DispelEffects(characterID uuid.UUID, dispelType string, maxCount int) ([]*models.StatusEffect, error) {
	removedEffects, err := s.effectRepo.DispelEffects(characterID, dispelType, maxCount)
	if err != nil {
		return nil, err
	}

	// Log des effets dispelled
	for _, effect := range removedEffects {
		if effect.SessionID != nil {
			logEntry := &models.CombatLog{
				ID:        uuid.New(),
				SessionID: *effect.SessionID,
				TargetID:  &effect.CharacterID,
				EventType: "effect_removed",
				Message:   fmt.Sprintf("dispelled %s", effect.Name),
				Color:     "#FFFFFF",
				Timestamp: time.Now(),
			}
			s.logRepo.CreateLog(logEntry)
		}
	}

	return removedEffects, nil
}