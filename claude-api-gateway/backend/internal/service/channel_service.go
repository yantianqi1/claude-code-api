package service

import (
	"fmt"

	"github.com/claude-api-gateway/backend/internal/model"
	"github.com/claude-api-gateway/backend/internal/repository"
)

// ChannelService handles channel business logic
type ChannelService struct {
	channelRepo *repository.ChannelRepository
}

// NewChannelService creates a new channel service
func NewChannelService() *ChannelService {
	return &ChannelService{
		channelRepo: repository.NewChannelRepository(),
	}
}

// Create creates a new channel
func (s *ChannelService) Create(create *model.ChannelCreate) (int64, error) {
	return s.channelRepo.Create(create)
}

// GetByID retrieves a channel by ID
func (s *ChannelService) GetByID(id int64) (*model.Channel, error) {
	return s.channelRepo.GetByID(id)
}

// List retrieves all channels
func (s *ChannelService) List() ([]*model.Channel, error) {
	return s.channelRepo.List()
}

// Update updates a channel
func (s *ChannelService) Update(id int64, update *model.ChannelUpdate) error {
	return s.channelRepo.Update(id, update)
}

// Activate activates a channel
func (s *ChannelService) Activate(id int64) error {
	return s.channelRepo.SetActive(id, true)
}

// Deactivate deactivates a channel
func (s *ChannelService) Deactivate(id int64) error {
	return s.channelRepo.SetActive(id, false)
}

// Delete deletes a channel
func (s *ChannelService) Delete(id int64) error {
	// Check if channel has mappings
	mappingRepo := repository.NewMappingRepository()
	mappings, err := mappingRepo.ListByChannel(id)
	if err == nil && len(mappings) > 0 {
		return fmt.Errorf("cannot delete channel with existing mappings")
	}

	return s.channelRepo.Delete(id)
}

// GetActiveCount returns the number of active channels
func (s *ChannelService) GetActiveCount() (int, error) {
	channels, err := s.channelRepo.ListActive()
	if err != nil {
		return 0, err
	}
	return len(channels), nil
}
