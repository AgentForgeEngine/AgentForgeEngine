package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type CatAgent struct {
	name string
}

func NewCatAgent() *CatAgent {
	return &CatAgent{name: "cat"}
}

func (a *CatAgent) Name() string {
	return a.name
}

func (a *CatAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *CatAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract path from input payload
	path, ok := input.Payload["path"].(string)
	if !ok || path == "" {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "Error: path parameter is required",
		}, nil
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("Error reading file %s: %v", path, err),
		}, nil
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"content": string(content),
			"path":    path,
			"size":    len(content),
		},
	}, nil
}

func (a *CatAgent) HealthCheck() error {
	return nil
}

func (a *CatAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewCatAgent()
