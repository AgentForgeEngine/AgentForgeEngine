package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type HTTPModel struct {
	config interfaces.ModelConfig
	client *http.Client
}

func NewHTTPModel(config interfaces.ModelConfig) *HTTPModel {
	return &HTTPModel{
		config: config,
		client: &http.Client{
			Timeout: time.Duration(getTimeout(config.Options)) * time.Second,
		},
	}
}

func (m *HTTPModel) Name() string {
	return m.config.Name
}

func (m *HTTPModel) Type() interfaces.ModelType {
	return interfaces.ModelTypeHTTP
}

func (m *HTTPModel) Initialize(config interfaces.ModelConfig) error {
	m.config = config

	// Test connection
	if err := m.HealthCheck(); err != nil {
		return fmt.Errorf("failed to connect to model %s: %w", m.config.Name, err)
	}

	return nil
}

func (m *HTTPModel) Generate(ctx context.Context, req interfaces.GenerationRequest) (*interfaces.GenerationResponse, error) {
	// Create request payload based on model type
	var payload interface{}
	var err error

	switch {
	case m.isLlamaCpp():
		payload, err = m.createLlamaCppPayload(req)
	default:
		payload, err = m.createGenericPayload(req)
	}

	if err != nil {
		return nil, err
	}

	// Serialize payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", m.config.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Parse response based on model type
	var response *interfaces.GenerationResponse
	switch {
	case m.isLlamaCpp():
		response, err = m.parseLlamaCppResponse(body)
	default:
		response, err = m.parseGenericResponse(body)
	}

	if err != nil {
		return nil, err
	}

	response.Model = m.config.Name
	return response, nil
}

func (m *HTTPModel) HealthCheck() error {
	// Simple health check by making a test request
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(getTimeout(m.config.Options))*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", m.config.Endpoint+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		// Status 404 is acceptable since not all models have health endpoints
		return nil
	}

	// Read body for debugging
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("health check failed with status: %d, body: %s", resp.StatusCode, string(body))
}

func (m *HTTPModel) Shutdown() error {
	// No cleanup needed for HTTP model
	return nil
}

func (m *HTTPModel) isLlamaCpp() bool {
	return m.config.Name == "llamacpp" ||
		containsIgnoreCase(m.config.Endpoint, "llamacpp") ||
		containsIgnoreCase(m.config.Endpoint, "8080") ||
		containsIgnoreCase(m.config.Endpoint, "8081")
}

func (m *HTTPModel) createLlamaCppPayload(req interfaces.GenerationRequest) (interface{}, error) {
	return map[string]interface{}{
		"prompt":      req.Prompt,
		"n_predict":   req.MaxTokens,
		"temperature": req.Temperature,
		"stop":        req.StopTokens,
		"stream":      req.Stream,
	}, nil
}

func (m *HTTPModel) createGenericPayload(req interfaces.GenerationRequest) (interface{}, error) {
	return map[string]interface{}{
		"prompt":      req.Prompt,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
		"stop":        req.StopTokens,
		"stream":      req.Stream,
	}, nil
}

func (m *HTTPModel) parseLlamaCppResponse(body []byte) (*interfaces.GenerationResponse, error) {
	var response struct {
		Content string `json:"content"`
		Stopped bool   `json:"stopped"`
		Tokens  int    `json:"tokens_predicted"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse llama.cpp response: %w", err)
	}

	return &interfaces.GenerationResponse{
		Text:     response.Content,
		Tokens:   response.Tokens,
		Finished: response.Stopped,
	}, nil
}

func (m *HTTPModel) parseGenericResponse(body []byte) (*interfaces.GenerationResponse, error) {
	var response struct {
		Text     string `json:"text"`
		Tokens   int    `json:"tokens"`
		Finished bool   `json:"finished"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &interfaces.GenerationResponse{
		Text:     response.Text,
		Tokens:   response.Tokens,
		Finished: response.Finished,
	}, nil
}

func getTimeout(options map[string]interface{}) int {
	if timeout, ok := options["timeout"].(int); ok {
		return timeout
	}
	return 30 // default timeout
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			containsIgnoreCase(s[1:], substr) ||
			(len(s) >= len(substr) &&
				strings.ToLower(s[:len(substr)]) == strings.ToLower(substr)))
}
