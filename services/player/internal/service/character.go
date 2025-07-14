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

// CharacterService gère la logique métier des personnages
type CharacterService struct {
	characterRepo repository.CharacterRepositoryInterface
	playerRepo    repository.PlayerRepositoryInterface
	config        *config.Config
}

// NewCharacterService crée un nouveau service de personnage
func NewCharacterService(
	characterRepo repository.CharacterRepositoryInterface,
	playerRepo repository.PlayerRepositoryInterface,
	config *config.Config,
) *CharacterService {
	return &CharacterService{
		characterRepo: characterRepo,
		playerRepo:    playerRepo,
		config:        config,
	}
}

// CreateCharacter crée un nouveau personnage
func (s *CharacterService) CreateCharacter(userID uuid.UUID, req models.CreateCharacterRequest) (*models.Character, error) {
	// Récupérer le joueur
	player, err := s.playerRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("player not found: %w", err)
	}

	// Vérifier si le joueur peut créer un nouveau personnage
	characterCount, err := s.playerRepo.GetCharacterCount(player.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character count: %w", err)
	}

	if characterCount >= s.config.Game.MaxCharactersPerPlayer {
		return nil, fmt.Errorf("maximum number of characters reached (%d)", s.config.Game.MaxCharactersPerPlayer)
	}

	// Validation des données
	if err := s.validateCreateCharacterRequest(req); err != nil {
		return nil, err
	}

	// Vérifier que le nom est unique
	existingByName, _ := s.characterRepo.GetByName(req.Name)
	if existingByName != nil {
		return nil, fmt.Errorf("character name already taken")
	}

	// Créer le nouveau personnage
	character := &models.Character{
		ID:         uuid.New(),
		PlayerID:   player.ID,
		Name:       req.Name,
		Class:      req.Class,
		Race:       req.Race,
		Gender:     req.Gender,
		Appearance: req.Appearance,
		Level:      s.config.Game.StartingLevel,
		Experience: 0,
		ZoneID:     "starting_zone",
		PositionX:  0,
		PositionY:  0,
		PositionZ:  0,
		Status:     "active",
		LastPlayed: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Sauvegarder en base (les stats seront créées automatiquement par le trigger)
	if err := s.characterRepo.Create(character); err != nil {
		return nil, fmt.Errorf("failed to create character: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"character_id": character.ID,
		"player_id":    player.ID,
		"name":         character.Name,
		"class":        character.Class,
		"race":         character.Race,
	}).Info("Character created successfully")

	return character, nil
}

// GetCharacter récupère un personnage par ID
func (s *CharacterService) GetCharacter(characterID uuid.UUID, userID uuid.UUID) (*models.Character, error) {
	// Vérifier que l'utilisateur est propriétaire du personnage
	if err := s.verifyCharacterOwnership(characterID, userID); err != nil {
		return nil, err
	}

	character, err := s.characterRepo.GetByID(characterID)
	if err != nil {
		return nil, fmt.Errorf("character not found: %w", err)
	}

	// Charger les stats
	stats, err := s.characterRepo.GetStats(characterID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to load character stats")
	} else {
		character.Stats = stats
	}

	return character, nil
}

// GetCharactersByPlayer récupère tous les personnages d'un joueur
func (s *CharacterService) GetCharactersByPlayer(userID uuid.UUID) ([]*models.Character, error) {
	player, err := s.playerRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("player not found: %w", err)
	}

	return s.characterRepo.GetByPlayerID(player.ID)
}

// UpdateCharacter met à jour un personnage
func (s *CharacterService) UpdateCharacter(characterID uuid.UUID, userID uuid.UUID, req models.UpdateCharacterRequest) (*models.Character, error) {
	// Vérifier la propriété
	if err := s.verifyCharacterOwnership(characterID, userID); err != nil {
		return nil, err
	}

	// Récupérer le personnage existant
	character, err := s.characterRepo.GetByID(characterID)
	if err != nil {
		return nil, fmt.Errorf("character not found: %w", err)
	}

	// Validation des données
	if err := s.validateUpdateCharacterRequest(req); err != nil {
		return nil, err
	}

	// Vérifier l'unicité du nom si changé
	if req.Name != "" && req.Name != character.Name {
		existingByName, _ := s.characterRepo.GetByName(req.Name)
		if existingByName != nil && existingByName.ID != character.ID {
			return nil, fmt.Errorf("character name already taken")
		}
		character.Name = req.Name
	}

	// Mettre à jour l'apparence
	character.Appearance = req.Appearance
	character.UpdatedAt = time.Now()

	// Sauvegarder
	if err := s.characterRepo.Update(character); err != nil {
		return nil, fmt.Errorf("failed to update character: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"character_id": character.ID,
		"name":         character.Name,
	}).Info("Character updated successfully")

	return character, nil
}

