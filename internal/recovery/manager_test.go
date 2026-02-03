package recovery

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/internal/loader"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type MockPluginManager struct {
	unloadCalled     bool
	loadLocalCalled  bool
	loadRemoteCalled bool
	loadedName       string
	loadedRepo       string
	loadedVersion    string
}

func (m *MockPluginManager) LoadLocalAgent(path, name string) error {
	m.loadLocalCalled = true
	m.loadedName = name
	return nil
}

func (m *MockPluginManager) LoadRemoteAgent(repo, version, entrypoint string) error {
	m.loadRemoteCalled = true
	m.loadedRepo = repo
	m.loadedVersion = version
	return nil
}

func (m *MockPluginManager) GetAgent(name string) (interfaces.Agent, bool) {
	return nil, false
}

func (m *MockPluginManager) ListAgents() []string {
	return []string{}
}

func (m *MockPluginManager) UnloadAgent(name string) error {
	m.unloadCalled = true
	return nil
}

func (m *MockPluginManager) ReloadAgent(name string) error {
	return nil
}

func (m *MockPluginManager) HealthCheckAll(ctx context.Context) map[string]error {
	return map[string]error{}
}

func TestManager_CalculateBackoff(t *testing.T) {
	config := interfaces.RecoveryConfig{
		BackoffSec: 5,
	}

	// Create a real plugin manager but we'll use it for the test
	realManager := loader.NewManager("", "")
	manager := NewManager(config, realManager)

	// Test exponential backoff
	tests := []struct {
		retryCount int
		expected   int // minimum expected seconds
	}{
		{0, 5},
		{1, 10},
		{2, 20},
		{3, 40},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("retry_%d", test.retryCount), func(t *testing.T) {
			manager.retryCounts["test"] = test.retryCount
			backoff := manager.calculateBackoff("test")

			expectedDuration := time.Duration(test.expected) * time.Second
			if backoff < expectedDuration {
				t.Errorf("Expected backoff >= %v, got %v", expectedDuration, backoff)
			}
		})
	}
}

func TestManager_ShouldRetry(t *testing.T) {
	config := interfaces.RecoveryConfig{
		MaxRetries: 3,
	}

	realManager := loader.NewManager("", "")
	manager := NewManager(config, realManager)

	// Test retry logic
	tests := []struct {
		retryCount int
		expected   bool
	}{
		{0, true},
		{1, true},
		{2, true},
		{3, false}, // Max retries reached
		{4, false},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("retry_%d", test.retryCount), func(t *testing.T) {
			manager.retryCounts["test"] = test.retryCount
			shouldRetry := manager.shouldRetry("test")

			if shouldRetry != test.expected {
				t.Errorf("Expected %v, got %v for retry count %d", test.expected, shouldRetry, test.retryCount)
			}
		})
	}
}

func TestManager_GetStatus(t *testing.T) {
	config := interfaces.RecoveryConfig{
		HotReload:   true,
		MaxRetries:  3,
		BackoffSec:  5,
		HealthCheck: 30,
	}

	realManager := loader.NewManager("", "")
	manager := NewManager(config, realManager)

	// Add some retry data
	manager.retryCounts["test-agent"] = 2
	manager.lastFailureTime["test-agent"] = time.Now()

	status := manager.GetStatus()

	if status["enabled"] != true {
		t.Errorf("Expected enabled=true, got %v", status["enabled"])
	}

	if status["max_retries"] != 3 {
		t.Errorf("Expected max_retries=3, got %v", status["max_retries"])
	}

	agents, ok := status["agents"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected agents to be a map")
	}

	agentStatus, ok := agents["test-agent"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected test-agent status to be a map")
	}

	if agentStatus["retry_count"] != 2 {
		t.Errorf("Expected retry_count=2, got %v", agentStatus["retry_count"])
	}
}
