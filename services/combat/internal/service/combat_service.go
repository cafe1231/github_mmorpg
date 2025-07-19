package service

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"combat/internal/config"
	"combat/internal/models"
	"combat/internal/repository"
)

// CombatServiceInterface définit les méthodes du service combat
type CombatServiceInterface interface {
	// Gestion des combats
	CreateCombat(req *models.CreateCombatRequest) (*models.CombatInstance, error)
	GetCombat(id uuid.UUID) (*models.CombatInstance, error)
	GetCombatStatus(id uuid.UUID, req *models.GetCombatStatusRequest) (*models.CombatStatusResponse, error)
	StartCombat(id uuid.UUID) error
	EndCombat(id uuid.UUID, req *models.EndCombatRequest) (*models.CombatResult, error)

	// Gestion des participants
	JoinCombat(combatID uuid.UUID, req *models.JoinCombatRequest) error
	LeaveCombat(combatID, characterID uuid.UUID, req *models.LeaveCombatRequest) error
	GetParticipants(combatID uuid.UUID) ([]*models.CombatParticipant, error)
	UpdateParticipant(participant *models.CombatParticipant) error

	// Actions de combat
	ExecuteAction(combatID, actorID uuid.UUID, req *models.ActionRequest) (*models.ActionResult, error)
	ValidateAction(combatID, actorID uuid.UUID, req *models.ValidateActionRequest) (*models.ValidationResponse, error)
	GetAvailableActions(combatID, actorID uuid.UUID) ([]*models.ActionTemplate, error)

	// Gestion des tours
	ProcessTurn(combatID uuid.UUID) error
	AdvanceTurn(combatID uuid.UUID) error
	GetCurrentTurn(combatID uuid.UUID) (*models.TurnInfo, error)

	// Recherche et historique
	SearchCombats(req *models.SearchCombatsRequest) (*models.CombatListResponse, error)
	GetCombatHistory(req *models.GetCombatHistoryRequest) (*models.CombatHistoryResponse, error)
	GetStatistics(req *models.GetStatisticsRequest) (*models.StatisticsResponse, error)

	// Maintenance
	CleanupExpiredCombats() error
	GetActiveCombatCount() (int, error)
}

// CombatService implémente l'interface CombatServiceInterface
type CombatService struct {
	combatRepo    repository.CombatRepositoryInterface
	actionRepo    repository.ActionRepositoryInterface
	effectRepo    repository.EffectRepositoryInterface
	actionService ActionServiceInterface
	effectService EffectServiceInterface
	antiCheat     AntiCheatServiceInterface
	config        *config.Config
}

// NewCombatService crée un nouveau service de combat
func NewCombatService(
	combatRepo repository.CombatRepositoryInterface,
	actionRepo repository.ActionRepositoryInterface,
	effectRepo repository.EffectRepositoryInterface,
	actionService ActionServiceInterface,
	effectService EffectServiceInterface,
	antiCheat AntiCheatServiceInterface,
	config *config.Config,
) CombatServiceInterface {
	return &CombatService{
		combatRepo:    combatRepo,
		actionRepo:    actionRepo,
		effectRepo:    effectRepo,
		actionService: actionService,
		effectService: effectService,
		antiCheat:     antiCheat,
		config:        config,
	}
}

