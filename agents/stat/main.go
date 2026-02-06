package main

import (
	"context"
	"fmt"
	"log"
	"os

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type StatAgent struct {
	name string
}

func NewStatAgent() *StatAgent {
	return &StatAgent{name: "stat"}
}

func (a *StatAgent) Name() string {
	return a.name
}

func (a *StatAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *StatAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract path from input
	path, ok := input["path"].(string)
	if !ok || path == "" {
		return interfaces.AgentOutput{
			Text:     "Error: path parameter is required",
			Finished: true,
		}, fmt.Errorf("path parameter is required")
	}

	// Get file stats
	stat, err := os.Stat(path)
	if err != nil {
		return interfaces.AgentOutput{
			Text:     fmt.Sprintf("Error getting stats for %s: %v", path, err),
			Finished: true,
		}, err
	}

	output := fmt.Sprintf("File stats for %s:\n", path)
	output += fmt.Sprintf("  Size: %d bytes\n", stat.Size())
	output += fmt.Sprintf("  Mode: %s\n", stat.Mode())
	output += fmt.Sprintf("  ModTime: %s\n", stat.ModTime())
	output += fmt.Sprintf("  IsDir: %t\n", stat.IsDir())

	return interfaces.AgentOutput{
		Text:     output,
		Finished: true,
	}, nil
}

func (a *StatAgent) HealthCheck() error {
	return nil
}

func (a *StatAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewStatAgent()
