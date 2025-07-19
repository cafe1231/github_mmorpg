package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/google/uuid"

	"combat/internal/config"
	"combat/internal/database"
	"combat/internal/models"
)

// CombatRepositoryInterface définit les méthodes du repository combat
type CombatRepositoryInterface interface {
	// CRUD de base
	Create(combat *models.CombatInstance) error
	GetByID(id uuid.UUID) (*models.CombatInstance, error)
	Update(combat *models.CombatInstance) error
	Delete(id uuid.UUID) error

	// Recherche et filtrage
	List(filters *models.SearchCombatsRequest) ([]*models.CombatInstance, int, error)
	GetByStatus(status models.CombatStatus) ([]*models.CombatInstance, error)
	GetByParticipant(participantID uuid.UUID) ([]*models.CombatInstance, error)
	GetByZone(zoneID string) ([]*models.CombatInstance, error)

	// Gestion des participants
	AddParticipant(participant *models.CombatParticipant) error
	RemoveParticipant(combatID, participantID uuid.UUID) error
	GetParticipants(combatID uuid.UUID) ([]*models.CombatParticipant, error)
	GetParticipant(combatID, characterID uuid.UUID) (*models.CombatParticipant, error)
	UpdateParticipant(participant *models.CombatParticipant) error

	// Statistiques et métriques
	GetStatistics(characterID uuid.UUID) (*models.CombatStatistics, error)
	UpdateStatistics(stats *models.CombatStatistics) error
	GetActiveCombatCount() (int, error)
	GetCombatHistory(req *models.GetCombatHistoryRequest) ([]*models.CombatHistoryEntry, int, error)

	// Nettoyage et maintenance
	CleanupExpiredCombats() error
	GetExpiredCombats() ([]*models.CombatInstance, error)
}

// CombatRepository implémente l'interface CombatRepositoryInterface
type CombatRepository struct {
	db *database.DB
}

// NewCombatRepository crée une nouvelle instance du repository combat
func NewCombatRepository(db *database.DB) CombatRepositoryInterface {
	return &CombatRepository{db: db}
}

