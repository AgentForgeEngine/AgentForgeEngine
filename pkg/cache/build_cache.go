package cache

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/userdirs"
	"gopkg.in/yaml.v3"
)

// BuildCache represents the main build cache structure
type BuildCache struct {
	Version       string              `yaml:"version"`
	CreatedAt     time.Time           `yaml:"created_at"`
	UpdatedAt     time.Time           `yaml:"updated_at"`
	AFEVersion    string              `yaml:"afe_version"`
	GoVersion     string              `yaml:"go_version"`
	Statistics    BuildStatistics     `yaml:"statistics"`
	Plugins       PluginRegistry      `yaml:"plugins"`
	BuildHistory  []BuildHistoryEntry `yaml:"build_history"`
	CacheSettings CacheSettings       `yaml:"cache_settings"`
	Integrity     IntegrityValidation `yaml:"integrity"`
}

// BuildStatistics tracks global build statistics
type BuildStatistics struct {
	TotalBuilds        int     `yaml:"total_builds"`
	CacheHits          int     `yaml:"cache_hits"`
	CacheMisses        int     `yaml:"cache_misses"`
	CacheHitRate       float64 `yaml:"cache_hit_rate"`
	TotalBuildTimeMs   int     `yaml:"total_build_time_ms"`
	AverageBuildTimeMs int     `yaml:"average_build_time_ms"`
	TotalCacheSizeMb   float64 `yaml:"total_cache_size_mb"`
}

// PluginRegistry contains all plugin information
type PluginRegistry struct {
	Providers map[string]PluginEntry `yaml:"providers"`
	Agents    map[string]PluginEntry `yaml:"agents"`
}

// PluginEntry represents a single plugin's cache entry
type PluginEntry struct {
	BuildInfo    PluginBuildInfo `yaml:"build_info"`
	SourceFiles  []SourceFile    `yaml:"source_files"`
	Dependencies []Dependency    `yaml:"dependencies"`
}

// PluginBuildInfo contains build information for a plugin
type PluginBuildInfo struct {
	SourceHash      string    `yaml:"source_hash"`
	GoModHash       string    `yaml:"go_mod_hash"`
	BuildConfigHash string    `yaml:"build_config_hash"`
	OutputPath      string    `yaml:"output_path"`
	OutputHash      string    `yaml:"output_hash"`
	BuildTime       time.Time `yaml:"build_time"`
	BuildDurationMs int       `yaml:"build_duration_ms"`
	PluginSizeBytes int64     `yaml:"plugin_size_bytes"`
	NeedsRebuild    bool      `yaml:"needs_rebuild"`
	LastUsed        time.Time `yaml:"last_used"`
	BuildCount      int       `yaml:"build_count"`
	CacheValid      bool      `yaml:"cache_valid"`
}

// SourceFile represents a source file in a plugin
type SourceFile struct {
	Path      string    `yaml:"path"`
	Hash      string    `yaml:"hash"`
	SizeBytes int       `yaml:"size_bytes"`
	Modified  time.Time `yaml:"modified"`
}

// Dependency represents a Go module dependency
type Dependency struct {
	Module  string `yaml:"module"`
	Version string `yaml:"version"`
	Hash    string `yaml:"hash"`
	Type    string `yaml:"type"`
}

// BuildHistoryEntry represents a single build operation
type BuildHistoryEntry struct {
	BuildID         string    `yaml:"build_id"`
	Timestamp       time.Time `yaml:"timestamp"`
	Command         string    `yaml:"command"`
	PluginsBuilt    []string  `yaml:"plugins_built"`
	PluginsCached   []string  `yaml:"plugins_cached"`
	TotalDurationMs int       `yaml:"total_duration_ms"`
	Success         bool      `yaml:"success"`
	CacheHitRate    float64   `yaml:"cache_hit_rate"`
}

// CacheSettings contains cache management settings
type CacheSettings struct {
	MaxSizeMb          int       `yaml:"max_size_mb"`
	RetentionDays      int       `yaml:"retention_days"`
	AutoCleanup        bool      `yaml:"auto_cleanup"`
	CompressionEnabled bool      `yaml:"compression_enabled"`
	ValidationInterval string    `yaml:"validation_interval"`
	LastCleanup        time.Time `yaml:"last_cleanup"`
}

