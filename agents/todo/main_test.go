package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/testing"
)

func TestTodoAgent_FunctionResponseFormat(t *testing.T) {
	agent := NewTodoAgent()
	suite := testing.NewAgentTestSuite(t, agent)

	// Test basic interface compliance
	suite.TestAgentInterface()

	// Initialize agent
	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	// Test function response format
	input := interfaces.AgentInput{
		Type: "create",
		Payload: map[string]interface{}{
			"steps": []string{"Step 1: Test setup", "Step 2: Execute test", "Step 3: Verify results"},
		},
	}

	suite.TestFunctionResponseFormat(input, "todo")
}

func TestTodoAgent_ParameterValidation(t *testing.T) {
	agent := NewTodoAgent()
	suite := testing.NewAgentTestSuite(t, agent)

	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	// Valid inputs
	validInputs := []interfaces.AgentInput{
		{
			Type: "create",
			Payload: map[string]interface{}{
				"steps": []string{"Single step"},
			},
		},
		{
			Type: "create",
			Payload: map[string]interface{}{
				"steps": []string{"Step 1", "Step 2", "Step 3"},
			},
		},
		{
			Type: "create",
			Payload: map[string]interface{}{
				"steps": []interface{}{"Step 1", "Step 2"},
			},
		},
	}

	// Invalid inputs
	invalidInputs := []interfaces.AgentInput{
		{
			Type: "create",
			Payload: map[string]interface{}{
				"steps": []string{},
			},
		},
		{
			Type: "create",
			Payload: map[string]interface{}{
				"steps": "not a list",
			},
		},
		{
			Type: "create",
			Payload: map[string]interface{}{
				"invalid": "parameter",
			},
		},
		{
			Type:    "create",
			Payload: map[string]interface{}{},
		},
	}

	suite.TestParameterValidation(validInputs, invalidInputs)
}

func TestTodoAgent_TestCases(t *testing.T) {
	agent := NewTodoAgent()
	suite := testing.NewAgentTestSuite(t, agent)

	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	testCases := []testing.AgentTestCase{
		{
			Name: "create_single_step",
			Input: interfaces.AgentInput{
				Type: "create",
				Payload: map[string]interface{}{
					"steps": []string{"Single step todo"},
				},
			},
			ExpectSuccess: true,
			ExpectedData: map[string]interface{}{
				"count": 1,
			},
		},
		{
			Name: "create_multiple_steps",
			Input: interfaces.AgentInput{
				Type: "create",
				Payload: map[string]interface{}{
					"steps": []string{"Step 1", "Step 2", "Step 3"},
				},
			},
			ExpectSuccess: true,
			ExpectedData: map[string]interface{}{
				"count": 3,
			},
		},
		{
			Name: "create_empty_steps",
			Input: interfaces.AgentInput{
				Type: "create",
				Payload: map[string]interface{}{
					"steps": []string{},
				},
			},
			ExpectSuccess: false, // Should fail with empty steps
		},
	}

	suite.RunTestCases(testCases)
}

func TestTodoAgent_ErrorHandling(t *testing.T) {
	agent := NewTodoAgent()
	suite := testing.NewAgentTestSuite(t, agent)

	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	invalidInputs := []interfaces.AgentInput{
		{
			Type: "create",
			Payload: map[string]interface{}{
				"steps": "not a list",
			},
		},
		{
			Type: "create",
			Payload: map[string]interface{}{
				"steps": []interface{}{123, 456}, // non-string elements
			},
		},
		{
			Type:    "create",
			Payload: map[string]interface{}{},
		},
	}

	suite.TestErrorHandling(invalidInputs)
}

