package loader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"runtime"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type Manager struct {
	registry   map[string]interfaces.Agent
	providers  map[string]interfaces.Provider
	pluginsDir string
	tempDir    string
}

func NewManager(pluginsDir, tempDir string) *Manager {
	if pluginsDir == "" {
		pluginsDir = "./plugins"
	}
	if tempDir == "" {
		tempDir = "./temp"
	}

	return &Manager{
		registry:   make(map[string]interfaces.Agent),
		providers:  make(map[string]interfaces.Provider),
		pluginsDir: pluginsDir,
		tempDir:    tempDir,
	}
}

func (pm *Manager) LoadLocalAgent(path, name string) error {
	// Build the plugin
	outputPath := filepath.Join(pm.pluginsDir, name+".so")
	if err := pm.buildPlugin(path, outputPath); err != nil {
		return fmt.Errorf("failed to build plugin %s: %w", name, err)
	}

	// Load the plugin
	return pm.loadPlugin(outputPath, name)
}

func (pm *Manager) LoadLocalProvider(path, name string) error {
	// Build the provider plugin
	outputPath := filepath.Join(pm.pluginsDir, name+"-provider.so")
	if err := pm.buildPlugin(path, outputPath); err != nil {
		return fmt.Errorf("failed to build provider plugin %s: %w", name, err)
	}

	// Load the provider
	return pm.loadPlugin(outputPath, name)
}

func (pm *Manager) LoadRemoteAgent(repo, version, entrypoint string) error {
	// Create temp directory for the repository
	tempDir := filepath.Join(pm.tempDir, "agent-"+filepath.Base(repo))
	if err := os.RemoveAll(tempDir); err != nil {
		return fmt.Errorf("failed to clean temp directory: %w", err)
	}
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Clone the repository
	if err := pm.cloneRepository(repo, tempDir, version); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Determine entrypoint
	if entrypoint == "" {
		entrypoint = "main.go"
	}

	// Build the plugin
	agentName := filepath.Base(repo)
	outputPath := filepath.Join(pm.pluginsDir, agentName+".so")
	if err := pm.buildPlugin(filepath.Join(tempDir, entrypoint), outputPath); err != nil {
		return fmt.Errorf("failed to build remote plugin: %w", err)
	}

	// Load the plugin
	return pm.loadPlugin(outputPath, agentName)
}

func (pm *Manager) GetAgent(name string) (interfaces.Agent, bool) {
	agent, exists := pm.registry[name]
	return agent, exists
}

func (pm *Manager) GetProvider(name string) (interfaces.Provider, bool) {
	provider, exists := pm.providers[name]
	return provider, exists
}

func (pm *Manager) ListAgents() []string {
	var agents []string
	for name := range pm.registry {
		agents = append(agents, name)
	}
	return agents
}

func (pm *Manager) ListProviders() []string {
	var providers []string
	for name := range pm.providers {
		providers = append(providers, name)
	}
	return providers
}

func (pm *Manager) UnloadAgent(name string) error {
	if _, exists := pm.registry[name]; !exists {
		return fmt.Errorf("agent %s not found", name)
	}

	delete(pm.registry, name)
	return nil
}

func (pm *Manager) UnloadProvider(name string) error {
	if _, exists := pm.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	provider := pm.providers[name]
	if err := provider.Shutdown(); err != nil {
		fmt.Printf("Error shutting down provider %s: %v", name, err)
	}

	delete(pm.providers, name)
	return nil
}

func (pm *Manager) ReloadAgent(name string) error {
	// For now, just unload - the calling code should handle reloading
	return pm.UnloadAgent(name)
}

func (pm *Manager) ReloadProvider(name string) error {
	// For now, just unload - the calling code should handle reloading
	return pm.UnloadProvider(name)
}

// LoadPluginFromFile loads a plugin from a specific file path
func (pm *Manager) LoadPluginFromFile(pluginPath, pluginName string) error {
	return pm.loadPlugin(pluginPath, pluginName)
}

// AddProviderToRegistry adds a provider to the registry (for hot reload)
func (pm *Manager) AddProviderToRegistry(name string, provider interfaces.Provider) {
	pm.providers[name] = provider
}

// AddAgentToRegistry adds an agent to the registry (for hot reload)
func (pm *Manager) AddAgentToRegistry(name string, agent interfaces.Agent) {
	pm.registry[name] = agent
}

func (pm *Manager) buildPlugin(source, output string) error {
	// Ensure plugins directory exists
	if err := os.MkdirAll(pm.pluginsDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugins directory: %w", err)
	}

	// Build the plugin
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", output, source)

	// Set environment variables for consistent build
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		"GOOS="+runtime.GOOS,
		"GOARCH="+runtime.GOARCH,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build failed: %w, output: %s", err, string(output))
	}

	return nil
}

func (pm *Manager) loadPlugin(path, name string) error {
	// Open the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look for the Agent symbol first
	symAgent, err := p.Lookup("Agent")
	if err == nil {
		// Type assert to Agent interface
		if agent, ok := symAgent.(interfaces.Agent); ok {
			// Register the agent
			pm.registry[name] = agent
			fmt.Printf("Successfully loaded agent: %s", name)
			return nil
		}
		return fmt.Errorf("invalid Agent type in plugin")
	}

	// Try Provider symbol if Agent not found
	symProvider, providerErr := p.Lookup("Provider")
	if providerErr != nil {
		return fmt.Errorf("plugin missing Agent and Provider symbols: %w", err)
	}

	// Type assert to Provider interface
	provider, ok := symProvider.(interfaces.Provider)
	if !ok {
		return fmt.Errorf("invalid Provider type in plugin")
	}

	// Register the provider
	pm.providers[name] = provider
	fmt.Printf("Successfully loaded provider: %s", name)
	return nil
}

func (pm *Manager) cloneRepository(repo, tempDir, version string) error {
	// Use git to clone the repository
	var args []string

	if version == "" || version == "latest" {
		args = []string{"clone", repo, tempDir}
	} else {
		args = []string{"clone", "--depth", "1", "--branch", version, repo, tempDir}
	}

	cmd := exec.Command("git", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %w, output: %s", err, string(output))
	}

	return nil
}

func (pm *Manager) downloadGitHubFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (pm *Manager) validateCompatibility(dir string) error {
	// Check for go.mod file
	goModPath := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return fmt.Errorf("go.mod file not found in repository")
	}

	// TODO: Add more compatibility checks
	// - Go version compatibility
	// - Dependency version checks
	// - Interface implementation verification

	return nil
}

// Health check all loaded agents and providers
func (pm *Manager) HealthCheckAll(ctx context.Context) map[string]error {
	results := make(map[string]error)

	// Check agents
	for name, agent := range pm.registry {
		select {
		case <-ctx.Done():
			results[name] = ctx.Err()
		default:
			results[name] = agent.HealthCheck()
		}
	}

	// Check providers
	for name, provider := range pm.providers {
		select {
		case <-ctx.Done():
			results[name] = ctx.Err()
		default:
			results[name] = provider.HealthCheck()
		}
	}

	return results
}
