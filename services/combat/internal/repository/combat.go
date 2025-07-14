// internal/repository/combat.go
package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"combat/internal/database"
	"combat/internal/models"
)

// CombatRepositoryInterface définit les méthodes du repository combat
type CombatRepositoryInterface interface {
	// Sessions de combat
	CreateSession(session *models.CombatSession) error
	GetSessionByID(id uuid.UUID) (*models.CombatSession, error)
	GetActiveSessions() ([]*models.CombatSession, error)
	GetSessionsByZone(zoneID string) ([]*models.CombatSession, error)
	UpdateSession(session *models.CombatSession) error
	EndSession(sessionID uuid.UUID) error
	
	// Participants
	AddParticipant(participant *models.CombatParticipant) error
	GetSessionParticipants(sessionID uuid.UUID) ([]*models.CombatParticipant, error)
	GetParticipantByCharacter(sessionID uuid.UUID, characterID uuid.UUID) (*models.CombatParticipant, error)
	UpdateParticipant(participant *models.CombatParticipant) error
	RemoveParticipant(sessionID uuid.UUID, characterID uuid.UUID) error
	
	// Actions de combat
	RecordAction(action *models.CombatAction) error
	GetSessionActions(sessionID uuid.UUID, limit int) ([]*models.CombatAction, error)
	GetRecentActions(sessionID uuid.UUID, since time.Time) ([]*models.CombatAction, error)
	
	// Statistiques
	GetCombatStatistics() (*models.CombatStatistics, error)
	CleanupInactiveSessions() (int, error)
}

// CombatRepository implémente l'interface CombatRepositoryInterface
type CombatRepository struct {
	db *database.DB
}

// NewCombatRepository crée une nouvelle instance du repository combat
func NewCombatRepository(db *database.DB) CombatRepositoryInterface {
	return &CombatRepository{db: db}
}

// CreateSession crée une nouvelle session de combat
func (r *CombatRepository) CreateSession(session *models.CombatSession) error {
	// Sérialiser les données JSON
	levelRangeJSON, err := json.Marshal(session.LevelRange)
	if err != nil {
		return fmt.Errorf("failed to marshal level range: %w", err)
	}

	rulesJSON, err := json.Marshal(session.Rules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	query := `
		INSERT INTO combat_sessions (
			id, type, status, zone_id, created_by, max_participants,
			is_private, level_range, rules, last_action_at, created_at, updated_at
		) VALUES (
			:id, :type, :status, :zone_id, :created_by, :max_participants,
			:is_private, :level_range, :rules, :last_action_at, :created_at, :updated_at
		)`

	data := map[string]interface{}{
		"id":               session.ID,
		"type":             session.Type,
		"status":           session.Status,
		"zone_id":          session.ZoneID,
		"created_by":       session.CreatedBy,
		"max_participants": session.MaxParticipants,
		"is_private":       session.IsPrivate,
		"level_range":      levelRangeJSON,
		"rules":            rulesJSON,
		"last_action_at":   session.LastActionAt,
		"created_at":       session.CreatedAt,
		"updated_at":       session.UpdatedAt,
	}

	_, err = r.db.NamedExec(query, data)
	if err != nil {
		return fmt.Errorf("failed to create combat session: %w", err)
	}

	return nil
}

// GetSessionByID récupère une session par son ID
func (r *CombatRepository) GetSessionByID(id uuid.UUID) (*models.CombatSession, error) {
	var session models.CombatSession
	var levelRangeJSON, rulesJSON []byte

	query := `
		SELECT id, type, status, zone_id, created_by, max_participants,
		       is_private, level_range, rules, started_at, ended_at,
		       last_action_at, created_at, updated_at
		FROM combat_sessions 
		WHERE id = $1`

	row := r.db.QueryRow(query, id)
	err := row.Scan(
		&session.ID, &session.Type, &session.Status, &session.ZoneID,
		&session.CreatedBy, &session.MaxParticipants, &session.IsPrivate,
		&levelRangeJSON, &rulesJSON, &session.StartedAt, &session.EndedAt,
		&session.LastActionAt, &session.CreatedAt, &session.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("combat session not found")
		}
		return nil, fmt.Errorf("failed to get combat session: %w", err)
	}

	// Désérialiser les données JSON
	if err := json.Unmarshal(levelRangeJSON, &session.LevelRange); err != nil {
		return nil, fmt.Errorf("failed to unmarshal level range: %w", err)
	}

	if err := json.Unmarshal(rulesJSON, &session.Rules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
	}

	return &session, nil
}

