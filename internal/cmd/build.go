package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/cache"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/userdirs"
	"github.com/spf13/cobra"
)

var (
	parallelBuilds bool
	forceBuild     bool
	cleanBuild     bool
	verboseBuild   bool
	buildName      string
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build AgentForge plugins",
	Long: `Build AgentForge plugins with intelligent caching and parallel compilation.
This command manages the complete build process for providers and agents,
storing build artifacts in ~/.afe/ and maintaining a YAML cache for
incremental builds.`,
}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Global build flags
	buildCmd.PersistentFlags().BoolVarP(&parallelBuilds, "parallel", "p", true, "Build plugins in parallel")
	buildCmd.PersistentFlags().BoolVar(&forceBuild, "force", false, "Force rebuild of all plugins")
	buildCmd.PersistentFlags().BoolVar(&cleanBuild, "clean", false, "Clean cache and rebuild all plugins")
	buildCmd.PersistentFlags().BoolVarP(&verboseBuild, "verbose", "v", false, "Verbose build output")
}

// runBuildCommand handles building specific plugin types
func runBuildCommand(pluginType string, args []string) error {
	if verboseBuild {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		if len(args) > 0 {
			fmt.Printf("🏗️  Building specific %s: %v\n", pluginType, args)
		} else {
			fmt.Printf("🏗️  Building all %ss...\n", pluginType)
		}
	}

	// Ensure user directories are initialized
	userDirs, err := userdirs.NewUserDirectories()
	if err != nil {
		return fmt.Errorf("failed to create user directories manager: %w", err)
	}

	if !userDirs.Exists() {
		fmt.Println("❌ AgentForge user directories not initialized")
		fmt.Println("💡 Run 'afe init' to initialize user directories")
		return fmt.Errorf("user directories not initialized")
	}

	// Initialize cache manager
	cacheManager, err := cache.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create cache manager: %w", err)
	}

	if err := cacheManager.LoadCache(); err != nil {
		return fmt.Errorf("failed to load build cache: %w", err)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Determine which plugins to build
	var pluginsToBuild []string
	if len(args) > 0 {
		pluginsToBuild = args
	} else {
		// Discover all plugins of this type
		pluginDir := filepath.Join(cwd, pluginType+"s")
		plugins, err := discoverPlugins(pluginType, pluginDir)
		if err != nil {
			return fmt.Errorf("failed to discover %ss: %w", pluginType, err)
		}
		pluginsToBuild = plugins
	}

	if len(pluginsToBuild) == 0 {
		fmt.Printf("❌ No %ss found to build\n", pluginType)
		return nil
	}

	if verboseBuild {
		fmt.Printf("📦 Found %d %ss to build\n", len(pluginsToBuild), pluginType)
	}

	// Create build plan for specific plugin type
	buildPlan := &BuildPlan{
		ProvidersToBuild: []string{},
		AgentsToBuild:    []string{},
		ProvidersCached:  []string{},
		AgentsCached:     []string{},
	}

	// Analyze plugins of the specified type
	for _, pluginName := range pluginsToBuild {
		pluginPath := filepath.Join(cwd, pluginType+"s", pluginName)
		shouldRebuild, reason, err := cacheManager.ShouldRebuild(pluginType, pluginName, pluginPath)
		if err != nil && verboseBuild {
			fmt.Printf("⚠️  Error checking %s %s: %v\n", pluginType, pluginName, err)
			shouldRebuild = true
		}

		if forceBuild || cleanBuild || shouldRebuild {
			if pluginType == "provider" {
				buildPlan.ProvidersToBuild = append(buildPlan.ProvidersToBuild, pluginName)
			} else {
				buildPlan.AgentsToBuild = append(buildPlan.AgentsToBuild, pluginName)
			}
			if verboseBuild {
				fmt.Printf("🔨 %s %s: %s\n", strings.Title(pluginType), pluginName, reason)
			}
		} else {
			if pluginType == "provider" {
				buildPlan.ProvidersCached = append(buildPlan.ProvidersCached, pluginName)
			} else {
				buildPlan.AgentsCached = append(buildPlan.AgentsCached, pluginName)
			}
			if verboseBuild {
				fmt.Printf("📦 %s %s: cached (unchanged)\n", strings.Title(pluginType), pluginName)
			}
		}
	}

	// Show build summary
	totalToBuild := len(buildPlan.ProvidersToBuild) + len(buildPlan.AgentsToBuild)
	totalCached := len(buildPlan.ProvidersCached) + len(buildPlan.AgentsCached)

	fmt.Printf("📊 Build Plan: %d to rebuild, %d cached\n", totalToBuild, totalCached)

	if totalToBuild == 0 && !forceBuild && !cleanBuild {
		fmt.Println("✅ All plugins are up-to-date")
		return nil
	}

	// Execute build
	startTime := time.Now()
	buildResult, err := executeBuild(buildPlan, cwd, userDirs, cacheManager)
	if err != nil {
		return fmt.Errorf("build execution failed: %w", err)
	}

	// Record build history
	totalDuration := int(time.Since(startTime).Milliseconds())
	commandName := fmt.Sprintf("afe build %s", pluginType)
	cacheManager.RecordBuildHistory(
		commandName,
		append(buildPlan.ProvidersToBuild, buildPlan.AgentsToBuild...),
		append(buildPlan.ProvidersCached, buildPlan.AgentsCached...),
		totalDuration,
		buildResult.Success,
	)

	// Save cache
	if err := cacheManager.SaveCache(); err != nil {
		return fmt.Errorf("failed to save build cache: %w", err)
	}

	// Show results
	if buildResult.Success {
		fmt.Printf("✅ Build completed: %d rebuilt, %d cached in %v\n",
			totalToBuild, totalCached, time.Since(startTime))

		// Hot reload updated plugins
		if totalToBuild > 0 {
			fmt.Println("🔄 Hot reloading updated plugins...")
			if err := triggerHotReload(buildPlan, userDirs); err != nil {
				fmt.Printf("⚠️  Hot reload failed: %v\n", err)
			} else {
				fmt.Println("✅ Hot reload completed successfully")
			}
		}

		fmt.Printf("🎉 %ss ready\n", strings.Title(pluginType))
	} else {
		fmt.Printf("❌ Build failed: %d successful, %d failed\n",
			buildResult.SuccessCount, buildResult.FailureCount)
		return fmt.Errorf("build failed")
	}

	return nil
}