// IntegrityValidation contains cache integrity information
type IntegrityValidation struct {
	CacheValid       bool      `yaml:"cache_valid"`
	LastValidation   time.Time `yaml:"last_validation"`
	ValidationErrors []string  `yaml:"validation_errors"`
	OrphanedEntries  []string  `yaml:"orphaned_entries"`
}

// Manager handles build cache operations
type Manager struct {
	userDirs *userdirs.UserDirectories
	cache    *BuildCache
}

// NewManager creates a new cache manager
func NewManager() (*Manager, error) {
	userDirs, err := userdirs.NewUserDirectories()
	if err != nil {
		return nil, fmt.Errorf("failed to create user directories: %w", err)
	}

	return &Manager{
		userDirs: userDirs,
	}, nil
}

// LoadCache loads the build cache from disk
func (m *Manager) LoadCache() error {
	cachePath := m.userDirs.GetCachePath()

	// Check if cache file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		// Create new empty cache
		m.cache = &BuildCache{
			Version:    "1.0",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			AFEVersion: "v0.1.0",
			Statistics: BuildStatistics{},
			Plugins: PluginRegistry{
				Providers: make(map[string]PluginEntry),
				Agents:    make(map[string]PluginEntry),
			},
			BuildHistory: []BuildHistoryEntry{},
			CacheSettings: CacheSettings{
				MaxSizeMb:          100,
				RetentionDays:      30,
				AutoCleanup:        true,
				CompressionEnabled: false,
				ValidationInterval: "24h",
				LastCleanup:        time.Now(),
			},
			Integrity: IntegrityValidation{
				CacheValid:     true,
				LastValidation: time.Now(),
			},
		}
		return nil
	}

	// Load existing cache
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache BuildCache
	if err := yaml.Unmarshal(data, &cache); err != nil {
		return fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	m.cache = &cache
	return nil
}

// SaveCache saves the build cache to disk
func (m *Manager) SaveCache() error {
	if m.cache == nil {
		return fmt.Errorf("cache not loaded")
	}

	m.cache.UpdatedAt = time.Now()

	cachePath := m.userDirs.GetCachePath()
	data, err := yaml.Marshal(m.cache)
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// ShouldRebuild determines if a plugin needs to be rebuilt
func (m *Manager) ShouldRebuild(pluginType, pluginName, pluginPath string) (bool, string, error) {
	if m.cache == nil {
		return true, "cache not loaded", nil
	}

	// Check if plugin exists in cache
	var pluginEntry PluginEntry
	var exists bool

	if pluginType == "provider" {
		pluginEntry, exists = m.cache.Plugins.Providers[pluginName]
	} else {
		pluginEntry, exists = m.cache.Plugins.Agents[pluginName]
	}

	if !exists {
		return true, "plugin not found in cache", nil
	}

	// Check if output file exists
	outputPath := m.userDirs.GetPluginOutputPath(pluginType, pluginName)
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return true, "output plugin file not found", nil
	}

	// Check source file changes
	currentHash, err := m.calculateSourceHash(pluginPath)
	if err != nil {
		return true, "failed to calculate current source hash", nil
	}

	if currentHash != pluginEntry.BuildInfo.SourceHash {
		return true, "source files modified", nil
	}

	// Check Go module changes
	goModPath := filepath.Join(pluginPath, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		currentGoModHash, err := m.calculateFileHash(goModPath)
		if err != nil {
			return true, "failed to calculate current go.mod hash", nil
		}

		if currentGoModHash != pluginEntry.BuildInfo.GoModHash {
			return true, "Go modules modified", nil
		}
	}

	// Check cache validity
	if !pluginEntry.BuildInfo.CacheValid {
		return true, "cache entry invalid", nil
	}

	return false, "plugin is up-to-date", nil
}

