package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/claude-api-gateway/backend/internal/model"
	"github.com/claude-api-gateway/backend/pkg/database"
)

// LogRepository handles request log data operations
type LogRepository struct {
	db *sql.DB
}

// NewLogRepository creates a new log repository
func NewLogRepository() *LogRepository {
	return &LogRepository{db: database.DB}
}

// Create creates a new request log
func (r *LogRepository) Create(log *model.RequestLog) (int64, error) {
	query := `
		INSERT INTO request_logs (channel_id, request_id, model_name, upstream_model,
		                          input_tokens, output_tokens, total_tokens, request_time,
		                          response_time, latency_ms, status, error_code, error_message, ip_address)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(
		query,
		log.ChannelID,
		log.RequestID,
		log.ModelName,
		log.UpstreamModel,
		log.InputTokens,
		log.OutputTokens,
		log.TotalTokens,
		log.RequestTime,
		log.ResponseTime,
		log.LatencyMs,
		log.Status,
		log.ErrorCode,
		log.ErrorMessage,
		log.IPAddress,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create log: %w", err)
	}
	return result.LastInsertId()
}

// GetByID retrieves a log by ID
func (r *LogRepository) GetByID(id int64) (*model.RequestLogWithChannel, error) {
	query := `
		SELECT l.id, l.channel_id, c.name as channel_name, l.request_id, l.model_name,
		       l.upstream_model, l.input_tokens, l.output_tokens, l.total_tokens,
		       l.request_time, l.response_time, l.latency_ms, l.status,
		       l.error_code, l.error_message, l.ip_address, l.created_at
		FROM request_logs l
		LEFT JOIN channels c ON l.channel_id = c.id
		WHERE l.id = ?
	`
	log := &model.RequestLogWithChannel{}
	err := r.db.QueryRow(query, id).Scan(
		&log.ID,
		&log.ChannelID,
		&log.ChannelName,
		&log.RequestID,
		&log.ModelName,
		&log.UpstreamModel,
		&log.InputTokens,
		&log.OutputTokens,
		&log.TotalTokens,
		&log.RequestTime,
		&log.ResponseTime,
		&log.LatencyMs,
		&log.Status,
		&log.ErrorCode,
		&log.ErrorMessage,
		&log.IPAddress,
		&log.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("log not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get log: %w", err)
	}
	return log, nil
}

// List retrieves all logs with pagination and filtering
func (r *LogRepository) List(filter *model.StatsFilter, page, pageSize int) ([]*model.RequestLogWithChannel, int64, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.StartDate != "" {
		whereClause += " AND l.request_time >= ?"
		args = append(args, filter.StartDate)
	}
	if filter.EndDate != "" {
		whereClause += " AND l.request_time <= ?"
		args = append(args, filter.EndDate)
	}
	if filter.ChannelID > 0 {
		whereClause += " AND l.channel_id = ?"
		args = append(args, filter.ChannelID)
	}
	if filter.ModelName != "" {
		whereClause += " AND l.model_name = ?"
		args = append(args, filter.ModelName)
	}
	if filter.Status != "" {
		whereClause += " AND l.status = ?"
		args = append(args, filter.Status)
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM request_logs l " + whereClause
	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count logs: %w", err)
	}

	// Get paginated data
	query := `
		SELECT l.id, l.channel_id, c.name as channel_name, l.request_id, l.model_name,
		       l.upstream_model, l.input_tokens, l.output_tokens, l.total_tokens,
		       l.request_time, l.response_time, l.latency_ms, l.status,
		       l.error_code, l.error_message, l.ip_address, l.created_at
		FROM request_logs l
		LEFT JOIN channels c ON l.channel_id = c.id
	` + whereClause + ` ORDER BY l.request_time DESC LIMIT ? OFFSET ?`

	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list logs: %w", err)
	}
	defer rows.Close()

	var logs []*model.RequestLogWithChannel
	for rows.Next() {
		log := &model.RequestLogWithChannel{}
		err := rows.Scan(
			&log.ID,
			&log.ChannelID,
			&log.ChannelName,
			&log.RequestID,
			&log.ModelName,
			&log.UpstreamModel,
			&log.InputTokens,
			&log.OutputTokens,
			&log.TotalTokens,
			&log.RequestTime,
			&log.ResponseTime,
			&log.LatencyMs,
			&log.Status,
			&log.ErrorCode,
			&log.ErrorMessage,
			&log.IPAddress,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan log: %w", err)
		}
		logs = append(logs, log)
	}
	return logs, total, nil
}

// GetChannelStats retrieves statistics per channel
func (r *LogRepository) GetChannelStats(filter *model.StatsFilter) ([]*model.ChannelStats, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.StartDate != "" {
		whereClause += " AND l.request_time >= ?"
		args = append(args, filter.StartDate)
	}
	if filter.EndDate != "" {
		whereClause += " AND l.request_time <= ?"
		args = append(args, filter.EndDate)
	}

	query := `
		SELECT
			l.channel_id,
			c.name as channel_name,
			COUNT(*) as total_requests,
			SUM(CASE WHEN l.status = 'success' THEN 1 ELSE 0 END) as success_requests,
			SUM(CASE WHEN l.status != 'success' THEN 1 ELSE 0 END) as failed_requests,
			COALESCE(SUM(l.input_tokens), 0) as input_tokens,
			COALESCE(SUM(l.output_tokens), 0) as output_tokens,
			COALESCE(SUM(l.total_tokens), 0) as total_tokens,
			COALESCE(AVG(l.latency_ms), 0) as avg_latency_ms
		FROM request_logs l
		LEFT JOIN channels c ON l.channel_id = c.id
	` + whereClause + ` GROUP BY l.channel_id, c.name ORDER BY total_requests DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel stats: %w", err)
	}
	defer rows.Close()

	var stats []*model.ChannelStats
	for rows.Next() {
		stat := &model.ChannelStats{}
		err := rows.Scan(
			&stat.ChannelID,
			&stat.ChannelName,
			&stat.TotalRequests,
			&stat.SuccessRequests,
			&stat.FailedRequests,
			&stat.InputTokens,
			&stat.OutputTokens,
			&stat.TotalTokens,
			&stat.AvgLatencyMs,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan channel stats: %w", err)
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

// GetDailyStats retrieves daily statistics
func (r *LogRepository) GetDailyStats(filter *model.StatsFilter) ([]*model.DailyStats, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.StartDate != "" {
		whereClause += " AND l.request_time >= ?"
		args = append(args, filter.StartDate)
	}
	if filter.EndDate != "" {
		whereClause += " AND l.request_time <= ?"
		args = append(args, filter.EndDate)
	}

	query := `
		SELECT
			DATE(l.request_time) as date,
			COUNT(*) as total_requests,
			COALESCE(SUM(l.input_tokens), 0) as input_tokens,
			COALESCE(SUM(l.output_tokens), 0) as output_tokens,
			COALESCE(SUM(l.total_tokens), 0) as total_tokens
		FROM request_logs l
	` + whereClause + ` GROUP BY DATE(l.request_time) ORDER BY date DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}
	defer rows.Close()

	var stats []*model.DailyStats
	for rows.Next() {
		stat := &model.DailyStats{}
		err := rows.Scan(
			&stat.Date,
			&stat.TotalRequests,
			&stat.InputTokens,
			&stat.OutputTokens,
			&stat.TotalTokens,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily stats: %w", err)
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

// GetModelStats retrieves statistics per model
func (r *LogRepository) GetModelStats(filter *model.StatsFilter) ([]*model.ModelStats, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.StartDate != "" {
		whereClause += " AND l.request_time >= ?"
		args = append(args, filter.StartDate)
	}
	if filter.EndDate != "" {
		whereClause += " AND l.request_time <= ?"
		args = append(args, filter.EndDate)
	}

	query := `
		SELECT
			l.model_name,
			COUNT(*) as total_requests,
			COALESCE(SUM(l.input_tokens), 0) as input_tokens,
			COALESCE(SUM(l.output_tokens), 0) as output_tokens,
			COALESCE(SUM(l.total_tokens), 0) as total_tokens
		FROM request_logs l
	` + whereClause + ` GROUP BY l.model_name ORDER BY total_requests DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get model stats: %w", err)
	}
	defer rows.Close()

	var stats []*model.ModelStats
	for rows.Next() {
		stat := &model.ModelStats{}
		err := rows.Scan(
			&stat.ModelName,
			&stat.TotalRequests,
			&stat.InputTokens,
			&stat.OutputTokens,
			&stat.TotalTokens,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model stats: %w", err)
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

// GetTotalRequests returns total requests in a date range
func (r *LogRepository) GetTotalRequests(startDate, endDate time.Time) (int64, error) {
	query := `SELECT COUNT(*) FROM request_logs WHERE request_time >= ? AND request_time <= ?`
	var total int64
	err := r.db.QueryRow(query, startDate, endDate).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total requests: %w", err)
	}
	return total, nil
}

// GetTotalTokens returns total tokens consumed in a date range
func (r *LogRepository) GetTotalTokens(startDate, endDate time.Time) (int64, error) {
	query := `SELECT COALESCE(SUM(total_tokens), 0) FROM request_logs WHERE request_time >= ? AND request_time <= ?`
	var total int64
	err := r.db.QueryRow(query, startDate, endDate).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total tokens: %w", err)
	}
	return total, nil
}