// Create crée une nouvelle instance de combat
func (r *CombatRepository) Create(combat *models.CombatInstance) error {
	// Sérialiser les paramètres
	settingsJSON, err := json.Marshal(combat.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		INSERT INTO combat_instances (
			id, combat_type, status, zone_id, max_participants,
			current_turn, turn_time_limit, max_duration, settings,
			created_at, updated_at
		) VALUES (
			:id, :combat_type, :status, :zone_id, :max_participants,
			:current_turn, :turn_time_limit, :max_duration, :settings,
			:created_at, :updated_at
		)`

	data := map[string]interface{}{
		"id":               combat.ID,
		"combat_type":      combat.CombatType,
		"status":           combat.Status,
		"zone_id":          combat.ZoneID,
		"max_participants": combat.MaxParticipants,
		"current_turn":     combat.CurrentTurn,
		"turn_time_limit":  combat.TurnTimeLimit,
		"max_duration":     combat.MaxDuration,
		"settings":         settingsJSON,
		"created_at":       combat.CreatedAt,
		"updated_at":       combat.UpdatedAt,
	}

	_, err = r.db.NamedExec(query, data)
	if err != nil {
		return fmt.Errorf("failed to create combat: %w", err)
	}

	return nil
}

// GetByID récupère un combat par son ID
func (r *CombatRepository) GetByID(id uuid.UUID) (*models.CombatInstance, error) {
	var combat models.CombatInstance
	var settingsJSON []byte

	query := `
		SELECT id, combat_type, status, zone_id, max_participants,
		       current_turn, turn_time_limit, max_duration, settings,
		       created_at, started_at, ended_at, updated_at
		FROM combat_instances 
		WHERE id = $1`

	row := r.db.QueryRow(query, id)
	err := row.Scan(
		&combat.ID, &combat.CombatType, &combat.Status, &combat.ZoneID,
		&combat.MaxParticipants, &combat.CurrentTurn, &combat.TurnTimeLimit,
		&combat.MaxDuration, &settingsJSON, &combat.CreatedAt,
		&combat.StartedAt, &combat.EndedAt, &combat.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("combat not found")
		}
		return nil, fmt.Errorf("failed to get combat: %w", err)
	}

	// Désérialiser les paramètres
	if err := json.Unmarshal(settingsJSON, &combat.Settings); err != nil {
		logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	return &combat, nil
}

// Update met à jour un combat
func (r *CombatRepository) Update(combat *models.CombatInstance) error {
	// Sérialiser les paramètres
	settingsJSON, err := json.Marshal(combat.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	combat.UpdatedAt = time.Now()

	query := `
		UPDATE combat_instances SET
			status = :status,
			current_turn = :current_turn,
			settings = :settings,
			started_at = :started_at,
			ended_at = :ended_at,
			updated_at = :updated_at
		WHERE id = :id`

	data := map[string]interface{}{
		"id":           combat.ID,
		"status":       combat.Status,
		"current_turn": combat.CurrentTurn,
		"settings":     settingsJSON,
		"started_at":   combat.StartedAt,
		"ended_at":     combat.EndedAt,
		"updated_at":   combat.UpdatedAt,
	}

	result, err := r.db.NamedExec(query, data)
	if err != nil {
		return fmt.Errorf("failed to update combat: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("combat not found")
	}

	return nil
}

// Delete supprime un combat
func (r *CombatRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM combat_instances WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete combat: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("combat not found")
	}

	return nil
}

// List récupère une liste de combats avec filtres
func (r *CombatRepository) List(filters *models.SearchCombatsRequest) ([]*models.CombatInstance, int, error) {
	var combats []*models.CombatInstance
	var total int

	// Construction de la requête avec filtres
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if filters.CombatType != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND combat_type = $%d", argCount)
		args = append(args, *filters.CombatType)
	}

	if filters.Status != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *filters.Status)
	}

	if filters.ZoneID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND zone_id = $%d", argCount)
		args = append(args, *filters.ZoneID)
	}

	if filters.CreatedAfter != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *filters.CreatedBefore)
	}

	if !filters.IncludeFinished {
		argCount++
		whereClause += fmt.Sprintf(" AND status NOT IN ($%d, $%d)", argCount, argCount+1)
		args = append(args, models.CombatStatusFinished, models.CombatStatusCancelled)
		argCount++
	}

	// Compter le total
	countQuery := "SELECT COUNT(*) FROM combat_instances " + whereClause
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count combats: %w", err)
	}

	// Récupérer les données
	argCount++
	limitClause := fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argCount)
	args = append(args, filters.Limit)

	if filters.Offset > 0 {
		argCount++
		limitClause += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filters.Offset)
	}

	query := `
		SELECT id, combat_type, status, zone_id, max_participants,
		       current_turn, turn_time_limit, max_duration, settings,
		       created_at, started_at, ended_at, updated_at
		FROM combat_instances ` + whereClause + limitClause

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query combats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var combat models.CombatInstance
		var settingsJSON []byte

		err := rows.Scan(
			&combat.ID, &combat.CombatType, &combat.Status, &combat.ZoneID,
			&combat.MaxParticipants, &combat.CurrentTurn, &combat.TurnTimeLimit,
			&combat.MaxDuration, &settingsJSON, &combat.CreatedAt,
			&combat.StartedAt, &combat.EndedAt, &combat.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan combat: %w", err)
		}

		// Désérialiser les paramètres
		if err := json.Unmarshal(settingsJSON, &combat.Settings); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal settings: %w", err)
		}

		combats = append(combats, &combat)
	}

	return combats, total, nil
}

// GetByStatus récupère les combats par statut
func (r *CombatRepository) GetByStatus(status models.CombatStatus) ([]*models.CombatInstance, error) {
	filters := &models.SearchCombatsRequest{
		Status:          &status,
		Limit:           config.DefaultVarianceDivisor,
		IncludeFinished: true,
	}

	combats, _, err := r.List(filters)
	return combats, err
}

// GetByParticipant récupère les combats d'un participant
func (r *CombatRepository) GetByParticipant(participantID uuid.UUID) ([]*models.CombatInstance, error) {
	query := `
		SELECT DISTINCT ci.id, ci.combat_type, ci.status, ci.zone_id, ci.max_participants,
		       ci.current_turn, ci.turn_time_limit, ci.max_duration, ci.settings,
		       ci.created_at, ci.started_at, ci.ended_at, ci.updated_at
		FROM combat_instances ci
		JOIN combat_participants cp ON ci.id = cp.combat_id
		WHERE cp.character_id = $1
		ORDER BY ci.created_at DESC`

	rows, err := r.db.Query(query, participantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query combats by participant: %w", err)
	}
	defer rows.Close()

	var combats []*models.CombatInstance
	for rows.Next() {
		var combat models.CombatInstance
		var settingsJSON []byte

		err := rows.Scan(
			&combat.ID, &combat.CombatType, &combat.Status, &combat.ZoneID,
			&combat.MaxParticipants, &combat.CurrentTurn, &combat.TurnTimeLimit,
			&combat.MaxDuration, &settingsJSON, &combat.CreatedAt,
			&combat.StartedAt, &combat.EndedAt, &combat.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan combat: %w", err)
		}

		// Désérialiser les paramètres
		if err := json.Unmarshal(settingsJSON, &combat.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}

		combats = append(combats, &combat)
	}

	return combats, nil
}

// GetByZone récupère les combats dans une zone
func (r *CombatRepository) GetByZone(zoneID string) ([]*models.CombatInstance, error) {
	filters := &models.SearchCombatsRequest{
		ZoneID:          &zoneID,
		Limit:           config.DefaultImprovementScore,
		IncludeFinished: false,
	}

	combats, _, err := r.List(filters)
	return combats, err
}

// AddParticipant ajoute un participant à un combat
func (r *CombatRepository) AddParticipant(participant *models.CombatParticipant) error {
	query := `
		INSERT INTO combat_participants (
			id, combat_id, character_id, user_id, team, position,
			health, max_health, mana, max_mana,
			physical_damage, magical_damage, physical_defense, magical_defense,
			critical_chance, attack_speed, is_alive, is_ready,
			damage_dealt, damage_taken, healing_done,
			created_at, updated_at
		) VALUES (
			:id, :combat_id, :character_id, :user_id, :team, :position,
			:health, :max_health, :mana, :max_mana,
			:physical_damage, :magical_damage, :physical_defense, :magical_defense,
			:critical_chance, :attack_speed, :is_alive, :is_ready,
			:damage_dealt, :damage_taken, :healing_done,
			:created_at, :updated_at
		)`

	_, err := r.db.NamedExec(query, participant)
	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}

// RemoveParticipant supprime un participant d'un combat
func (r *CombatRepository) RemoveParticipant(combatID, participantID uuid.UUID) error {
	query := `DELETE FROM combat_participants WHERE combat_id = $1 AND character_id = $2`

	result, err := r.db.Exec(query, combatID, participantID)
	if err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}

// GetParticipants récupère tous les participants d'un combat
func (r *CombatRepository) GetParticipants(combatID uuid.UUID) ([]*models.CombatParticipant, error) {
	query := `
		SELECT id, combat_id, character_id, user_id, team, position,
		       health, max_health, mana, max_mana,
		       physical_damage, magical_damage, physical_defense, magical_defense,
		       critical_chance, attack_speed, is_alive, is_ready,
		       last_action_at, damage_dealt, damage_taken, healing_done,
		       created_at, updated_at
		FROM combat_participants 
		WHERE combat_id = $1 
		ORDER BY team, position`

	var participants []*models.CombatParticipant
	err := r.db.Select(&participants, query, combatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}

	return participants, nil
}

// GetParticipant récupère un participant spécifique
func (r *CombatRepository) GetParticipant(combatID, characterID uuid.UUID) (*models.CombatParticipant, error) {
	var participant models.CombatParticipant

	query := `
		SELECT id, combat_id, character_id, user_id, team, position,
		       health, max_health, mana, max_mana,
		       physical_damage, magical_damage, physical_defense, magical_defense,
		       critical_chance, attack_speed, is_alive, is_ready,
		       last_action_at, damage_dealt, damage_taken, healing_done,
		       created_at, updated_at
		FROM combat_participants 
		WHERE combat_id = $1 AND character_id = $2`

	err := r.db.Get(&participant, query, combatID, characterID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("participant not found")
		}
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}

	return &participant, nil
}

// UpdateParticipant met à jour un participant
func (r *CombatRepository) UpdateParticipant(participant *models.CombatParticipant) error {
	participant.UpdatedAt = time.Now()

	query := `
		UPDATE combat_participants SET
			health = :health,
			mana = :mana,
			is_alive = :is_alive,
			is_ready = :is_ready,
			last_action_at = :last_action_at,
			damage_dealt = :damage_dealt,
			damage_taken = :damage_taken,
			healing_done = :healing_done,
			updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExec(query, participant)
	if err != nil {
		return fmt.Errorf("failed to update participant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}

// GetStatistics récupère les statistiques de combat d'un personnage
func (r *CombatRepository) GetStatistics(characterID uuid.UUID) (*models.CombatStatistics, error) {
	var stats models.CombatStatistics

	query := `
		SELECT id, character_id, user_id,
		       pve_battles_won, pve_battles_lost, monsters_killed, bosses_killed,
		       pvp_battles_won, pvp_battles_lost, pvp_draws, pvp_rating,
		       total_damage_dealt, total_damage_taken, total_healing_done, total_deaths,
		       highest_damage_dealt, longest_combat_duration,
		       created_at, updated_at
		FROM combat_statistics 
		WHERE character_id = $1`

	err := r.db.Get(&stats, query, characterID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Créer des statistiques par défaut
			return r.createDefaultStatistics(characterID)
		}
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return &stats, nil
}

// createDefaultStatistics crée des statistiques par défaut pour un personnage
func (r *CombatRepository) createDefaultStatistics(characterID uuid.UUID) (*models.CombatStatistics, error) {
	stats := &models.CombatStatistics{
		ID:          uuid.New(),
		CharacterID: characterID,
		PvPRating:   config.DefaultPvPRating, // Rating par défaut
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO combat_statistics (
			id, character_id, user_id, pvp_rating, created_at, updated_at
		) VALUES (
			:id, :character_id, :user_id, :pvp_rating, :created_at, :updated_at
		)`

	_, err := r.db.NamedExec(query, stats)
	if err != nil {
		return nil, fmt.Errorf("failed to create default statistics: %w", err)
	}

	return stats, nil
}

// UpdateStatistics met à jour les statistiques de combat
func (r *CombatRepository) UpdateStatistics(stats *models.CombatStatistics) error {
	stats.UpdatedAt = time.Now()

	query := `
		UPDATE combat_statistics SET
			pve_battles_won = :pve_battles_won,
			pve_battles_lost = :pve_battles_lost,
			monsters_killed = :monsters_killed,
			bosses_killed = :bosses_killed,
			pvp_battles_won = :pvp_battles_won,
			pvp_battles_lost = :pvp_battles_lost,
			pvp_draws = :pvp_draws,
			pvp_rating = :pvp_rating,
			total_damage_dealt = :total_damage_dealt,
			total_damage_taken = :total_damage_taken,
			total_healing_done = :total_healing_done,
			total_deaths = :total_deaths,
			highest_damage_dealt = :highest_damage_dealt,
			longest_combat_duration = :longest_combat_duration,
			updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExec(query, stats)
	if err != nil {
		return fmt.Errorf("failed to update statistics: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("statistics not found")
	}

	return nil
}

// GetActiveCombatCount retourne le nombre de combats actifs
func (r *CombatRepository) GetActiveCombatCount() (int, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM combat_instances 
		WHERE status IN ($1, $2)`

	err := r.db.Get(&count, query, models.CombatStatusWaiting, models.CombatStatusActive)
	if err != nil {
		return 0, fmt.Errorf("failed to get active combat count: %w", err)
	}

	return count, nil
}

// GetCombatHistory récupère l'historique de combat
func (r *CombatRepository) GetCombatHistory(req *models.GetCombatHistoryRequest) ([]*models.CombatHistoryEntry, int, error) {
	// Construction de la requête avec filtres
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if req.CharacterID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND cp.character_id = $%d", argCount)
		args = append(args, *req.CharacterID)
	}

	if req.UserID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND cp.user_id = $%d", argCount)
		args = append(args, *req.UserID)
	}

	if req.CombatType != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND ci.combat_type = $%d", argCount)
		args = append(args, *req.CombatType)
	}

	if req.DateFrom != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND ci.created_at >= $%d", argCount)
		args = append(args, *req.DateFrom)
	}

	if req.DateTo != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND ci.created_at <= $%d", argCount)
		args = append(args, *req.DateTo)
	}

	// Filtre sur les victories/défaites
	if req.WinsOnly || req.LossesOnly {
		// Cette logique nécessiterait une table de résultats ou une logique plus complexe
		// Pour l'instant, on l'ignore
		logrus.Debug("WinsOnly/LossesOnly filtering not yet implemented")
	}

	// Compter le total
	countQuery := `
		SELECT COUNT(DISTINCT ci.id)
		FROM combat_instances ci
		JOIN combat_participants cp ON ci.id = cp.combat_id ` + whereClause

	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count combat history: %w", err)
	}

	// Récupérer les données
	argCount++
	limitClause := fmt.Sprintf(" ORDER BY ci.created_at DESC LIMIT $%d", argCount)
	args = append(args, req.Limit)

	if req.Offset > 0 {
		argCount++
		limitClause += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, req.Offset)
	}

	query := `
		SELECT DISTINCT ci.id, ci.combat_type, ci.created_at, ci.ended_at,
		       cp.damage_dealt, cp.damage_taken, cp.healing_done
		FROM combat_instances ci
		JOIN combat_participants cp ON ci.id = cp.combat_id ` + whereClause + limitClause

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query combat history: %w", err)
	}
	defer rows.Close()

	var history []*models.CombatHistoryEntry
	for rows.Next() {
		var entry models.CombatHistoryEntry
		var endedAt sql.NullTime

		err := rows.Scan(
			&entry.CombatID, &entry.CombatType, &entry.CreatedAt, &endedAt,
			&entry.DamageDealt, &entry.DamageTaken, &entry.HealingDone,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan combat history: %w", err)
		}

		// Calculer la durée
		if endedAt.Valid {
			entry.Duration = endedAt.Time.Sub(entry.CreatedAt)
		}

		// TODO: Déterminer le résultat (victoire/défaite/égalité)
		entry.Result = "unknown"

		history = append(history, &entry)
	}

	return history, total, nil
}

// CleanupExpiredCombats nettoie les combats expirés
func (r *CombatRepository) CleanupExpiredCombats() error {
	query := `
		UPDATE combat_instances 
		SET status = $1, ended_at = $2, updated_at = $2
		WHERE status IN ($3, $4) 
		AND created_at < $5`

	expiredTime := time.Now().Add(-24 * time.Hour) // Expire après 24h
	now := time.Now()

	_, err := r.db.Exec(query,
		models.CombatStatusCancelled, now,
		models.CombatStatusWaiting, models.CombatStatusActive,
		expiredTime)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired combats: %w", err)
	}

	return nil
}

// GetExpiredCombats récupère les combats expirés
func (r *CombatRepository) GetExpiredCombats() ([]*models.CombatInstance, error) {
	expiredTime := time.Now().Add(-24 * time.Hour)

	filters := &models.SearchCombatsRequest{
		CreatedBefore:   &expiredTime,
		Limit:           config.DefaultVarianceDivisor,
		IncludeFinished: false,
	}

	combats, _, err := r.List(filters)
	return combats, err
}
