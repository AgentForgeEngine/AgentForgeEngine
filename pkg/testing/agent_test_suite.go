package testing

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

// FunctionResponse represents the expected format for agent responses
type FunctionResponse struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// AgentTestSuite provides comprehensive testing utilities for agents
type AgentTestSuite struct {
	t     *testing.T
	agent interfaces.Agent
}

// NewAgentTestSuite creates a new test suite for an agent
func NewAgentTestSuite(t *testing.T, agent interfaces.Agent) *AgentTestSuite {
	return &AgentTestSuite{
		t:     t,
		agent: agent,
	}
}

// TestFunctionResponseFormat tests that agent output conforms to the expected function response format
func (ats *AgentTestSuite) TestFunctionResponseFormat(input interfaces.AgentInput, expectedAgentName string) {
	ctx := ats.t.Context()

	// Process the input
	output, err := ats.agent.Process(ctx, input)
	if err != nil {
		ats.t.Fatalf("Agent processing failed: %v", err)
	}

	// Verify output structure
	if !output.Success {
		ats.t.Errorf("Expected successful output, got error: %s", output.Error)
	}

	// Check if we have data that can be formatted as function response
	if output.Data == nil {
		ats.t.Error("Expected data in output for function response formatting")
		return
	}

	// Test the function response format
	functionResp := &FunctionResponse{
		Name:      expectedAgentName,
		Arguments: output.Data,
	}

	// Convert to the expected XML-like format
	xmlOutput := ats.formatFunctionResponse(functionResp)

	// Validate the XML format
	ats.validateFunctionResponseXML(xmlOutput, expectedAgentName)

	// Test JSON serialization
	jsonOutput, err := json.Marshal(functionResp)
	if err != nil {
		ats.t.Errorf("Failed to marshal function response to JSON: %v", err)
	}

	// Validate JSON structure
	ats.validateFunctionResponseJSON(jsonOutput, expectedAgentName)
}

// TestAgentInterface tests that the agent properly implements all interface methods
func (ats *AgentTestSuite) TestAgentInterface() {
	// Test Name method
	name := ats.agent.Name()
	if name == "" {
		ats.t.Error("Agent name should not be empty")
	}

	// Test Initialize method
	err := ats.agent.Initialize(nil)
	if err != nil {
		ats.t.Errorf("Agent initialization failed: %v", err)
	}

	// Test HealthCheck method
	err = ats.agent.HealthCheck()
	if err != nil {
		ats.t.Errorf("Agent health check failed: %v", err)
	}

	// Test Shutdown method
	err = ats.agent.Shutdown()
	if err != nil {
		ats.t.Errorf("Agent shutdown failed: %v", err)
	}
}

// TestErrorHandling tests agent error handling capabilities
func (ats *AgentTestSuite) TestErrorHandling(invalidInputs []interfaces.AgentInput) {
	ctx := ats.t.Context()

	for i, input := range invalidInputs {
		output, err := ats.agent.Process(ctx, input)

		// Should either return an error or a failed output
		if err == nil && output.Success {
			ats.t.Errorf("Test %d: Expected error or failed output for invalid input", i+1)
		}
	}
}

// TestParameterValidation tests agent parameter validation
func (ats *AgentTestSuite) TestParameterValidation(validInputs, invalidInputs []interfaces.AgentInput) {
	ctx := ats.t.Context()

	// Test valid inputs
	for i, input := range validInputs {
		output, err := ats.agent.Process(ctx, input)
		if err != nil {
			ats.t.Errorf("Valid input %d: Unexpected error: %v", i+1, err)
		}
		if !output.Success {
			ats.t.Errorf("Valid input %d: Expected success, got error: %s", i+1, output.Error)
		}
	}

	// Test invalid inputs
	ats.TestErrorHandling(invalidInputs)
}

// formatFunctionResponse converts a FunctionResponse to the expected XML-like format
func (ats *AgentTestSuite) formatFunctionResponse(resp *FunctionResponse) string {
	argsJSON, _ := json.Marshal(resp.Arguments)
	return fmt.Sprintf(`<function_response name="%s">%s</function_response>`, resp.Name, string(argsJSON))
}

