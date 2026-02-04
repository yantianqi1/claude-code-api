package service

import (
	"github.com/claude-api-gateway/backend/internal/model"
	"github.com/claude-api-gateway/backend/internal/repository"
)

// MappingService handles model mapping business logic
type MappingService struct {
	mappingRepo *repository.MappingRepository
	channelRepo *repository.ChannelRepository
}

// NewMappingService creates a new mapping service
func NewMappingService() *MappingService {
	return &MappingService{
		mappingRepo: repository.NewMappingRepository(),
		channelRepo: repository.NewChannelRepository(),
	}
}

// Create creates a new model mapping
func (s *MappingService) Create(create *model.MappingCreate) (int64, error) {
	// Verify channel exists
	_, err := s.channelRepo.GetByID(create.ChannelID)
	if err != nil {
		return 0, err
	}

	return s.mappingRepo.Create(create)
}

// GetByID retrieves a mapping by ID
func (s *MappingService) GetByID(id int64) (*model.ModelMapping, error) {
	return s.mappingRepo.GetByID(id)
}

// List retrieves all mappings with channel info
func (s *MappingService) List() ([]*model.ModelMappingWithChannel, error) {
	return s.mappingRepo.List()
}

// ListByChannel retrieves all mappings for a specific channel
func (s *MappingService) ListByChannel(channelID int64) ([]*model.ModelMapping, error) {
	return s.mappingRepo.ListByChannel(channelID)
}

// Update updates a mapping
func (s *MappingService) Update(id int64, update *model.MappingUpdate) error {
	return s.mappingRepo.Update(id, update)
}

// Delete deletes a mapping
func (s *MappingService) Delete(id int64) error {
	return s.mappingRepo.Delete(id)
}
