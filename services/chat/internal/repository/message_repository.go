package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"chat/internal/models"
)

type messageRepository struct {
	db *sql.DB
}

// NewMessageRepository crée une nouvelle instance du repository des messages
func NewMessageRepository(db *sql.DB) MessageRepository {
	return &messageRepository{db: db}
}

// Create crée un nouveau message
func (r *messageRepository) Create(ctx context.Context, message *models.Message) error {
	if message.ID == uuid.Nil {
		message.ID = uuid.New()
	}

	mentionsJSON, _ := json.Marshal(message.Mentions)
	attachmentsJSON, _ := json.Marshal(message.Attachments)
	reactionsJSON, _ := json.Marshal(message.Reactions)

	query := `
		INSERT INTO messages (
			id, channel_id, user_id, content, type, mentions, attachments, reactions,
			is_edited, edited_at, is_deleted, deleted_at, is_moderated, mod_reason,
			reply_to_id, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	message.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		message.ID, message.ChannelID, message.UserID, message.Content, message.Type,
		mentionsJSON, attachmentsJSON, reactionsJSON, message.IsEdited, message.EditedAt,
		message.IsDeleted, message.DeletedAt, message.IsModerated, message.ModReason,
		message.ReplyToID, message.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"message_id": message.ID,
		"channel_id": message.ChannelID,
		"user_id":    message.UserID,
		"type":       message.Type,
	}).Info("Message created successfully")

	return nil
}

// GetByID récupère un message par son ID
func (r *messageRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	query := `
		SELECT 
			id, channel_id, user_id, content, type, mentions, attachments, reactions,
			is_edited, edited_at, is_deleted, deleted_at, is_moderated, mod_reason,
			reply_to_id, created_at
		FROM messages
		WHERE id = $1 AND is_deleted = false
	`

	var message models.Message
	var mentionsJSON, attachmentsJSON, reactionsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&message.ID, &message.ChannelID, &message.UserID, &message.Content, &message.Type,
		&mentionsJSON, &attachmentsJSON, &reactionsJSON, &message.IsEdited, &message.EditedAt,
		&message.IsDeleted, &message.DeletedAt, &message.IsModerated, &message.ModReason,
		&message.ReplyToID, &message.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	// Unmarshal JSON fields
	json.Unmarshal(mentionsJSON, &message.Mentions)
	json.Unmarshal(attachmentsJSON, &message.Attachments)
	json.Unmarshal(reactionsJSON, &message.Reactions)

	return &message, nil
}

// Update met à jour un message
func (r *messageRepository) Update(ctx context.Context, message *models.Message) error {
	query := `
		UPDATE messages SET
			content = $2, is_edited = true, edited_at = $3
		WHERE id = $1 AND is_deleted = false
	`

	message.IsEdited = true
	message.EditedAt = &[]time.Time{time.Now()}[0]

	result, err := r.db.ExecContext(ctx, query, message.ID, message.Content, message.EditedAt)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("message not found or already deleted")
	}

	return nil
}

// Delete supprime définitivement un message
func (r *messageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM messages WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}

// SoftDelete effectue une suppression douce d'un message
func (r *messageRepository) SoftDelete(ctx context.Context, id uuid.UUID, reason string) error {
	query := `
		UPDATE messages SET
			is_deleted = true, deleted_at = $2, mod_reason = $3
		WHERE id = $1 AND is_deleted = false
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, id, now, reason)
	if err != nil {
		return fmt.Errorf("failed to soft delete message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("message not found or already deleted")
	}

	return nil
}

// GetChannelMessages récupère les messages d'un channel
func (r *messageRepository) GetChannelMessages(ctx context.Context, channelID uuid.UUID, req *models.GetMessagesRequest) ([]*models.Message, error) {
	query := `
		SELECT 
			id, channel_id, user_id, content, type, mentions, attachments, reactions,
			is_edited, edited_at, is_deleted, deleted_at, is_moderated, mod_reason,
			reply_to_id, created_at
		FROM messages
		WHERE channel_id = $1 AND is_deleted = false
	`

	args := []interface{}{channelID}
	argIndex := 2

	// Ajouter des conditions selon la requête
	if req.Before != nil {
		query += fmt.Sprintf(" AND created_at < (SELECT created_at FROM messages WHERE id = $%d)", argIndex)
		args = append(args, *req.Before)
		argIndex++
	}

	if req.After != nil {
		query += fmt.Sprintf(" AND created_at > (SELECT created_at FROM messages WHERE id = $%d)", argIndex)
		args = append(args, *req.After)
		argIndex++
	}

	if req.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *req.UserID)
		argIndex++
	}

	if req.Type != nil {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, *req.Type)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, req.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel messages: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// GetUserMessages récupère les messages d'un utilisateur