// validateFunctionResponseXML validates the XML-like format of function responses
func (ats *AgentTestSuite) validateFunctionResponseXML(xmlOutput, expectedName string) {
	// Check opening tag
	openingTag := fmt.Sprintf(`<function_response name="%s">`, expectedName)
	if !strings.Contains(xmlOutput, openingTag) {
		ats.t.Errorf("Expected opening tag '%s' in output: %s", openingTag, xmlOutput)
	}

	// Check closing tag
	if !strings.Contains(xmlOutput, `</function_response>`) {
		ats.t.Errorf("Expected closing tag '</function_response>' in output: %s", xmlOutput)
	}

	// Validate JSON content within tags
	jsonPattern := regexp.MustCompile(`<function_response name="[^"]+">(.+?)</function_response>`)
	matches := jsonPattern.FindStringSubmatch(xmlOutput)
	if len(matches) < 2 {
		ats.t.Errorf("Could not extract JSON content from XML output: %s", xmlOutput)
		return
	}

	jsonContent := strings.TrimSpace(matches[1])
	var testJSON interface{}
	if err := json.Unmarshal([]byte(jsonContent), &testJSON); err != nil {
		ats.t.Errorf("Invalid JSON content in function response: %v\nJSON: %s", err, jsonContent)
	}
}

// validateFunctionResponseJSON validates the JSON structure of function responses
func (ats *AgentTestSuite) validateFunctionResponseJSON(jsonOutput []byte, expectedName string) {
	var resp FunctionResponse
	if err := json.Unmarshal(jsonOutput, &resp); err != nil {
		ats.t.Errorf("Failed to unmarshal function response JSON: %v", err)
	}

	if resp.Name != expectedName {
		ats.t.Errorf("Expected agent name '%s', got '%s'", expectedName, resp.Name)
	}

	if resp.Arguments == nil {
		ats.t.Error("Expected arguments in function response")
	}
}

// MockModelResponse simulates a model response that would trigger an agent
type MockModelResponse struct {
	FunctionCall string `json:"function_call"`
}

// CreateMockModelResponse creates a mock model response for testing
func CreateMockModelResponse(agentName string, arguments map[string]interface{}) *MockModelResponse {
	argsJSON, _ := json.Marshal(arguments)
	return &MockModelResponse{
		FunctionCall: fmt.Sprintf(`<function_call name="%s">%s</function_call>`, agentName, string(argsJSON)),
	}
}

// ParseFunctionCall parses a function call from model response
func ParseFunctionCall(modelResponse string) (agentName string, arguments map[string]interface{}, err error) {
	// Extract function call content
	pattern := regexp.MustCompile(`<function_call name="([^"]+)">(.*?)</function_call>`)
	matches := pattern.FindStringSubmatch(modelResponse)
	if len(matches) < 3 {
		return "", nil, fmt.Errorf("invalid function call format")
	}

	agentName = matches[1]
	jsonArgs := matches[2]

	err = json.Unmarshal([]byte(jsonArgs), &arguments)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return agentName, arguments, nil
}

// AgentTestCase represents a test case for an agent
type AgentTestCase struct {
	Name          string
	Input         interfaces.AgentInput
	ExpectError   bool
	ExpectSuccess bool
	ExpectedData  map[string]interface{}
}

// RunTestCases runs multiple test cases for an agent
func (ats *AgentTestSuite) RunTestCases(testCases []AgentTestCase) {
	ctx := ats.t.Context()

	for _, tc := range testCases {
		ats.t.Run(tc.Name, func(t *testing.T) {
			output, err := ats.agent.Process(ctx, tc.Input)

			if tc.ExpectError && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tc.ExpectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tc.ExpectSuccess && !output.Success {
				t.Errorf("Expected success but got error: %s", output.Error)
				return
			}

			if !tc.ExpectSuccess && output.Success {
				t.Error("Expected failure but got success")
				return
			}

			// Validate expected data
			if tc.ExpectedData != nil && output.Data != nil {
				for key, expectedValue := range tc.ExpectedData {
					if actualValue, exists := output.Data[key]; exists {
						if !ats.compareValues(expectedValue, actualValue) {
							t.Errorf("Expected %s to be %v, got %v", key, expectedValue, actualValue)
						}
					} else {
						t.Errorf("Expected key '%s' in output data", key)
					}
				}
			}
		})
	}
}

// compareValues compares two values for test validation
func (ats *AgentTestSuite) compareValues(expected, actual interface{}) bool {
	// Simple comparison for now - can be enhanced for different types
	return fmt.Sprintf("%v", expected) == fmt.Sprintf("%v", actual)
}
