package hotreload

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/internal/loader"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/userdirs"
)

// Manager handles hot reloading of plugins
type Manager struct {
	pluginManager   *loader.Manager
	userDirs        *userdirs.UserDirectories
	reloadQueue     chan ReloadRequest
	reloadWorkers   int
	active          bool
	mu              sync.RWMutex
	reloadCallbacks map[string]func(string, error)
}

// ReloadRequest represents a hot reload request
type ReloadRequest struct {
	PluginType string // "agent" or "provider"
	PluginName string
	PluginPath string
	Force      bool
	Timestamp  time.Time
}

// ReloadResult represents the result of a hot reload operation
type ReloadResult struct {
	PluginName string
	PluginType string
	Success    bool
	Error      error
	Duration   time.Duration
	Restarted  bool
	OldVersion string
	NewVersion string
}

// NewManager creates a new hot reload manager
func NewManager(pluginManager *loader.Manager, userDirs *userdirs.UserDirectories) *Manager {
	return &Manager{
		pluginManager:   pluginManager,
		userDirs:        userDirs,
		reloadQueue:     make(chan ReloadRequest, 100),
		reloadWorkers:   4, // Number of concurrent reload workers
		active:          false,
		reloadCallbacks: make(map[string]func(string, error)),
	}
}

// Start starts the hot reload manager
func (hrm *Manager) Start() error {
	hrm.mu.Lock()
	defer hrm.mu.Unlock()

	if hrm.active {
		return fmt.Errorf("hot reload manager is already active")
	}

	hrm.active = true

	// Start reload workers
	for i := 0; i < hrm.reloadWorkers; i++ {
		go hrm.reloadWorker(i)
	}

	log.Println("🔄 Hot reload manager started")
	return nil
}

// Stop stops the hot reload manager
func (hrm *Manager) Stop() {
	hrm.mu.Lock()
	defer hrm.mu.Unlock()

	if !hrm.active {
		return
	}

	hrm.active = false
	close(hrm.reloadQueue)

	log.Println("🔄 Hot reload manager stopped")
}

