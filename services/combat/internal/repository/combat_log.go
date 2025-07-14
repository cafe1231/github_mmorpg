// internal/repository/combat_log.go
package repository

import (
"encoding/json"
"fmt"
"time"

"github.com/google/uuid"

"combat/internal/database"
"combat/internal/models"
)

// CombatLogRepositoryInterface définit les méthodes du repository combat log
type CombatLogRepositoryInterface interface {
CreateLog(log *models.CombatLog) error
GetSessionLogs(sessionID uuid.UUID, limit int) ([]*models.CombatLog, error)
GetCharacterLogs(characterID uuid.UUID, limit int) ([]*models.CombatLog, error)
GetLogsByType(sessionID uuid.UUID, eventType string) ([]*models.CombatLog, error)
DeleteOldLogs(olderThan time.Time) (int, error)
CreateCombatSummary(sessionID uuid.UUID) error
}

// CombatLogRepository implémente l'interface CombatLogRepositoryInterface
type CombatLogRepository struct {
db *database.DB
}

// NewCombatLogRepository crée une nouvelle instance du repository combat log
func NewCombatLogRepository(db *database.DB) CombatLogRepositoryInterface {
return &CombatLogRepository{db: db}
}

// CreateLog crée un nouveau log de combat
func (r *CombatLogRepository) CreateLog(log *models.CombatLog) error {
rawDataJSON, err := json.Marshal(log.RawData)
if err != nil {
return fmt.Errorf("failed to marshal raw data: %w", err)
}

query := `INSERT INTO combat_logs (
id, session_id, action_id, actor_id, target_id, event_type,
message, raw_data, value, old_value, new_value,
is_critical, is_resisted, is_absorbed, color, icon, priority, timestamp
) VALUES (
, , , , , , , , , , , , , , , , , 
)`

_, err = r.db.Exec(query,
log.ID, log.SessionID, log.ActionID, log.ActorID, log.TargetID,
log.EventType, log.Message, rawDataJSON, log.Value, log.OldValue,
log.NewValue, log.IsCritical, log.IsResisted, log.IsAbsorbed,
log.Color, log.Icon, log.Priority, log.Timestamp,
)

if err != nil {
return fmt.Errorf("failed to create combat log: %w", err)
}

return nil
}

// Implémentations basiques pour les autres méthodes
func (r *CombatLogRepository) GetSessionLogs(sessionID uuid.UUID, limit int) ([]*models.CombatLog, error) {
return []*models.CombatLog{}, nil
}

func (r *CombatLogRepository) GetCharacterLogs(characterID uuid.UUID, limit int) ([]*models.CombatLog, error) {
return []*models.CombatLog{}, nil
}

func (r *CombatLogRepository) GetLogsByType(sessionID uuid.UUID, eventType string) ([]*models.CombatLog, error) {
return []*models.CombatLog{}, nil
}

func (r *CombatLogRepository) DeleteOldLogs(olderThan time.Time) (int, error) {
return 0, nil
}

func (r *CombatLogRepository) CreateCombatSummary(sessionID uuid.UUID) error {
return nil
}
