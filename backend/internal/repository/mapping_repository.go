package repository

import (
	"database/sql"
	"fmt"

	"github.com/claude-api-gateway/backend/internal/model"
	"github.com/claude-api-gateway/backend/pkg/database"
)

// MappingRepository handles model mapping data operations
type MappingRepository struct {
	db *sql.DB
}

// NewMappingRepository creates a new mapping repository
func NewMappingRepository() *MappingRepository {
	return &MappingRepository{db: database.DB}
}

// Create creates a new model mapping
func (r *MappingRepository) Create(mapping *model.MappingCreate) (int64, error) {
	query := `
		INSERT INTO model_mappings (channel_id, upstream_model, display_model)
		VALUES (?, ?, ?)
	`
	result, err := r.db.Exec(query, mapping.ChannelID, mapping.UpstreamModel, mapping.DisplayModel)
	if err != nil {
		return 0, fmt.Errorf("failed to create mapping: %w", err)
	}
	return result.LastInsertId()
}

// GetByID retrieves a mapping by ID
func (r *MappingRepository) GetByID(id int64) (*model.ModelMapping, error) {
	query := `
		SELECT id, channel_id, upstream_model, display_model, is_enabled, created_at, updated_at
		FROM model_mappings WHERE id = ?
	`
	mapping := &model.ModelMapping{}
	err := r.db.QueryRow(query, id).Scan(
		&mapping.ID,
		&mapping.ChannelID,
		&mapping.UpstreamModel,
		&mapping.DisplayModel,
		&mapping.IsEnabled,
		&mapping.CreatedAt,
		&mapping.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("mapping not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get mapping: %w", err)
	}
	return mapping, nil
}

// List retrieves all mappings with channel info
func (r *MappingRepository) List() ([]*model.ModelMappingWithChannel, error) {
	query := `
		SELECT m.id, m.channel_id, c.name as channel_name, m.upstream_model,
		       m.display_model, m.is_enabled, m.created_at, m.updated_at
		FROM model_mappings m
		LEFT JOIN channels c ON m.channel_id = c.id
		ORDER BY m.id ASC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list mappings: %w", err)
	}
	defer rows.Close()

	var mappings []*model.ModelMappingWithChannel
	for rows.Next() {
		mapping := &model.ModelMappingWithChannel{}
		err := rows.Scan(
			&mapping.ID,
			&mapping.ChannelID,
			&mapping.ChannelName,
			&mapping.UpstreamModel,
			&mapping.DisplayModel,
			&mapping.IsEnabled,
			&mapping.CreatedAt,
			&mapping.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mapping: %w", err)
		}
		mappings = append(mappings, mapping)
	}
	return mappings, nil
}

// ListByChannel retrieves all mappings for a specific channel
func (r *MappingRepository) ListByChannel(channelID int64) ([]*model.ModelMapping, error) {
	query := `
		SELECT id, channel_id, upstream_model, display_model, is_enabled, created_at, updated_at
		FROM model_mappings WHERE channel_id = ?
	`
	rows, err := r.db.Query(query, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to list mappings by channel: %w", err)
	}
	defer rows.Close()

	var mappings []*model.ModelMapping
	for rows.Next() {
		mapping := &model.ModelMapping{}
		err := rows.Scan(
			&mapping.ID,
			&mapping.ChannelID,
			&mapping.UpstreamModel,
			&mapping.DisplayModel,
			&mapping.IsEnabled,
			&mapping.CreatedAt,
			&mapping.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mapping: %w", err)
		}
		mappings = append(mappings, mapping)
	}
	return mappings, nil
}

// FindByDisplayModel finds an enabled mapping for a display model
func (r *MappingRepository) FindByDisplayModel(displayModel string) ([]*model.ModelMappingWithChannel, error) {
	query := `
		SELECT m.id, m.channel_id, c.name as channel_name, m.upstream_model,
		       m.display_model, m.is_enabled, m.created_at, m.updated_at
		FROM model_mappings m
		LEFT JOIN channels c ON m.channel_id = c.id
		WHERE m.display_model = ? AND m.is_enabled = 1 AND c.is_active = 1
		ORDER BY c.priority DESC
	`
	rows, err := r.db.Query(query, displayModel)
	if err != nil {
		return nil, fmt.Errorf("failed to find mappings by display model: %w", err)
	}
	defer rows.Close()

	var mappings []*model.ModelMappingWithChannel
	for rows.Next() {
		mapping := &model.ModelMappingWithChannel{}
		err := rows.Scan(
			&mapping.ID,
			&mapping.ChannelID,
			&mapping.ChannelName,
			&mapping.UpstreamModel,
			&mapping.DisplayModel,
			&mapping.IsEnabled,
			&mapping.CreatedAt,
			&mapping.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mapping: %w", err)
		}
		mappings = append(mappings, mapping)
	}
	return mappings, nil
}

// Update updates a mapping
func (r *MappingRepository) Update(id int64, update *model.MappingUpdate) error {
	query := `
		UPDATE model_mappings
		SET upstream_model = COALESCE(?, upstream_model),
		    display_model = COALESCE(?, display_model),
		    is_enabled = COALESCE(?, is_enabled),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	result, err := r.db.Exec(
		query,
		update.UpstreamModel,
		update.DisplayModel,
		update.IsEnabled,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update mapping: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("mapping not found")
	}
	return nil
}

// Delete deletes a mapping
func (r *MappingRepository) Delete(id int64) error {
	query := `DELETE FROM model_mappings WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete mapping: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("mapping not found")
	}
	return nil
}
