package main

import (
	"context"
	"fmt"
	"log"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type TodoAgent struct {
	name string
}

func NewTodoAgent() *TodoAgent {
	return &TodoAgent{name: "todo"}
}

func (a *TodoAgent) Name() string {
	return a.name
}

func (a *TodoAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *TodoAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract steps from input payload
	steps, ok := input.Payload["steps"].([]interface{})
	if !ok {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "Error: steps parameter is required and must be a list",
		}, nil
	}

	// Convert interface{} to string slice
	var stepStrings []string
	for _, step := range steps {
		if stepStr, ok := step.(string); ok {
			stepStrings = append(stepStrings, stepStr)
		}
	}

	output := fmt.Sprintf("Created TODO list with %d steps:\n", len(stepStrings))
	for i, step := range stepStrings {
		output += fmt.Sprintf("%d. %s\n", i+1, step)
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"steps":     stepStrings,
			"count":     len(stepStrings),
			"formatted": output,
		},
	}, nil
}

func (a *TodoAgent) HealthCheck() error {
	return nil
}

func (a *TodoAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewTodoAgent()