// CreateCombat crée un nouveau combat
func (s *CombatService) CreateCombat(req *models.CreateCombatRequest) (*models.CombatInstance, error) {
	// Validation de la demande
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Vérifier la limite de combats actifs
	activeCount, err := s.combatRepo.GetActiveCombatCount()
	if err != nil {
		return nil, fmt.Errorf("failed to check active combat count: %w", err)
	}

	if activeCount >= s.config.Combat.MaxConcurrent {
		return nil, fmt.Errorf("maximum concurrent combats reached: %d", s.config.Combat.MaxConcurrent)
	}

	// Créer l'instance de combat
	combat := &models.CombatInstance{
		ID:              uuid.New(),
		CombatType:      req.CombatType,
		Status:          models.CombatStatusWaiting,
		ZoneID:          &req.ZoneID,
		MaxParticipants: req.MaxParticipants,
		CurrentTurn:     0,
		TurnTimeLimit:   req.TurnTimeLimit,
		MaxDuration:     req.MaxDuration,
		Settings:        *req.Settings,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Valeurs par défaut
	if combat.MaxParticipants == 0 {
		combat.MaxParticipants = 4
	}
	if combat.TurnTimeLimit == 0 {
		combat.TurnTimeLimit = int(s.config.Combat.TurnTimeout.Seconds())
	}
	if combat.MaxDuration == 0 {
		combat.MaxDuration = int(s.config.Combat.MaxDuration.Seconds())
	}
	if req.Settings == nil {
		defaultSettings := models.GetDefaultCombatSettings()
		combat.Settings = defaultSettings
	}

	// Sauvegarder en base
	if err := s.combatRepo.Create(combat); err != nil {
		return nil, fmt.Errorf("failed to create combat: %w", err)
	}

	// Ajouter les participants initiaux
	for _, participantReq := range req.Participants {
		participant := &models.CombatParticipant{
			ID:          uuid.New(),
			CombatID:    combat.ID,
			CharacterID: participantReq.CharacterID,
			UserID:      participantReq.UserID,
			Team:        participantReq.Team,
			Position:    participantReq.Position,
			IsAlive:     true,
			IsReady:     false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// TODO: Récupérer les stats du personnage depuis le service player
		// Pour l'instant, utiliser des valeurs par défaut
		participant.Health = 100
		participant.MaxHealth = 100
		participant.Mana = 50
		participant.MaxMana = 50
		participant.PhysicalDamage = 20
		participant.MagicalDamage = 15
		participant.PhysicalDefense = 10
		participant.MagicalDefense = 8
		participant.CriticalChance = 0.05
		participant.AttackSpeed = 1.0

		if err := s.combatRepo.AddParticipant(participant); err != nil {
			return nil, fmt.Errorf("failed to add participant: %w", err)
		}
	}

	logrus.WithFields(logrus.Fields{
		"combat_id":    combat.ID,
		"combat_type":  combat.CombatType,
		"zone_id":      combat.ZoneID,
		"participants": len(req.Participants),
	}).Info("Combat created")

	return combat, nil
}

// GetCombat récupère un combat par son ID
func (s *CombatService) GetCombat(id uuid.UUID) (*models.CombatInstance, error) {
	combat, err := s.combatRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("combat not found: %w", err)
	}

	// Charger les participants
	participants, err := s.combatRepo.GetParticipants(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load participants: %w", err)
	}
	combat.Participants = participants

	return combat, nil
}

// GetCombatStatus récupère le statut détaillé d'un combat
func (s *CombatService) GetCombatStatus(id uuid.UUID, req *models.GetCombatStatusRequest) (*models.CombatStatusResponse, error) {
	combat, err := s.GetCombat(id)
	if err != nil {
		return nil, err
	}

	response := &models.CombatStatusResponse{
		Combat: combat,
	}

	// Charger les participants si demandé
	if req.IncludeParticipants {
		participants, err := s.combatRepo.GetParticipants(id)
		if err != nil {
			return nil, fmt.Errorf("failed to load participants: %w", err)
		}
		response.Participants = participants
	}

	// Charger les actions récentes si demandé
	if req.IncludeActions {
		actions, err := s.actionRepo.GetRecentActions(id, config.DefaultMinIntervals)
		if err != nil {
			return nil, fmt.Errorf("failed to load recent actions: %w", err)
		}
		response.RecentActions = actions
	}

	// Charger les effets actifs si demandé
	if req.IncludeEffects {
		effects, err := s.effectRepo.GetActiveByCombat(id)
		if err != nil {
			return nil, fmt.Errorf("failed to load active effects: %w", err)
		}
		response.ActiveEffects = effects
	}

	// Charger les logs si demandé
	if req.IncludeLogs {
		// TODO: Implémenter la récupération des logs
		response.Logs = []*models.CombatLog{}
	}

	// Informations du tour actuel
	if combat.Status == models.CombatStatusActive {
		turnInfo, err := s.GetCurrentTurn(id)
		if err == nil {
			response.CurrentTurn = turnInfo
		}
	}

	return response, nil
}

// StartCombat démarre un combat
func (s *CombatService) StartCombat(id uuid.UUID) error {
	combat, err := s.combatRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("combat not found: %w", err)
	}

	if combat.Status != models.CombatStatusWaiting {
		return fmt.Errorf("combat cannot be started, current status: %s", combat.Status)
	}

	// Vérifier qu'il y a au moins 2 participants
	participants, err := s.combatRepo.GetParticipants(id)
	if err != nil {
		return fmt.Errorf("failed to get participants: %w", err)
	}

	if len(participants) < config.DefaultMinParticipants {
		return fmt.Errorf("au moins %d participants requis", config.DefaultMinParticipants)
	}

	// Vérifier que tous les participants sont prêts
	for _, p := range participants {
		if !p.IsReady {
			return fmt.Errorf("all participants must be ready")
		}
	}

	// Démarrer le combat
	now := time.Now()
	combat.Status = models.CombatStatusActive
	combat.StartedAt = &now
	combat.CurrentTurn = 1

	if err := s.combatRepo.Update(combat); err != nil {
		return fmt.Errorf("failed to start combat: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"combat_id":    combat.ID,
		"participants": len(participants),
	}).Info("Combat started")

	return nil
}

