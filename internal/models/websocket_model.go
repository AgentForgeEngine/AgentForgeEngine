package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type WebSocketModel struct {
	config   interfaces.ModelConfig
	endpoint string
	client   *http.Client
}

func NewWebSocketModel(config interfaces.ModelConfig) *WebSocketModel {
	return &WebSocketModel{
		config: config,
		client: &http.Client{
			Timeout: time.Duration(getTimeout(config.Options)) * time.Second,
		},
	}
}

func (m *WebSocketModel) Name() string {
	return m.config.Name
}

func (m *WebSocketModel) Type() interfaces.ModelType {
	return interfaces.ModelTypeWebSocket
}

func (m *WebSocketModel) Initialize(config interfaces.ModelConfig) error {
	m.config = config

	// Convert ws:// to http:// for API calls
	endpoint := m.config.Endpoint
	if len(endpoint) >= 5 && endpoint[:5] == "ws://" {
		m.endpoint = "http://" + endpoint[5:]
	} else if len(endpoint) >= 6 && endpoint[:6] == "wss://" {
		m.endpoint = "https://" + endpoint[6:]
	} else {
		m.endpoint = endpoint
	}

	// Test connection
	if err := m.HealthCheck(); err != nil {
		return fmt.Errorf("failed to connect to model %s: %w", m.config.Name, err)
	}

	return nil
}

func (m *WebSocketModel) Generate(ctx context.Context, req interfaces.GenerationRequest) (*interfaces.GenerationResponse, error) {
	// Use HTTP API for Ollama (it doesn't really use WebSocket for generation)
	if m.isOllama() {
		return m.generateOllama(ctx, req)
	}

	// For other WebSocket models, fall back to HTTP for now
	// WebSocket implementation would require a more complex streaming solution
	return m.generateGeneric(ctx, req)
}

func (m *WebSocketModel) generateOllama(ctx context.Context, req interfaces.GenerationRequest) (*interfaces.GenerationResponse, error) {
	// Create Ollama request payload
	payload := map[string]interface{}{
		"model":  "default",
		"prompt": req.Prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": req.Temperature,
			"num_predict": req.MaxTokens,
			"stop":        req.StopTokens,
		},
	}

	// Make API call
	resp, err := m.makeAPIRequest(ctx, "/api/generate", payload)
	if err != nil {
		return nil, err
	}

	// Parse Ollama response
	var response struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}

	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	return &interfaces.GenerationResponse{
		Text:     response.Response,
		Finished: response.Done,
		Model:    m.config.Name,
	}, nil
}

func (m *WebSocketModel) generateGeneric(ctx context.Context, req interfaces.GenerationRequest) (*interfaces.GenerationResponse, error) {
	// Fallback to HTTP-based generation
	payload := map[string]interface{}{
		"prompt":      req.Prompt,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
		"stop":        req.StopTokens,
	}

	resp, err := m.makeAPIRequest(ctx, "/generate", payload)
	if err != nil {
		return nil, err
	}

	// Parse generic response
	var response struct {
		Text     string `json:"text"`
		Tokens   int    `json:"tokens"`
		Finished bool   `json:"finished"`
	}

	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &interfaces.GenerationResponse{
		Text:     response.Text,
		Tokens:   response.Tokens,
		Finished: response.Finished,
		Model:    m.config.Name,
	}, nil
}

func (m *WebSocketModel) makeAPIRequest(ctx context.Context, path string, payload interface{}) ([]byte, error) {
	// Build URL
	u, err := url.JoinPath(m.endpoint, path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Serialize payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := m.client.Do(req)
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

	return body, nil
}

func (m *WebSocketModel) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to get available models (Ollama health check)
	if m.isOllama() {
		u, err := url.JoinPath(m.endpoint, "/api/tags")
		if err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
		if err != nil {
			return err
		}

		resp, err := m.client.Do(req)
		if err != nil {
			return fmt.Errorf("health check failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}
	}

	// Generic health check
	req, err := http.NewRequestWithContext(ctx, "GET", m.endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
}

func (m *WebSocketModel) Shutdown() error {
	// No cleanup needed for HTTP-based WebSocket model
	return nil
}

func (m *WebSocketModel) isOllama() bool {
	return m.config.Name == "ollama" ||
		containsIgnoreCase(m.config.Endpoint, "ollama") ||
		containsIgnoreCase(m.config.Endpoint, "11434")
}