// UpdatePlugin updates a plugin's cache entry after a successful build
func (m *Manager) UpdatePlugin(pluginType, pluginName, pluginPath string, buildDurationMs int, pluginSizeBytes int64) error {
	if m.cache == nil {
		return fmt.Errorf("cache not loaded")
	}

	// Calculate hashes
	sourceHash, err := m.calculateSourceHash(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to calculate source hash: %w", err)
	}

	goModHash := ""
	goModPath := filepath.Join(pluginPath, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		goModHash, err = m.calculateFileHash(goModPath)
		if err != nil {
			return fmt.Errorf("failed to calculate go.mod hash: %w", err)
		}
	}

	outputPath := m.userDirs.GetPluginOutputPath(pluginType, pluginName)
	outputHash, err := m.calculateFileHash(outputPath)
	if err != nil {
		return fmt.Errorf("failed to calculate output hash: %w", err)
	}

	// Get source files
	sourceFiles, err := m.getSourceFiles(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to get source files: %w", err)
	}

	// Create plugin entry
	pluginEntry := PluginEntry{
		BuildInfo: PluginBuildInfo{
			SourceHash:      sourceHash,
			GoModHash:       goModHash,
			BuildConfigHash: "default", // TODO: Implement build config hashing
			OutputPath:      outputPath,
			OutputHash:      outputHash,
			BuildTime:       time.Now(),
			BuildDurationMs: buildDurationMs,
			PluginSizeBytes: pluginSizeBytes,
			NeedsRebuild:    false,
			LastUsed:        time.Now(),
			CacheValid:      true,
		},
		SourceFiles: sourceFiles,
	}

	// Update build count
	if pluginType == "provider" {
		if existing, exists := m.cache.Plugins.Providers[pluginName]; exists {
			pluginEntry.BuildInfo.BuildCount = existing.BuildInfo.BuildCount + 1
		} else {
			pluginEntry.BuildInfo.BuildCount = 1
		}
		m.cache.Plugins.Providers[pluginName] = pluginEntry
	} else {
		if existing, exists := m.cache.Plugins.Agents[pluginName]; exists {
			pluginEntry.BuildInfo.BuildCount = existing.BuildInfo.BuildCount + 1
		} else {
			pluginEntry.BuildInfo.BuildCount = 1
		}
		m.cache.Plugins.Agents[pluginName] = pluginEntry
	}

	// Update statistics
	m.cache.Statistics.TotalBuilds++
	m.cache.Statistics.TotalBuildTimeMs += buildDurationMs
	if m.cache.Statistics.TotalBuilds > 0 {
		m.cache.Statistics.AverageBuildTimeMs = m.cache.Statistics.TotalBuildTimeMs / m.cache.Statistics.TotalBuilds
	}

	return nil
}

// RecordBuildHistory records a build operation in the history
func (m *Manager) RecordBuildHistory(command string, pluginsBuilt, pluginsCached []string, totalDurationMs int, success bool) {
	if m.cache == nil {
		return
	}

	buildID := fmt.Sprintf("build_%s", time.Now().Format("20060102_150405"))

	cacheHitRate := 0.0
	totalPlugins := len(pluginsBuilt) + len(pluginsCached)
	if totalPlugins > 0 {
		cacheHitRate = float64(len(pluginsCached)) / float64(totalPlugins) * 100
	}

	entry := BuildHistoryEntry{
		BuildID:         buildID,
		Timestamp:       time.Now(),
		Command:         command,
		PluginsBuilt:    pluginsBuilt,
		PluginsCached:   pluginsCached,
		TotalDurationMs: totalDurationMs,
		Success:         success,
		CacheHitRate:    cacheHitRate,
	}

	// Add to history (keep last 100 entries)
	m.cache.BuildHistory = append(m.cache.BuildHistory, entry)
	if len(m.cache.BuildHistory) > 100 {
		m.cache.BuildHistory = m.cache.BuildHistory[1:]
	}

	// Update cache statistics
	if success {
		m.cache.Statistics.CacheHits += len(pluginsCached)
		m.cache.Statistics.CacheMisses += len(pluginsBuilt)
		totalBuilds := m.cache.Statistics.CacheHits + m.cache.Statistics.CacheMisses
		if totalBuilds > 0 {
			m.cache.Statistics.CacheHitRate = float64(m.cache.Statistics.CacheHits) / float64(totalBuilds) * 100
		}
	}
}

