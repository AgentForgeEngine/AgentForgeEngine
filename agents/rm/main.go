package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type RmAgent struct {
	name string
}

func NewRmAgent() *RmAgent {
	return &RmAgent{name: "rm"}
}

func (a *RmAgent) Name() string {
	return a.name
}

func (a *RmAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *RmAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract path from input payload
	path, ok := input.Payload["path"].(string)
	if !ok || path == "" {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "Error: path parameter is required",
		}, nil
	}

	// Check if path exists
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return interfaces.AgentOutput{
				Success: false,
				Error:   fmt.Sprintf("Error: path %s does not exist", path),
			}, nil
		}
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("Error checking path %s: %v", path, err),
		}, nil
	}

	// Determine if it's a directory or file
	isDir := fileInfo.IsDir()
	var removedItems []string

	if isDir {
		// Remove directory and all contents
		err = os.RemoveAll(path)
		if err != nil {
			return interfaces.AgentOutput{
				Success: false,
				Error:   fmt.Sprintf("Error removing directory %s: %v", path, err),
			}, nil
		}
		removedItems = append(removedItems, fmt.Sprintf("Directory: %s", path))
	} else {
		// Remove single file
		err = os.Remove(path)
		if err != nil {
			return interfaces.AgentOutput{
				Success: false,
				Error:   fmt.Sprintf("Error removing file %s: %v", path, err),
			}, nil
		}
		removedItems = append(removedItems, fmt.Sprintf("File: %s", path))
	}

	// Get absolute path for reporting
	absPath, _ := filepath.Abs(path)

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"path":          path,
			"absolute_path": absPath,
			"type":          map[bool]string{true: "directory", false: "file"}[isDir],
			"removed":       removedItems,
			"success":       true,
		},
	}, nil
}

func (a *RmAgent) HealthCheck() error {
	return nil
}

func (a *RmAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewRmAgent()
