package main

import (
	"context"
	"fmt"
	"log"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type ChatAgent struct {
	name string
}

func NewChatAgent() *ChatAgent {
	return &ChatAgent{name: "chat"}
}

func (a *ChatAgent) Name() string {
	return a.name
}

func (a *ChatAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *ChatAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract message from input
	message, ok := input["message"].(string)
	if !ok {
		return interfaces.AgentOutput{
			Text:     "Error: message parameter is required",
			Finished: true,
		}, fmt.Errorf("message parameter is required")
	}

	// For chat agent, we'll just return the message as-is
	// In a real implementation, this might send to a chat service
	output := fmt.Sprintf("Message received: %s", message)
	
	return interfaces.AgentOutput{
		Text:     output,
		Finished: true,
	}, nil
}

func (a *ChatAgent) HealthCheck() error {
	return nil
}

func (a *ChatAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewChatAgent()