// GetActiveSessions récupère toutes les sessions actives
func (r *CombatRepository) GetActiveSessions() ([]*models.CombatSession, error) {
	var sessions []*models.CombatSession

	query := `
		SELECT id, type, status, zone_id, created_by, max_participants,
		       is_private, level_range, rules, started_at, ended_at,
		       last_action_at, created_at, updated_at
		FROM combat_sessions 
		WHERE status IN ('waiting', 'active')
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var session models.CombatSession
		var levelRangeJSON, rulesJSON []byte

		err := rows.Scan(
			&session.ID, &session.Type, &session.Status, &session.ZoneID,
			&session.CreatedBy, &session.MaxParticipants, &session.IsPrivate,
			&levelRangeJSON, &rulesJSON, &session.StartedAt, &session.EndedAt,
			&session.LastActionAt, &session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		// Désérialiser les données JSON
		if err := json.Unmarshal(levelRangeJSON, &session.LevelRange); err != nil {
			return nil, fmt.Errorf("failed to unmarshal level range: %w", err)
		}

		if err := json.Unmarshal(rulesJSON, &session.Rules); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
		}

		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// GetSessionsByZone récupère les sessions d'une zone
func (r *CombatRepository) GetSessionsByZone(zoneID string) ([]*models.CombatSession, error) {
	var sessions []*models.CombatSession

	query := `
		SELECT id, type, status, zone_id, created_by, max_participants,
		       is_private, level_range, rules, started_at, ended_at,
		       last_action_at, created_at, updated_at
		FROM combat_sessions 
		WHERE zone_id = $1 AND status IN ('waiting', 'active')
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by zone: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var session models.CombatSession
		var levelRangeJSON, rulesJSON []byte

		err := rows.Scan(
			&session.ID, &session.Type, &session.Status, &session.ZoneID,
			&session.CreatedBy, &session.MaxParticipants, &session.IsPrivate,
			&levelRangeJSON, &rulesJSON, &session.StartedAt, &session.EndedAt,
			&session.LastActionAt, &session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		// Désérialiser les données JSON
		if err := json.Unmarshal(levelRangeJSON, &session.LevelRange); err != nil {
			return nil, fmt.Errorf("failed to unmarshal level range: %w", err)
		}

		if err := json.Unmarshal(rulesJSON, &session.Rules); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
		}

		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// UpdateSession met à jour une session de combat
func (r *CombatRepository) UpdateSession(session *models.CombatSession) error {
	rulesJSON, err := json.Marshal(session.Rules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	query := `
		UPDATE combat_sessions 
		SET status = $1, rules = $2, started_at = $3, ended_at = $4,
		    last_action_at = $5, updated_at = NOW()
		WHERE id = $6`

	_, err = r.db.Exec(query, session.Status, rulesJSON, session.StartedAt,
		session.EndedAt, session.LastActionAt, session.ID)
	if err != nil {
		return fmt.Errorf("failed to update combat session: %w", err)
	}

	return nil
}

// EndSession termine une session de combat
func (r *CombatRepository) EndSession(sessionID uuid.UUID) error {
	query := `
		UPDATE combat_sessions 
		SET status = 'ended', ended_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status != 'ended'`

	result, err := r.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to end combat session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("combat session not found or already ended")
	}

	return nil
}

// AddParticipant ajoute un participant à un combat
func (r *CombatRepository) AddParticipant(participant *models.CombatParticipant) error {
	positionJSON, err := json.Marshal(participant.Position)
	if err != nil {
		return fmt.Errorf("failed to marshal position: %w", err)
	}

	query := `
		INSERT INTO combat_participants (
			id, session_id, character_id, player_id, team, position,
			status, current_health, max_health, current_mana, max_mana,
			damage, defense, crit_chance, attack_speed, joined_at
		) VALUES (
			:id, :session_id, :character_id, :player_id, :team, :position,
			:status, :current_health, :max_health, :current_mana, :max_mana,
			:damage, :defense, :crit_chance, :attack_speed, :joined_at
		)`

	data := map[string]interface{}{
		"id":             participant.ID,
		"session_id":     participant.SessionID,
		"character_id":   participant.CharacterID,
		"player_id":      participant.PlayerID,
		"team":           participant.Team,
		"position":       positionJSON,
		"status":         participant.Status,
		"current_health": participant.CurrentHealth,
		"max_health":     participant.MaxHealth,
		"current_mana":   participant.CurrentMana,
		"max_mana":       participant.MaxMana,
		"damage":         participant.Damage,
		"defense":        participant.Defense,
		"crit_chance":    participant.CritChance,
		"attack_speed":   participant.AttackSpeed,
		"joined_at":      participant.JoinedAt,
	}

	_, err = r.db.NamedExec(query, data)
	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}

// GetSessionParticipants récupère tous les participants d'une session
func (r *CombatRepository) GetSessionParticipants(sessionID uuid.UUID) ([]*models.CombatParticipant, error) {
	var participants []*models.CombatParticipant

	query := `
		SELECT id, session_id, character_id, player_id, team, position,
		       status, current_health, max_health, current_mana, max_mana,
		       damage, defense, crit_chance, attack_speed, joined_at, last_action_at
		FROM combat_participants 
		WHERE session_id = $1
		ORDER BY team, joined_at`

	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session participants: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var participant models.CombatParticipant
		var positionJSON []byte

		err := rows.Scan(
			&participant.ID, &participant.SessionID, &participant.CharacterID,
			&participant.PlayerID, &participant.Team, &positionJSON,
			&participant.Status, &participant.CurrentHealth, &participant.MaxHealth,
			&participant.CurrentMana, &participant.MaxMana, &participant.Damage,
			&participant.Defense, &participant.CritChance, &participant.AttackSpeed,
			&participant.JoinedAt, &participant.LastActionAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}

		if err := json.Unmarshal(positionJSON, &participant.Position); err != nil {
			return nil, fmt.Errorf("failed to unmarshal position: %w", err)
		}

		participants = append(participants, &participant)
	}

	return participants, nil
}

