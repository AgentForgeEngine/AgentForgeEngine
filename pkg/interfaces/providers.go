package interfaces

import "context"

// Provider represents a model connection provider that can generate text
type Provider interface {
	Name() string
	Initialize(config map[string]interface{}) error
	Generate(ctx context.Context, input GenerationRequest) (*GenerationResponse, error)
	HealthCheck() error
	Shutdown() error
}