// DeleteCharacter supprime un personnage
func (s *CharacterService) DeleteCharacter(characterID uuid.UUID, userID uuid.UUID) error {
	// Vérifier la propriété
	if err := s.verifyCharacterOwnership(characterID, userID); err != nil {
		return err
	}

	// Supprimer le personnage (soft delete)
	if err := s.characterRepo.Delete(characterID); err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}

	logrus.WithField("character_id", characterID).Info("Character deleted successfully")
	return nil
}

// GetCharacterStats récupère les statistiques complètes d'un personnage
func (s *CharacterService) GetCharacterStats(characterID uuid.UUID, userID uuid.UUID) (*models.StatsResponse, error) {
	// Vérifier la propriété
	if err := s.verifyCharacterOwnership(characterID, userID); err != nil {
		return nil, err
	}

	// Récupérer les stats de base
	baseStats, err := s.characterRepo.GetStats(characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character stats: %w", err)
	}

	// Récupérer les stats de combat
	combatStats, err := s.characterRepo.GetCombatStats(characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get combat stats: %w", err)
	}

	// Récupérer les modificateurs actifs
	modifiers, err := s.characterRepo.GetActiveModifiers(characterID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get active modifiers")
		modifiers = []*models.StatModifier{}
	}

	return &models.StatsResponse{
		BaseStats:   baseStats,
		CombatStats: combatStats,
		Modifiers:   modifiers,
	}, nil
}

// UpdateCharacterStats met à jour les statistiques d'un personnage
func (s *CharacterService) UpdateCharacterStats(characterID uuid.UUID, userID uuid.UUID, req models.UpdateStatsRequest) (*models.CharacterStats, error) {
	// Vérifier la propriété
	if err := s.verifyCharacterOwnership(characterID, userID); err != nil {
		return nil, err
	}

	// Récupérer les stats actuelles
	stats, err := s.characterRepo.GetStats(characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character stats: %w", err)
	}

	// Calculer le total de points à dépenser
	totalPointsToSpend := 0
	if req.Strength != nil {
		totalPointsToSpend += *req.Strength
	}
	if req.Agility != nil {
		totalPointsToSpend += *req.Agility
	}
	if req.Intelligence != nil {
		totalPointsToSpend += *req.Intelligence
	}
	if req.Vitality != nil {
		totalPointsToSpend += *req.Vitality
	}

	// Vérifier que le joueur a assez de points
	if stats.StatPoints < totalPointsToSpend {
		return nil, fmt.Errorf("not enough stat points (have %d, need %d)", stats.StatPoints, totalPointsToSpend)
	}

	// Appliquer les changements
	if req.Strength != nil && *req.Strength > 0 {
		stats.SpendStatPoint("strength", *req.Strength)
	}
	if req.Agility != nil && *req.Agility > 0 {
		stats.SpendStatPoint("agility", *req.Agility)
	}
	if req.Intelligence != nil && *req.Intelligence > 0 {
		stats.SpendStatPoint("intelligence", *req.Intelligence)
	}
	if req.Vitality != nil && *req.Vitality > 0 {
		stats.SpendStatPoint("vitality", *req.Vitality)
	}

	// Sauvegarder
	if err := s.characterRepo.UpdateStats(stats); err != nil {
		return nil, fmt.Errorf("failed to update stats: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"character_id":   characterID,
		"points_spent":   totalPointsToSpend,
		"remaining_points": stats.StatPoints,
	}).Info("Character stats updated successfully")

	return stats, nil
}

// AddExperience ajoute de l'expérience à un personnage
func (s *CharacterService) AddExperience(characterID uuid.UUID, experience int64) error {
	character, err := s.characterRepo.GetByID(characterID)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	character.Experience += experience
	levelsGained := 0

	// Vérifier les montées de niveau
	for character.CanLevelUp() && character.Level < s.config.Game.MaxLevel {
		character.LevelUp()
		levelsGained++
		
		// Ajouter des points de stats et de compétences
		stats, err := s.characterRepo.GetStats(characterID)
		if err == nil {
			stats.AddStatPoints(5) // 5 points par niveau
			stats.SkillPoints += 2  // 2 points de compétence par niveau
			s.characterRepo.UpdateStats(stats)
		}
	}

	// Sauvegarder le personnage
	if err := s.characterRepo.Update(character); err != nil {
		return fmt.Errorf("failed to update character: %w", err)
	}

	if levelsGained > 0 {
		logrus.WithFields(logrus.Fields{
			"character_id":   characterID,
			"levels_gained":  levelsGained,
			"new_level":      character.Level,
			"experience":     character.Experience,
		}).Info("Character leveled up")
	}

	return nil
}

