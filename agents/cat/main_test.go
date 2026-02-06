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

func TestCatAgent_FunctionResponseFormat(t *testing.T) {
	agent := NewCatAgent()
	suite := testing.NewAgentTestSuite(t, agent)

	// Test basic interface compliance
	suite.TestAgentInterface()

	// Initialize agent
	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	// Create temporary file with test content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!\nThis is a test file."
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test function response format
	input := interfaces.AgentInput{
		Type: "read",
		Payload: map[string]interface{}{
			"path": testFile,
		},
	}

	suite.TestFunctionResponseFormat(input, "cat")
}

func TestCatAgent_ParameterValidation(t *testing.T) {
	agent := NewCatAgent()
	suite := testing.NewAgentTestSuite(t, agent)

	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	// Create temporary file for valid tests
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "valid.txt")
	err = os.WriteFile(testFile, []byte("valid content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Valid inputs
	validInputs := []interfaces.AgentInput{
		{
			Type: "read",
			Payload: map[string]interface{}{
				"path": testFile,
			},
		},
	}

	// Invalid inputs
	invalidInputs := []interfaces.AgentInput{
		{
			Type: "read",
			Payload: map[string]interface{}{
				"path": "",
			},
		},
		{
			Type: "read",
			Payload: map[string]interface{}{
				"path": "/non/existent/file.txt",
			},
		},
		{
			Type: "read",
			Payload: map[string]interface{}{
				"invalid": "parameter",
			},
		},
	}

	suite.TestParameterValidation(validInputs, invalidInputs)
}

func TestCatAgent_TestCases(t *testing.T) {
	agent := NewCatAgent()
	suite := testing.NewAgentTestSuite(t, agent)

	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	// Create temporary directory and files
	tmpDir := t.TempDir()
	smallFile := filepath.Join(tmpDir, "small.txt")
	largeFile := filepath.Join(tmpDir, "large.txt")

	smallContent := "Small content"
	largeContent := strings.Repeat("Large content line\n", 100)

	err = os.WriteFile(smallFile, []byte(smallContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create small test file: %v", err)
	}

	err = os.WriteFile(largeFile, []byte(largeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}

	testCases := []testing.AgentTestCase{
		{
			Name: "read_small_file",
			Input: interfaces.AgentInput{
				Type: "read",
				Payload: map[string]interface{}{
					"path": smallFile,
				},
			},
			ExpectSuccess: true,
			ExpectedData: map[string]interface{}{
				"path": smallFile,
				"size": len(smallContent),
			},
		},
		{
			Name: "read_large_file",
			Input: interfaces.AgentInput{
				Type: "read",
				Payload: map[string]interface{}{
					"path": largeFile,
				},
			},
			ExpectSuccess: true,
			ExpectedData: map[string]interface{}{
				"path": largeFile,
				"size": len(largeContent),
			},
		},
	}

	suite.RunTestCases(testCases)
}

func TestCatAgent_ErrorHandling(t *testing.T) {
	agent := NewCatAgent()
	suite := testing.NewAgentTestSuite(t, agent)

	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	invalidInputs := []interfaces.AgentInput{
		{
			Type: "read",
			Payload: map[string]interface{}{
				"path": "/non/existent/file.txt",
			},
		},
		{
			Type: "read",
			Payload: map[string]interface{}{
				"path": "",
			},
		},
		{
			Type:    "read",
			Payload: map[string]interface{}{},
		},
	}

	suite.TestErrorHandling(invalidInputs)
}

func TestCatAgent_ModelResponseIntegration(t *testing.T) {
	agent := NewCatAgent()

	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	// Create temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "integration.txt")
	testContent := "Integration test content for cat agent"
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Simulate model response that would trigger this agent
	modelResponse := testing.CreateMockModelResponse("cat", map[string]interface{}{
		"path": testFile,
	})

	// Parse the function call
	agentName, arguments, err := testing.ParseFunctionCall(modelResponse.FunctionCall)
	if err != nil {
		t.Fatalf("Failed to parse function call: %v", err)
	}

	// Verify agent name
	if agentName != "cat" {
		t.Errorf("Expected agent name 'cat', got '%s'", agentName)
	}

	// Create input from parsed arguments
	input := interfaces.AgentInput{
		Type:    "read",
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

	// Verify content
	content, ok := output.Data["content"].(string)
	if !ok {
		t.Fatal("Expected content in output data")
	}

	if content != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, content)
	}

	// Verify we can format the response as function response
	functionResp := &testing.FunctionResponse{
		Name:      "cat",
		Arguments: output.Data,
	}

	// Test XML formatting
	argsJSON, _ := json.Marshal(functionResp.Arguments)
	xmlOutput := fmt.Sprintf(`<function_response name="%s">%s</function_response>`,
		functionResp.Name,
		string(argsJSON))

	// Validate XML format
	if !strings.Contains(xmlOutput, `<function_response name="cat">`) {
		t.Errorf("Expected function response tag not found in: %s", xmlOutput)
	}

	if !strings.Contains(xmlOutput, `</function_response>`) {
		t.Errorf("Expected closing function response tag not found in: %s", xmlOutput)
	}

	// Validate JSON content within tags
	if !strings.Contains(xmlOutput, testContent) {
		t.Errorf("Expected test content not found in XML output: %s", xmlOutput)
	}
}

func TestCatAgent_EdgeCases(t *testing.T) {
	agent := NewCatAgent()

	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	ctx := context.Background()

	// Test reading empty file
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.txt")
	err = os.WriteFile(emptyFile, []byte{}, 0644)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	input := interfaces.AgentInput{
		Type: "read",
		Payload: map[string]interface{}{
			"path": emptyFile,
		},
	}

	output, err := agent.Process(ctx, input)
	if err != nil {
		t.Fatalf("Agent processing failed: %v", err)
	}

	if !output.Success {
		t.Errorf("Expected successful processing for empty file, got error: %s", output.Error)
	}

	content, ok := output.Data["content"].(string)
	if !ok {
		t.Fatal("Expected content in output data")
	}

	if content != "" {
		t.Errorf("Expected empty content, got '%s'", content)
	}

	size, ok := output.Data["size"].(int)
	if !ok {
		t.Fatal("Expected size in output data")
	}

	if size != 0 {
		t.Errorf("Expected size 0, got %d", size)
	}
}
