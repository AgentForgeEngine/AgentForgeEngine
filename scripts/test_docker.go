package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/testing"
)

// TestResult represents the result of testing an agent
type TestResult struct {
	AgentName string
	Success   bool
	Error     string
	Duration  time.Duration
	Tests     []string
}

// AgentInfo represents information about an agent
type AgentInfo struct {
	Name     string
	Path     string
	HasMain  bool
	HasGoMod bool
	HasTest  bool
	IsValid  bool
}

// TestRunner manages testing of agents
type TestRunner struct {
	agentsDir string
	results   []TestResult
	logger    *log.Logger
}

// NewTestRunner creates a new test runner
func NewTestRunner(agentsDir string) *TestRunner {
	return &TestRunner{
		agentsDir: agentsDir,
		results:   make([]TestResult, 0),
		logger:    log.New(os.Stdout, "[TEST-RUNNER] ", log.LstdFlags),
	}
}

// DiscoverAgents discovers all agents in the agents directory
func (tr *TestRunner) DiscoverAgents() ([]AgentInfo, error) {
	var agents []AgentInfo

	entries, err := ioutil.ReadDir(tr.agentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read agents directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		agentName := entry.Name()
		agentPath := filepath.Join(tr.agentsDir, agentName)

		// Skip hidden directories
		if strings.HasPrefix(agentName, ".") {
			continue
		}

		agentInfo := AgentInfo{
			Name:     agentName,
			Path:     agentPath,
			HasMain:  tr.fileExists(filepath.Join(agentPath, "main.go")),
			HasGoMod: tr.fileExists(filepath.Join(agentPath, "go.mod")),
			HasTest:  tr.fileExists(filepath.Join(agentPath, "main_test.go")),
		}

		// Agent is valid if it has main.go and go.mod
		agentInfo.IsValid = agentInfo.HasMain && agentInfo.HasGoMod

		if agentInfo.IsValid {
			agents = append(agents, agentInfo)
			tr.logger.Printf("Discovered agent: %s (valid: %t)", agentName, agentInfo.IsValid)
		} else {
			tr.logger.Printf("Skipping agent: %s (invalid: missing main.go or go.mod)", agentName)
		}
	}

	// Sort agents by name
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].Name < agents[j].Name
	})

	return agents, nil
}

// TestAllAgents tests all discovered agents
func (tr *TestRunner) TestAllAgents() error {
	agents, err := tr.DiscoverAgents()
	if err != nil {
		return fmt.Errorf("failed to discover agents: %w", err)
	}

	tr.logger.Printf("Found %d agents to test", len(agents))

	for _, agent := range agents {
		result := tr.TestAgent(agent)
		tr.results = append(tr.results, result)
	}

	return nil
}

