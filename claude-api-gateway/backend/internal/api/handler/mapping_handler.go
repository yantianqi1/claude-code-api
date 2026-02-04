package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/claude-api-gateway/backend/internal/model"
	"github.com/claude-api-gateway/backend/internal/service"
)

// MappingHandler handles model mapping API requests
type MappingHandler struct {
	mappingService *service.MappingService
}

// NewMappingHandler creates a new mapping handler
func NewMappingHandler() *MappingHandler {
	return &MappingHandler{
		mappingService: service.NewMappingService(),
	}
}

// Create creates a new model mapping
func (h *MappingHandler) Create(c *gin.Context) {
	var req model.MappingCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.mappingService.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mapping, _ := h.mappingService.GetByID(id)
	c.JSON(http.StatusCreated, mapping)
}

// List returns all mappings
func (h *MappingHandler) List(c *gin.Context) {
	mappings, err := h.mappingService.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  mappings,
		"total": len(mappings),
	})
}

// Get returns a single mapping by ID
func (h *MappingHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mapping id"})
		return
	}

	mapping, err := h.mappingService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mapping not found"})
		return
	}

	c.JSON(http.StatusOK, mapping)
}

// Update updates a mapping
func (h *MappingHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mapping id"})
		return
	}

	var req model.MappingUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.mappingService.Update(id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mapping, _ := h.mappingService.GetByID(id)
	c.JSON(http.StatusOK, mapping)
}

// Delete deletes a mapping
func (h *MappingHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mapping id"})
		return
	}

	if err := h.mappingService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "mapping deleted"})
}

// ListByChannel returns all mappings for a specific channel
func (h *MappingHandler) ListByChannel(c *gin.Context) {
	channelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	mappings, err := h.mappingService.ListByChannel(channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  mappings,
		"total": len(mappings),
	})
}
