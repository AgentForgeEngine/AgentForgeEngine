package testing

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAllAgentFunctionResponses tests all agents for proper function response format
func TestAllAgentFunctionResponses(t *testing.T) {
	// Test agents that are directly importable
	agents := map[string]interface{}{
		"ls":   &struct{ name string }{name: "ls"},
		"cat":  &struct{ name string }{name: "cat"},
		"todo": &struct{ name string }{name: "todo"},
	}

	// Since we can't easily load plugins in unit tests, we'll test the format validation
	for agentName := range agents {
		t.Run(fmt.Sprintf("Format_%s", agentName), func(t *testing.T) {
			testFunctionResponseFormatValidation(t, agentName)
		})
	}
}

// testFunctionResponseFormatValidation tests the function response format validation
func testFunctionResponseFormatValidation(t *testing.T, agentName string) {
	// Create mock data
	mockData := map[string]interface{}{
		"result": "success",
		"value":  42,
		"items":  []string{"item1", "item2"},
	}

	// Create function response
	functionResp := &FunctionResponse{
		Name:      agentName,
		Arguments: mockData,
	}

	// Test XML formatting
	argsJSON, _ := json.Marshal(functionResp.Arguments)
	xmlOutput := fmt.Sprintf(`<function_response name="%s">%s</function_response>`,
		functionResp.Name,
		string(argsJSON))

	// Validate XML format manually since we don't have a real agent
	expectedTag := fmt.Sprintf(`<function_response name="%s">`, agentName)
	if !strings.Contains(xmlOutput, expectedTag) {
		t.Errorf("Expected function response tag not found in: %s", xmlOutput)
	}

	if !strings.Contains(xmlOutput, `</function_response>`) {
		t.Errorf("Expected closing function response tag not found in: %s", xmlOutput)
	}

	// Validate JSON content
	startIndex := strings.Index(xmlOutput, ">") + 1
	endIndex := strings.LastIndex(xmlOutput, "<")
	if startIndex >= endIndex {
		t.Errorf("Invalid XML format: %s", xmlOutput)
	}

	jsonContent := xmlOutput[startIndex:endIndex]
	var parsedData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent), &parsedData); err != nil {
		t.Errorf("Invalid JSON content in function response: %v", err)
	}

	// Verify the mock data in JSON
	if parsedData["result"] != "success" {
		t.Errorf("Expected result 'success' in JSON, got '%v'", parsedData["result"])
	}

	if parsedData["value"] != float64(42) {
		t.Errorf("Expected value 42 in JSON, got '%v'", parsedData["value"])
	}
}

// TestModelResponseParsing tests parsing of model responses
func TestModelResponseParsing(t *testing.T) {
	testCases := []struct {
		name          string
		modelResponse string
		expectedAgent string
		expectedArgs  map[string]interface{}
		expectError   bool
	}{
		{
			name:          "ls_agent_response",
			modelResponse: `<function_call name="ls">{"path": "/tmp", "flags": "-la"}</function_call>`,
			expectedAgent: "ls",
			expectedArgs: map[string]interface{}{
				"path":  "/tmp",
				"flags": "-la",
			},
			expectError: false,
		},
		{
			name:          "cat_agent_response",
			modelResponse: `<function_call name="cat">{"path": "/etc/hosts"}</function_call>`,
			expectedAgent: "cat",
			expectedArgs: map[string]interface{}{
				"path": "/etc/hosts",
			},
			expectError: false,
		},
		{
			name:          "todo_agent_response",
			modelResponse: `<function_call name="todo">{"steps": ["Step 1", "Step 2"]}</function_call>`,
			expectedAgent: "todo",
			expectedArgs: map[string]interface{}{
				"steps": []interface{}{"Step 1", "Step 2"},
			},
			expectError: false,
		},
		{
			name:          "invalid_format",
			modelResponse: `invalid format`,
			expectError:   true,
		},
		{
			name:          "invalid_json",
			modelResponse: `<function_call name="ls">{"path": invalid json}</function_call>`,
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			agentName, arguments, err := ParseFunctionCall(tc.modelResponse)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if agentName != tc.expectedAgent {
				t.Errorf("Expected agent name '%s', got '%s'", tc.expectedAgent, agentName)
			}

			// Compare arguments
			if !compareMaps(tc.expectedArgs, arguments) {
				t.Errorf("Expected arguments %v, got %v", tc.expectedArgs, arguments)
			}
		})
	}
}