func (r *messageRepository) GetUserMessages(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	query := `
		SELECT 
			id, channel_id, user_id, content, type, mentions, attachments, reactions,
			is_edited, edited_at, is_deleted, deleted_at, is_moderated, mod_reason,
			reply_to_id, created_at
		FROM messages
		WHERE user_id = $1 AND is_deleted = false
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user messages: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// GetReplies récupère les réponses à un message
func (r *messageRepository) GetReplies(ctx context.Context, messageID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	query := `
		SELECT 
			id, channel_id, user_id, content, type, mentions, attachments, reactions,
			is_edited, edited_at, is_deleted, deleted_at, is_moderated, mod_reason,
			reply_to_id, created_at
		FROM messages
		WHERE reply_to_id = $1 AND is_deleted = false
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, messageID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get replies: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// Search recherche des messages
func (r *messageRepository) Search(ctx context.Context, req *models.SearchMessagesRequest) ([]*models.Message, error) {
	query := `
		SELECT 
			id, channel_id, user_id, content, type, mentions, attachments, reactions,
			is_edited, edited_at, is_deleted, deleted_at, is_moderated, mod_reason,
			reply_to_id, created_at
		FROM messages
		WHERE LOWER(content) LIKE LOWER($1) AND is_deleted = false
	`

	args := []interface{}{"%" + req.Query + "%"}
	argIndex := 2

	if req.ChannelID != nil {
		query += fmt.Sprintf(" AND channel_id = $%d", argIndex)
		args = append(args, *req.ChannelID)
		argIndex++
	}

	if req.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *req.UserID)
		argIndex++
	}

	if req.Type != nil {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, *req.Type)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, req.Limit)
		argIndex++
	}

	if req.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, req.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// SearchInChannel recherche des messages dans un channel spécifique
func (r *messageRepository) SearchInChannel(ctx context.Context, channelID uuid.UUID, query string, limit, offset int) ([]*models.Message, error) {
	searchQuery := `
		SELECT 
			id, channel_id, user_id, content, type, mentions, attachments, reactions,
			is_edited, edited_at, is_deleted, deleted_at, is_moderated, mod_reason,
			reply_to_id, created_at
		FROM messages
		WHERE channel_id = $1 AND LOWER(content) LIKE LOWER($2) AND is_deleted = false
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	searchTerm := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, channelID, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages in channel: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// Implémentations simplifiées pour les autres méthodes de l'interface
func (r *messageRepository) AddReaction(ctx context.Context, messageID uuid.UUID, userID uuid.UUID, emoji string) error {
	// Implementation simplifiée - à compléter plus tard
	return nil
}

func (r *messageRepository) RemoveReaction(ctx context.Context, messageID uuid.UUID, userID uuid.UUID, emoji string) error {
	// Implementation simplifiée - à compléter plus tard
	return nil
}

func (r *messageRepository) GetReactions(ctx context.Context, messageID uuid.UUID) ([]models.Reaction, error) {
	// Implementation simplifiée - à compléter plus tard
	return []models.Reaction{}, nil
}

func (r *messageRepository) GetMentions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	// Implementation simplifiée - à compléter plus tard
	return []*models.Message{}, nil
}

func (r *messageRepository) MarkMentionAsRead(ctx context.Context, messageID, userID uuid.UUID) error {
	// Implementation simplifiée - à compléter plus tard
	return nil
}

func (r *messageRepository) DeleteOldMessages(ctx context.Context, days int) (int, error) {
	query := `DELETE FROM messages WHERE created_at < NOW() - INTERVAL '%d days'`
	result, err := r.db.ExecContext(ctx, fmt.Sprintf(query, days))
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

func (r *messageRepository) GetMessageCount(ctx context.Context, channelID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM messages WHERE channel_id = $1 AND is_deleted = false`

	var count int
	err := r.db.QueryRowContext(ctx, query, channelID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get message count: %w", err)
	}

	return count, nil
}

// scanMessages aide à scanner les résultats des requêtes de messages
func (r *messageRepository) scanMessages(rows *sql.Rows) ([]*models.Message, error) {
	var messages []*models.Message

	for rows.Next() {
		var message models.Message
		var mentionsJSON, attachmentsJSON, reactionsJSON []byte

		err := rows.Scan(
			&message.ID, &message.ChannelID, &message.UserID, &message.Content, &message.Type,
			&mentionsJSON, &attachmentsJSON, &reactionsJSON, &message.IsEdited, &message.EditedAt,
			&message.IsDeleted, &message.DeletedAt, &message.IsModerated, &message.ModReason,
			&message.ReplyToID, &message.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		// Unmarshal JSON fields
		json.Unmarshal(mentionsJSON, &message.Mentions)
		json.Unmarshal(attachmentsJSON, &message.Attachments)
		json.Unmarshal(reactionsJSON, &message.Reactions)

		messages = append(messages, &message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return messages, nil
}
