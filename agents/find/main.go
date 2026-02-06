package main

import (
	"context"
	"fmt"
	"log"
	"os/exec

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type FindAgent struct {
	name string
}

func NewFindAgent() *FindAgent {
	return &FindAgent{name: "find"}
}

func (a *FindAgent) Name() string {
	return a.name
}

func (a *FindAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *FindAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract path and name from input
	path, _ := input["path"].(string)
	name, _ := input["name"].(string)

	// Build find command
	args := []string{}
	if path != "" {
		args = append(args, path)
	} else {
		args = append(args, ".")
	}

	if name != "" {
		args = append(args, "-name", name)
	}

	cmd := exec.CommandContext(ctx, "find", args...)
	output, err := cmd.Output()
	if err != nil {
		return interfaces.AgentOutput{
			Text:     fmt.Sprintf("Error executing find: %v", err),
			Finished: true,
		}, err
	}

	return interfaces.AgentOutput{
		Text:     string(output),
		Finished: true,
	}, nil
}

func (a *FindAgent) HealthCheck() error {
	return nil
}

func (a *FindAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewFindAgent()