// EndCombat termine un combat
func (s *CombatService) EndCombat(id uuid.UUID, req *models.EndCombatRequest) (*models.CombatResult, error) {
	combat, err := s.combatRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("combat not found: %w", err)
	}

	if combat.IsFinished() {
		return nil, fmt.Errorf("combat already finished")
	}

	// Charger les participants
	participants, err := s.combatRepo.GetParticipants(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}

	// Terminer le combat
	now := time.Now()
	combat.Status = models.CombatStatusFinished
	combat.EndedAt = &now

	if err := s.combatRepo.Update(combat); err != nil {
		return nil, fmt.Errorf("failed to end combat: %w", err)
	}

	// Calculer les résultats
	result := s.calculateCombatResult(combat, participants, req)

	// Mettre à jour les statistiques des participants
	if err := s.updateParticipantStatistics(combat, participants, result); err != nil {
		logrus.WithError(err).Error("Failed to update participant statistics")
	}

	logrus.WithFields(logrus.Fields{
		"combat_id":    combat.ID,
		"duration":     result.Duration,
		"winning_team": result.WinningTeam,
		"end_reason":   result.EndReason,
	}).Info("Combat ended")

	return result, nil
}

// JoinCombat ajoute un participant à un combat
func (s *CombatService) JoinCombat(combatID uuid.UUID, req *models.JoinCombatRequest) error {
	combat, err := s.combatRepo.GetByID(combatID)
	if err != nil {
		return fmt.Errorf("combat not found: %w", err)
	}

	if combat.Status != models.CombatStatusWaiting {
		return fmt.Errorf("cannot join combat with status: %s", combat.Status)
	}

	// Vérifier la limite de participants
	participants, err := s.combatRepo.GetParticipants(combatID)
	if err != nil {
		return fmt.Errorf("failed to get participants: %w", err)
	}

	if len(participants) >= combat.MaxParticipants {
		return fmt.Errorf("combat is full (%d/%d)", len(participants), combat.MaxParticipants)
	}

	// Vérifier que le personnage n'est pas déjà dans le combat
	for _, p := range participants {
		if p.CharacterID == req.CharacterID {
			return fmt.Errorf("character already in combat")
		}
	}

	// Créer le participant
	participant := &models.CombatParticipant{
		ID:          uuid.New(),
		CombatID:    combatID,
		CharacterID: req.CharacterID,
		Team:        req.Team,
		Position:    req.Position,
		IsAlive:     true,
		IsReady:     false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// TODO: Récupérer les stats du personnage depuis le service player
	// Pour l'instant, utiliser des valeurs par défaut
	participant.Health = 100
	participant.MaxHealth = 100
	participant.Mana = 50
	participant.MaxMana = 50
	participant.PhysicalDamage = 20
	participant.MagicalDamage = 15
	participant.PhysicalDefense = 10
	participant.MagicalDefense = 8
	participant.CriticalChance = 0.05
	participant.AttackSpeed = 1.0

	if err := s.combatRepo.AddParticipant(participant); err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"combat_id":    combatID,
		"character_id": req.CharacterID,
		"team":         req.Team,
	}).Info("Player joined combat")

	return nil
}

