# Agent Testing Framework Documentation

## Overview

This comprehensive testing framework ensures that all AgentForgeEngine agents work correctly with the `<function_response name="agent_name">{JSON}</function_response>` format when communicating with models. The framework provides unit tests, integration tests, and validation utilities without requiring any model dependencies.

## Key Features

### ✅ **Function Response Format Validation**
- Validates XML-like format: `<function_response name="agent_name">{JSON}</function_response>`
- Ensures proper JSON structure within function responses
- Tests round-trip model communication simulation

### ✅ **Comprehensive Agent Testing**
- Unit tests for individual agent functionality
- Interface compliance testing
- Parameter validation and error handling
- Model response integration testing

### ✅ **Model-Independent Testing**
- No dependencies on running models
- Mock model responses for testing
- Complete agent behavior validation

### ✅ **Automated Test Runner**
- Bash script for testing all agents
- Individual agent testing capabilities
- Integration test suite

## Architecture

### Core Components

1. **AgentTestSuite** (`pkg/testing/agent_test_suite.go`)
   - Primary testing utilities
   - Function response format validation
   - Mock model response creation

2. **AgentIntegrationTestSuite** (`pkg/testing/integration_test_suite.go`)
   - Plugin loading capabilities
   - Multi-agent testing
   - Integration test management

3. **Test Files**
   - `pkg/testing/agent_integration_test.go` - Main integration tests
   - Individual agent test files (e.g., `agents/ls/main_test.go`)

## Usage

### Running Tests

#### 1. Integration Tests (Recommended)
```bash
./scripts/test_agents.sh integration
```

#### 2. Test All Agents
```bash
./scripts/test_agents.sh all
```

#### 3. Test Specific Agent
```bash
./scripts/test_agents.sh agent ls
```

#### 4. Run Tests Manually
```bash
# Run integration tests
go test -v ./pkg/testing

# Run specific agent tests
go test -v ./agents/ls
```

### Test Framework API

#### Creating Agent Tests

```go
package main

import (
    "testing"
    "github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
    "github.com/AgentForgeEngine/AgentForgeEngine/pkg/testing"
)

func TestYourAgent_FunctionResponseFormat(t *testing.T) {
    agent := NewYourAgent()
    suite := testing.NewAgentTestSuite(t, agent)
    
    // Test basic interface compliance
    suite.TestAgentInterface()
    
    // Test function response format
    input := interfaces.AgentInput{
        Type: "your_action",
        Payload: map[string]interface{}{
            "param1": "value1",
            "param2": "value2",
        },
    }
    
    suite.TestFunctionResponseFormat(input, "your_agent_name")
}
```

#### Testing Model Response Integration

```go
func TestYourAgent_ModelResponseIntegration(t *testing.T) {
    agent := NewYourAgent()
    
    // Create mock model response
    modelResponse := testing.CreateMockModelResponse("your_agent_name", map[string]interface{}{
        "param1": "value1",
        "param2": "value2",
    })
    
    // Parse the function call
    agentName, arguments, err := testing.ParseFunctionCall(modelResponse.FunctionCall)
    if err != nil {
        t.Fatalf("Failed to parse function call: %v", err)
    }
    
    // Verify agent name
    if agentName != "your_agent_name" {
        t.Errorf("Expected agent name 'your_agent_name', got '%s'", agentName)
    }
    
    // Create input and test processing
    input := interfaces.AgentInput{
        Type:    "your_action",
        Payload: arguments,
    }
    
    ctx := context.Background()
    output, err := agent.Process(ctx, input)
    if err != nil {
        t.Fatalf("Agent processing failed: %v", err)
    }
    
    // Verify success and format response
    if !output.Success {
        t.Errorf("Expected successful processing, got error: %s", output.Error)
    }
    
    // Test function response formatting
    functionResp := &testing.FunctionResponse{
        Name:      agentName,
        Arguments: output.Data,
    }
    
    // Validate XML format
    argsJSON, _ := json.Marshal(functionResp.Arguments)
    xmlOutput := fmt.Sprintf(`<function_response name="%s">%s</function_response>`, 
        functionResp.Name, 
        string(argsJSON))
    
    // Validate XML structure
    expectedTag := fmt.Sprintf(`<function_response name="%s">`, agentName)
    if !strings.Contains(xmlOutput, expectedTag) {
        t.Errorf("Expected function response tag not found in: %s", xmlOutput)
    }
}
```

