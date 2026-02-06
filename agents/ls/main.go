package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type LsAgent struct {
	name string
}

func NewLsAgent() *LsAgent {
	return &LsAgent{name: "ls"}
}

func (a *LsAgent) Name() string {
	return a.name
}

func (a *LsAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *LsAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract path and flags from input payload
	path, _ := input.Payload["path"].(string)
	flags, _ := input.Payload["flags"].(string)

	// Build command
	args := []string{}
	if flags != "" {
		args = append(args, strings.Split(flags, " ")...)
	}
	if path != "" {
		args = append(args, path)
	} else {
		args = append(args, ".")
	}

	cmd := exec.CommandContext(ctx, "ls", args...)
	output, err := cmd.Output()
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("Error executing ls: %v", err),
		}, nil
	}

	// Parse the output and return structured data
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []interface{}
	var dirs []interface{}

	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"output": string(output),
			"files":  files,
			"dirs":   dirs,
			"path":   path,
			"flags":  flags,
		},
	}, nil
}

func (a *LsAgent) HealthCheck() error {
	return nil
}

func (a *LsAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewLsAgent()
