package main

import (
	"context"
	"fmt"
	"log"
	"os/exec

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type DuAgent struct {
	name string
}

func NewDuAgent() *DuAgent {
	return &DuAgent{name: "du"}
}

func (a *DuAgent) Name() string {
	return a.name
}

func (a *DuAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *DuAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract path from input
	path, _ := input["path"].(string)

	// Build du command
	args := []string{"-h"}
	if path != "" {
		args = append(args, path)
	} else {
		args = append(args, ".")
	}

	cmd := exec.CommandContext(ctx, "du", args...)
	output, err := cmd.Output()
	if err != nil {
		return interfaces.AgentOutput{
			Text:     fmt.Sprintf("Error executing du: %v", err),
			Finished: true,
		}, err
	}

	return interfaces.AgentOutput{
		Text:     string(output),
		Finished: true,
	}, nil
}

func (a *DuAgent) HealthCheck() error {
	return nil
}

func (a *DuAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewDuAgent()