## Function Response Format

### Expected Format

Agents must respond in this exact format:

```xml
<function_response name="agent_name">{JSON_DATA}</function_response>
```

### Example

```xml
<function_response name="ls">{"output": "file1.txt\nfile2.txt", "files": ["file1.txt", "file2.txt"], "path": "/tmp", "flags": "-la"}</function_response>
```

### Validation Rules

1. **Opening Tag**: Must be `<function_response name="agent_name">`
2. **Closing Tag**: Must be `</function_response>`
3. **JSON Content**: Must be valid JSON within the tags
4. **Agent Name**: Must match the actual agent name

## Agent Interface Compliance

### Required Methods

All agents must implement the `interfaces.Agent` interface:

```go
type Agent interface {
    Name() string
    Initialize(config map[string]interface{}) error
    Process(ctx context.Context, input AgentInput) (AgentOutput, error)
    HealthCheck() error
    Shutdown() error
}
```

### Input/Output Format

#### AgentInput
```go
type AgentInput struct {
    Type     string                 `json:"type"`
    Payload  map[string]interface{} `json:"payload"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

#### AgentOutput
```go
type AgentOutput struct {
    Success bool                   `json:"success"`
    Data    map[string]interface{} `json:"data,omitempty"`
    Error   string                 `json:"error,omitempty"`
}
```

## Test Categories

### 1. Unit Tests
- Individual agent functionality
- Parameter validation
- Error handling
- Interface compliance

### 2. Function Response Tests
- XML format validation
- JSON structure validation
- Round-trip testing

### 3. Integration Tests
- Model response simulation
- Multi-agent coordination
- End-to-end workflow testing

### 4. Error Handling Tests
- Invalid input handling
- Malformed data handling
- Edge case testing

## Mock Model Responses

### Creating Mock Responses

```go
// Simple mock response
modelResponse := testing.CreateMockModelResponse("ls", map[string]interface{}{
    "path": ".",
    "flags": "-la",
})

// This creates: <function_call name="ls">{"path": ".", "flags": "-la"}</function_call>
```

### Parsing Mock Responses

```go
agentName, arguments, err := testing.ParseFunctionCall(modelResponse.FunctionCall)
// agentName = "ls"
// arguments = map[string]interface{}{"path": ".", "flags": "-la"}
```

## Test Results

### Successful Test Output

```
✅ Integration tests passed
✅ Function response format test passed for ls
✅ Unit tests passed for ls
✅ Built ls plugin
```

### Test Coverage

The framework validates:

- ✅ Function call parsing
- ✅ Function response formatting
- ✅ JSON structure validation
- ✅ XML tag validation
- ✅ Agent interface compliance
- ✅ Parameter validation
- ✅ Error handling
- ✅ Model response integration

## Best Practices

### 1. Agent Development
- Use the correct `AgentInput`/`AgentOutput` interface
- Return structured data in the `Data` field
- Handle errors gracefully with proper error messages
- Validate input parameters

### 2. Test Development
- Test both success and failure scenarios
- Validate function response format
- Use mock model responses for integration testing
- Test edge cases and error conditions

### 3. Function Response Format
- Always use the exact XML format
- Ensure JSON is valid and properly structured
- Include relevant data in the response
- Handle errors in the JSON response format

## Troubleshooting

### Common Issues

1. **Interface Mismatch**
   - Ensure agents use `AgentInput`/`AgentOutput` structs
   - Check that payload data is accessed via `input.Payload`

2. **Function Response Format**
   - Verify XML tags match exactly
   - Ensure JSON content is valid
   - Check agent name matches expected

3. **Test Failures**
   - Run tests individually for detailed output
   - Check test logs for specific error messages
   - Verify mock data matches expected format

### Debug Commands

```bash
# Run specific test with verbose output
go test -v -run TestSpecificFunction ./agents/ls

# Run integration tests only
go test -v ./pkg/testing

# Test function response validation
go test -v -run TestFunctionResponse ./pkg/testing
```

## Contributing

When adding new agents:

1. Create agent in `agents/agent_name/`
2. Implement the `Agent` interface correctly
3. Add comprehensive tests using the testing framework
4. Validate function response format
5. Update the test runner script if needed

When modifying the testing framework:

1. Ensure backward compatibility
2. Add tests for new functionality
3. Update documentation
4. Test with existing agents

This framework ensures that all agents work seamlessly with the model communication protocol while maintaining high code quality and reliability.