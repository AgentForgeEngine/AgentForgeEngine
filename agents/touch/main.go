package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type TouchAgent struct {
	name string
}

func NewTouchAgent() *TouchAgent {
	return &TouchAgent{name: "touch"}
}

func (a *TouchAgent) Name() string {
	return a.name
}

func (a *TouchAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *TouchAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract file from input payload
	file, ok := input.Payload["file"].(string)
	if !ok || file == "" {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "Error: file parameter is required",
		}, nil
	}

	// Create empty file or update timestamp
	err := os.WriteFile(file, []byte{}, 0644)
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("Error creating file %s: %v", file, err),
		}, nil
	}

	// Get file info
	fileInfo, err := os.Stat(file)
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("Error getting file info for %s: %v", file, err),
		}, nil
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"file":     file,
			"size":     fileInfo.Size(),
			"modified": fileInfo.ModTime().Format(time.RFC3339),
			"mode":     fileInfo.Mode(),
			"created":  true,
		},
	}, nil
}

func (a *TouchAgent) HealthCheck() error {
	return nil
}

func (a *TouchAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewTouchAgent()
