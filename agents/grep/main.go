package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type GrepAgent struct {
	name string
}

func NewGrepAgent() *GrepAgent {
	return &GrepAgent{name: "grep"}
}

func (a *GrepAgent) Name() string {
	return a.name
}

func (a *GrepAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *GrepAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract pattern and path from input
	pattern, ok := input["pattern"].(string)
	if !ok || pattern == "" {
		return interfaces.AgentOutput{
			Text:     "Error: pattern parameter is required",
			Finished: true,
		}, fmt.Errorf("pattern parameter is required")
	}

	path, ok := input["path"].(string)
	if !ok || path == "" {
		return interfaces.AgentOutput{
			Text:     "Error: path parameter is required",
			Finished: true,
		}, fmt.Errorf("path parameter is required")
	}

	// Build grep command
	cmd := exec.CommandContext(ctx, "grep", "-n", pattern, path)
	output, err := cmd.Output()
	if err != nil {
		// grep returns non-zero exit code when no matches found, which is not an error
		if strings.Contains(string(output), "No such file or directory") {
			return interfaces.AgentOutput{
				Text:     fmt.Sprintf("Error: %s", string(output)),
				Finished: true,
			}, err
		}
		// If no matches, just return empty output
		return interfaces.AgentOutput{
			Text:     "",
			Finished: true,
		}, nil
	}

	return interfaces.AgentOutput{
		Text:     string(output),
		Finished: true,
	}, nil
}

func (a *GrepAgent) HealthCheck() error {
	return nil
}

func (a *GrepAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewGrepAgent()
