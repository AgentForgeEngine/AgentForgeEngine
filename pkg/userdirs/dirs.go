package userdirs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// UserDirectories manages the ~/.afe directory structure
type UserDirectories struct {
	HomeDir      string
	AFEDir       string
	ProvidersDir string
	AgentsDir    string
	CacheDir     string
	ConfigDir    string
	LogsDir      string
}

// NewUserDirectories creates a new UserDirectories instance
func NewUserDirectories() (*UserDirectories, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	afeDir := filepath.Join(homeDir, ".afe")

	return &UserDirectories{
		HomeDir:      homeDir,
		AFEDir:       afeDir,
		ProvidersDir: filepath.Join(afeDir, "providers"),
		AgentsDir:    filepath.Join(afeDir, "agents"),
		CacheDir:     filepath.Join(afeDir, "cache"),
		ConfigDir:    filepath.Join(afeDir, "config"),
		LogsDir:      filepath.Join(afeDir, "logs"),
	}, nil
}

// EnsureDirectories creates the ~/.afe directory structure
func (ud *UserDirectories) EnsureDirectories() error {
	dirs := []string{
		ud.AFEDir,
		ud.ProvidersDir,
		ud.AgentsDir,
		ud.CacheDir,
		filepath.Join(ud.CacheDir, "plugin_hashes"),
		filepath.Join(ud.CacheDir, "plugin_hashes", "providers"),
		filepath.Join(ud.CacheDir, "plugin_hashes", "agents"),
		filepath.Join(ud.CacheDir, "build_metadata"),
		ud.ConfigDir,
		ud.LogsDir,
		filepath.Join(ud.AFEDir, "accounts"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// Exists checks if ~/.afe directory structure exists
func (ud *UserDirectories) Exists() bool {
	if _, err := os.Stat(ud.AFEDir); os.IsNotExist(err) {
		return false
	}

	// Check key subdirectories
	keyDirs := []string{ud.ProvidersDir, ud.AgentsDir, ud.CacheDir, ud.ConfigDir}
	for _, dir := range keyDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return false
		}
	}

	return true
}

// GetDefaultConfig returns the default user configuration
func (ud *UserDirectories) GetDefaultConfig() string {
	return fmt.Sprintf(`# AgentForgeEngine User Configuration
# This configuration applies to ALL your projects

build:
  plugins_dir: "%s"
  agents_dir: "%s"
  cache_dir: "%s"
  go_version_min: "1.21"
  build_flags: ["-ldflags=-s -w"]
  parallel_builds: true
  clean_before_build: false
  timeout: 300
  max_workers: 0  # 0 = auto-detect CPU cores

cache:
  enabled: true
  max_size_mb: 100
  retention_days: 30
  auto_cleanup: true
  compression_enabled: false
  validation_interval: "24h"

logging:
  build_log: "%s"
  cache_log: "%s"
  verbose: false
  max_log_size_mb: 10
`, ud.ProvidersDir, ud.AgentsDir, ud.CacheDir,
		filepath.Join(ud.LogsDir, "build.log"),
		filepath.Join(ud.LogsDir, "cache.log"))
}

// CreateDefaultConfig creates the default configuration file
func (ud *UserDirectories) CreateDefaultConfig() error {
	configPath := filepath.Join(ud.ConfigDir, "build_config.yaml")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // Config already exists
	}

	configContent := ud.GetDefaultConfig()

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create default config: %w", err)
	}

	return nil
}

// GetCachePath returns the path to the main build cache file
func (ud *UserDirectories) GetCachePath() string {
	return filepath.Join(ud.CacheDir, "build_cache.yaml")
}

// GetPluginHashPath returns the path to a plugin's hash file
func (ud *UserDirectories) GetPluginHashPath(pluginType, pluginName string) string {
	return filepath.Join(ud.CacheDir, "plugin_hashes", pluginType, pluginName+".yaml")
}

// GetBuildMetadataPath returns the path to build metadata files
func (ud *UserDirectories) GetBuildMetadataPath(filename string) string {
	return filepath.Join(ud.CacheDir, "build_metadata", filename)
}

// GetPluginOutputPath returns the path where a plugin should be built
func (ud *UserDirectories) GetPluginOutputPath(pluginType, pluginName string) string {
	if pluginType == "provider" {
		return filepath.Join(ud.ProvidersDir, pluginName+".so")
	}
	return filepath.Join(ud.AgentsDir, pluginName+".so")
}

// GetSystemInfo returns system information for build configuration
func (ud *UserDirectories) GetSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"cpu_cores":  runtime.NumCPU(),
		"go_root":    os.Getenv("GOROOT"),
		"gopath":     os.Getenv("GOPATH"),
		"home_dir":   ud.HomeDir,
		"afe_dir":    ud.AFEDir,
		"created_at": time.Now().Format(time.RFC3339),
	}
}

// MigrateFromOldStructure migrates plugins from old ./plugins/ directory
func (ud *UserDirectories) MigrateFromOldStructure(projectDir string) (*MigrationResult, error) {
	result := &MigrationResult{
		StartTime: time.Now(),
	}

	oldPluginsDir := filepath.Join(projectDir, "plugins")

	// Check if old plugins directory exists
	if _, err := os.Stat(oldPluginsDir); os.IsNotExist(err) {
		result.Success = true
		result.Message = "No old plugins directory found - nothing to migrate"
		result.EndTime = time.Now()
		return result, nil
	}

	// Find all .so files in old plugins directory
	soFiles, err := filepath.Glob(filepath.Join(oldPluginsDir, "*.so"))
	if err != nil {
		return nil, fmt.Errorf("failed to scan old plugins directory: %w", err)
	}

	if len(soFiles) == 0 {
		result.Success = true
		result.Message = "No plugin files found in old plugins directory"
		result.EndTime = time.Now()
		return result, nil
	}

	// Migrate each .so file
	for _, soFile := range soFiles {
		filename := filepath.Base(soFile)

		// Determine if it's a provider or agent based on naming convention
		var targetDir string
		if strings.Contains(filename, "provider") ||
			strings.Contains(filename, "bridge") ||
			strings.Contains(filename, "qwen3") {
			targetDir = ud.ProvidersDir
			result.ProvidersMigrated = append(result.ProvidersMigrated, filename)
		} else {
			targetDir = ud.AgentsDir
			result.AgentsMigrated = append(result.AgentsMigrated, filename)
		}

		targetPath := filepath.Join(targetDir, filename)

		// Copy the file
		if err := ud.copyFile(soFile, targetPath); err != nil {
			return nil, fmt.Errorf("failed to copy plugin %s: %w", filename, err)
		}

		result.FilesMigrated++
	}

	result.Success = true
	result.Message = fmt.Sprintf("Successfully migrated %d plugin files", result.FilesMigrated)
	result.EndTime = time.Now()

	return result, nil
}

// copyFile copies a file from src to dst
func (ud *UserDirectories) copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}

// MigrationResult contains the results of a migration operation
type MigrationResult struct {
	StartTime         time.Time
	EndTime           time.Time
	Success           bool
	Message           string
	FilesMigrated     int
	ProvidersMigrated []string
	AgentsMigrated    []string
}

// Duration returns the duration of the migration
func (mr *MigrationResult) Duration() time.Duration {
	return mr.EndTime.Sub(mr.StartTime)
}
