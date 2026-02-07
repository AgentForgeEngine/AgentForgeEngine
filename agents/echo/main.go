package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type EchoAgent struct {
	name string
}

func NewEchoAgent() *EchoAgent {
	return &EchoAgent{name: "echo"}
}

func (a *EchoAgent) Name() string {
	return a.name
}

func (a *EchoAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *EchoAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract message and file from input payload
	message, _ := input.Payload["message"].(string)
	file, _ := input.Payload["file"].(string)

	var output string
	var err error

	if file != "" {
		// Echo to file
		err = os.WriteFile(file, []byte(message), 0644)
		if err != nil {
			return interfaces.AgentOutput{
				Success: false,
				Error:   fmt.Sprintf("Error writing to file %s: %v", file, err),
			}, nil
		}
		output = fmt.Sprintf("Message written to file: %s", file)
	} else {
		// Echo to stdout
		output = message
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"message": message,
			"file":    file,
			"output":  output,
		},
	}, nil
}

func (a *EchoAgent) HealthCheck() error {
	return nil
}

func (a *EchoAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewEchoAgent()
