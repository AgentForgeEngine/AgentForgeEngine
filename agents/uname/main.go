package main

import (
	"context"
	"fmt"
	"log"
	"os/exec

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type UnameAgent struct {
	name string
}

func NewUnameAgent() *UnameAgent {
	return &UnameAgent{name: "uname"}
}

func (a *UnameAgent) Name() string {
	return a.name
}

func (a *UnameAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *UnameAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Build uname command
	cmd := exec.CommandContext(ctx, "uname", "-a")
	output, err := cmd.Output()
	if err != nil {
		return interfaces.AgentOutput{
			Text:     fmt.Sprintf("Error executing uname: %v", err),
			Finished: true,
		}, err
	}

	return interfaces.AgentOutput{
		Text:     string(output),
		Finished: true,
	}, nil
}

func (a *UnameAgent) HealthCheck() error {
	return nil
}

func (a *UnameAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewUnameAgent()
