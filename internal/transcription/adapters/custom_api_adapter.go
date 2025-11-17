Вот полное содержимое файла custom_api_adapter.go (скопируйте ВСЁ):

package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"scriberr/internal/transcription/interfaces"
	"scriberr/pkg/logger"
)

// CustomAPIAdapter implements the TranscriptionAdapter interface for OpenAI-compatible APIs
type CustomAPIAdapter struct {
	*BaseAdapter
	baseURL   string
	apiKey    string
	modelName string
	client    *http.Client
}

// NewCustomAPIAdapter creates a new Custom API adapter
func NewCustomAPIAdapter(baseURL, apiKey, modelName string) *CustomAPIAdapter {
	// Ensure baseURL doesn't have trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	capabilities := interfaces.ModelCapabilities{
		ModelID:     "custom_api",
		ModelFamily: "custom",
		DisplayName: "Custom API",
		Description: "OpenAI-compatible transcription API",
		Version:     "1.0.0",
		SupportedLanguages: []string{"auto"}, // Depends on the actual API
		SupportedFormats:   []string{"wav", "mp3", "flac", "m4a", "ogg", "webm"},
		RequiresGPU:        false, // API-based, no local GPU needed
		MemoryRequirement:  0,     // API-based, no local memory needed
		Features: map[string]bool{
			"timestamps":          true,
			"word_level":          false, // May not be supported by all APIs
			"language_detection":  true,
		},
		Metadata: map[string]string{
			"engine":      "custom_api",
			"api_type":    "openai_compatible",
			"base_url":    baseURL,
			"model_name":  modelName,
		},
	}

	schema := []interfaces.ParameterSchema{
		{
			Name:        "language",
			Type:        "string",
			Required:    false,
			Default:     nil,
			Description: "Language code (auto-detect if not specified)",
			Group:       "basic",
		},
		{
			Name:        "prompt",
			Type:        "string",
			Required:    false,
			Default:     nil,
			Description: "Optional text to guide the model's style",
			Group:       "advanced",
		},
		{
			Name:        "temperature",
			Type:        "float",
			Required:    false,
			Default:     0.0,
			Min:         &[]float64{0.0}[0],
			Max:         &[]float64{1.0}[0],
			Description: "Sampling temperature",
			Group:       "quality",
		},
		{
			Name:        "response_format",
			Type:        "string",
			Required:    false,
			Default:     "verbose_json",
			Options:     []string{"json", "text", "srt", "verbose_json", "vtt"},
			Description: "Response format",
			Group:       "advanced",
		},
	}

	baseAdapter := NewBaseAdapter("custom_api", "", capabilities, schema)

	adapter := &CustomAPIAdapter{
		BaseAdapter: baseAdapter,
		baseURL:     baseURL,
		apiKey:      apiKey,
		modelName:   modelName,
		client: &http.Client{
			Timeout: 10 * time.Minute, // Allow for long audio files
		},
	}

	return adapter
}

// GetSupportedModels returns the list of models (single custom model)
func (c *CustomAPIAdapter) GetSupportedModels() []string {
	return []string{c.modelName}
}

// PrepareEnvironment checks if the API is accessible
func (c *CustomAPIAdapter) PrepareEnvironment(ctx context.Context) error {
	logger.Info("Preparing Custom API environment", "base_url", c.baseURL, "model", c.modelName)

	// Simple health check - try to make a request to the base URL
	// Some APIs might have a /v1/models endpoint
	healthURL := c.baseURL + "/v1/models"

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		logger.Warn("Failed to create health check request", "error", err)
		// Don't fail - API might not have a models endpoint
		c.initialized = true
		return nil
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		logger.Warn("Custom API health check failed (API might still work)", "error", err)
		// Don't fail - we'll check again during actual transcription
		c.initialized = true
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		logger.Info("Custom API health check passed")
	} else {
		logger.Warn("Custom API health check returned non-OK status", "status", resp.StatusCode)
	}

	c.initialized = true
	return nil
}

// IsReady checks if the adapter is ready
func (c *CustomAPIAdapter) IsReady(ctx context.Context) bool {
	return c.initialized && c.baseURL != "" && c.modelName != ""
}

