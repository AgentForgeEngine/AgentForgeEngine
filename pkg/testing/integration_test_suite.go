package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"testing"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

// AgentTestConfig represents configuration for testing agents
type AgentTestConfig struct {
	AgentPath     string
	AgentName     string
	TestCases     []AgentTestCase
	ExpectedTools []string
}

// AgentIntegrationTestSuite provides comprehensive integration testing for all agents
type AgentIntegrationTestSuite struct {
	t      *testing.T
	agents map[string]interfaces.Agent
}

// NewAgentIntegrationTestSuite creates a new integration test suite
func NewAgentIntegrationTestSuite(t *testing.T) *AgentIntegrationTestSuite {
	return &AgentIntegrationTestSuite{
		t:      t,
		agents: make(map[string]interfaces.Agent),
	}
}

// LoadAgentFromPlugin loads an agent from a .so plugin file
func (aits *AgentIntegrationTestSuite) LoadAgentFromPlugin(pluginPath, agentName string) error {
	// Load the plugin
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin %s: %w", pluginPath, err)
	}

	// Look for the exported Agent symbol
	symbol, err := plug.Lookup("Agent")
	if err != nil {
		return fmt.Errorf("failed to find Agent symbol in plugin %s: %w", pluginPath, err)
	}

	// Type assert to interfaces.Agent
	agent, ok := symbol.(interfaces.Agent)
	if !ok {
		return fmt.Errorf("plugin %s does not implement Agent interface", pluginPath)
	}

	aits.agents[agentName] = agent
	return nil
}

// TestAllAgents tests all loaded agents with comprehensive test cases
func (aits *AgentIntegrationTestSuite) TestAllAgents() {
	for agentName, agent := range aits.agents {
		aits.t.Run(fmt.Sprintf("Agent_%s", agentName), func(t *testing.T) {
			suite := NewAgentTestSuite(t, agent)

			// Test basic interface compliance
			suite.TestAgentInterface()

			// Test function response format
			aits.testFunctionResponseFormat(t, agent, agentName)

			// Test model response integration
			aits.testModelResponseIntegration(t, agent, agentName)

			// Test error handling
			aits.testErrorHandling(t, agent, agentName)
		})
	}
}

// testFunctionResponseFormat tests function response format for a specific agent
func (aits *AgentIntegrationTestSuite) testFunctionResponseFormat(t *testing.T, agent interfaces.Agent, agentName string) {
	// Initialize agent
	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent %s: %v", agentName, err)
	}

	// Get test input for this agent
	input := aits.getTestInputForAgent(agentName)
	if input == nil {
		t.Skipf("No test input defined for agent %s", agentName)
		return
	}

	// Create test suite
	suite := NewAgentTestSuite(t, agent)

	// Test function response format
	suite.TestFunctionResponseFormat(*input, agentName)
}

// testModelResponseIntegration tests model response integration for a specific agent
func (aits *AgentIntegrationTestSuite) testModelResponseIntegration(t *testing.T, agent interfaces.Agent, agentName string) {
	// Initialize agent
	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent %s: %v", agentName, err)
	}

	// Get test arguments for this agent
	arguments := aits.getTestArgumentsForAgent(agentName)
	if arguments == nil {
		t.Skipf("No test arguments defined for agent %s", agentName)
		return
	}

	// Create mock model response
	modelResponse := CreateMockModelResponse(agentName, arguments)

	// Parse the function call
	parsedAgentName, parsedArgs, err := ParseFunctionCall(modelResponse.FunctionCall)
	if err != nil {
		t.Fatalf("Failed to parse function call: %v", err)
	}

	// Verify agent name
	if parsedAgentName != agentName {
		t.Errorf("Expected agent name '%s', got '%s'", agentName, parsedAgentName)
	}

	// Create input from parsed arguments
	input := interfaces.AgentInput{
		Type:    aits.getInputTypeForAgent(agentName),
		Payload: parsedArgs,
	}

	// Process the input
	ctx := context.Background()
	output, err := agent.Process(ctx, input)
	if err != nil {
		t.Fatalf("Agent processing failed: %v", err)
	}

	// Verify success
	if !output.Success {
		t.Errorf("Expected successful processing, got error: %s", output.Error)
	}

	// Verify we can format the response as function response
	functionResp := &FunctionResponse{
		Name:      agentName,
		Arguments: output.Data,
	}

	// Test XML formatting
	argsJSON, _ := json.Marshal(functionResp.Arguments)
	xmlOutput := fmt.Sprintf(`<function_response name="%s">%s</function_response>`,
		functionResp.Name,
		string(argsJSON))

	// Validate XML format
	expectedTag := fmt.Sprintf(`<function_response name="%s">`, agentName)
	if !strings.Contains(xmlOutput, expectedTag) {
		t.Errorf("Expected function response tag not found in: %s", xmlOutput)
	}

	if !strings.Contains(xmlOutput, `</function_response>`) {
		t.Errorf("Expected closing function response tag not found in: %s", xmlOutput)
	}

	// Validate JSON content
	var parsedData map[string]interface{}
	if err := json.Unmarshal([]byte(xmlOutput[strings.Index(xmlOutput, ">")+1:strings.LastIndex(xmlOutput, "<")]), &parsedData); err != nil {
		t.Errorf("Invalid JSON content in function response: %v", err)
	}
}

