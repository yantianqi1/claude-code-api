package repository

import (
	"database/sql"
	"fmt"

	"github.com/claude-api-gateway/backend/internal/model"
	"github.com/claude-api-gateway/backend/pkg/database"
)

// ChannelRepository handles channel data operations
type ChannelRepository struct {
	db *sql.DB
}

// NewChannelRepository creates a new channel repository
func NewChannelRepository() *ChannelRepository {
	return &ChannelRepository{db: database.DB}
}

// Create creates a new channel
func (r *ChannelRepository) Create(channel *model.ChannelCreate) (int64, error) {
	query := `
		INSERT INTO channels (name, base_url, api_key, provider, priority, max_retries, timeout, rate_limit)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(
		query,
		channel.Name,
		channel.BaseURL,
		channel.APIKey,
		channel.Provider,
		channel.Priority,
		channel.MaxRetries,
		channel.Timeout,
		channel.RateLimit,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create channel: %w", err)
	}
	return result.LastInsertId()
}

// GetByID retrieves a channel by ID
func (r *ChannelRepository) GetByID(id int64) (*model.Channel, error) {
	query := `
		SELECT id, name, base_url, api_key, provider, is_active, priority,
		       max_retries, timeout, rate_limit, created_at, updated_at
		FROM channels WHERE id = ?
	`
	channel := &model.Channel{}
	err := r.db.QueryRow(query, id).Scan(
		&channel.ID,
		&channel.Name,
		&channel.BaseURL,
		&channel.APIKey,
		&channel.Provider,
		&channel.IsActive,
		&channel.Priority,
		&channel.MaxRetries,
		&channel.Timeout,
		&channel.RateLimit,
		&channel.CreatedAt,
		&channel.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("channel not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	return channel, nil
}

// List retrieves all channels
func (r *ChannelRepository) List() ([]*model.Channel, error) {
	query := `
		SELECT id, name, base_url, api_key, provider, is_active, priority,
		       max_retries, timeout, rate_limit, created_at, updated_at
		FROM channels ORDER BY priority DESC, id ASC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list channels: %w", err)
	}
	defer rows.Close()

	var channels []*model.Channel
	for rows.Next() {
		channel := &model.Channel{}
		err := rows.Scan(
			&channel.ID,
			&channel.Name,
			&channel.BaseURL,
			&channel.APIKey,
			&channel.Provider,
			&channel.IsActive,
			&channel.Priority,
			&channel.MaxRetries,
			&channel.Timeout,
			&channel.RateLimit,
			&channel.CreatedAt,
			&channel.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan channel: %w", err)
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

// ListActive retrieves all active channels ordered by priority
func (r *ChannelRepository) ListActive() ([]*model.Channel, error) {
	query := `
		SELECT id, name, base_url, api_key, provider, is_active, priority,
		       max_retries, timeout, rate_limit, created_at, updated_at
		FROM channels WHERE is_active = 1 ORDER BY priority DESC, id ASC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active channels: %w", err)
	}
	defer rows.Close()

	var channels []*model.Channel
	for rows.Next() {
		channel := &model.Channel{}
		err := rows.Scan(
			&channel.ID,
			&channel.Name,
			&channel.BaseURL,
			&channel.APIKey,
			&channel.Provider,
			&channel.IsActive,
			&channel.Priority,
			&channel.MaxRetries,
			&channel.Timeout,
			&channel.RateLimit,
			&channel.CreatedAt,
			&channel.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan channel: %w", err)
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

// Update updates a channel
func (r *ChannelRepository) Update(id int64, update *model.ChannelUpdate) error {
	query := `
		UPDATE channels
		SET name = COALESCE(?, name),
		    base_url = COALESCE(?, base_url),
		    api_key = COALESCE(?, api_key),
		    provider = COALESCE(?, provider),
		    is_active = COALESCE(?, is_active),
		    priority = COALESCE(?, priority),
		    max_retries = COALESCE(?, max_retries),
		    timeout = COALESCE(?, timeout),
		    rate_limit = COALESCE(?, rate_limit),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	result, err := r.db.Exec(
		query,
		update.Name,
		update.BaseURL,
		update.APIKey,
		update.Provider,
		update.IsActive,
		update.Priority,
		update.MaxRetries,
		update.Timeout,
		update.RateLimit,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("channel not found")
	}
	return nil
}

// SetActive sets the active status of a channel
func (r *ChannelRepository) SetActive(id int64, isActive bool) error {
	query := `UPDATE channels SET is_active = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	result, err := r.db.Exec(query, isActive, id)
	if err != nil {
		return fmt.Errorf("failed to set channel active status: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("channel not found")
	}
	return nil
}

// Delete deletes a channel
func (r *ChannelRepository) Delete(id int64) error {
	query := `DELETE FROM channels WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("channel not found")
	}
	return nil
}
