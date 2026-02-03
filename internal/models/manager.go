package models

import (
	"context"
	"fmt"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type Manager struct {
	models map[string]interfaces.Model
}

func NewManager() *Manager {
	return &Manager{
		models: make(map[string]interfaces.Model),
	}
}

func (m *Manager) InitializeModels(configs []interfaces.ModelConfig) error {
	for _, config := range configs {
		if err := m.InitializeModel(config); err != nil {
			return fmt.Errorf("failed to initialize model %s: %w", config.Name, err)
		}
	}
	return nil
}

func (m *Manager) InitializeModel(config interfaces.ModelConfig) error {
	var model interfaces.Model

	switch config.Type {
	case interfaces.ModelTypeHTTP:
		model = NewHTTPModel(config)
	case interfaces.ModelTypeWebSocket:
		model = NewWebSocketModel(config)
	default:
		return fmt.Errorf("unsupported model type: %s", config.Type)
	}

	if err := model.Initialize(config); err != nil {
		return fmt.Errorf("failed to initialize model %s: %w", config.Name, err)
	}

	m.models[config.Name] = model
	return nil
}

func (m *Manager) GetModel(name string) (interfaces.Model, bool) {
	model, exists := m.models[name]
	return model, exists
}

func (m *Manager) ListModels() []string {
	var names []string
	for name := range m.models {
		names = append(names, name)
	}
	return names
}

func (m *Manager) Generate(ctx context.Context, modelName string, req interfaces.GenerationRequest) (*interfaces.GenerationResponse, error) {
	model, exists := m.GetModel(modelName)
	if !exists {
		return nil, fmt.Errorf("model %s not found", modelName)
	}

	return model.Generate(ctx, req)
}

func (m *Manager) HealthCheckAll(ctx context.Context) map[string]error {
	results := make(map[string]error)

	for name, model := range m.models {
		select {
		case <-ctx.Done():
			results[name] = ctx.Err()
		default:
			results[name] = model.HealthCheck()
		}
	}

	return results
}

func (m *Manager) Shutdown() error {
	var lastErr error

	for name, model := range m.models {
		if err := model.Shutdown(); err != nil {
			lastErr = fmt.Errorf("failed to shutdown model %s: %w", name, err)
		}
	}

	return lastErr
}
