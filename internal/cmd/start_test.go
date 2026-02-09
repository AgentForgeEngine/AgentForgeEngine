package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/internal/config"
	"github.com/AgentForgeEngine/AgentForgeEngine/internal/loader"
	"github.com/AgentForgeEngine/AgentForgeEngine/internal/orchestrator"
	"github.com/AgentForgeEngine/AgentForgeEngine/internal/response"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/status"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/userdirs"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start AgentForgeEngine",
	Long:  "Start AgentForgeEngine with the specified configuration",
	RunE:  runStart,
}

var serverCtx context.Context
var serverCancel context.CancelFunc
var statusManager *status.Manager
var pluginManager *loader.Manager
var orchestratorManager *orchestrator.Manager

func runStart(cmd *cobra.Command, args []string) error {
	// Initialize user directories and status manager
	userDirs, err := userdirs.NewUserDirectories()
	if err != nil {
		return fmt.Errorf("failed to create user directories: %w", err)
	}

	statusManager = status.NewManager(userDirs)

	// Initialize plugin manager
	pluginManager = loader.NewManager(userDirs.AgentsDir, userDirs.CacheDir)

	// Load available agents
	agentConfigs := []config.AgentConfig{
		{Name: "ls", Type: "local", Path: "agents/ls"},
		{Name: "cat", Type: "local", Path: "agents/cat"},
		{Name: "echo", Type: "local", Path: "agents/echo"},
		{Name: "touch", Type: "local", Path: "agents/touch"},
	}

	for _, agentConfig := range agentConfigs {
		err := pluginManager.LoadLocalAgent(agentConfig.Path, agentConfig.Name)
		if err != nil {
			log.Printf("Failed to load agent %s: %v", agentConfig.Name, err)
		} else {
			log.Printf("Loaded agent: %s", agentConfig.Name)
		}
	}

	// Initialize orchestrator
	orchestratorConfig := map[string]interface{}{
		"enabled":              true,
		"max_concurrent_tasks": 10,
		"task_timeout":         "5m",
		"retry_attempts":       3,
	}

	orchestratorManager = orchestrator.NewManager(pluginManager, orchestratorConfig)

	if len(args) > 0 && args[0] == "--test" {
		return runTests()
	}

	// Start server components
	serverConfig := config.ServerConfig()
	log.Printf("Server starting on %s:%d with %d agents loaded", serverConfig.Host, serverConfig.Port, len(agentConfigs))
	log.Printf("Orchestrator manager initialized with functions: %v", orchestratorManager.GetAvailableOrchestrators())

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		serverCancel()
		time.Sleep(2 * time.Second) // Give components time to cleanup

		if err := statusManager.Cleanup(); err != nil && verbose {
			log.Printf("Cleanup error: %v", err)
		}

		fmt.Println("AgentForgeEngine stopped")
	}()

	return nil
}

func runTests() error {
	fmt.Println("=== Testing AgentForgeEngine ===")

	// Test 1: Response Formatter
	testResponseFormatter()

	// Test 2: Todo Parser
	testTodoParser()

	// Test 3: Orchestrator Basic
	testOrchestrator()

	// Test 4: End-to-End Workflow
	testEndToEndWorkflow()

	fmt.Println("=== All Tests Passed ===")
	return nil
}

func testResponseFormatter() {
	fmt.Println("\n1. Testing Response Formatter...")

	// Test JSON formatter
	formatter := response.NewJSONFormatter()
	output := interfaces.AgentOutput{
		Success: true,
		Data:    map[string]interface{}{"test": "data", "value": 123},
	}

	formatted, err := formatter.FormatAgentOutput("test-agent", output)
	if err != nil {
		fmt.Printf("❌ JSON formatter error: %v\n", err)
		return
	}

	expected := `<function_response name="test-agent">{"test":"data","value":123}</function_response>`
	if formatted != expected {
		fmt.Printf("❌ JSON formatter failed: expected %s, got %s\n", expected, formatted)
		return
	}

	fmt.Println("✅ Response formatter working correctly")
}

func testTodoParser() {
	fmt.Println("\n2. Testing Todo Parser...")

	parser := orchestrator.NewTodoParser()

	testCases := []struct {
		name     string
		todo     string
		expected string
	}{
		{"List directory", "[ ] List directory", "ls-agent"},
		{"grep go files", "[ ] grep go files", "grep-agent"},
		{"create file", "[ ] create test.txt", "touch-agent"},
		{"invalid todo", "invalid todo", "unknown"},
	}

	for i, tc := range testCases {
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

	fmt.Println("✅ Todo parser working correctly")
}

func testOrchestrator() {
	fmt.Println("\n3. Testing Orchestrator...")

	// Create mock plugin manager
	mockPluginMgr := &mockPluginManager{
		agents: map[string]interface{}{
			"ls-agent":    &mockAgent{name: "ls-agent"},
			"touch-agent": &mockAgent{name: "touch-agent"},
		},
	}

	// Test orchestrator creation
	config := map[string]interface{}{
		"enabled":              true,
		"max_concurrent_tasks": 5,
		"task_timeout":         "1m",
	}

	orchestrator := orchestrator.NewManager(mockPluginMgr, config)
	if orchestrator == nil {
		fmt.Println("❌ Failed to create orchestrator")
		return
	}

	fmt.Println("✅ Orchestrator created successfully")
}

func testEndToEndWorkflow() {
	fmt.Println("\n4. Testing End-to-End Workflow...")

	mockPluginMgr := &mockPluginManager{
		agents: map[string]interface{}{
			"ls-agent": &mockAgent{
				name: "ls-agent",
				processFunc: func(input interfaces.AgentInput) (interfaces.AgentOutput, error) {
					return interfaces.AgentOutput{
						Success: true,
						Data:    map[string]interface{}{"output": "file1.txt\nfile2.txt\n"},
					}, nil
				},
			},
		},
	}

	orchestrator := orchestrator.NewManager(mockPluginMgr, map[string]interface{}{})
	if orchestrator == nil {
		fmt.Println("❌ Failed to create orchestrator")
		return
	}

	// Test workflow execution
	input := interfaces.AgentInput{
		Type: "manager",
		Payload: map[string]interface{}{
			"todos": []string{"[ ] List directory", "[ ] create test.txt"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	output, err := orchestrator.Process(ctx, input)
	if err != nil {
		fmt.Printf("❌ Workflow execution failed: %v\n", err)
		return
	}

	if !output.Success {
		fmt.Printf("❌ Workflow failed: %s\n", output.Error)
		return
	}

	// Check for function response
	responseStr, ok := output.Data["function_response"].(string)
	if !ok {
		fmt.Println("❌ No function response in output")
		return
	}

	// Should contain ls-agent and touch-agent calls
	if !strings.Contains(responseStr, "ls-agent") || !strings.Contains(responseStr, "touch-agent") {
		fmt.Printf("❌ Expected function calls not found in: %s\n", responseStr)
		return
	}

	fmt.Println("✅ End-to-end workflow test passed")
}

type mockAgent struct {
	name        string
	processFunc func(interfaces.AgentInput) (interfaces.AgentOutput, error)
}

type mockPluginManager struct {
	agents map[string]interface{}
}

func (m *mockPluginManager) GetAgent(name string) (interfaces.Agent, bool) {
	agent, exists := m.agents[name]
	if !exists {
		return nil, false
	}
	return agent.(interfaces.Agent), true
}

func (m *mockPluginManager) ListAgents() []string {
	var names []string
	for name := range m.agents {
		names = append(names, name)
	}
	return names
}
