package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"player/internal/config"
	"player/internal/models"
	"player/internal/repository"
)

// PlayerService gère la logique métier des joueurs
type PlayerService struct {
	playerRepo    repository.PlayerRepositoryInterface
	characterRepo repository.CharacterRepositoryInterface
	config        *config.Config
}

// NewPlayerService crée un nouveau service de joueur
func NewPlayerService(
	playerRepo repository.PlayerRepositoryInterface,
	characterRepo repository.CharacterRepositoryInterface,
	config *config.Config,
) *PlayerService {
	return &PlayerService{
		playerRepo:    playerRepo,
		characterRepo: characterRepo,
		config:        config,
	}
}

// CreatePlayer crée un nouveau profil joueur
func (s *PlayerService) CreatePlayer(userID uuid.UUID, req models.CreatePlayerRequest) (*models.Player, error) {
	// Validation des données
	if err := s.validateCreatePlayerRequest(req); err != nil {
		return nil, err
	}

	// Vérifier que l'utilisateur n'a pas déjà un profil joueur
	existingPlayer, _ := s.playerRepo.GetByUserID(userID)
	if existingPlayer != nil {
		return nil, fmt.Errorf("player profile already exists for this user")
	}

	// Vérifier que le nom d'affichage est unique
	existingByName, _ := s.playerRepo.GetByDisplayName(req.DisplayName)
	if existingByName != nil {
		return nil, fmt.Errorf("display name already taken")
	}

	// Créer le nouveau joueur
	player := &models.Player{
		ID:            uuid.New(),
		UserID:        userID,
		DisplayName:   req.DisplayName,
		Avatar:        req.Avatar,
		Title:         "",
		GuildID:       nil,
		TotalPlayTime: 0,
		LastSeen:      time.Now(),
		Preferences:   models.GetDefaultPreferences(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Sauvegarder en base
	if err := s.playerRepo.Create(player); err != nil {
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"player_id":    player.ID,
		"user_id":      userID,
		"display_name": player.DisplayName,
	}).Info("Player profile created successfully")

	return player, nil
}

// GetPlayer récupère un profil joueur par ID utilisateur
func (s *PlayerService) GetPlayer(userID uuid.UUID) (*models.PlayerResponse, error) {
	// Récupérer le joueur
	player, err := s.playerRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("player not found: %w", err)
	}

	// Récupérer les personnages
	characters, err := s.characterRepo.GetByPlayerID(player.ID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get player characters")
		characters = []*models.Character{}
	}

	// Récupérer les statistiques
	stats, err := s.playerRepo.GetStats(player.ID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get player stats")
	}

	// Mettre à jour le nombre de personnages
	player.CharacterCount = len(characters)
	player.Characters = characters

	return &models.PlayerResponse{
		Player:     player,
		Stats:      stats,
		Characters: characters,
	}, nil
}

// GetPlayerByID récupère un profil joueur par son ID
func (s *PlayerService) GetPlayerByID(playerID uuid.UUID) (*models.Player, error) {
	return s.playerRepo.GetByID(playerID)
}

// UpdatePlayer met à jour un profil joueur
func (s *PlayerService) UpdatePlayer(userID uuid.UUID, req models.UpdatePlayerRequest) (*models.Player, error) {
	// Récupérer le joueur existing
	player, err := s.playerRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("player not found: %w", err)
	}

	// Validation des données
	if err := s.validateUpdatePlayerRequest(req); err != nil {
		return nil, err
	}

	// Vérifier l'unicité du nom d'affichage si changé
	if req.DisplayName != "" && req.DisplayName != player.DisplayName {
		existingByName, _ := s.playerRepo.GetByDisplayName(req.DisplayName)
		if existingByName != nil && existingByName.ID != player.ID {
			return nil, fmt.Errorf("display name already taken")
		}
		player.DisplayName = req.DisplayName
	}

	// Mettre à jour les champs
	if req.Avatar != "" {
		player.Avatar = req.Avatar
	}
	if req.Title != "" {
		player.Title = req.Title
	}

	// Mettre à jour les préférences
	player.Preferences = req.Preferences
	player.UpdatedAt = time.Now()

	// Sauvegarder
	if err := s.playerRepo.Update(player); err != nil {
		return nil, fmt.Errorf("failed to update player: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"player_id":    player.ID,
		"user_id":      userID,
		"display_name": player.DisplayName,
	}).Info("Player profile updated successfully")

	return player, nil
}

// UpdatePlayTime met à jour le temps de jeu d'un joueur
func (s *PlayerService) UpdatePlayTime(userID uuid.UUID, minutes int) error {
	player, err := s.playerRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("player not found: %w", err)
	}

	player.UpdatePlayTime(minutes)

	if err := s.playerRepo.Update(player); err != nil {
		return fmt.Errorf("failed to update play time: %w", err)
	}

	return nil
}

// UpdateLastSeen met à jour la dernière connection d'un joueur
func (s *PlayerService) UpdateLastSeen(userID uuid.UUID) error {
	player, err := s.playerRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("player not found: %w", err)
	}

	player.LastSeen = time.Now()
	player.UpdatedAt = time.Now()

	if err := s.playerRepo.Update(player); err != nil {
		return fmt.Errorf("failed to update last seen: %w", err)
	}

	return nil
}