// LeaveCombat retire un participant d'un combat
func (s *CombatService) LeaveCombat(combatID, characterID uuid.UUID, req *models.LeaveCombatRequest) error {
	combat, err := s.combatRepo.GetByID(combatID)
	if err != nil {
		return fmt.Errorf("combat not found: %w", err)
	}

	// Vérifier si le combat permet de partir
	if combat.Status == models.CombatStatusActive && !combat.Settings.AllowFlee {
		return fmt.Errorf("fleeing not allowed in this combat")
	}

	if err := s.combatRepo.RemoveParticipant(combatID, characterID); err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"combat_id":    combatID,
		"character_id": characterID,
		"reason":       req.Reason,
	}).Info("Player left combat")

	// Vérifier si le combat doit se terminer
	participants, err := s.combatRepo.GetParticipants(combatID)
	if err != nil {
		return fmt.Errorf("failed to get participants: %w", err)
	}

	if len(participants) < 2 && combat.Status == models.CombatStatusActive {
		// Terminer le combat automatiquement
		endReq := &models.EndCombatRequest{
			Reason:   "insufficient_participants",
			ForceEnd: true,
		}
		_, err := s.EndCombat(combatID, endReq)
		if err != nil {
			logrus.WithError(err).Error("Failed to auto-end combat")
		}
	}

	return nil
}

// GetParticipants récupère les participants d'un combat
func (s *CombatService) GetParticipants(combatID uuid.UUID) ([]*models.CombatParticipant, error) {
	return s.combatRepo.GetParticipants(combatID)
}

// UpdateParticipant met à jour un participant
func (s *CombatService) UpdateParticipant(participant *models.CombatParticipant) error {
	return s.combatRepo.UpdateParticipant(participant)
}

// ExecuteAction exécute une action de combat
func (s *CombatService) ExecuteAction(combatID, actorID uuid.UUID, req *models.ActionRequest) (*models.ActionResult, error) {
	// Vérifier que le combat est actif
	combat, err := s.combatRepo.GetByID(combatID)
	if err != nil {
		return nil, fmt.Errorf("combat not found: %w", err)
	}

	if combat.Status != models.CombatStatusActive {
		return nil, fmt.Errorf("combat is not active")
	}

	// Vérifier que l'acteur peut agir
	actor, err := s.combatRepo.GetParticipant(combatID, actorID)
	if err != nil {
		return nil, fmt.Errorf("participant not found: %w", err)
	}

	if !actor.IsAlive {
		return nil, fmt.Errorf("dead participants cannot act")
	}

	// Validation anti-cheat
	if validation := s.antiCheat.ValidateAction(actor, req); !validation.Valid {
		logrus.WithFields(logrus.Fields{
			"combat_id":   combatID,
			"actor_id":    actorID,
			"action_type": req.ActionType,
			"errors":      validation.Errors,
			"suspicious":  validation.AntiCheat.Suspicious,
		}).Warn("Suspicious action detected")

		if validation.AntiCheat.Action == "block" {
			return &models.ActionResult{
				Success: false,
				Error:   "Action blocked by anti-cheat system",
			}, nil
		}
	}

	// Déléguer l'exécution au service d'actions
	return s.actionService.ExecuteAction(combat, actor, req)
}

// ValidateAction valide une action sans l'exécuter
func (s *CombatService) ValidateAction(combatID, actorID uuid.UUID, req *models.ValidateActionRequest) (*models.ValidationResponse, error) {
	combat, err := s.combatRepo.GetByID(combatID)
	if err != nil {
		return nil, fmt.Errorf("combat not found: %w", err)
	}

	actor, err := s.combatRepo.GetParticipant(combatID, actorID)
	if err != nil {
		return nil, fmt.Errorf("participant not found: %w", err)
	}

	return s.actionService.ValidateAction(combat, actor, req)
}

// GetAvailableActions récupère les actions disponibles pour un participant
func (s *CombatService) GetAvailableActions(combatID, actorID uuid.UUID) ([]*models.ActionTemplate, error) {
	combat, err := s.combatRepo.GetByID(combatID)
	if err != nil {
		return nil, fmt.Errorf("combat not found: %w", err)
	}

	actor, err := s.combatRepo.GetParticipant(combatID, actorID)
	if err != nil {
		return nil, fmt.Errorf("participant not found: %w", err)
	}

	return s.actionService.GetAvailableActions(combat, actor)
}

