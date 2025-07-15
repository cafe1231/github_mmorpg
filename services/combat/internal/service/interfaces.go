package service

import (
	"github.com/google/uuid"
	"combat/internal/models"
)

// CombatServiceInterface définit les méthodes du service de combat
type CombatServiceInterface interface {
	CreateCombatSession(creatorID uuid.UUID, req models.StartCombatRequest) (*models.CombatSession, error)
	GetCombatSession(sessionID uuid.UUID) (*models.CombatSession, error)
	JoinCombatSession(sessionID, characterID uuid.UUID) error
	LeaveCombatSession(sessionID, characterID uuid.UUID) error
	ExecuteAction(actionReq models.CombatActionRequest) (*models.CombatActionResult, error)
	AddParticipant(participant *models.CombatParticipant) error
	GetParticipants(sessionID uuid.UUID) ([]*models.CombatParticipant, error)
	EndCombatSession(sessionID uuid.UUID) error
}