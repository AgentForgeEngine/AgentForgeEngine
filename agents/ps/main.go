package main

import (
	"context"
	"fmt"
	"log"
	"os/exec

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type PsAgent struct {
	name string
}

func NewPsAgent() *PsAgent {
	return &PsAgent{name: "ps"}
}

func (a *PsAgent) Name() string {
	return a.name
}

func (a *PsAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *PsAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Build ps command
	cmd := exec.CommandContext(ctx, "ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return interfaces.AgentOutput{
			Text:     fmt.Sprintf("Error executing ps: %v", err),
			Finished: true,
		}, err
	}

	return interfaces.AgentOutput{
		Text:     string(output),
		Finished: true,
	}, nil
}

func (a *PsAgent) HealthCheck() error {
	return nil
}

func (a *PsAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewPsAgent()