// ProcessTurn traite un tour de combat
func (s *CombatService) ProcessTurn(combatID uuid.UUID) error {
	combat, err := s.combatRepo.GetByID(combatID)
	if err != nil {
		return fmt.Errorf("combat not found: %w", err)
	}

	if combat.Status != models.CombatStatusActive {
		return fmt.Errorf("combat is not active")
	}

	// Traiter les effets actifs
	participants, err := s.combatRepo.GetParticipants(combatID)
	if err != nil {
		return fmt.Errorf("failed to get participants: %w", err)
	}

	for _, participant := range participants {
		if err := s.effectService.ProcessEffects(participant); err != nil {
			logrus.WithError(err).WithField("participant_id", participant.ID).Error("Failed to process effects")
		}
	}

	// Vérifier les conditions de victoire
	if winner := s.checkWinConditions(participants); winner != nil {
		endReq := &models.EndCombatRequest{
			Reason:   "victory_condition",
			WinnerID: winner, // winner est déjà un *int
		}
		_, err := s.EndCombat(combatID, endReq)
		return err
	}

	return nil
}

// AdvanceTurn avance au tour suivant
func (s *CombatService) AdvanceTurn(combatID uuid.UUID) error {
	combat, err := s.combatRepo.GetByID(combatID)
	if err != nil {
		return fmt.Errorf("combat not found: %w", err)
	}

	combat.CurrentTurn++
	return s.combatRepo.Update(combat)
}

// GetCurrentTurn récupère les informations du tour actuel
func (s *CombatService) GetCurrentTurn(combatID uuid.UUID) (*models.TurnInfo, error) {
	combat, err := s.combatRepo.GetByID(combatID)
	if err != nil {
		return nil, fmt.Errorf("combat not found: %w", err)
	}

	// TODO: Implémenter la logique de tour plus sophistiquée
	return &models.TurnInfo{
		TurnNumber:      combat.CurrentTurn,
		TimeRemaining:   combat.TurnTimeLimit,
		TurnStartTime:   time.Now(),
		ActionsThisTurn: 0,
		CanAct:          true,
	}, nil
}

// SearchCombats recherche des combats
func (s *CombatService) SearchCombats(req *models.SearchCombatsRequest) (*models.CombatListResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	combats, total, err := s.combatRepo.List(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search combats: %w", err)
	}

	// Convertir en résumés
	summaries := make([]*models.CombatListItem, len(combats))
	for i, combat := range combats {
		summaries[i] = s.combatToSummary(combat)
	}

	return &models.CombatListResponse{
		Combats:  summaries,
		Total:    total,
		Page:     req.Offset/req.Limit + 1,
		PageSize: req.Limit,
		HasMore:  req.Offset+req.Limit < total,
	}, nil
}

// GetCombatHistory récupère l'historique de combat
func (s *CombatService) GetCombatHistory(req *models.GetCombatHistoryRequest) (*models.CombatHistoryResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	history, total, err := s.combatRepo.GetCombatHistory(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get combat history: %w", err)
	}

	// Calculer le résumé
	summary := s.calculateHistorySummary(history)

	return &models.CombatHistoryResponse{
		History:  history,
		Total:    total,
		Page:     req.Offset/req.Limit + 1,
		PageSize: req.Limit,
		HasMore:  req.Offset+req.Limit < total,
		Summary:  summary,
	}, nil
}