// UpdatePosition met à jour la position d'un personnage
func (s *CharacterService) UpdatePosition(characterID uuid.UUID, zoneID string, x, y, z float64) error {
	character, err := s.characterRepo.GetByID(characterID)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	character.ZoneID = zoneID
	character.PositionX = x
	character.PositionY = y
	character.PositionZ = z
	character.LastPlayed = time.Now()
	character.UpdatedAt = time.Now()

	if err := s.characterRepo.Update(character); err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}

	return nil
}

// GetAvailableClassesAndRaces récupère les classes et races disponibles
func (s *CharacterService) GetAvailableClassesAndRaces() map[string]interface{} {
	return map[string]interface{}{
		"classes": models.CharacterClasses,
		"races":   models.CharacterRaces,
		"config": map[string]interface{}{
			"max_level":      s.config.Game.MaxLevel,
			"starting_level": s.config.Game.StartingLevel,
			"starting_stats": s.config.Game.StartingStats,
		},
	}
}

// CleanupExpiredModifiers nettoie les modificateurs expirés
func (s *CharacterService) CleanupExpiredModifiers() error {
	return s.characterRepo.CleanupExpiredModifiers()
}

// Close ferme le service de personnage
func (s *CharacterService) Close() error {
	logrus.Info("Character service closed")
	return nil
}

// Méthodes privées

// verifyCharacterOwnership vérifie que l'utilisateur est propriétaire du personnage
func (s *CharacterService) verifyCharacterOwnership(characterID uuid.UUID, userID uuid.UUID) error {
	character, err := s.characterRepo.GetByID(characterID)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	player, err := s.playerRepo.GetByID(character.PlayerID)
	if err != nil {
		return fmt.Errorf("player not found: %w", err)
	}

	if player.UserID != userID {
		return fmt.Errorf("access denied: not the owner of this character")
	}

	return nil
}

// validateCreateCharacterRequest valide une demande de création de personnage
func (s *CharacterService) validateCreateCharacterRequest(req models.CreateCharacterRequest) error {
	// Validation du nom
	if len(strings.TrimSpace(req.Name)) < 3 {
		return fmt.Errorf("character name must be at least 3 characters long")
	}
	if len(req.Name) > 20 {
		return fmt.Errorf("character name must be less than 20 characters long")
	}
	if !isValidCharacterName(req.Name) {
		return fmt.Errorf("character name contains invalid characters")
	}

	// Validation de la classe
	if !models.IsValidClass(req.Class) {
		return fmt.Errorf("invalid class: %s", req.Class)
	}

	// Validation de la race
	if !models.IsValidRace(req.Race) {
		return fmt.Errorf("invalid race: %s", req.Race)
	}

	// Validation du genre
	if !models.IsValidGender(req.Gender) {
		return fmt.Errorf("invalid gender: %s", req.Gender)
	}

	return nil
}

// validateUpdateCharacterRequest valide une demande de mise à jour
func (s *CharacterService) validateUpdateCharacterRequest(req models.UpdateCharacterRequest) error {
	// Validation du nom si fourni
	if req.Name != "" {
		if len(strings.TrimSpace(req.Name)) < 3 {
			return fmt.Errorf("character name must be at least 3 characters long")
		}
		if len(req.Name) > 20 {
			return fmt.Errorf("character name must be less than 20 characters long")
		}
		if !isValidCharacterName(req.Name) {
			return fmt.Errorf("character name contains invalid characters")
		}
	}

	return nil
}

// isValidCharacterName vérifie si un nom de personnage est valide
func isValidCharacterName(name string) bool {
	// Autorise lettres, chiffres et quelques caractères spéciaux
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '\'' || char == '-') {
			return false
		}
	}
	
	// Pas d'espaces consécutifs ou de caractères spéciaux consécutifs
	return !strings.Contains(name, "--") && !strings.Contains(name, "''")
}