// BuildPlan contains the plan for what to build
type BuildPlan struct {
	ProvidersToBuild []string
	AgentsToBuild    []string
	ProvidersCached  []string
	AgentsCached     []string
}

// BuildResult contains the results of a build operation
type BuildResult struct {
	Success       bool
	SuccessCount  int
	FailureCount  int
	BuildTime     time.Duration
	BuiltPlugins  []string
	FailedPlugins []string
	Errors        []error
}

// discoverPlugins finds all plugins of a given type in a directory
func discoverPlugins(pluginType, pluginDir string) ([]string, error) {
	var plugins []string

	// Check if directory exists
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return plugins, nil
	}

	// Read directory entries
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin directory: %w", err)
	}

	// Find plugin directories (contain go.mod files)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginPath := filepath.Join(pluginDir, entry.Name())
		goModPath := filepath.Join(pluginPath, "go.mod")

		// Check if this is a valid plugin (has go.mod)
		if _, err := os.Stat(goModPath); err == nil {
			plugins = append(plugins, entry.Name())
		}
	}

	return plugins, nil
}

// executeBuild executes the build plan
func executeBuild(plan *BuildPlan, projectDir string, userDirs *userdirs.UserDirectories, cacheManager *cache.Manager) (*BuildResult, error) {
	result := &BuildResult{
		BuiltPlugins:  []string{},
		FailedPlugins: []string{},
		Errors:        []error{},
	}

	startTime := time.Now()

	// Determine maximum workers
	maxWorkers := runtime.NumCPU()
	if maxWorkers > 8 {
		maxWorkers = 8 // Reasonable upper limit
	}

	if !parallelBuilds {
		maxWorkers = 1
	}

	// Create worker pool
	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Build providers
	for _, provider := range plan.ProvidersToBuild {
		wg.Add(1)
		go func(pluginName string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := buildPlugin("provider", pluginName, projectDir, userDirs, cacheManager); err != nil {
				mu.Lock()
				result.FailedPlugins = append(result.FailedPlugins, pluginName)
				result.Errors = append(result.Errors, err)
				mu.Unlock()
			} else {
				mu.Lock()
				result.BuiltPlugins = append(result.BuiltPlugins, pluginName)
				result.SuccessCount++
				mu.Unlock()
			}
		}(provider)
	}

	// Build agents
	for _, agent := range plan.AgentsToBuild {
		wg.Add(1)
		go func(pluginName string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := buildPlugin("agent", pluginName, projectDir, userDirs, cacheManager); err != nil {
				mu.Lock()
				result.FailedPlugins = append(result.FailedPlugins, pluginName)
				result.Errors = append(result.Errors, err)
				mu.Unlock()
			} else {
				mu.Lock()
				result.BuiltPlugins = append(result.BuiltPlugins, pluginName)
				result.SuccessCount++
				mu.Unlock()
			}
		}(agent)
	}

	wg.Wait()

	result.BuildTime = time.Since(startTime)
	result.Success = result.FailureCount == 0

	return result, nil
}

// buildPlugin builds a single plugin
func buildPlugin(pluginType, pluginName, projectDir string, userDirs *userdirs.UserDirectories, cacheManager *cache.Manager) error {
	startTime := time.Now()

	if verboseBuild {
		fmt.Printf("🔨 Building %s %s...\n", pluginType, pluginName)
	}

	// Determine source and output paths
	sourcePath := filepath.Join(projectDir, pluginType+"s", pluginName)
	outputPath := userDirs.GetPluginOutputPath(pluginType, pluginName)

	if verboseBuild {
		fmt.Printf("   Source: %s\n", sourcePath)
		fmt.Printf("   Output: %s\n", outputPath)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build the plugin using go build directly
	if err := buildGoPlugin(sourcePath, outputPath); err != nil {
		if verboseBuild {
			fmt.Printf("❌ Build failed: %v\n", err)
		}
		return fmt.Errorf("failed to build %s %s: %w", pluginType, pluginName, err)
	}

	// Get plugin size
	stat, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("failed to get plugin size: %w", err)
	}

	// Update cache
	buildDuration := int(time.Since(startTime).Milliseconds())
	if err := cacheManager.UpdatePlugin(pluginType, pluginName, sourcePath, buildDuration, stat.Size()); err != nil {
		return fmt.Errorf("failed to update cache: %w", err)
	}

	if verboseBuild {
		fmt.Printf("✅ Built %s %s (%v, %s)\n", pluginType, pluginName, time.Since(startTime), formatBytes(stat.Size()))
	}

	return nil
}

// buildGoPlugin builds a Go plugin using the go build command
func buildGoPlugin(source, output string) error {
	// Build the plugin - change directory to source and build .
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", output, ".")
	cmd.Dir = source

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

// formatBytes formats bytes in human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
