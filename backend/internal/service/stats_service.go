package service

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/claude-api-gateway/backend/internal/model"
	"github.com/claude-api-gateway/backend/internal/repository"
)

// StatsService handles statistics and reporting
type StatsService struct {
	channelRepo *repository.ChannelRepository
	logRepo     *repository.LogRepository
}

// NewStatsService creates a new stats service
func NewStatsService() *StatsService {
	return &StatsService{
		channelRepo: repository.NewChannelRepository(),
		logRepo:     repository.NewLogRepository(),
	}
}

// GetOverallStats returns overall statistics
func (s *StatsService) GetOverallStats(filter *model.StatsFilter) (*model.OverallStats, error) {
	channels, err := s.channelRepo.List()
	if err != nil {
		return nil, err
	}

	activeCount := 0
	for _, ch := range channels {
		if ch.IsActive {
			activeCount++
		}
	}

	// Parse dates if provided
	var startDate, endDate time.Time
	if filter.StartDate != "" {
		startDate, _ = time.Parse("2006-01-02", filter.StartDate)
	}
	if filter.EndDate != "" {
		endDate, _ = time.Parse("2006-01-02", filter.EndDate)
		endDate = endDate.Add(24 * time.Hour) // Include full day
	}

	// Default to last 30 days if no filter
	if filter.StartDate == "" && filter.EndDate == "" {
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
		filter.StartDate = startDate.Format("2006-01-02")
		filter.EndDate = endDate.Format("2006-01-02")
	}

	totalRequests, _ := s.logRepo.GetTotalRequests(startDate, endDate)
	totalTokens, _ := s.logRepo.GetTotalTokens(startDate, endDate)

	channelStatsPtrs, _ := s.logRepo.GetChannelStats(filter)
	dailyStatsPtrs, _ := s.logRepo.GetDailyStats(filter)
	modelStatsPtrs, _ := s.logRepo.GetModelStats(filter)

	// Convert pointer slices to value slices
	channelStats := make([]model.ChannelStats, len(channelStatsPtrs))
	for i, s := range channelStatsPtrs {
		channelStats[i] = *s
	}
	dailyStats := make([]model.DailyStats, len(dailyStatsPtrs))
	for i, s := range dailyStatsPtrs {
		dailyStats[i] = *s
	}
	modelStats := make([]model.ModelStats, len(modelStatsPtrs))
	for i, s := range modelStatsPtrs {
		modelStats[i] = *s
	}

	return &model.OverallStats{
		TotalChannels:  int64(len(channels)),
		ActiveChannels: int64(activeCount),
		TotalRequests:  totalRequests,
		TotalTokens:    totalTokens,
		ChannelStats:   channelStats,
		DailyStats:     dailyStats,
		ModelStats:     modelStats,
	}, nil
}

// GetRequestLogs retrieves request logs with pagination
func (s *StatsService) GetRequestLogs(filter *model.StatsFilter, page, pageSize int) ([]*model.RequestLogWithChannel, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.logRepo.List(filter, page, pageSize)
}

// ExportToCSV exports logs to CSV format
func (s *StatsService) ExportToCSV(filter *model.StatsFilter) ([]byte, error) {
	logs, _, err := s.logRepo.List(filter, 1, 10000)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"ID", "Channel", "Request ID", "Model", "Upstream Model",
		"Input Tokens", "Output Tokens", "Total Tokens",
		"Request Time", "Response Time", "Latency (ms)",
		"Status", "Error Code", "Error Message", "IP Address",
	}
	writer.Write(header)

	// Write data rows
	for _, log := range logs {
		responseTime := ""
		if log.ResponseTime != nil {
			responseTime = log.ResponseTime.Format("2006-01-02 15:04:05")
		}

		row := []string{
			fmt.Sprintf("%d", log.ID),
			log.ChannelName,
			log.RequestID,
			log.ModelName,
			log.UpstreamModel,
			fmt.Sprintf("%d", log.InputTokens),
			fmt.Sprintf("%d", log.OutputTokens),
			fmt.Sprintf("%d", log.TotalTokens),
			log.RequestTime.Format("2006-01-02 15:04:05"),
			responseTime,
			fmt.Sprintf("%d", log.LatencyMs),
			log.Status,
			log.ErrorCode,
			log.ErrorMessage,
			log.IPAddress,
		}
		writer.Write(row)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