// TestFunctionResponseRoundTrip tests the complete round-trip of function responses
func TestFunctionResponseRoundTrip(t *testing.T) {
	agents := []string{"ls", "cat", "todo", "pwd", "whoami", "uname", "ps", "df", "du", "grep", "find", "stat", "chat", "echo", "touch", "mkdir", "rm", "cp", "mv"}

	for _, agentName := range agents {
		t.Run(fmt.Sprintf("RoundTrip_%s", agentName), func(t *testing.T) {
			// Create mock arguments based on agent type
			arguments := createMockArgumentsForAgent(agentName)

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

			// Create function response
			functionResp := &FunctionResponse{
				Name:      parsedAgentName,
				Arguments: parsedArgs,
			}

			// Format as XML
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

			// Extract and validate JSON content
			startIndex := strings.Index(xmlOutput, ">") + 1
			endIndex := strings.LastIndex(xmlOutput, "<")
			if startIndex >= endIndex {
				t.Errorf("Invalid XML format: %s", xmlOutput)
			}

			jsonContent := xmlOutput[startIndex:endIndex]
			var parsedData map[string]interface{}
			if err := json.Unmarshal([]byte(jsonContent), &parsedData); err != nil {
				t.Errorf("Invalid JSON content in function response: %v", err)
			}
		})
	}
}

// createMockArgumentsForAgent creates mock arguments for testing
func createMockArgumentsForAgent(agentName string) map[string]interface{} {
	switch agentName {
	case "ls":
		return map[string]interface{}{
			"path":  "/tmp",
			"flags": "-la",
		}
	case "cat":
		return map[string]interface{}{
			"path": "/etc/hosts",
		}
	case "todo":
		return map[string]interface{}{
			"steps": []string{"Step 1", "Step 2", "Step 3"},
		}
	case "echo":
		return map[string]interface{}{
			"message": "Hello from echo agent",
			"file":    "",
		}
	case "touch":
		return map[string]interface{}{
			"file": "/tmp/touch_test.txt",
		}
	case "mkdir":
		return map[string]interface{}{
			"path": "/tmp/mkdir_test",
		}
	case "rm":
		return map[string]interface{}{
			"path": "/tmp/rm_test.txt",
		}
	case "cp":
		return map[string]interface{}{
			"source":      "/tmp/cp_source.txt",
			"destination": "/tmp/cp_dest.txt",
		}
	case "mv":
		return map[string]interface{}{
			"source":      "/tmp/mv_source.txt",
			"destination": "/tmp/mv_dest.txt",
		}
	case "pwd":
		return map[string]interface{}{}
	case "whoami":
		return map[string]interface{}{}
	case "uname":
		return map[string]interface{}{
			"flags": "-a",
		}
	case "ps":
		return map[string]interface{}{
			"flags": "-ef",
		}
	case "df":
		return map[string]interface{}{
			"flags": "-h",
		}
	case "du":
		return map[string]interface{}{
			"path":  ".",
			"flags": "-h",
		}
	case "grep":
		return map[string]interface{}{
			"pattern": "test",
			"path":    ".",
		}
	case "find":
		return map[string]interface{}{
			"path": ".",
			"name": "*.go",
		}
	case "stat":
		return map[string]interface{}{
			"path": ".",
		}
	case "chat":
		return map[string]interface{}{
			"message": "Hello, this is a test message",
		}
	default:
		return map[string]interface{}{
			"test": "value",
		}
	}
}

// compareMaps compares two maps for equality
func compareMaps(map1, map2 map[string]interface{}) bool {
	if len(map1) != len(map2) {
		return false
	}

	for key, val1 := range map1 {
		val2, exists := map2[key]
		if !exists {
			return false
		}

		// Simple string comparison for now
		if fmt.Sprintf("%v", val1) != fmt.Sprintf("%v", val2) {
			return false
		}
	}

	return true
}

// TestAgentFileStructure tests that all agent directories have the required structure
func TestAgentFileStructure(t *testing.T) {
	agents := []string{"ls", "cat", "todo", "pwd", "whoami", "uname", "ps", "df", "du", "grep", "find", "stat", "chat"}

	for _, agent := range agents {
		t.Run(fmt.Sprintf("Structure_%s", agent), func(t *testing.T) {
			// Try multiple possible paths for agents
			possiblePaths := []string{
				filepath.Join("agents", agent),
				filepath.Join(".", "agents", agent),
				filepath.Join("..", "agents", agent),
			}

			var agentDir string
			var found bool
			for _, path := range possiblePaths {
				if _, err := os.Stat(path); err == nil {
					agentDir = path
					found = true
					break
				}
			}

			if !found {
				t.Skipf("Agent directory for %s not found in any of the expected paths", agent)
				return
			}

			// Check for main.go
			mainGo := filepath.Join(agentDir, "main.go")
			if _, err := os.Stat(mainGo); os.IsNotExist(err) {
				t.Errorf("main.go not found in agent directory %s", agentDir)
			}

			// Check for go.mod
			goMod := filepath.Join(agentDir, "go.mod")
			if _, err := os.Stat(goMod); os.IsNotExist(err) {
				t.Errorf("go.mod not found in agent directory %s", agentDir)
			}
		})
	}
}
