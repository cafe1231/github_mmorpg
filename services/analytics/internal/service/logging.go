package service

import (
	"context"
	"encoding/json"

	"github.com/dan-2/github_mmorpg/services/analytics/internal/models"
	"github.com/dan-2/github_mmorpg/services/analytics/internal/repository"
)

const (
	EmptyContextJSON = "{}"
)

// loggingService implémente LoggingService
type loggingService struct {
	logRepo repository.LogRepository
}

// NewLoggingService crée une nouvelle instance de LoggingService
func NewLoggingService(logRepo repository.LogRepository) LoggingService {
	return &loggingService{
		logRepo: logRepo,
	}
}

// Log crée une nouvelle entrée de log
func (s *loggingService) Log(ctx context.Context, level, message string, context map[string]interface{}) error {
	contextJSON := EmptyContextJSON
	if context != nil {
		if contextBytes, err := json.Marshal(context); err == nil {
			contextJSON = string(contextBytes)
		}
	}

	logEntry := &models.LogEntry{
		Level:   level,
		Message: message,
		Context: contextJSON,
	}

	return s.logRepo.Create(ctx, logEntry)
}

// GetLogs récupère les logs avec filtres et pagination
func (s *loggingService) GetLogs(ctx context.Context, req *models.GetLogsRequest) ([]*models.LogEntry, int, error) {
	return s.logRepo.List(ctx, req.Level, req.From, req.To, req.Context, req.Page, req.Limit)
}
