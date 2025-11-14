package api

import (
	"net/http"
	"strconv"

	"scriberr/internal/database"
	"scriberr/internal/models"

	"github.com/gin-gonic/gin"
)

// STTModelRequest represents a request to create/update an STT model
type STTModelRequest struct {
	Name      string               `json:"name" binding:"required"`
	Type      models.STTModelType  `json:"type" binding:"required,oneof=whisperx openai custom"`
	BaseURL   *string              `json:"base_url,omitempty"`
	APIKey    *string              `json:"api_key,omitempty"`
	ModelName *string              `json:"model_name,omitempty"`
	IsActive  *bool                `json:"is_active,omitempty"`
	IsDefault *bool                `json:"is_default,omitempty"`
}

// @Summary Get all STT models
// @Description Get list of all STT models (admin only)
// @Tags admin
// @Produce json
// @Success 200 {array} models.STTModelResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/stt-models [get]
func (h *Handler) GetSTTModels(c *gin.Context) {
	var sttModels []models.STTModel

	if err := database.DB.Order("is_default DESC, name ASC").Find(&sttModels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch STT models"})
		return
	}

	// Convert to response format (hide API keys)
	response := make([]models.STTModelResponse, len(sttModels))
	for i, model := range sttModels {
		response[i] = model.ToResponse()
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Create STT model
// @Description Create a new STT model configuration (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param request body STTModelRequest true "STT model details"
// @Success 201 {object} models.STTModelResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/stt-models [post]
func (h *Handler) CreateSTTModel(c *gin.Context) {
	var req STTModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate type-specific requirements
	if req.Type == models.STTModelTypeOpenAI || req.Type == models.STTModelTypeCustom {
		if req.BaseURL == nil || *req.BaseURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "base_url is required for OpenAI and custom models"})
			return
		}
		if req.ModelName == nil || *req.ModelName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "model_name is required for OpenAI and custom models"})
			return
		}
	}

	// If setting as default, unset other defaults
	if req.IsDefault != nil && *req.IsDefault {
		if err := database.DB.Model(&models.STTModel{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update existing defaults"})
			return
		}
	}

	// Create model
	sttModel := models.STTModel{
		Name:      req.Name,
		Type:      req.Type,
		BaseURL:   req.BaseURL,
		APIKey:    req.APIKey,
		ModelName: req.ModelName,
		IsActive:  true, // Default to active
		IsDefault: false,
	}

	if req.IsActive != nil {
		sttModel.IsActive = *req.IsActive
	}
	if req.IsDefault != nil {
		sttModel.IsDefault = *req.IsDefault
	}

	if err := database.DB.Create(&sttModel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create STT model"})
		return
	}

	c.JSON(http.StatusCreated, sttModel.ToResponse())
}

// @Summary Update STT model
// @Description Update an existing STT model (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "Model ID"
// @Param request body STTModelRequest true "STT model details"
// @Success 200 {object} models.STTModelResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/stt-models/{id} [put]
func (h *Handler) UpdateSTTModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid model ID"})
		return
	}

	var req STTModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Find existing model
	var sttModel models.STTModel
	if err := database.DB.First(&sttModel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "STT model not found"})
		return
	}

	// If setting as default, unset other defaults
	if req.IsDefault != nil && *req.IsDefault && !sttModel.IsDefault {
		if err := database.DB.Model(&models.STTModel{}).Where("is_default = ? AND id != ?", true, id).Update("is_default", false).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update existing defaults"})
			return
		}
	}

	// Update fields
	sttModel.Name = req.Name
	sttModel.Type = req.Type
	sttModel.BaseURL = req.BaseURL
	sttModel.ModelName = req.ModelName

	// Only update API key if provided
	if req.APIKey != nil {
		sttModel.APIKey = req.APIKey
	}

	if req.IsActive != nil {
		sttModel.IsActive = *req.IsActive
	}
	if req.IsDefault != nil {
		sttModel.IsDefault = *req.IsDefault
	}

	if err := database.DB.Save(&sttModel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update STT model"})
		return
	}

	c.JSON(http.StatusOK, sttModel.ToResponse())
}

// @Summary Delete STT model
// @Description Delete an STT model (admin only)
// @Tags admin
// @Produce json
// @Param id path int true "Model ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/stt-models/{id} [delete]
func (h *Handler) DeleteSTTModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid model ID"})
		return
	}

	// Check if model exists
	var sttModel models.STTModel
	if err := database.DB.First(&sttModel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "STT model not found"})
		return
	}

	// Don't allow deletion of default model
	if sttModel.IsDefault {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete default STT model. Set another model as default first."})
		return
	}

	if err := database.DB.Delete(&sttModel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete STT model"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "STT model deleted successfully"})
}

// @Summary Get active STT models for users
// @Description Get list of active STT models available for transcription
// @Tags transcription
// @Produce json
// @Success 200 {array} models.STTModelResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/stt-models [get]
func (h *Handler) GetActiveSTTModels(c *gin.Context) {
	var sttModels []models.STTModel

	if err := database.DB.Where("is_active = ?", true).Order("is_default DESC, name ASC").Find(&sttModels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch STT models"})
		return
	}

	// Convert to response format (hide API keys)
	response := make([]models.STTModelResponse, len(sttModels))
	for i, model := range sttModels {
		response[i] = model.ToResponse()
	}

	c.JSON(http.StatusOK, response)
}