// testErrorHandling tests error handling for a specific agent
func (aits *AgentIntegrationTestSuite) testErrorHandling(t *testing.T, agent interfaces.Agent, agentName string) {
	// Initialize agent
	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent %s: %v", agentName, err)
	}

	// Get invalid test inputs for this agent
	invalidInputs := aits.getInvalidInputsForAgent(agentName)
	if len(invalidInputs) == 0 {
		t.Skipf("No invalid inputs defined for agent %s", agentName)
		return
	}

	ctx := context.Background()

	for i, input := range invalidInputs {
		output, err := agent.Process(ctx, input)

		// Should either return an error or a failed output
		if err == nil && output.Success {
			t.Errorf("Invalid input %d: Expected error or failed output", i+1)
		}
	}
}

// getTestInputForAgent returns test input for a specific agent
func (aits *AgentIntegrationTestSuite) getTestInputForAgent(agentName string) *interfaces.AgentInput {
	switch agentName {
	case "ls":
		return &interfaces.AgentInput{
			Type: "list",
			Payload: map[string]interface{}{
				"path":  ".",
				"flags": "-la",
			},
		}
	case "cat":
		// Create a temporary test file
		tmpDir := os.TempDir()
		testFile := filepath.Join(tmpDir, "cat_test.txt")
		os.WriteFile(testFile, []byte("Test content for cat agent"), 0644)

		return &interfaces.AgentInput{
			Type: "read",
			Payload: map[string]interface{}{
				"path": testFile,
			},
		}
	case "todo":
		return &interfaces.AgentInput{
			Type: "create",
			Payload: map[string]interface{}{
				"steps": []string{"Test step 1", "Test step 2"},
			},
		}
	case "pwd":
		return &interfaces.AgentInput{
			Type:    "get",
			Payload: map[string]interface{}{},
		}
	case "whoami":
		return &interfaces.AgentInput{
			Type:    "get",
			Payload: map[string]interface{}{},
		}
	case "uname":
		return &interfaces.AgentInput{
			Type: "get",
			Payload: map[string]interface{}{
				"flags": "-a",
			},
		}
	case "ps":
		return &interfaces.AgentInput{
			Type: "list",
			Payload: map[string]interface{}{
				"flags": "-ef",
			},
		}
	case "df":
		return &interfaces.AgentInput{
			Type: "get",
			Payload: map[string]interface{}{
				"flags": "-h",
			},
		}
	case "du":
		return &interfaces.AgentInput{
			Type: "get",
			Payload: map[string]interface{}{
				"path":  ".",
				"flags": "-h",
			},
		}
	case "grep":
		return &interfaces.AgentInput{
			Type: "search",
			Payload: map[string]interface{}{
				"pattern": "test",
				"path":    ".",
			},
		}
	case "find":
		return &interfaces.AgentInput{
			Type: "search",
			Payload: map[string]interface{}{
				"path": ".",
				"name": "*.go",
			},
		}
	case "stat":
		return &interfaces.AgentInput{
			Type: "get",
			Payload: map[string]interface{}{
				"path": ".",
			},
		}
	case "chat":
		return &interfaces.AgentInput{
			Type: "message",
			Payload: map[string]interface{}{
				"message": "Hello, this is a test message",
			},
		}
	default:
		return nil
	}
}

// getTestArgumentsForAgent returns test arguments for a specific agent
func (aits *AgentIntegrationTestSuite) getTestArgumentsForAgent(agentName string) map[string]interface{} {
	input := aits.getTestInputForAgent(agentName)
	if input == nil {
		return nil
	}
	return input.Payload
}

// getInputTypeForAgent returns input type for a specific agent
func (aits *AgentIntegrationTestSuite) getInputTypeForAgent(agentName string) string {
	input := aits.getTestInputForAgent(agentName)
	if input == nil {
		return "process"
	}
	return input.Type
}

// getInvalidInputsForAgent returns invalid inputs for a specific agent
func (aits *AgentIntegrationTestSuite) getInvalidInputsForAgent(agentName string) []interfaces.AgentInput {
	switch agentName {
	case "ls":
		return []interfaces.AgentInput{
			{
				Type: "list",
				Payload: map[string]interface{}{
					"path": "/non/existent/path",
				},
			},
		}
	case "cat":
		return []interfaces.AgentInput{
			{
				Type: "read",
				Payload: map[string]interface{}{
					"path": "/non/existent/file.txt",
				},
			},
			{
				Type:    "read",
				Payload: map[string]interface{}{},
			},
		}
	case "todo":
		return []interfaces.AgentInput{
			{
				Type: "create",
				Payload: map[string]interface{}{
					"steps": "not a list",
				},
			},
			{
				Type:    "create",
				Payload: map[string]interface{}{},
			},
		}
	default:
		return []interfaces.AgentInput{}
	}
}

// BuildAllAgents builds all agent plugins for testing
func BuildAllAgents(buildDir string) error {
	agents := []string{
		"ls", "cat", "todo", "pwd", "whoami", "uname",
		"ps", "df", "du", "grep", "find", "stat", "chat",
	}

	for _, agent := range agents {
		agentDir := filepath.Join("agents", agent)
		pluginPath := filepath.Join(buildDir, agent+".so")

		// Build the agent plugin
		cmd := fmt.Sprintf("cd %s && go build -buildmode=plugin -o %s .", agentDir, pluginPath)
		if err := runShellCommand(cmd); err != nil {
			return fmt.Errorf("failed to build agent %s: %w", agent, err)
		}
	}

	return nil
}

// runShellCommand runs a shell command (simplified version)
func runShellCommand(command string) error {
	// This is a placeholder - in a real implementation you'd use os/exec
	fmt.Printf("Running: %s\n", command)
	return nil
}
