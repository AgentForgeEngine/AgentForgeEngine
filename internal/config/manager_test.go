package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManager_Load(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")

	configContent := `
server:
  host: "test-host"
  port: 9090

models:
  - name: "test-model"
    type: "http"
    endpoint: "http://localhost:8080"

agents:
  local:
    - name: "test-agent"
      path: "./test-agent"
  remote:
    - name: "remote-agent"
      repo: "github.com/test/test-agent"
      version: "latest"

recovery:
  hot_reload: false
  max_retries: 5
  backoff_seconds: 10
  health_check_interval: 60
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test loading
	manager := NewManager()
	err = manager.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test server config
	serverConfig := manager.GetServerConfig()
	if serverConfig.Host != "test-host" {
		t.Errorf("Expected host 'test-host', got '%s'", serverConfig.Host)
	}
	if serverConfig.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", serverConfig.Port)
	}

	// Test model configs
	modelConfigs := manager.GetModelConfigs()
	if len(modelConfigs) != 1 {
		t.Errorf("Expected 1 model config, got %d", len(modelConfigs))
	}
	if modelConfigs[0].Name != "test-model" {
		t.Errorf("Expected model name 'test-model', got '%s'", modelConfigs[0].Name)
	}

	// Test agent configs
	agentConfigs := manager.GetAgentConfigs()
	if len(agentConfigs) != 2 {
		t.Errorf("Expected 2 agent configs, got %d", len(agentConfigs))
	}

	// Test recovery config
	recoveryConfig := manager.GetRecoveryConfig()
	if recoveryConfig.HotReload {
		t.Errorf("Expected hot_reload false, got true")
	}
	if recoveryConfig.MaxRetries != 5 {
		t.Errorf("Expected max_retries 5, got %d", recoveryConfig.MaxRetries)
	}
	if recoveryConfig.BackoffSec != 10 {
		t.Errorf("Expected backoff_sec 10, got %d", recoveryConfig.BackoffSec)
	}
	if recoveryConfig.HealthCheck != 60 {
		t.Errorf("Expected health_check 60, got %d", recoveryConfig.HealthCheck)
	}
}

func TestManager_Defaults(t *testing.T) {
	// Test loading with non-existent file (should use defaults)
	manager := NewManager()
	err := manager.Load("/non/existent/file.yaml")
	if err != nil {
		t.Fatalf("Failed to load non-existent config: %v", err)
	}

	// Test default values
	serverConfig := manager.GetServerConfig()
	if serverConfig.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got '%s'", serverConfig.Host)
	}
	if serverConfig.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", serverConfig.Port)
	}

	recoveryConfig := manager.GetRecoveryConfig()
	if !recoveryConfig.HotReload {
		t.Errorf("Expected default hot_reload true, got false")
	}
}
