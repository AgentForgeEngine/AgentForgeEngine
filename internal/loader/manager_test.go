package loader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

func TestManager_LoadLocalAgent(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := filepath.Join(tmpDir, "plugins")

	// Create a simple test plugin
	agentDir := filepath.Join(tmpDir, "test-agent")
	err := os.MkdirAll(agentDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create agent directory: %v", err)
	}

	// Create test plugin code
	pluginCode := `
package main

import (
	"context"
	"fmt"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type TestAgent struct {
	name string
}

func (ta *TestAgent) Name() string { return "test-agent" }
func (ta *TestAgent) Initialize(config map[string]interface{}) error { return nil }
func (ta *TestAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"message": "test response",
		},
	}, nil
}
func (ta *TestAgent) HealthCheck() error { return nil }
func (ta *TestAgent) Shutdown() error { return nil }

var Agent interfaces.Agent = &TestAgent{name: "test-agent"}
`

	err = os.WriteFile(filepath.Join(agentDir, "main.go"), []byte(pluginCode), 0644)
	if err != nil {
		t.Fatalf("Failed to write plugin code: %v", err)
	}

	// Initialize go module for test agent
	err = os.WriteFile(filepath.Join(agentDir, "go.mod"), []byte("module test-agent\nrequire github.com/AgentForgeEngine/AgentForgeEngine v0.0.0\nreplace github.com/AgentForgeEngine/AgentForgeEngine => ../../"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create plugin manager
	manager := NewManager(pluginsDir, filepath.Join(tmpDir, "temp"))

	// Test loading local agent
	err = manager.LoadLocalAgent(agentDir, "test-agent")
	if err != nil {
		t.Fatalf("Failed to load local agent: %v", err)
	}

	// Verify agent was loaded
	agent, exists := manager.GetAgent("test-agent")
	if !exists {
		t.Fatal("Agent not found in registry")
	}

	if agent.Name() != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got '%s'", agent.Name())
	}

	// Test agent functionality
	ctx := context.Background()
	response, err := agent.Process(ctx, interfaces.AgentInput{
		Type: "test",
		Payload: map[string]interface{}{
			"test": "data",
		},
	})

	if err != nil {
		t.Fatalf("Agent process failed: %v", err)
	}

	if !response.Success {
		t.Error("Expected successful response")
	}

	if response.Data["message"] != "test response" {
		t.Errorf("Expected 'test response', got '%s'", response.Data["message"])
	}
}

func TestManager_ListAgents(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "plugins"), filepath.Join(tmpDir, "temp"))

	// Initially empty
	agents := manager.ListAgents()
	if len(agents) != 0 {
		t.Errorf("Expected 0 agents, got %d", len(agents))
	}

	// Create and load a mock agent
	agent := &mockAgent{name: "test-agent"}
	manager.registry["test-agent"] = agent

	agents = manager.ListAgents()
	if len(agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(agents))
	}

	if agents[0] != "test-agent" {
		t.Errorf("Expected 'test-agent', got '%s'", agents[0])
	}
}

func TestManager_UnloadAgent(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "plugins"), filepath.Join(tmpDir, "temp"))

	// Add a mock agent
	agent := &mockAgent{name: "test-agent"}
	manager.registry["test-agent"] = agent

	// Verify it exists
	_, exists := manager.GetAgent("test-agent")
	if !exists {
		t.Fatal("Agent should exist before unload")
	}

	// Unload
	err := manager.UnloadAgent("test-agent")
	if err != nil {
		t.Fatalf("Failed to unload agent: %v", err)
	}

	// Verify it's gone
	_, exists = manager.GetAgent("test-agent")
	if exists {
		t.Error("Agent should not exist after unload")
	}

	// Test unloading non-existent agent
	err = manager.UnloadAgent("non-existent")
	if err == nil {
		t.Error("Expected error when unloading non-existent agent")
	}
}

func TestManager_HealthCheckAll(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "plugins"), filepath.Join(tmpDir, "temp"))

	// Add mock agents
	healthyAgent := &mockAgent{name: "healthy-agent", healthy: true}
	unhealthyAgent := &mockAgent{name: "unhealthy-agent", healthy: false}

	manager.registry["healthy-agent"] = healthyAgent
	manager.registry["unhealthy-agent"] = unhealthyAgent

	// Run health checks
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results := manager.HealthCheckAll(ctx)

	if len(results) != 2 {
		t.Errorf("Expected 2 health check results, got %d", len(results))
	}

	if results["healthy-agent"] != nil {
		t.Errorf("Expected healthy agent to have no error, got: %v", results["healthy-agent"])
	}

	if results["unhealthy-agent"] == nil {
		t.Error("Expected unhealthy agent to have error")
	}
}

// Mock agent for testing
type mockAgent struct {
	name    string
	healthy bool
}

func (ma *mockAgent) Name() string                                   { return ma.name }
func (ma *mockAgent) Initialize(config map[string]interface{}) error { return nil }
func (ma *mockAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	return interfaces.AgentOutput{Success: true}, nil
}
func (ma *mockAgent) HealthCheck() error {
	if ma.healthy {
		return nil
	}
	return fmt.Errorf("agent is unhealthy")
}
func (ma *mockAgent) Shutdown() error { return nil }
