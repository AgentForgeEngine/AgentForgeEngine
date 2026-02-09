package main

import (
	"fmt"
	"github.com/AgentForgeEngine/AgentForgeEngine/internal/orchestrator"
)

func main() {
	fmt.Println("=== Testing Orchestrator System ===")

	// Test 1: Todo Parser
	fmt.Println("\n1. Testing Todo Parser...")
	parser := orchestrator.NewTodoParser()

	testCases := []struct {
		todo     string
		expected string
	}{
		{"List directory", "ls-agent"},
		{"grep go files", "grep-agent"},
		{"create file", "touch-agent"},
	}

	for _, tc := range testCases {
		parsed, err := parser.ParseTodo(tc.todo)
		if err != nil {
			fmt.Printf("❌ Parser error for '%s': %v\n", tc.todo, err)
			continue
		}

		if parsed.AgentName != tc.expected {
			fmt.Printf("❌ Parser failed for '%s': expected %s, got %s\n", tc.todo, tc.expected, parsed.AgentName)
			continue
		}

		fmt.Printf("✅ Parser success: '%s' → %s\n", tc.todo, parsed.AgentName)
	}

	// Test 2: Response Formatter
	fmt.Println("\n2. Testing Response Formatter...")
	formatter := response.NewAutoFormatter()

	output := &interfaces.AgentOutput{
		Success: true,
		Data:    map[string]interface{}{"test": "data"},
	}

	formatted, err := formatter.FormatAgentOutput("ls", output)
	if err != nil {
		fmt.Printf("❌ Formatter error: %v\n", err)
		return
	}

	fmt.Printf("✅ Response formatted: %s\n", formatted)

	// Test 3: Simple Orchestrator
	fmt.Println("\n3. Testing Simple Orchestrator...")

	mockPluginMgr := &MockPluginManager{
		agents: map[string]interface{}{
			"ls":    &MockAgent{},
			"touch": &MockAgent{},
		},
	}

	config := map[string]interface{}{
		"enabled":              true,
		"max_concurrent_tasks": 5,
	}

	orch := orchestrator.NewManager(mockPluginMgr, config)
	if orch == nil {
		fmt.Println("❌ Failed to create orchestrator")
		return
	}

	fmt.Println("✅ Simple orchestrator created successfully")
	fmt.Println("   Available functions:")
	available := orch.GetAvailableOrchestrators()
	for name, desc := range available {
		fmt.Printf("   - orch.%s: %s\n", name, desc)
	}
}

// MockAgent is a minimal mock for testing
type MockAgent struct {
	name string
}

func (m *MockAgent) Name() string {
	return m.name
}

func (m *MockAgent) Initialize(config map[string]interface{}) error {
	return nil
}

func (m *MockAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	return interfaces.AgentOutput{
		Success: true,
		Data:    map[string]interface{}{"mock_output": fmt.Sprintf("%s executed", m.name)},
	}, nil
}

func (m *MockAgent) HealthCheck() error {
	return nil
}

func (m *MockAgent) Shutdown() error {
	return nil
}

// MockPluginManager implements a minimal plugin manager for testing
type MockPluginManager struct {
	agents map[string]interface{}
}

func (m *MockPluginManager) GetAgent(name string) (interfaces.Agent, bool) {
	agent, exists := m.agents[name]
	if !exists {
		return nil, false
	}
	return agent.(interfaces.Agent), true
}

func (m *MockPluginManager) ListAgents() []string {
	var names []string
	for name := range m.agents {
		names = append(names, name)
	}
	return names
}

func (m *MockPluginManager) LoadLocalAgent(path, name string) error {
	// Mock implementation - just add to the map
	m.agents[name] = &MockAgent{name: name}
	return nil
}
