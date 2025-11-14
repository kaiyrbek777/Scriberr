package models

import (
	"time"
)

// STTModelType represents the type of STT model
type STTModelType string

const (
	STTModelTypeWhisperX STTModelType = "whisperx" // Local WhisperX
	STTModelTypeOpenAI   STTModelType = "openai"   // OpenAI Whisper API
	STTModelTypeCustom   STTModelType = "custom"   // Custom OpenAI-compatible API
)

// STTModel represents a Speech-to-Text model configuration
type STTModel struct {
	ID        uint          `json:"id" gorm:"primaryKey"`
	Name      string        `json:"name" gorm:"type:varchar(255);not null"`        // Display name (e.g., "WhisperX Large", "OpenAI Whisper")
	Type      STTModelType  `json:"type" gorm:"type:varchar(20);not null"`         // whisperx, openai, custom
	BaseURL   *string       `json:"base_url,omitempty" gorm:"type:text"`           // API base URL (for openai/custom)
	APIKey    *string       `json:"-" gorm:"type:text"`                            // Encrypted API key (for openai/custom)
	ModelName *string       `json:"model_name,omitempty" gorm:"type:varchar(255)"` // Model identifier (e.g., "whisper-1", "large-v3")
	IsActive  bool          `json:"is_active" gorm:"type:boolean;default:true"`    // Is this model available for use
	IsDefault bool          `json:"is_default" gorm:"type:boolean;default:false"`  // Is this the default model
	Config    *string       `json:"config,omitempty" gorm:"type:text"`             // JSON config for additional parameters
	CreatedAt time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}

// STTModelResponse is the API response (excludes sensitive fields)
type STTModelResponse struct {
	ID        uint         `json:"id"`
	Name      string       `json:"name"`
	Type      STTModelType `json:"type"`
	BaseURL   *string      `json:"base_url,omitempty"`
	ModelName *string      `json:"model_name,omitempty"`
	IsActive  bool         `json:"is_active"`
	IsDefault bool         `json:"is_default"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// ToResponse converts STTModel to STTModelResponse (hides API key)
func (m *STTModel) ToResponse() STTModelResponse {
	return STTModelResponse{
		ID:        m.ID,
		Name:      m.Name,
		Type:      m.Type,
		BaseURL:   m.BaseURL,
		ModelName: m.ModelName,
		IsActive:  m.IsActive,
		IsDefault: m.IsDefault,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