// GetCacheStatus returns current cache status information
func (m *Manager) GetCacheStatus() (*CacheStatus, error) {
	if m.cache == nil {
		return nil, fmt.Errorf("cache not loaded")
	}

	status := &CacheStatus{
		Version:          m.cache.Version,
		CreatedAt:        m.cache.CreatedAt,
		UpdatedAt:        m.cache.UpdatedAt,
		TotalBuilds:      m.cache.Statistics.TotalBuilds,
		CacheHits:        m.cache.Statistics.CacheHits,
		CacheMisses:      m.cache.Statistics.CacheMisses,
		CacheHitRate:     m.cache.Statistics.CacheHitRate,
		AverageBuildTime: m.cache.Statistics.AverageBuildTimeMs,
		TotalCacheSize:   m.cache.Statistics.TotalCacheSizeMb,
		CacheValid:       m.cache.Integrity.CacheValid,
		LastValidation:   m.cache.Integrity.LastValidation,
		ProvidersCached:  len(m.cache.Plugins.Providers),
		AgentsCached:     len(m.cache.Plugins.Agents),
	}

	return status, nil
}

// CacheStatus contains cache status information
type CacheStatus struct {
	Version          string    `yaml:"version"`
	CreatedAt        time.Time `yaml:"created_at"`
	UpdatedAt        time.Time `yaml:"updated_at"`
	TotalBuilds      int       `yaml:"total_builds"`
	CacheHits        int       `yaml:"cache_hits"`
	CacheMisses      int       `yaml:"cache_misses"`
	CacheHitRate     float64   `yaml:"cache_hit_rate"`
	AverageBuildTime int       `yaml:"average_build_time_ms"`
	TotalCacheSize   float64   `yaml:"total_cache_size_mb"`
	CacheValid       bool      `yaml:"cache_valid"`
	LastValidation   time.Time `yaml:"last_validation"`
	ProvidersCached  int       `yaml:"providers_cached"`
	AgentsCached     int       `yaml:"agents_cached"`
}

// Helper methods

func (m *Manager) calculateSourceHash(pluginPath string) (string, error) {
	// Get all .go files in the plugin directory
	goFiles, err := filepath.Glob(filepath.Join(pluginPath, "*.go"))
	if err != nil {
		return "", fmt.Errorf("failed to find Go files: %w", err)
	}

	if len(goFiles) == 0 {
		return "", fmt.Errorf("no Go files found in %s", pluginPath)
	}

	// Calculate combined hash of all source files
	hasher := sha256.New()
	for _, goFile := range goFiles {
		fileHash, err := m.calculateFileHash(goFile)
		if err != nil {
			return "", fmt.Errorf("failed to calculate hash for %s: %w", goFile, err)
		}
		hasher.Write([]byte(fileHash))
	}

	return fmt.Sprintf("sha256:%x", hasher.Sum(nil)), nil
}

func (m *Manager) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to calculate file hash: %w", err)
	}

	return fmt.Sprintf("sha256:%x", hasher.Sum(nil)), nil
}

func (m *Manager) getSourceFiles(pluginPath string) ([]SourceFile, error) {
	goFiles, err := filepath.Glob(filepath.Join(pluginPath, "*.go"))
	if err != nil {
		return nil, fmt.Errorf("failed to find Go files: %w", err)
	}

	var sourceFiles []SourceFile
	for _, goFile := range goFiles {
		stat, err := os.Stat(goFile)
		if err != nil {
			continue // Skip files we can't stat
		}

		hash, err := m.calculateFileHash(goFile)
		if err != nil {
			continue // Skip files we can't hash
		}

		sourceFiles = append(sourceFiles, SourceFile{
			Path:      strings.TrimPrefix(goFile, pluginPath+"/"),
			Hash:      hash,
			SizeBytes: int(stat.Size()),
			Modified:  stat.ModTime(),
		})
	}

	return sourceFiles, nil
}
