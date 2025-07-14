// internal/service/pvp.go
package service

import (
"combat/internal/config"
"combat/internal/models"
"github.com/google/uuid"
)

// PvPServiceInterface définit les méthodes du service PvP
type PvPServiceInterface interface {
CreateChallenge(challengerID, challengedID uuid.UUID, req models.ChallengePvPRequest) error
AcceptChallenge(challengeID uuid.UUID) error
DeclineChallenge(challengeID uuid.UUID) error
GetRankings(season string, limit int) ([]*models.PvPRanking, error)
}

// PvPService implémente l'interface PvPServiceInterface
type PvPService struct {
config *config.Config
}

// NewPvPService crée une nouvelle instance du service PvP
func NewPvPService(cfg *config.Config) PvPServiceInterface {
return &PvPService{
config: cfg,
}
}

// CreateChallenge crée un défi PvP
func (s *PvPService) CreateChallenge(challengerID, challengedID uuid.UUID, req models.ChallengePvPRequest) error {
return nil // TODO: Implement
}

// AcceptChallenge accepte un défi
func (s *PvPService) AcceptChallenge(challengeID uuid.UUID) error {
return nil // TODO: Implement
}

// DeclineChallenge refuse un défi
func (s *PvPService) DeclineChallenge(challengeID uuid.UUID) error {
return nil // TODO: Implement
}

// GetRankings récupère les classements
func (s *PvPService) GetRankings(season string, limit int) ([]*models.PvPRanking, error) {
return []*models.PvPRanking{}, nil // TODO: Implement
}