func TestTodoAgent_ModelResponseIntegration(t *testing.T) {
	agent := NewTodoAgent()

	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	// Simulate model response that would trigger this agent
	modelResponse := testing.CreateMockModelResponse("todo", map[string]interface{}{
		"steps": []string{"Set up test environment", "Run test cases", "Verify results", "Clean up"},
	})

	// Parse the function call
	agentName, arguments, err := testing.ParseFunctionCall(modelResponse.FunctionCall)
	if err != nil {
		t.Fatalf("Failed to parse function call: %v", err)
	}

	// Verify agent name
	if agentName != "todo" {
		t.Errorf("Expected agent name 'todo', got '%s'", agentName)
	}

	// Create input from parsed arguments
	input := interfaces.AgentInput{
		Type:    "create",
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

	// Verify steps data
	steps, ok := output.Data["steps"].([]string)
	if !ok {
		t.Fatal("Expected steps array in output data")
	}

	expectedSteps := []string{"Set up test environment", "Run test cases", "Verify results", "Clean up"}
	if len(steps) != len(expectedSteps) {
		t.Errorf("Expected %d steps, got %d", len(expectedSteps), len(steps))
	}

	for i, step := range steps {
		if i < len(expectedSteps) && step != expectedSteps[i] {
			t.Errorf("Expected step '%s', got '%s'", expectedSteps[i], step)
		}
	}

	// Verify count
	count, ok := output.Data["count"].(int)
	if !ok {
		t.Fatal("Expected count in output data")
	}

	if count != len(expectedSteps) {
		t.Errorf("Expected count %d, got %d", len(expectedSteps), count)
	}

	// Verify formatted output
	formatted, ok := output.Data["formatted"].(string)
	if !ok {
		t.Fatal("Expected formatted output in output data")
	}

	if !strings.Contains(formatted, "Created TODO list with 4 steps") {
		t.Errorf("Expected formatted output to contain step count, got: %s", formatted)
	}

	// Verify we can format the response as function response
	functionResp := &testing.FunctionResponse{
		Name:      "todo",
		Arguments: output.Data,
	}

	// Test XML formatting
	argsJSON, _ := json.Marshal(functionResp.Arguments)
	xmlOutput := fmt.Sprintf(`<function_response name="%s">%s</function_response>`,
		functionResp.Name,
		string(argsJSON))

	// Validate XML format
	if !strings.Contains(xmlOutput, `<function_response name="todo">`) {
		t.Errorf("Expected function response tag not found in: %s", xmlOutput)
	}

	if !strings.Contains(xmlOutput, `</function_response>`) {
		t.Errorf("Expected closing function response tag not found in: %s", xmlOutput)
	}

	// Validate JSON content within tags
	if !strings.Contains(xmlOutput, "Set up test environment") {
		t.Errorf("Expected first step not found in XML output: %s", xmlOutput)
	}
}

func TestTodoAgent_EdgeCases(t *testing.T) {
	agent := NewTodoAgent()

	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	ctx := context.Background()

	// Test with mixed interface types (some strings, some not)
	input := interfaces.AgentInput{
		Type: "create",
		Payload: map[string]interface{}{
			"steps": []interface{}{
				"Valid string step",
				123, // invalid number
				"Another valid step",
				nil, // invalid nil
			},
		},
	}

	output, err := agent.Process(ctx, input)
	if err != nil {
		t.Fatalf("Agent processing failed: %v", err)
	}

	if !output.Success {
		t.Errorf("Expected successful processing with mixed types, got error: %s", output.Error)
	}

	// Should only contain the valid string steps
	steps, ok := output.Data["steps"].([]string)
	if !ok {
		t.Fatal("Expected steps array in output data")
	}

	expectedValidSteps := []string{"Valid string step", "Another valid step"}
	if len(steps) != len(expectedValidSteps) {
		t.Errorf("Expected %d valid steps, got %d", len(expectedValidSteps), len(steps))
	}

	count, ok := output.Data["count"].(int)
	if !ok {
		t.Fatal("Expected count in output data")
	}

	if count != len(expectedValidSteps) {
		t.Errorf("Expected count %d, got %d", len(expectedValidSteps), count)
	}
}

func TestTodoAgent_FormattedOutput(t *testing.T) {
	agent := NewTodoAgent()

	err := agent.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	ctx := context.Background()

	steps := []string{"First task", "Second task", "Third task"}
	input := interfaces.AgentInput{
		Type: "create",
		Payload: map[string]interface{}{
			"steps": steps,
		},
	}

	output, err := agent.Process(ctx, input)
	if err != nil {
		t.Fatalf("Agent processing failed: %v", err)
	}

	if !output.Success {
		t.Fatalf("Expected successful processing, got error: %s", output.Error)
	}

	formatted, ok := output.Data["formatted"].(string)
	if !ok {
		t.Fatal("Expected formatted output in output data")
	}

	// Verify formatted output structure
	expectedLines := []string{
		"Created TODO list with 3 steps:",
		"1. First task",
		"2. Second task",
		"3. Third task",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(formatted, expectedLine) {
			t.Errorf("Expected formatted output to contain '%s', got: %s", expectedLine, formatted)
		}
	}
}