// GetStatistics récupère les statistiques de combat
func (s *CombatService) GetStatistics(req *models.GetStatisticsRequest) (*models.StatisticsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var characterID uuid.UUID
	if req.CharacterID != nil {
		characterID = *req.CharacterID
	}

	stats, err := s.combatRepo.GetStatistics(characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	// Convertir en réponse formatée
	return s.formatStatisticsResponse(stats, req), nil
}

// CleanupExpiredCombats nettoie les combats expirés
func (s *CombatService) CleanupExpiredCombats() error {
	return s.combatRepo.CleanupExpiredCombats()
}

// GetActiveCombatCount retourne le nombre de combats actifs
func (s *CombatService) GetActiveCombatCount() (int, error) {
	return s.combatRepo.GetActiveCombatCount()
}

// Helper methods

func (s *CombatService) calculateCombatResult(
	combat *models.CombatInstance,
	participants []*models.CombatParticipant,
	req *models.EndCombatRequest,
) *models.CombatResult {
	result := &models.CombatResult{
		CombatID:     combat.ID,
		Status:       combat.Status,
		Duration:     combat.GetDuration(),
		Participants: participants,
		EndReason:    req.Reason,
	}

	// Déterminer l'équipe gagnante
	if req.WinnerID != nil {
		result.WinningTeam = req.WinnerID
	} else {
		// Déterminer automatiquement le gagnant
		teamAlive := make(map[int]int)
		for _, p := range participants {
			if p.IsAlive {
				teamAlive[p.Team]++
			}
		}
		for team, count := range teamAlive {
			if count > 0 {
				result.WinningTeam = &team
				break
			}
		}
	}

	// Calculer les récompenses
	result.Rewards = s.calculateRewards(combat, participants, result.WinningTeam)

	// Créer le résumé
	result.Summary = s.createCombatSummary(combat, participants)

	return result
}

func (s *CombatService) calculateRewards(
	combat *models.CombatInstance,
	participants []*models.CombatParticipant,
	winningTeam *int,
) map[uuid.UUID]*models.CombatReward {
	rewards := make(map[uuid.UUID]*models.CombatReward)

	for _, p := range participants {
		reward := &models.CombatReward{
			Experience: config.DefaultBaseExperience, // Base XP
			Gold:       config.DefaultBaseGold,       // Base gold
		}

		// Bonus pour les gagnants
		if winningTeam != nil && p.Team == *winningTeam {
			reward.Experience *= 2
			reward.Gold *= 2
		}

		// Bonus basé sur les performances
		if p.DamageDealt > config.DefaultMaxDamage {
			p.DamageDealt = config.DefaultMaxDamage
		}
		if p.HealingDone > config.DefaultMaxHealing {
			p.HealingDone = config.DefaultMaxHealing
		}

		rewards[p.CharacterID] = reward
	}

	return rewards
}

func (s *CombatService) createCombatSummary(combat *models.CombatInstance, participants []*models.CombatParticipant) *models.CombatSummary {
	var totalDamage, totalHealing int64
	playerStats := make(map[uuid.UUID]*models.ParticipantStats)

	for _, p := range participants {
		totalDamage += int64(p.DamageDealt)
		totalHealing += int64(p.HealingDone)

		playerStats[p.CharacterID] = &models.ParticipantStats{
			DamageDealt: p.DamageDealt,
			DamageTaken: p.DamageTaken,
			HealingDone: p.HealingDone,
			// TODO: Ajouter d'autres statistiques
		}
	}

	return &models.CombatSummary{
		TotalTurns:   combat.CurrentTurn,
		TotalDamage:  totalDamage,
		TotalHealing: totalHealing,
		PlayerStats:  playerStats,
	}
}

func (s *CombatService) updateParticipantStatistics(
	combat *models.CombatInstance,
	participants []*models.CombatParticipant,
	result *models.CombatResult,
) error {
	for _, participant := range participants {
		// Récupérer les statistiques existing
		stats, err := s.combatRepo.GetStatistics(participant.CharacterID)
		if err != nil {
			logrus.WithError(err).WithField("character_id", participant.CharacterID).Error("Failed to get statistics")
			continue
		}

		// Mettre à jour les statistiques générales
		stats.TotalDamageDealt += int64(participant.DamageDealt)
		stats.TotalDamageTaken += int64(participant.DamageTaken)
		stats.TotalHealingDone += int64(participant.HealingDone)

		if !participant.IsAlive {
			stats.TotalDeaths++
		}

		if participant.DamageDealt > stats.HighestDamageDealt {
			stats.HighestDamageDealt = participant.DamageDealt
		}

		duration := int(result.Duration.Seconds())
		if duration > stats.LongestCombatDuration {
			stats.LongestCombatDuration = duration
		}

		// Mettre à jour les statistiques selon le type de combat
		isWinner := result.WinningTeam != nil && participant.Team == *result.WinningTeam

		switch combat.CombatType {
		case models.CombatTypePvE:
			if isWinner {
				stats.PvEBattlesWon++
				stats.MonstersKilled++ // Simplifié pour l'example
			} else {
				stats.PvEBattlesLost++
			}

		case models.CombatTypePvP:
			if isWinner {
				stats.PvPBattlesWon++
			} else if result.WinningTeam == nil {
				stats.PvPDraws++
			} else {
				stats.PvPBattlesLost++
			}

			// Calcul du rating PvP (système Elo simplifié)
			if result.WinningTeam != nil {
				ratingChange := s.calculatePvPRatingChange(participant, participants, isWinner)
				stats.PvPRating += ratingChange
				if stats.PvPRating < 0 {
					stats.PvPRating = 0
				}
			}
		}

		// Sauvegarder les statistiques
		if err := s.combatRepo.UpdateStatistics(stats); err != nil {
			logrus.WithError(err).WithField("character_id", participant.CharacterID).Error("Failed to update statistics")
		}
	}

	return nil
}

func (s *CombatService) calculatePvPRatingChange(
	participant *models.CombatParticipant,
	allParticipants []*models.CombatParticipant,
	isWinner bool,
) int {
	// Système Elo simplifié
	k := 32.0 // Facteur K

	// Pour simplifier, on prend la moyenne des ratings adverses
	var opponentRatings []int
	for _, p := range allParticipants {
		if p.Team != participant.Team {
			// On devrait récupérer le rating réel ici
			opponentRatings = append(opponentRatings, config.DefaultPvPRating) // Rating par défaut
		}
	}

	if len(opponentRatings) == 0 {
		return 0
	}

	avgOpponentRating := 0
	for _, rating := range opponentRatings {
		avgOpponentRating += rating
	}
	avgOpponentRating /= len(opponentRatings)

	// Calcul Elo
	expectedScore := 1.0 / (1.0 + math.Pow(config.DefaultEloBase, float64(avgOpponentRating-config.DefaultPvPRating)/config.DefaultEloDivisor))

	actualScore := 0.0
	if isWinner {
		actualScore = 1.0
	}

	change := k * (actualScore - expectedScore)
	return int(change)
}

func (s *CombatService) checkWinConditions(participants []*models.CombatParticipant) *int {
	// Compter les joueurs vivants par équipe
	teamAlive := make(map[int]int)

	for _, p := range participants {
		if p.IsAlive {
			teamAlive[p.Team]++
		}
	}

	// Si une seule équipe a des survivants, elle gagne
	aliveTeams := 0
	var winner int
	for team, count := range teamAlive {
		if count > 0 {
			aliveTeams++
			winner = team
		}
	}

	if aliveTeams == 1 {
		return &winner
	}

	return nil
}

func (s *CombatService) combatToSummary(combat *models.CombatInstance) *models.CombatListItem {
	summary := &models.CombatListItem{ // <- Changer le type
		ID:              combat.ID,
		CombatType:      combat.CombatType,
		Status:          combat.Status,
		MaxParticipants: combat.MaxParticipants,
		Duration:        combat.GetDuration(),
		CreatedAt:       combat.CreatedAt,
		StartedAt:       combat.StartedAt,
		EndedAt:         combat.EndedAt,
		ZoneID:          combat.ZoneID,
	}

	// Compter les participants
	if combat.Participants != nil {
		summary.ParticipantCount = len(combat.Participants)
	}

	return summary
}

func (s *CombatService) calculateHistorySummary(history []*models.CombatHistoryEntry) *models.HistorySummary {
	summary := &models.HistorySummary{
		TotalCombats: len(history),
	}

	if len(history) == 0 {
		return summary
	}

	var totalDuration int64
	var totalDamage int64
	var totalHealing int64

	for _, entry := range history {
		switch entry.Result {
		case "win":
			summary.Wins++
		case "loss":
			summary.Losses++
		case "draw":
			summary.Draws++
		}

		totalDuration += int64(entry.Duration.Seconds())
		totalDamage += int64(entry.DamageDealt)
		totalHealing += int64(entry.HealingDone)
	}

	// Calculer les moyennes
	if summary.TotalCombats > 0 {
		summary.WinRate = float64(summary.Wins) / float64(summary.TotalCombats) * config.DefaultPercentageMultiplier
		summary.AverageDuration = float64(totalDuration) / float64(summary.TotalCombats)
	}

	summary.TotalDamage = totalDamage
	summary.TotalHealing = totalHealing

	// TODO: Calculer les streaks
	summary.BestStreak = 0
	summary.CurrentStreak = 0

	return summary
}

func (s *CombatService) formatStatisticsResponse(
	stats *models.CombatStatistics,
	req *models.GetStatisticsRequest,
) *models.StatisticsResponse {
	response := &models.StatisticsResponse{
		CharacterID: stats.CharacterID,
		UserID:      stats.UserID,
		Period:      req.Period,
		GeneratedAt: time.Now(),
	}

	// Statistiques générales
	totalCombats := stats.PvEBattlesWon + stats.PvEBattlesLost + stats.PvPBattlesWon + stats.PvPBattlesLost + stats.PvPDraws
	response.General = &models.GeneralStats{
		TotalCombats:       totalCombats,
		TotalDamageDealt:   stats.TotalDamageDealt,
		TotalDamageTaken:   stats.TotalDamageTaken,
		TotalHealingDone:   stats.TotalHealingDone,
		TotalDeaths:        stats.TotalDeaths,
		HighestDamageDealt: stats.HighestDamageDealt,
		LongestCombat:      time.Duration(stats.LongestCombatDuration) * time.Second,
	}

	if totalCombats > 0 {
		response.General.AverageDuration = time.Duration(stats.LongestCombatDuration/totalCombats) * time.Second
	}

	// Statistiques PvE
	pveBattles := stats.PvEBattlesWon + stats.PvEBattlesLost
	response.PvE = &models.PvEStats{
		BattlesWon:     stats.PvEBattlesWon,
		BattlesLost:    stats.PvEBattlesLost,
		MonstersKilled: stats.MonstersKilled,
		BossesKilled:   stats.BossesKilled,
	}

	if pveBattles > 0 {
		response.PvE.WinRate = float64(stats.PvEBattlesWon) / float64(pveBattles) * config.DefaultPercentageMultiplier
	}

	// Statistiques PvP
	pvpBattles := stats.PvPBattlesWon + stats.PvPBattlesLost + stats.PvPDraws
	response.PvP = &models.PvPStats{
		BattlesWon:    stats.PvPBattlesWon,
		BattlesLost:   stats.PvPBattlesLost,
		Draws:         stats.PvPDraws,
		CurrentRating: stats.PvPRating,
		HighestRating: stats.PvPRating, // Devrait être stocké séparément
		RankName:      models.GetRankFromRating(stats.PvPRating),
	}

	if pvpBattles > 0 {
		response.PvP.WinRate = float64(stats.PvPBattlesWon) / float64(pvpBattles) * config.DefaultPercentageMultiplier
	}

	// Statistiques détaillées si demandées
	if req.Detailed {
		response.Combat = &models.DetailedCombatStats{
			// TODO: Implémenter les statistiques détaillées
		}

		response.Trends = &models.StatsTrends{
			WinRateTrend:     "stable",
			DamageTrend:      "stable",
			PerformanceTrend: "stable",
			ActivityTrend:    "stable",
			ImprovementScore: config.DefaultImprovementScore,
		}

		// TODO: Charger les achievements
		response.Achievements = []*models.Achievement{}
	}

	return response
}

// StartCombatCleanupRoutine démarre la routine de nettoyage des combats
func (s *CombatService) StartCombatCleanupRoutine() {
	ticker := time.NewTicker(s.config.Combat.CleanupInterval)
	go func() {
		for range ticker.C {
			if err := s.CleanupExpiredCombats(); err != nil {
				logrus.WithError(err).Error("Failed to cleanup expired combats")
			}
		}
	}()
}

// IsParticipantInCombat vérifie si un personnage participate à un combat actif
func (s *CombatService) IsParticipantInCombat(characterID uuid.UUID) (bool, *models.CombatInstance, error) {
	combats, err := s.combatRepo.GetByParticipant(characterID)
	if err != nil {
		return false, nil, err
	}

	for _, combat := range combats {
		if combat.IsActive() {
			return true, combat, nil
		}
	}

	return false, nil, nil
}

// GetCombatMetrics récupère les métriques du service combat
func (s *CombatService) GetCombatMetrics() (*models.CombatMetrics, error) {
	activeCount, err := s.GetActiveCombatCount()
	if err != nil {
		return nil, err
	}

	// TODO: Implémenter d'autres métriques
	metrics := &models.CombatMetrics{
		ActiveCombats:    activeCount,
		TotalCombats:     0, // À calculer
		ActionsPerSecond: 0, // À calculer
	}

	return metrics, nil
}