// GetParticipantByCharacter récupère un participant par son character ID
func (r *CombatRepository) GetParticipantByCharacter(sessionID uuid.UUID, characterID uuid.UUID) (*models.CombatParticipant, error) {
	var participant models.CombatParticipant
	var positionJSON []byte

	query := `
		SELECT id, session_id, character_id, player_id, team, position,
		       status, current_health, max_health, current_mana, max_mana,
		       damage, defense, crit_chance, attack_speed, joined_at, last_action_at
		FROM combat_participants 
		WHERE session_id = $1 AND character_id = $2`

	err := r.db.QueryRow(query, sessionID, characterID).Scan(
		&participant.ID, &participant.SessionID, &participant.CharacterID,
		&participant.PlayerID, &participant.Team, &positionJSON,
		&participant.Status, &participant.CurrentHealth, &participant.MaxHealth,
		&participant.CurrentMana, &participant.MaxMana, &participant.Damage,
		&participant.Defense, &participant.CritChance, &participant.AttackSpeed,
		&participant.JoinedAt, &participant.LastActionAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("participant not found")
		}
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}

	if err := json.Unmarshal(positionJSON, &participant.Position); err != nil {
		return nil, fmt.Errorf("failed to unmarshal position: %w", err)
	}

	return &participant, nil
}

// UpdateParticipant met à jour un participant
func (r *CombatRepository) UpdateParticipant(participant *models.CombatParticipant) error {
	positionJSON, err := json.Marshal(participant.Position)
	if err != nil {
		return fmt.Errorf("failed to marshal position: %w", err)
	}

	query := `
		UPDATE combat_participants 
		SET team = $1, position = $2, status = $3, current_health = $4,
		    max_health = $5, current_mana = $6, max_mana = $7,
		    damage = $8, defense = $9, crit_chance = $10, attack_speed = $11,
		    last_action_at = $12
		WHERE id = $13`

	_, err = r.db.Exec(query,
		participant.Team, positionJSON, participant.Status,
		participant.CurrentHealth, participant.MaxHealth,
		participant.CurrentMana, participant.MaxMana,
		participant.Damage, participant.Defense,
		participant.CritChance, participant.AttackSpeed,
		participant.LastActionAt, participant.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update participant: %w", err)
	}

	return nil
}

// RemoveParticipant supprime un participant d'une session
func (r *CombatRepository) RemoveParticipant(sessionID uuid.UUID, characterID uuid.UUID) error {
	query := `DELETE FROM combat_participants WHERE session_id = $1 AND character_id = $2`

	result, err := r.db.Exec(query, sessionID, characterID)
	if err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}

// RecordAction enregistre une action de combat
func (r *CombatRepository) RecordAction(action *models.CombatAction) error {
	actionDataJSON, err := json.Marshal(action.ActionData)
	if err != nil {
		return fmt.Errorf("failed to marshal action data: %w", err)
	}

	resultsJSON, err := json.Marshal(action.Results)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	query := `
		INSERT INTO combat_actions (
			id, session_id, actor_id, type, action_data, targets,
			results, success, critical_hit, executed_at, duration
		) VALUES (
			:id, :session_id, :actor_id, :type, :action_data, :targets,
			:results, :success, :critical_hit, :executed_at, :duration
		)`

	data := map[string]interface{}{
		"id":           action.ID,
		"session_id":   action.SessionID,
		"actor_id":     action.ActorID,
		"type":         action.Type,
		"action_data":  actionDataJSON,
		"targets":      pq.Array(action.Targets),
		"results":      resultsJSON,
		"success":      action.Success,
		"critical_hit": action.CriticalHit,
		"executed_at":  action.ExecutedAt,
		"duration":     action.Duration,
	}

	_, err = r.db.NamedExec(query, data)
	if err != nil {
		return fmt.Errorf("failed to record action: %w", err)
	}

	return nil
}

