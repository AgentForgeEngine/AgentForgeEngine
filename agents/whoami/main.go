package main

import (
	"context"
	"fmt"
	"log"
	"os/exec

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type WhoamiAgent struct {
	name string
}

func NewWhoamiAgent() *WhoamiAgent {
	return &WhoamiAgent{name: "whoami"}
}

func (a *WhoamiAgent) Name() string {
	return a.name
}

func (a *WhoamiAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *WhoamiAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Build whoami command
	cmd := exec.CommandContext(ctx, "whoami")
	output, err := cmd.Output()
	if err != nil {
		return interfaces.AgentOutput{
			Text:     fmt.Sprintf("Error executing whoami: %v", err),
			Finished: true,
		}, err
	}

	return interfaces.AgentOutput{
		Text:     string(output),
		Finished: true,
	}, nil
}

func (a *WhoamiAgent) HealthCheck() error {
	return nil
}

func (a *WhoamiAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewWhoamiAgent()
