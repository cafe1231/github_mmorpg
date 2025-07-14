package service

import (
"github.com/google/uuid"
"combat/internal/config"
"combat/internal/models"
"combat/internal/repository"
)

// SpellServiceInterface définit les méthodes du service spell
type SpellServiceInterface interface {
GetCharacterSpells(characterID uuid.UUID) ([]*models.CharacterSpell, error)
LearnSpell(characterID uuid.UUID, spellID uuid.UUID) error
CastSpell(req models.CastSpellRequest) error
GetSpellCooldowns(characterID uuid.UUID) ([]*models.SpellCooldown, error)
GetAvailableSpells(characterLevel int, characterClass string) ([]*models.Spell, error)
}

// SpellService implémente l'interface SpellServiceInterface
type SpellService struct {
config    *config.Config
spellRepo repository.SpellRepositoryInterface
}

// NewSpellService crée une nouvelle instance du service spell
func NewSpellService(cfg *config.Config, spellRepo repository.SpellRepositoryInterface) SpellServiceInterface {
return &SpellService{
config:    cfg,
spellRepo: spellRepo,
}
}

// GetCharacterSpells récupère les sorts d'un personnage
func (s *SpellService) GetCharacterSpells(characterID uuid.UUID) ([]*models.CharacterSpell, error) {
return s.spellRepo.GetCharacterSpells(characterID)
}

// LearnSpell fait apprendre un sort à un personnage
func (s *SpellService) LearnSpell(characterID uuid.UUID, spellID uuid.UUID) error {
// TODO: Implement spell learning logic
return nil
}

// CastSpell lance un sort
func (s *SpellService) CastSpell(req models.CastSpellRequest) error {
// TODO: Implement spell casting logic
return nil
}

// GetSpellCooldowns récupère les cooldowns actifs
func (s *SpellService) GetSpellCooldowns(characterID uuid.UUID) ([]*models.SpellCooldown, error) {
return s.spellRepo.GetActiveCooldowns(characterID)
}

// GetAvailableSpells récupère les sorts disponibles
func (s *SpellService) GetAvailableSpells(characterLevel int, characterClass string) ([]*models.Spell, error) {
return s.spellRepo.GetAvailableSpells(characterLevel, characterClass)
}