// ReloadPlugin queues a plugin for hot reloading
func (hrm *Manager) ReloadPlugin(pluginType, pluginName string, force bool) error {
	if !hrm.active {
		return fmt.Errorf("hot reload manager is not active")
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	pluginPath := filepath.Join(cwd, pluginType+"s", pluginName)

	request := ReloadRequest{
		PluginType: pluginType,
		PluginName: pluginName,
		PluginPath: pluginPath,
		Force:      force,
		Timestamp:  time.Now(),
	}

	select {
	case hrm.reloadQueue <- request:
		log.Printf("🔄 Queued hot reload for %s %s", pluginType, pluginName)
		return nil
	default:
		return fmt.Errorf("reload queue is full")
	}
}

// ReloadPlugins queues multiple plugins for hot reloading
func (hrm *Manager) ReloadPlugins(pluginNames []string, pluginType string) error {
	for _, pluginName := range pluginNames {
		if err := hrm.ReloadPlugin(pluginType, pluginName, false); err != nil {
			return fmt.Errorf("failed to queue reload for %s: %w", pluginName, err)
		}
	}
	return nil
}

// reloadWorker processes reload requests
func (hrm *Manager) reloadWorker(workerID int) {
	log.Printf("🔄 Hot reload worker %d started", workerID)

	for request := range hrm.reloadQueue {
		result := hrm.processReload(request)
		hrm.logReloadResult(result)
		hrm.notifyCallbacks(request.PluginName, result.Error)
	}

	log.Printf("🔄 Hot reload worker %d stopped", workerID)
}

// processReload processes a single reload request
func (hrm *Manager) processReload(request ReloadRequest) ReloadResult {
	startTime := time.Now()

	result := ReloadResult{
		PluginName: request.PluginName,
		PluginType: request.PluginType,
		Success:    false,
		Duration:   0,
		Restarted:  false,
	}

	log.Printf("🔄 Processing hot reload for %s %s", request.PluginType, request.PluginName)

	// Get old plugin info
	oldVersion := hrm.getPluginVersion(request.PluginType, request.PluginName)

	// Step 1: Unload the existing plugin
	if err := hrm.unloadPlugin(request.PluginType, request.PluginName); err != nil {
		result.Error = fmt.Errorf("failed to unload plugin: %w", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// Step 2: Load the new plugin
	if err := hrm.loadPlugin(request.PluginType, request.PluginName, request.PluginPath); err != nil {
		result.Error = fmt.Errorf("failed to load plugin: %w", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// Step 3: Get new plugin info
	result.NewVersion = hrm.getPluginVersion(request.PluginType, request.PluginName)
	result.OldVersion = oldVersion
	result.Restarted = true
	result.Success = true
	result.Duration = time.Since(startTime)

	log.Printf("✅ Successfully hot reloaded %s %s (%v)",
		request.PluginType, request.PluginName, result.Duration)

	return result
}

// unloadPlugin unloads a plugin
func (hrm *Manager) unloadPlugin(pluginType, pluginName string) error {
	switch pluginType {
	case "agent":
		return hrm.pluginManager.UnloadAgent(pluginName)
	case "provider":
		return hrm.pluginManager.UnloadProvider(pluginName)
	default:
		return fmt.Errorf("unknown plugin type: %s", pluginType)
	}
}

// loadPlugin loads a plugin from the user directories
func (hrm *Manager) loadPlugin(pluginType, pluginName, pluginPath string) error {
	// Get the output path from user directories
	outputPath := hrm.userDirs.GetPluginOutputPath(pluginType, pluginName)

	// Use the plugin manager's LoadPluginFromFile method
	return hrm.pluginManager.LoadPluginFromFile(outputPath, pluginName)
}

// getPluginVersion gets version information for a plugin
func (hrm *Manager) getPluginVersion(pluginType, pluginName string) string {
	// For now, return a simple timestamp-based version
	// TODO: Implement proper version tracking from build cache
	return fmt.Sprintf("v%d", time.Now().Unix())
}

// logReloadResult logs the result of a hot reload operation
func (hrm *Manager) logReloadResult(result ReloadResult) {
	if result.Success {
		log.Printf("✅ Hot reload successful: %s %s (%v)",
			result.PluginType, result.PluginName, result.Duration)
		if result.Restarted {
			log.Printf("🔄 Plugin restarted: %s -> %s",
				result.OldVersion, result.NewVersion)
		}
	} else {
		log.Printf("❌ Hot reload failed: %s %s (%v): %v",
			result.PluginType, result.PluginName, result.Duration, result.Error)
	}
}

// notifyCallbacks notifies registered callbacks about reload results
func (hrm *Manager) notifyCallbacks(pluginName string, err error) {
	hrm.mu.RLock()
	defer hrm.mu.RUnlock()

	for _, callback := range hrm.reloadCallbacks {
		// Call the callback - it returns an error but we handle it internally
		callback(pluginName, err)
	}
}

// RegisterCallback registers a callback for reload events
func (hrm *Manager) RegisterCallback(id string, callback func(string, error)) {
	hrm.mu.Lock()
	defer hrm.mu.Unlock()

	hrm.reloadCallbacks[id] = callback
	log.Printf("🔄 Registered reload callback: %s", id)
}

// UnregisterCallback unregisters a reload callback
func (hrm *Manager) UnregisterCallback(id string) {
	hrm.mu.Lock()
	defer hrm.mu.Unlock()

	delete(hrm.reloadCallbacks, id)
	log.Printf("🔄 Unregistered reload callback: %s", id)
}

// GetStatus returns the current status of the hot reload manager
func (hrm *Manager) GetStatus() *HotReloadStatus {
	hrm.mu.RLock()
	defer hrm.mu.RUnlock()

	return &HotReloadStatus{
		Active:              hrm.active,
		QueueLength:         len(hrm.reloadQueue),
		Workers:             hrm.reloadWorkers,
		RegisteredCallbacks: len(hrm.reloadCallbacks),
		LoadedAgents:        hrm.pluginManager.ListAgents(),
		LoadedProviders:     hrm.pluginManager.ListProviders(),
	}
}

// HotReloadStatus contains the status of the hot reload manager
type HotReloadStatus struct {
	Active              bool
	QueueLength         int
	Workers             int
	RegisteredCallbacks int
	LoadedAgents        []string
	LoadedProviders     []string
}

// IsReloadable checks if a plugin can be hot reloaded
func (hrm *Manager) IsReloadable(pluginType, pluginName string) bool {
	switch pluginType {
	case "agent":
		_, exists := hrm.pluginManager.GetAgent(pluginName)
		return exists
	case "provider":
		_, exists := hrm.pluginManager.GetProvider(pluginName)
		return exists
	default:
		return false
	}
}
