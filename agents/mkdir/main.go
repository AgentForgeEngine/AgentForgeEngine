package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type MkdirAgent struct {
	name string
}

func NewMkdirAgent() *MkdirAgent {
	return &MkdirAgent{name: "mkdir"}
}

func (a *MkdirAgent) Name() string {
	return a.name
}

func (a *MkdirAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *MkdirAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract path from input payload
	path, ok := input.Payload["path"].(string)
	if !ok || path == "" {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "Error: path parameter is required",
		}, nil
	}

	// Create directory with default permissions
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("Error creating directory %s: %v", path, err),
		}, nil
	}

	// Get directory info
	dirInfo, err := os.Stat(path)
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("Error getting directory info for %s: %v", path, err),
		}, nil
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"path":     path,
			"created":  true,
			"mode":     dirInfo.Mode(),
			"modified": dirInfo.ModTime().Format(time.RFC3339),
		},
	}, nil
}

func (a *MkdirAgent) HealthCheck() error {
	return nil
}

func (a *MkdirAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewMkdirAgent()