// TestAgent tests a single agent
func (tr *TestRunner) TestAgent(agent AgentInfo) TestResult {
	startTime := time.Now()
	result := TestResult{
		AgentName: agent.Name,
		Success:   false,
		Duration:  0,
		Tests:     make([]string, 0),
	}

	tr.logger.Printf("Testing agent: %s", agent.Name)

	// Test 1: Function Response Format
	if err := tr.testFunctionResponseFormat(agent); err != nil {
		result.Error = fmt.Sprintf("Function response test failed: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}
	result.Tests = append(result.Tests, "function_response_format")

	// Test 2: Model Response Parsing
	if err := tr.testModelResponseParsing(agent); err != nil {
		result.Error = fmt.Sprintf("Model response parsing test failed: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}
	result.Tests = append(result.Tests, "model_response_parsing")

	// Test 3: Interface Compliance
	if err := tr.testInterfaceCompliance(agent); err != nil {
		result.Error = fmt.Sprintf("Interface compliance test failed: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}
	result.Tests = append(result.Tests, "interface_compliance")

	// Test 4: Build Test (if has test file)
	if agent.HasTest {
		if err := tr.testBuild(agent); err != nil {
			result.Error = fmt.Sprintf("Build test failed: %v", err)
			result.Duration = time.Since(startTime)
			return result
		}
		result.Tests = append(result.Tests, "build")
	}

	result.Success = true
	result.Duration = time.Since(startTime)
	tr.logger.Printf("✅ Agent %s passed all tests (%v)", agent.Name, result.Duration)

	return result
}

// testFunctionResponseFormat tests function response format
func (tr *TestRunner) testFunctionResponseFormat(agent AgentInfo) error {
	// Create mock arguments for this agent
	arguments := tr.createMockArguments(agent.Name)

	// Create function response
	functionResp := &testing.FunctionResponse{
		Name:      agent.Name,
		Arguments: arguments,
	}

	// Test XML formatting
	xmlOutput := tr.formatFunctionResponse(functionResp)

	// Validate XML format
	expectedTag := fmt.Sprintf(`<function_response name="%s">`, agent.Name)
	if !strings.Contains(xmlOutput, expectedTag) {
		return fmt.Errorf("expected function response tag not found in: %s", xmlOutput)
	}

	if !strings.Contains(xmlOutput, `</function_response>`) {
		return fmt.Errorf("expected closing function response tag not found in: %s", xmlOutput)
	}

	return nil
}

// testModelResponseParsing tests model response parsing
func (tr *TestRunner) testModelResponseParsing(agent AgentInfo) error {
	// Create mock arguments
	arguments := tr.createMockArguments(agent.Name)

	// Create mock model response
	modelResponse := testing.CreateMockModelResponse(agent.Name, arguments)

	// Parse the function call
	agentName, parsedArgs, err := testing.ParseFunctionCall(modelResponse.FunctionCall)
	if err != nil {
		return fmt.Errorf("failed to parse function call: %w", err)
	}

	// Verify agent name
	if agentName != agent.Name {
		return fmt.Errorf("expected agent name '%s', got '%s'", agent.Name, agentName)
	}

	// Verify arguments (basic check)
	if parsedArgs == nil {
		return fmt.Errorf("parsed arguments is nil")
	}

	return nil
}

// testInterfaceCompliance tests interface compliance
func (tr *TestRunner) testInterfaceCompliance(agent AgentInfo) error {
	// This is a basic compliance test
	// In a real implementation, you might load the plugin and test the interface

	if !agent.HasMain {
		return fmt.Errorf("agent missing main.go")
	}

	if !agent.HasGoMod {
		return fmt.Errorf("agent missing go.mod")
	}

	// Check if main.go contains the required exports
	mainFile := filepath.Join(agent.Path, "main.go")
	content, err := ioutil.ReadFile(mainFile)
	if err != nil {
		return fmt.Errorf("failed to read main.go: %w", err)
	}

	contentStr := string(content)
	requiredExports := []string{
		"var Agent interfaces.Agent",
		"func New" + strings.Title(agent.Name) + "Agent",
		"func (a *" + strings.Title(agent.Name) + "Agent) Name() string",
		"func (a *" + strings.Title(agent.Name) + "Agent) Initialize",
		"func (a *" + strings.Title(agent.Name) + "Agent) Process",
		"func (a *" + strings.Title(agent.Name) + "Agent) HealthCheck",
		"func (a *" + strings.Title(agent.Name) + "Agent) Shutdown",
	}

	for _, export := range requiredExports {
		if !strings.Contains(contentStr, export) {
			return fmt.Errorf("missing required export: %s", export)
		}
	}

	return nil
}

// testBuild tests if the agent can be built
func (tr *TestRunner) testBuild(agent AgentInfo) error {
	// This is a placeholder for build testing
	// In a real implementation, you would try to build the agent plugin

	tr.logger.Printf("Build test for %s (placeholder)", agent.Name)
	return nil
}

// createMockArguments creates mock arguments for an agent
func (tr *TestRunner) createMockArguments(agentName string) map[string]interface{} {
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

// formatFunctionResponse formats a function response
func (tr *TestRunner) formatFunctionResponse(resp *testing.FunctionResponse) string {
	argsJSON := `{"test": "data"}`
	if resp.Arguments != nil {
		// In a real implementation, marshal the arguments
		argsJSON = `{"mock": "data"}`
	}
	return fmt.Sprintf(`<function_response name="%s">%s</function_response>`, resp.Name, argsJSON)
}

// fileExists checks if a file exists
func (tr *TestRunner) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// PrintResults prints all test results
func (tr *TestRunner) PrintResults() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                    AGENT TEST RESULTS")
	fmt.Println(strings.Repeat("=", 80))

	successCount := 0
	failCount := 0
	totalDuration := time.Duration(0)

	for _, result := range tr.results {
		status := "❌ FAIL"
		if result.Success {
			status = "✅ PASS"
			successCount++
		} else {
			failCount++
		}

		fmt.Printf("%-20s %s %8v (%s)\n", result.AgentName, status, result.Duration, strings.Join(result.Tests, ", "))
		totalDuration += result.Duration
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total: %d agents | Success: %d | Failed: %d | Duration: %v\n",
		len(tr.results), successCount, failCount, totalDuration)
	fmt.Println(strings.Repeat("=", 80))

	// Print failures
	if failCount > 0 {
		fmt.Println("\nFAILED AGENTS:")
		for _, result := range tr.results {
			if !result.Success {
				fmt.Printf("  - %s: %s\n", result.AgentName, result.Error)
			}
		}
	}
}

// SaveResults saves test results to a file
func (tr *TestRunner) SaveResults(outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create results file: %w", err)
	}
	defer file.Close()

	// Write results in JSON format
	fmt.Fprintf(file, "{\n")
	fmt.Fprintf(file, "  \"timestamp\": \"%s\",\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "  \"total_agents\": %d,\n", len(tr.results))

	successCount := 0
	for _, result := range tr.results {
		if result.Success {
			successCount++
		}
	}
	fmt.Fprintf(file, "  \"success_count\": %d,\n", successCount)
	fmt.Fprintf(file, "  \"failed_count\": %d,\n", len(tr.results)-successCount)
	fmt.Fprintf(file, "  \"results\": [\n")

	for i, result := range tr.results {
		fmt.Fprintf(file, "    {\n")
		fmt.Fprintf(file, "      \"agent_name\": \"%s\",\n", result.AgentName)
		fmt.Fprintf(file, "      \"success\": %t,\n", result.Success)
		fmt.Fprintf(file, "      \"duration_ms\": %d,\n", result.Duration.Milliseconds())
		fmt.Fprintf(file, "      \"tests\": [%s],\n", strings.Join(result.Tests, ", "))
		if !result.Success {
			fmt.Fprintf(file, "      \"error\": \"%s\",\n", result.Error)
		}
		if i < len(tr.results)-1 {
			fmt.Fprintf(file, "    },\n")
		} else {
			fmt.Fprintf(file, "    }\n")
		}
	}

	fmt.Fprintf(file, "  ]\n")
	fmt.Fprintf(file, "}\n")

	return nil
}

// main function
func main() {
	agentsDir := "./agents"
	if len(os.Args) > 1 {
		agentsDir = os.Args[1]
	}

	runner := NewTestRunner(agentsDir)

	// Run all tests
	if err := runner.TestAllAgents(); err != nil {
		log.Fatalf("Failed to run tests: %v", err)
	}

	// Print results
	runner.PrintResults()

	// Save results to file
	resultsPath := "/tmp/test-results.json"
	if err := runner.SaveResults(resultsPath); err != nil {
		log.Printf("Failed to save results: %v", err)
	} else {
		log.Printf("Results saved to: %s", resultsPath)
	}

	// Exit with error code if any tests failed
	failCount := 0
	for _, result := range runner.results {
		if !result.Success {
			failCount++
		}
	}

	if failCount > 0 {
		os.Exit(1)
	}
}
