package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

func TestManager_InitializeModels(t *testing.T) {
	manager := NewManager()

	// Test with empty config
	err := manager.InitializeModels([]interfaces.ModelConfig{})
	if err != nil {
		t.Errorf("Expected no error with empty config, got: %v", err)
	}

	// Test with HTTP model config
	httpConfig := interfaces.ModelConfig{
		Name:     "test-http",
		Type:     interfaces.ModelTypeHTTP,
		Endpoint: "http://localhost:8080",
		Options: map[string]interface{}{
			"timeout": 30,
		},
	}

	err = manager.InitializeModel(httpConfig)
	if err != nil {
		t.Errorf("Failed to initialize HTTP model: %v", err)
	}

	// Verify model was added
	model, exists := manager.GetModel("test-http")
	if !exists {
		t.Error("HTTP model not found after initialization")
	}

	if model.Name() != "test-http" {
		t.Errorf("Expected model name 'test-http', got '%s'", model.Name())
	}

	if model.Type() != interfaces.ModelTypeHTTP {
		t.Errorf("Expected model type %s, got %s", interfaces.ModelTypeHTTP, model.Type())
	}

	// Test with WebSocket model config
	wsConfig := interfaces.ModelConfig{
		Name:     "test-ws",
		Type:     interfaces.ModelTypeWebSocket,
		Endpoint: "ws://localhost:11434",
		Options: map[string]interface{}{
			"timeout": 30,
		},
	}

	err = manager.InitializeModel(wsConfig)
	if err != nil {
		t.Errorf("Failed to initialize WebSocket model: %v", err)
	}

	// Verify WebSocket model was added
	model, exists = manager.GetModel("test-ws")
	if !exists {
		t.Error("WebSocket model not found after initialization")
	}

	if model.Type() != interfaces.ModelTypeWebSocket {
		t.Errorf("Expected model type %s, got %s", interfaces.ModelTypeWebSocket, model.Type())
	}
}

func TestManager_InitializeModel_UnsupportedType(t *testing.T) {
	manager := NewManager()

	config := interfaces.ModelConfig{
		Name:     "test-unsupported",
		Type:     interfaces.ModelType("unsupported"),
		Endpoint: "http://localhost:8080",
	}

	err := manager.InitializeModel(config)
	if err == nil {
		t.Error("Expected error for unsupported model type")
	}
}

func TestManager_ListModels(t *testing.T) {
	manager := NewManager()

	// Initially empty
	models := manager.ListModels()
	if len(models) != 0 {
		t.Errorf("Expected 0 models, got %d", len(models))
	}

	// Add mock models
	manager.models["model1"] = &mockModel{name: "model1"}
	manager.models["model2"] = &mockModel{name: "model2"}

	models = manager.ListModels()
	if len(models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(models))
	}
}

func TestManager_HealthCheckAll(t *testing.T) {
	manager := NewManager()

	// Add mock models
	healthyModel := &mockModel{name: "healthy-model", healthy: true}
	unhealthyModel := &mockModel{name: "unhealthy-model", healthy: false}

	manager.models["healthy-model"] = healthyModel
	manager.models["unhealthy-model"] = unhealthyModel

	// Run health checks
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results := manager.HealthCheckAll(ctx)

	if len(results) != 2 {
		t.Errorf("Expected 2 health check results, got %d", len(results))
	}

	if results["healthy-model"] != nil {
		t.Errorf("Expected healthy model to have no error, got: %v", results["healthy-model"])
	}

	if results["unhealthy-model"] == nil {
		t.Error("Expected unhealthy model to have error")
	}
}

func TestHTTPModel_CreatePayload(t *testing.T) {
	model := HTTPModel{}

	// Test llama.cpp payload
	model.config = interfaces.ModelConfig{Name: "llamacpp"}
	req := interfaces.GenerationRequest{
		Prompt:      "test prompt",
		MaxTokens:   100,
		Temperature: 0.7,
		StopTokens:  []string{"\n"},
	}

	payload, err := model.createLlamaCppPayload(req)
	if err != nil {
		t.Fatalf("Failed to create llama.cpp payload: %v", err)
	}

	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		t.Fatal("Payload should be a map")
	}

	if payloadMap["prompt"] != "test prompt" {
		t.Errorf("Expected prompt 'test prompt', got '%s'", payloadMap["prompt"])
	}

	if payloadMap["n_predict"] != float64(100) {
		t.Errorf("Expected n_predict 100, got %v", payloadMap["n_predict"])
	}
}

func TestHTTPModel_IsLlamaCpp(t *testing.T) {
	tests := []struct {
		name     string
		config   interfaces.ModelConfig
		expected bool
	}{
		{
			name: "llamacpp name",
			config: interfaces.ModelConfig{
				Name: "llamacpp",
			},
			expected: true,
		},
		{
			name: "llamacpp endpoint",
			config: interfaces.ModelConfig{
				Endpoint: "http://localhost:8081/llamacpp",
			},
			expected: true,
		},
		{
			name: "other model",
			config: interfaces.ModelConfig{
				Name:     "other",
				Endpoint: "http://localhost:8080",
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := HTTPModel{config: test.config}
			result := model.isLlamaCpp()
			if result != test.expected {
				t.Errorf("Expected %v, got %v for config: %+v", test.expected, result, test.config)
			}
		})
	}
}

// Mock model for testing
type mockModel struct {
	name    string
	healthy bool
}

func (mm *mockModel) Name() string                                   { return mm.name }
func (mm *mockModel) Type() interfaces.ModelType                     { return interfaces.ModelTypeHTTP }
func (mm *mockModel) Initialize(config interfaces.ModelConfig) error { return nil }
func (mm *mockModel) Generate(ctx context.Context, req interfaces.GenerationRequest) (*interfaces.GenerationResponse, error) {
	return &interfaces.GenerationResponse{
		Text:     "test response",
		Finished: true,
		Model:    mm.name,
	}, nil
}
func (mm *mockModel) HealthCheck() error {
	if mm.healthy {
		return nil
	}
	return fmt.Errorf("model is unhealthy")
}
func (mm *mockModel) Shutdown() error { return nil }
