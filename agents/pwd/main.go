package main

import (
	"context"
	"fmt"
	"log"
	"os/exec

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type PwdAgent struct {
	name string
}

func NewPwdAgent() *PwdAgent {
	return &PwdAgent{name: "pwd"}
}

func (a *PwdAgent) Name() string {
	return a.name
}

func (a *PwdAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *PwdAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Build pwd command
	cmd := exec.CommandContext(ctx, "pwd")
	output, err := cmd.Output()
	if err != nil {
		return interfaces.AgentOutput{
			Text:     fmt.Sprintf("Error executing pwd: %v", err),
			Finished: true,
		}, err
	}

	return interfaces.AgentOutput{
		Text:     string(output),
		Finished: true,
	}, nil
}

func (a *PwdAgent) HealthCheck() error {
	return nil
}

func (a *PwdAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewPwdAgent()
