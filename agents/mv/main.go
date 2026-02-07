package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type MvAgent struct {
	name string
}

func NewMvAgent() *MvAgent {
	return &MvAgent{name: "mv"}
}

func (a *MvAgent) Name() string {
	return a.name
}

func (a *MvAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *MvAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract source and destination from input payload
	source, ok := input.Payload["source"].(string)
	if !ok || source == "" {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "Error: source parameter is required",
		}, nil
	}

	destination, ok := input.Payload["destination"].(string)
	if !ok || destination == "" {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "Error: destination parameter is required",
		}, nil
	}

	// Check if source exists
	sourceInfo, err := os.Stat(source)
	if err != nil {
		if os.IsNotExist(err) {
			return interfaces.AgentOutput{
				Success: false,
				Error:   fmt.Sprintf("Error: source %s does not exist", source),
			}, nil
		}
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("Error checking source %s: %v", source, err),
		}, nil
	}

	// Perform the move operation
	err = os.Rename(source, destination)
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("Error moving %s to %s: %v", source, destination, err),
		}, nil
	}

	// Get moved item info
	movedInfo, err := os.Stat(destination)
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("Error getting moved item info for %s: %v", destination, err),
		}, nil
	}

	// Get absolute paths for reporting
	absSource, _ := filepath.Abs(source)
	absDestination, _ := filepath.Abs(destination)

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"source":               source,
			"destination":          destination,
			"absolute_source":      absSource,
			"absolute_destination": absDestination,
			"type":                 map[bool]string{true: "directory", false: "file"}[sourceInfo.IsDir()],
			"moved":                fmt.Sprintf("%s -> %s", source, destination),
			"size":                 movedInfo.Size(),
			"modified":             movedInfo.ModTime().Format("2006-01-02 15:04:05"),
			"success":              true,
		},
	}, nil
}

func (a *MvAgent) HealthCheck() error {
	return nil
}

func (a *MvAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewMvAgent()