// Transcribe processes audio using the Custom API
func (c *CustomAPIAdapter) Transcribe(ctx context.Context, input interfaces.AudioInput, params map[string]interface{}, procCtx interfaces.ProcessingContext) (*interfaces.TranscriptResult, error) {
	startTime := time.Now()
	c.LogProcessingStart(input, procCtx)
	defer func() {
		c.LogProcessingEnd(procCtx, time.Since(startTime), nil)
	}()

	// Validate input
	if err := c.ValidateAudioInput(input); err != nil {
		return nil, fmt.Errorf("invalid audio input: %w", err)
	}

	// Validate parameters
	if err := c.ValidateParameters(params); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Build API request
	apiURL := c.baseURL + "/v1/audio/transcriptions"

	// Open audio file
	audioFile, err := os.Open(input.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer audioFile.Close()

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(input.FilePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, audioFile); err != nil {
		return nil, fmt.Errorf("failed to copy audio file: %w", err)
	}

	// Add model
	if err := writer.WriteField("model", c.modelName); err != nil {
		return nil, fmt.Errorf("failed to write model field: %w", err)
	}

	// Add language if specified
	if language := c.GetStringParameter(params, "language"); language != "" {
		if err := writer.WriteField("language", language); err != nil {
			return nil, fmt.Errorf("failed to write language field: %w", err)
		}
	}

	// Add prompt if specified
	if prompt := c.GetStringParameter(params, "prompt"); prompt != "" {
		if err := writer.WriteField("prompt", prompt); err != nil {
			return nil, fmt.Errorf("failed to write prompt field: %w", err)
		}
	}

	// Add temperature
	temperature := c.GetFloatParameter(params, "temperature")
	if err := writer.WriteField("temperature", fmt.Sprintf("%.2f", temperature)); err != nil {
		return nil, fmt.Errorf("failed to write temperature field: %w", err)
	}

	// Add response format
	responseFormat := c.GetStringParameter(params, "response_format")
	if responseFormat == "" {
		responseFormat = "verbose_json"
	}
	if err := writer.WriteField("response_format", responseFormat); err != nil {
		return nil, fmt.Errorf("failed to write response_format field: %w", err)
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	logger.Info("Sending request to Custom API", "url", apiURL, "model", c.modelName)

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse response
	result, err := c.parseAPIResponse(responseBody, responseFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	result.ProcessingTime = time.Since(startTime)
	result.ModelUsed = c.modelName
	if result.Metadata == nil {
		result.Metadata = make(map[string]string)
	}
	result.Metadata["api_url"] = c.baseURL
	result.Metadata["model_name"] = c.modelName

	logger.Info("Custom API transcription completed",
		"segments", len(result.Segments),
		"words", len(result.WordSegments),
		"processing_time", result.ProcessingTime)

	return result, nil
}

// parseAPIResponse parses the API response based on format
func (c *CustomAPIAdapter) parseAPIResponse(responseBody []byte, format string) (*interfaces.TranscriptResult, error) {
	if format == "verbose_json" || format == "json" {
		// Parse OpenAI-style JSON response
		var apiResponse struct {
			Task     string  `json:"task"`
			Language string  `json:"language"`
			Duration float64 `json:"duration"`
			Text     string  `json:"text"`
			Segments []struct {
				ID               int     `json:"id"`
				Seek             int     `json:"seek"`
				Start            float64 `json:"start"`
				End              float64 `json:"end"`
				Text             string  `json:"text"`
				Tokens           []int   `json:"tokens,omitempty"`
				Temperature      float64 `json:"temperature,omitempty"`
				AvgLogprob       float64 `json:"avg_logprob,omitempty"`
				CompressionRatio float64 `json:"compression_ratio,omitempty"`
				NoSpeechProb     float64 `json:"no_speech_prob,omitempty"`
			} `json:"segments,omitempty"`
			Words []struct {
				Word  string  `json:"word"`
				Start float64 `json:"start"`
				End   float64 `json:"end"`
			} `json:"words,omitempty"`
		}

		if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
			return nil, fmt.Errorf("failed to parse JSON response: %w", err)
		}

		// Convert to standard format
		result := &interfaces.TranscriptResult{
			Text:         apiResponse.Text,
			Language:     apiResponse.Language,
			Segments:     make([]interfaces.TranscriptSegment, len(apiResponse.Segments)),
			WordSegments: make([]interfaces.TranscriptWord, len(apiResponse.Words)),
			Confidence:   0.0,
			Metadata:     map[string]string{},
		}

		// Convert segments
		for i, seg := range apiResponse.Segments {
			result.Segments[i] = interfaces.TranscriptSegment{
				Start: seg.Start,
				End:   seg.End,
				Text:  seg.Text,
			}
		}

		// Convert words if available
		for i, word := range apiResponse.Words {
			result.WordSegments[i] = interfaces.TranscriptWord{
				Start: word.Start,
				End:   word.End,
				Word:  word.Word,
				Score: 1.0, // Default score if not provided
			}
		}

		return result, nil
	}

	// For text format, just return the text
	if format == "text" {
		return &interfaces.TranscriptResult{
			Text:     string(responseBody),
			Language: "unknown",
			Segments: []interfaces.TranscriptSegment{
				{
					Start: 0,
					End:   0,
					Text:  string(responseBody),
				},
			},
			WordSegments: []interfaces.TranscriptWord{},
			Metadata:     map[string]string{},
		}, nil
	}

	// For SRT/VTT formats, we'd need to parse them
	// For now, return an error
	return nil, fmt.Errorf("response format %s not yet supported", format)
}

// GetEstimatedProcessingTime provides API-specific time estimation
func (c *CustomAPIAdapter) GetEstimatedProcessingTime(input interfaces.AudioInput) time.Duration {
	// API processing time depends on network and remote processing
	// Estimate conservatively: 30% of audio duration + 10 seconds for network overhead
	audioDuration := input.Duration
	if audioDuration == 0 {
		// Fallback estimation
		estimatedMinutes := float64(input.Size) / (1024 * 1024)
		audioDuration = time.Duration(estimatedMinutes * float64(time.Minute))
	}

	processingTime := time.Duration(float64(audioDuration)*0.3) + 10*time.Second

	// Add minimum processing time
	minProcessingTime := 15 * time.Second
	if processingTime < minProcessingTime {
		processingTime = minProcessingTime
	}

	return processingTime
}