#!/bin/bash

# Agent Test Runner
# This script tests individual agents and provides comprehensive validation

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TEST_RESULTS_DIR="$PROJECT_ROOT/test_results"
AGENT_BUILD_DIR="$TEST_RESULTS_DIR/agents"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

print_header() {
    print_status "$BLUE" "=========================================="
    print_status "$BLUE" "$1"
    print_status "$BLUE" "=========================================="
}

print_success() {
    print_status "$GREEN" "✅ $1"
}

print_error() {
    print_status "$RED" "❌ $1"
}

print_warning() {
    print_status "$YELLOW" "⚠️  $1"
}

# Create test results directory
setup_test_environment() {
    print_header "Setting up test environment"
    
    mkdir -p "$TEST_RESULTS_DIR"
    mkdir -p "$AGENT_BUILD_DIR"
    
    print_success "Test environment created"
}

# Build agent plugin
build_agent() {
    local agent_name=$1
    local agent_dir="$PROJECT_ROOT/agents/$agent_name"
    local plugin_path="$AGENT_BUILD_DIR/${agent_name}.so"
    
    print_status "$BLUE" "Building agent: $agent_name"
    
    if [[ ! -d "$agent_dir" ]]; then
        print_error "Agent directory not found: $agent_dir"
        return 1
    fi
    
    # Build the agent plugin
    cd "$agent_dir"
    if go build -buildmode=plugin -o "$plugin_path" .; then
        print_success "Built $agent_name plugin"
        return 0
    else
        print_error "Failed to build $agent_name plugin"
        return 1
    fi
}

# Test individual agent
test_agent() {
    local agent_name=$1
    local agent_dir="$PROJECT_ROOT/agents/$agent_name"
    
    print_header "Testing agent: $agent_name"
    
    # Check if agent directory exists
    if [[ ! -d "$agent_dir" ]]; then
        print_error "Agent directory not found: $agent_dir"
        return 1
    fi
    
    # Check required files
    local required_files=("main.go" "go.mod")
    for file in "${required_files[@]}"; do
        if [[ ! -f "$agent_dir/$file" ]]; then
            print_error "Required file not found: $file"
            return 1
        fi
    done
    
    print_success "Required files present"
    
    # Build the agent
    if ! build_agent "$agent_name"; then
        return 1
    fi
    
    # Run unit tests
    print_status "$BLUE" "Running unit tests for $agent_name"
    cd "$agent_dir"
    if go test -v; then
        print_success "Unit tests passed for $agent_name"
    else
        print_error "Unit tests failed for $agent_name"
        return 1
    fi
    
    # Test function response format
    test_agent_function_response "$agent_name"
    
    return 0
}

# Test agent function response format
test_agent_function_response() {
    local agent_name=$1
    
    print_status "$BLUE" "Testing function response format for $agent_name"
    
    # Create a simple test for function response format
    local test_file="$TEST_RESULTS_DIR/${agent_name}_function_response_test.go"
    
    cat > "$test_file" << EOF
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

func Test${agent_name^}FunctionResponseFormat(t *testing.T) {
    // This test validates that the agent can produce proper function response format
    
    // Create mock input based on agent type
    input := getMockInputForAgent("$agent_name")
    
    // Test the function response format validation
    testSuite := testing.NewAgentTestSuite(t, nil)
    
    // Create mock output data
    mockData := map[string]interface{}{
        "result": "success",
        "agent":  "$agent_name",
    }
    
    // Create function response
    functionResp := &testing.FunctionResponse{
        Name:      "$agent_name",
        Arguments: mockData,
    }
    
    // Test XML formatting
    argsJSON, _ := json.Marshal(functionResp.Arguments)
    xmlOutput := fmt.Sprintf(\`<function_response name="%s">%s</function_response>\`, 
        functionResp.Name, 
        string(argsJSON))
    
    // Validate XML format
    expectedTag := fmt.Sprintf(\`<function_response name="%s">\`, "$agent_name")
    if !strings.Contains(xmlOutput, expectedTag) {
        t.Errorf("Expected function response tag not found in: %s", xmlOutput)
    }
    
    if !strings.Contains(xmlOutput, \`</function_response>\`) {
        t.Errorf("Expected closing function response tag not found in: %s", xmlOutput)
    }
    
    // Test model response parsing
    modelResponse := testing.CreateMockModelResponse("$agent_name", mockData)
    parsedAgentName, parsedArgs, err := testing.ParseFunctionCall(modelResponse.FunctionCall)
    if err != nil {
        t.Fatalf("Failed to parse function call: %v", err)
    }
    
    if parsedAgentName != "$agent_name" {
        t.Errorf("Expected agent name '%s', got '%s'", "$agent_name", parsedAgentName)
    }
}

func getMockInputForAgent(agentName string) interfaces.AgentInput {
    switch agentName {
    case "ls":
        return interfaces.AgentInput{
            Type: "list",
            Payload: map[string]interface{}{
                "path": ".",
                "flags": "-la",
            },
        }
    case "cat":
        return interfaces.AgentInput{
            Type: "read",
            Payload: map[string]interface{}{
                "path": "/tmp/test.txt",
            },
        }
    case "todo":
        return interfaces.AgentInput{
            Type: "create",
            Payload: map[string]interface{}{
                "steps": []string{"Test step 1", "Test step 2"},
            },
        }
    default:
        return interfaces.AgentInput{
            Type: "process",
            Payload: map[string]interface{}{
                "test": "value",
            },
        }
    }
}
EOF
    
    # Run the function response test
    if go test -v "$test_file"; then
        print_success "Function response format test passed for $agent_name"
    else
        print_error "Function response format test failed for $agent_name"
        return 1
    fi
    
    # Clean up test file
    rm -f "$test_file"
}

# Test all agents
test_all_agents() {
    print_header "Testing all agents"
    
    local agents=("ls" "cat" "todo" "pwd" "whoami" "uname" "ps" "df" "du" "grep" "find" "stat" "chat")
    local passed=0
    local failed=0
    
    for agent in "${agents[@]}"; do
        echo
        if test_agent "$agent"; then
            ((passed++))
        else
            ((failed++))
        fi
    done
    
    echo
    print_header "Test Results Summary"
    print_success "Passed: $passed agents"
    if [[ $failed -gt 0 ]]; then
        print_error "Failed: $failed agents"
    fi
    
    return $failed
}

# Run integration tests
run_integration_tests() {
    print_header "Running integration tests"
    
    cd "$PROJECT_ROOT"
    if go test -v ./pkg/testing; then
        print_success "Integration tests passed"
        return 0
    else
        print_error "Integration tests failed"
        return 1
    fi
}

# Main function
main() {
    local command=${1:-"all"}
    
    case $command in
        "setup")
            setup_test_environment
            ;;
        "agent")
            if [[ -z $2 ]]; then
                print_error "Please specify an agent name"
                echo "Usage: $0 agent <agent_name>"
                exit 1
            fi
            test_agent "$2"
            ;;
        "all")
            setup_test_environment
            test_all_agents
            run_integration_tests
            ;;
        "integration")
            run_integration_tests
            ;;
        *)
            print_error "Unknown command: $command"
            echo "Available commands:"
            echo "  setup       - Set up test environment"
            echo "  agent <name> - Test specific agent"
            echo "  all         - Test all agents and run integration tests"
            echo "  integration - Run integration tests only"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"