// GetPlayerStats récupère les statistiques d'un joueur
func (s *PlayerService) GetPlayerStats(userID uuid.UUID) (*models.PlayerStats, error) {
	player, err := s.playerRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("player not found: %w", err)
	}

	return s.playerRepo.GetStats(player.ID)
}

// CanCreateCharacter vérifie si un joueur peut créer un nouveau personnage
func (s *PlayerService) CanCreateCharacter(userID uuid.UUID) (bool, error) {
	player, err := s.playerRepo.GetByUserID(userID)
	if err != nil {
		return false, fmt.Errorf("player not found: %w", err)
	}

	characterCount, err := s.playerRepo.GetCharacterCount(player.ID)
	if err != nil {
		return false, fmt.Errorf("failed to get character count: %w", err)
	}

	return characterCount < s.config.Game.MaxCharactersPerPlayer, nil
}

// ListPlayers récupère une liste de joueurs (admin/debug)
func (s *PlayerService) ListPlayers(limit, offset int) ([]*models.Player, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // Valeur par défaut
	}
	if offset < 0 {
		offset = 0
	}

	return s.playerRepo.List(limit, offset)
}

// SearchPlayersByDisplayName recherche des joueurs par nom d'affichage
func (s *PlayerService) SearchPlayersByDisplayName(query string) ([]*models.Player, error) {
	if strings.TrimSpace(query) == "" {
		return []*models.Player{}, nil
	}

	// Pour l'instant, implémentation simple
	// En production, on pourrait utiliser une recherche plus sophistiquée
	players, err := s.playerRepo.List(50, 0)
	if err != nil {
		return nil, err
	}

	var results []*models.Player
	queryLower := strings.ToLower(query)

	for _, player := range players {
		if strings.Contains(strings.ToLower(player.DisplayName), queryLower) {
			results = append(results, player)
		}
	}

	return results, nil
}

// DeletePlayer supprime un profil joueur (soft delete)
func (s *PlayerService) DeletePlayer(userID uuid.UUID) error {
	player, err := s.playerRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("player not found: %w", err)
	}

	// Supprimer tous les personnages d'abord
	characters, err := s.characterRepo.GetByPlayerID(player.ID)
	if err != nil {
		return fmt.Errorf("failed to get characters: %w", err)
	}

	for _, character := range characters {
		if err := s.characterRepo.Delete(character.ID); err != nil {
			logrus.WithError(err).WithField("character_id", character.ID).Error("Failed to delete character")
		}
	}

	// Supprimer le joueur
	if err := s.playerRepo.Delete(player.ID); err != nil {
		return fmt.Errorf("failed to delete player: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"player_id":    player.ID,
		"user_id":      userID,
		"display_name": player.DisplayName,
	}).Info("Player profile deleted successfully")

	return nil
}

// Close ferme le service joueur
func (s *PlayerService) Close() error {
	logrus.Info("Player service closed")
	return nil
}

// Méthodes de validation privées

// validateCreatePlayerRequest valide une demande de création de joueur
func (s *PlayerService) validateCreatePlayerRequest(req models.CreatePlayerRequest) error {
	// Validation du nom d'affichage
	if len(strings.TrimSpace(req.DisplayName)) < 3 {
		return fmt.Errorf("display name must be at least 3 characters long")
	}
	if len(req.DisplayName) > 20 {
		return fmt.Errorf("display name must be less than 20 characters long")
	}

	// Caractères autorisés pour le nom d'affichage
	if !isValidDisplayName(req.DisplayName) {
		return fmt.Errorf("display name contains invalid characters")
	}

	// Validation de l'avatar (optionnel)
	if req.Avatar != "" && len(req.Avatar) > 255 {
		return fmt.Errorf("avatar URL too long")
	}

	return nil
}

// validateUpdatePlayerRequest valide une demande de mise à jour
func (s *PlayerService) validateUpdatePlayerRequest(req models.UpdatePlayerRequest) error {
	// Validation du nom d'affichage si fourni
	if req.DisplayName != "" {
		if len(strings.TrimSpace(req.DisplayName)) < 3 {
			return fmt.Errorf("display name must be at least 3 characters long")
		}
		if len(req.DisplayName) > 20 {
			return fmt.Errorf("display name must be less than 20 characters long")
		}
		if !isValidDisplayName(req.DisplayName) {
			return fmt.Errorf("display name contains invalid characters")
		}
	}

	// Validation de l'avatar
	if req.Avatar != "" && len(req.Avatar) > 255 {
		return fmt.Errorf("avatar URL too long")
	}

	// Validation du titre
	if req.Title != "" && len(req.Title) > 50 {
		return fmt.Errorf("title too long")
	}

	return nil
}

// isValidDisplayName vérifie si un nom d'affichage est valide
func isValidDisplayName(displayName string) bool {
	// Autorise lettres, chiffres, espaces, tirets et underscores
	for _, char := range displayName {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == ' ' || char == '-' || char == '_') {
			return false
		}
	}

	// Vérifier qu'il n'y a pas d'espaces consécutifs
	return !strings.Contains(displayName, "  ")
}