// GetSessionActions récupère les actions d'une session
func (r *CombatRepository) GetSessionActions(sessionID uuid.UUID, limit int) ([]*models.CombatAction, error) {
	var actions []*models.CombatAction

	query := `
		SELECT id, session_id, actor_id, type, action_data, targets,
		       results, success, critical_hit, executed_at, duration
		FROM combat_actions 
		WHERE session_id = $1
		ORDER BY executed_at DESC
		LIMIT $2`

	rows, err := r.db.Query(query, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get session actions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var action models.CombatAction
		var actionDataJSON, resultsJSON []byte

		err := rows.Scan(
			&action.ID, &action.SessionID, &action.ActorID, &action.Type,
			&actionDataJSON, pq.Array(&action.Targets), &resultsJSON,
			&action.Success, &action.CriticalHit, &action.ExecutedAt, &action.Duration,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan action: %w", err)
		}

		if err := json.Unmarshal(actionDataJSON, &action.ActionData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal action data: %w", err)
		}

		if err := json.Unmarshal(resultsJSON, &action.Results); err != nil {
			return nil, fmt.Errorf("failed to unmarshal results: %w", err)
		}

		actions = append(actions, &action)
	}

	return actions, nil
}

// GetRecentActions récupère les actions récentes depuis une date
func (r *CombatRepository) GetRecentActions(sessionID uuid.UUID, since time.Time) ([]*models.CombatAction, error) {
	var actions []*models.CombatAction

	query := `
		SELECT id, session_id, actor_id, type, action_data, targets,
		       results, success, critical_hit, executed_at, duration
		FROM combat_actions 
		WHERE session_id = $1 AND executed_at > $2
		ORDER BY executed_at ASC`

	rows, err := r.db.Query(query, sessionID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent actions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var action models.CombatAction
		var actionDataJSON, resultsJSON []byte

		err := rows.Scan(
			&action.ID, &action.SessionID, &action.ActorID, &action.Type,
			&actionDataJSON, pq.Array(&action.Targets), &resultsJSON,
			&action.Success, &action.CriticalHit, &action.ExecutedAt, &action.Duration,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan action: %w", err)
		}

		if err := json.Unmarshal(actionDataJSON, &action.ActionData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal action data: %w", err)
		}

		if err := json.Unmarshal(resultsJSON, &action.Results); err != nil {
			return nil, fmt.Errorf("failed to unmarshal results: %w", err)
		}

		actions = append(actions, &action)
	}

	return actions, nil
}

// GetCombatStatistics récupère les statistiques de combat
func (r *CombatRepository) GetCombatStatistics() (*models.CombatStatistics, error) {
	stats := &models.CombatStatistics{}

	// Compter les combats totaux et actifs
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status IN ('waiting', 'active') THEN 1 END) as active
		FROM combat_sessions`

	err := r.db.QueryRow(query).Scan(&stats.TotalCombats, &stats.ActiveCombats)
	if err != nil {
		return nil, fmt.Errorf("failed to get combat counts: %w", err)
	}

	// Calculer la durée moyenne des sessions terminées
	query = `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (ended_at - started_at))), 0)
		FROM combat_sessions 
		WHERE status = 'ended' AND started_at IS NOT NULL AND ended_at IS NOT NULL`

	var avgSeconds float64
	err = r.db.QueryRow(query).Scan(&avgSeconds)
	if err != nil {
		return nil, fmt.Errorf("failed to get average session time: %w", err)
	}
	stats.AverageSessionTime = time.Duration(avgSeconds) * time.Second

	// Compter les matches PvP et encounters PvE
	query = `
		SELECT 
			COUNT(CASE WHEN type = 'pvp' THEN 1 END) as pvp,
			COUNT(CASE WHEN type = 'pve' THEN 1 END) as pve
		FROM combat_sessions`

	err = r.db.QueryRow(query).Scan(&stats.PvPMatches, &stats.PvEEncounters)
	if err != nil {
		return nil, fmt.Errorf("failed to get match type counts: %w", err)
	}

	return stats, nil
}

// CleanupInactiveSessions nettoie les sessions inactives
func (r *CombatRepository) CleanupInactiveSessions() (int, error) {
	// Utiliser la fonction PostgreSQL définie dans les migrations
	var cleanupCount int
	err := r.db.QueryRow("SELECT cleanup_inactive_sessions()").Scan(&cleanupCount)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup inactive sessions: %w", err)
	}

	return cleanupCount, nil
}