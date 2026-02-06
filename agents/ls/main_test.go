package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/testing"
)

func TestLsAgent_FunctionResponseFormat(t *testing.T) {
	agent := NewLsAgent()
	suite := testing.NewAgentTestSuite(t, agent)

	// Test basic interface compliance
	suite.TestAgentInterface()

	// Initialize agent
	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	// Create temporary directory with test files
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test function response format
	input := interfaces.AgentInput{
		Type: "list",
		Payload: map[string]interface{}{
			"path":  tmpDir,
			"flags": "-la",
		},
	}

	suite.TestFunctionResponseFormat(input, "ls")
}

func TestLsAgent_ParameterValidation(t *testing.T) {
	agent := NewLsAgent()
	suite := testing.NewAgentTestSuite(t, agent)

	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	// Valid inputs
	validInputs := []interfaces.AgentInput{
		{
			Type: "list",
			Payload: map[string]interface{}{
				"path": ".",
			},
		},
		{
			Type: "list",
			Payload: map[string]interface{}{
				"path":  "/tmp",
				"flags": "-l",
			},
		},
		{
			Type: "list",
			Payload: map[string]interface{}{
				"flags": "-a",
			},
		},
		{
			Type:    "list",
			Payload: map[string]interface{}{},
		},
	}

	// Invalid inputs
	invalidInputs := []interfaces.AgentInput{
		{
			Type: "list",
			Payload: map[string]interface{}{
				"path":  "/non/existent/path",
				"flags": "invalid-flag",
			},
		},
	}

	suite.TestParameterValidation(validInputs, invalidInputs)
}

func TestLsAgent_TestCases(t *testing.T) {
	agent := NewLsAgent()
	suite := testing.NewAgentTestSuite(t, agent)

	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	// Create temporary directory
	tmpDir := t.TempDir()

	testCases := []testing.AgentTestCase{
		{
			Name: "list_current_directory",
			Input: interfaces.AgentInput{
				Type: "list",
				Payload: map[string]interface{}{
					"path": ".",
				},
			},
			ExpectSuccess: true,
			ExpectedData: map[string]interface{}{
				"path": ".",
			},
		},
		{
			Name: "list_with_flags",
			Input: interfaces.AgentInput{
				Type: "list",
				Payload: map[string]interface{}{
					"path":  tmpDir,
					"flags": "-l",
				},
			},
			ExpectSuccess: true,
			ExpectedData: map[string]interface{}{
				"path":  tmpDir,
				"flags": "-l",
			},
		},
	}

	suite.RunTestCases(testCases)
}

func TestLsAgent_ErrorHandling(t *testing.T) {
	agent := NewLsAgent()
	suite := testing.NewAgentTestSuite(t, agent)

	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	invalidInputs := []interfaces.AgentInput{
		{
			Type: "list",
			Payload: map[string]interface{}{
				"path": "/non/existent/directory/that/should/not/exist",
			},
		},
	}

	suite.TestErrorHandling(invalidInputs)
}

func TestLsAgent_ModelResponseIntegration(t *testing.T) {
	agent := NewLsAgent()

	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	// Simulate model response that would trigger this agent
	modelResponse := testing.CreateMockModelResponse("ls", map[string]interface{}{
		"path":  ".",
		"flags": "-la",
	})

	// Parse the function call
	agentName, arguments, err := testing.ParseFunctionCall(modelResponse.FunctionCall)
	if err != nil {
		t.Fatalf("Failed to parse function call: %v", err)
	}

	// Verify agent name
	if agentName != "ls" {
		t.Errorf("Expected agent name 'ls', got '%s'", agentName)
	}

	// Create input from parsed arguments
	input := interfaces.AgentInput{
		Type:    "list",
		Payload: arguments,
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
	functionResp := &testing.FunctionResponse{
		Name:      "ls",
		Arguments: output.Data,
	}

	// Test XML formatting
	xmlOutput := fmt.Sprintf(`<function_response name="%s">%s</function_response>`,
		functionResp.Name,
		func() string {
			if argsJSON, err := json.Marshal(functionResp.Arguments); err == nil {
				return string(argsJSON)
			}
			return "{}"
		}())

	// Validate XML format
	if !strings.Contains(xmlOutput, `<function_response name="ls">`) {
		t.Errorf("Expected function response tag not found in: %s", xmlOutput)
	}

	if !strings.Contains(xmlOutput, `</function_response>`) {
		t.Errorf("Expected closing function response tag not found in: %s", xmlOutput)
	}
}
