package repository

import (
	"database/sql"
	"fmt"

	"github.com/claude-api-gateway/backend/internal/model"
	"github.com/claude-api-gateway/backend/pkg/database"
)

// ConfigRepository handles system configuration data operations
type ConfigRepository struct {
	db *sql.DB
}

// NewConfigRepository creates a new config repository
func NewConfigRepository() *ConfigRepository {
	return &ConfigRepository{db: database.DB}
}

// Create creates a new system config
func (r *ConfigRepository) Create(config *model.ConfigCreate) (int64, error) {
	query := `INSERT INTO system_configs (key, value, description) VALUES (?, ?, ?)`
	result, err := r.db.Exec(query, config.Key, config.Value, config.Description)
	if err != nil {
		return 0, fmt.Errorf("failed to create config: %w", err)
	}
	return result.LastInsertId()
}

// GetByKey retrieves a config by key
func (r *ConfigRepository) GetByKey(key string) (*model.SystemConfig, error) {
	query := `SELECT id, key, value, description, created_at, updated_at FROM system_configs WHERE key = ?`
	config := &model.SystemConfig{}
	err := r.db.QueryRow(query, key).Scan(
		&config.ID,
		&config.Key,
		&config.Value,
		&config.Description,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("config not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	return config, nil
}

// GetOrDefault retrieves a config by key or returns default value
func (r *ConfigRepository) GetOrDefault(key, defaultValue string) string {
	config, err := r.GetByKey(key)
	if err != nil {
		return defaultValue
	}
	return config.Value
}

// List retrieves all configs
func (r *ConfigRepository) List() ([]*model.SystemConfig, error) {
	query := `SELECT id, key, value, description, created_at, updated_at FROM system_configs ORDER BY key ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}
	defer rows.Close()

	var configs []*model.SystemConfig
	for rows.Next() {
		config := &model.SystemConfig{}
		err := rows.Scan(
			&config.ID,
			&config.Key,
			&config.Value,
			&config.Description,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan config: %w", err)
		}
		configs = append(configs, config)
	}
	return configs, nil
}

// Update updates a config
func (r *ConfigRepository) Update(key string, update *model.ConfigUpdate) error {
	query := `
		UPDATE system_configs
		SET value = COALESCE(?, value),
		    description = COALESCE(?, description),
		    updated_at = CURRENT_TIMESTAMP
		WHERE key = ?
	`
	result, err := r.db.Exec(query, update.Value, update.Description, key)
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("config not found")
	}
	return nil
}

// Delete deletes a config
func (r *ConfigRepository) Delete(key string) error {
	query := `DELETE FROM system_configs WHERE key = ?`
	result, err := r.db.Exec(query, key)
	if err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("config not found")
	}
	return nil
}
