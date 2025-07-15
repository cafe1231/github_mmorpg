// internal/service/interfaces.go
package service

import (
	"time"
	
	"github.com/google/uuid"
	"combat/internal/models"
)

// CombatServiceInterface définit les méthodes du service de combat
type CombatServiceInterface interface {
	CreateCombatSession(creatorID uuid.UUID, req models.StartCombatRequest) (*models.CombatSession, error)
	GetCombatSession(sessionID uuid.UUID) (*models.CombatSession, error)
	JoinCombatSession(sessionID, characterID uuid.UUID) error
	LeaveCombatSession(sessionID, characterID uuid.UUID) error
	ExecuteAction(actionReq models.PerformActionRequest) (*models.CombatActionResult, error) // Corrigé: utilise PerformActionRequest
	AddParticipant(participant *models.CombatParticipant) error
	GetParticipants(sessionID uuid.UUID) ([]*models.CombatParticipant, error)
	EndCombatSession(sessionID uuid.UUID) error
}

// AntiCheatServiceInterface définit les méthodes du service anti-cheat
type AntiCheatServiceInterface interface {
	ValidateCombatAction(action *models.PerformActionRequest, participant *models.CombatParticipant) error // Corrigé
	ValidateMovement(characterID uuid.UUID, oldPos, newPos map[string]interface{}, deltaTime time.Duration) error
	ValidateSpellCast(characterID uuid.UUID, spell *models.Spell, castTime time.Duration) error
	ValidateDamage(attacker, defender *models.CombatParticipant, damage int, damageType string) error
	CheckSuspiciousActivity(characterID uuid.UUID) (*models.SuspicionReport, error)
	ReportViolation(violation *models.AntiCheatViolation) error
	GetPlayerViolations(characterID uuid.UUID) ([]*models.AntiCheatViolation, error)
	IsPlayerBanned(characterID uuid.UUID) (bool, string, error)
}