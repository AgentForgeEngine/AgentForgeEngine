package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/userdirs"
	"github.com/spf13/cobra"
)

var (
	forceInit   bool
	migrateInit bool
	verboseInit bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AgentForge user directories",
	Long: `Initialize the ~/.afe directory structure for storing plugins,
cache, and configuration. This creates a clean separation between
project files and user-specific build artifacts.

By default, this command creates the necessary directories and default
configuration files in your home directory under ~/.afe/`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&forceInit, "force", false, "Force reinitialization, overwriting existing files")
	initCmd.Flags().BoolVar(&migrateInit, "migrate", true, "Migrate existing plugins from ./plugins/ directory")
	initCmd.Flags().BoolVarP(&verboseInit, "verbose", "v", false, "Verbose output")
}

func runInit(cmd *cobra.Command, args []string) error {
	if verboseInit {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		fmt.Println("🚀 Initializing AgentForge user directories...")
	}

	// Create user directories manager
	userDirs, err := userdirs.NewUserDirectories()
	if err != nil {
		return fmt.Errorf("failed to create user directories manager: %w", err)
	}

	// Check if already initialized
	if userDirs.Exists() && !forceInit {
		fmt.Printf("✅ AgentForge user directories already exist at: %s\n", userDirs.AFEDir)
		fmt.Println("Use --force to reinitialize")
		return nil
	}

	// Create directory structure
	if err := userDirs.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to create user directories: %w", err)
	}

	if verboseInit {
		fmt.Printf("✅ Created directory structure at: %s\n", userDirs.AFEDir)
	}

	// Create default configuration
	if err := userDirs.CreateDefaultConfig(); err != nil {
		return fmt.Errorf("failed to create default configuration: %w", err)
	}

	if verboseInit {
		fmt.Printf("✅ Created default configuration at: %s\n",
			filepath.Join(userDirs.ConfigDir, "build_config.yaml"))
	}

	// Get current working directory for migration
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Migrate existing plugins if requested
	if migrateInit {
		if verboseInit {
			fmt.Println("🔄 Checking for existing plugins to migrate...")
		}

		migrationResult, err := userDirs.MigrateFromOldStructure(cwd)
		if err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		if verboseInit {
			fmt.Printf("✅ Migration completed: %s\n", migrationResult.Message)
			fmt.Printf("   Duration: %v\n", migrationResult.Duration())
			if len(migrationResult.ProvidersMigrated) > 0 {
				fmt.Printf("   Providers migrated: %v\n", migrationResult.ProvidersMigrated)
			}
			if len(migrationResult.AgentsMigrated) > 0 {
				fmt.Printf("   Agents migrated: %v\n", migrationResult.AgentsMigrated)
			}
		} else {
			if migrationResult.FilesMigrated > 0 {
				fmt.Printf("✅ Migrated %d existing plugins\n", migrationResult.FilesMigrated)
			}
		}
	}

	// Create initial cache file
	if err := createInitialCache(userDirs); err != nil {
		return fmt.Errorf("failed to create initial cache: %w", err)
	}

	if verboseInit {
		fmt.Printf("✅ Created initial build cache at: %s\n", userDirs.GetCachePath())
	}

	// Show system information
	if verboseInit {
		systemInfo := userDirs.GetSystemInfo()
		fmt.Println("\n📊 System Information:")
		fmt.Printf("   OS: %s\n", systemInfo["os"])
		fmt.Printf("   Architecture: %s\n", systemInfo["arch"])
		fmt.Printf("   CPU Cores: %v\n", systemInfo["cpu_cores"])
		fmt.Printf("   Go Root: %v\n", systemInfo["go_root"])
		fmt.Printf("   Home Directory: %v\n", systemInfo["home_dir"])
		fmt.Printf("   AFE Directory: %v\n", systemInfo["afe_dir"])
	}

	// Show success message
	fmt.Printf("\n🎉 AgentForge user directories initialized successfully!\n")
	fmt.Printf("📁 Location: %s\n", userDirs.AFEDir)
	fmt.Printf("⚙️  Configuration: %s\n", filepath.Join(userDirs.ConfigDir, "build_config.yaml"))
	fmt.Printf("🏗️  Build Cache: %s\n", userDirs.GetCachePath())

	if migrateInit {
		fmt.Println("🔄 Existing plugins have been migrated to the new structure")
	}

	fmt.Println("\n📋 Next Steps:")
	fmt.Println("   1. Run 'afe build all' to build all plugins with intelligent caching")
	fmt.Println("   2. Run 'afe cache status' to view cache statistics")
	fmt.Println("   3. Run 'afe build providers --name <name>' to build specific plugins")

	return nil
}

// createInitialCache creates the initial build cache file
func createInitialCache(userDirs *userdirs.UserDirectories) error {
	cachePath := userDirs.GetCachePath()

	// Check if cache already exists
	if _, err := os.Stat(cachePath); err == nil {
		return nil // Cache already exists
	}

	systemInfo := userDirs.GetSystemInfo()

	initialCache := fmt.Sprintf(`# AgentForge Engine Build Cache
# Schema Version: 1.0
version: "1.0"
created_at: "%s"
updated_at: "%s"
afe_version: "v0.1.0"
go_version: "%s"

# Global build statistics
statistics:
  total_builds: 0
  cache_hits: 0
  cache_misses: 0
  cache_hit_rate: 0.0
  total_build_time_ms: 0
  average_build_time_ms: 0
  total_cache_size_mb: 0.0

# Plugin registry (will be populated as plugins are built)
plugins:
  providers: {}
  agents: {}

# Build history (will be populated as builds occur)
build_history: []

# Cache management settings
cache_settings:
  max_size_mb: 100
  retention_days: 30
  auto_cleanup: true
  compression_enabled: false
  validation_interval: "24h"
  last_cleanup: "%s"

# Integrity validation
integrity:
  cache_valid: true
  last_validation: "%s"
  validation_errors: []
  orphaned_entries: []
`,
		systemInfo["created_at"],
		systemInfo["created_at"],
		systemInfo["go_root"],
		systemInfo["created_at"],
		systemInfo["created_at"],
	)

	return os.WriteFile(cachePath, []byte(initialCache), 0644)
}
