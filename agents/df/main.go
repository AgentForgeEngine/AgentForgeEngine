package main

import (
	"context"
	"fmt"
	"log"
	"os/exec

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type DfAgent struct {
	name string
}

func NewDfAgent() *DfAgent {
	return &DfAgent{name: "df"}
}

func (a *DfAgent) Name() string {
	return a.name
}

func (a *DfAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *DfAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Build df command
	cmd := exec.CommandContext(ctx, "df", "-h")
	output, err := cmd.Output()
	if err != nil {
		return interfaces.AgentOutput{
			Text:     fmt.Sprintf("Error executing df: %v", err),
			Finished: true,
		}, err
	}

	return interfaces.AgentOutput{
		Text:     string(output),
		Finished: true,
	}, nil
}

func (a *DfAgent) HealthCheck() error {
	return nil
}

func (a *DfAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewDfAgent()
