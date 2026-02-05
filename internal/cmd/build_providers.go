package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/cache"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/userdirs"
	"github.com/spf13/cobra"
)

// providersCmd represents the 'afe build providers' command
var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "Build provider plugins",
	Long: `Build provider plugins with intelligent caching.
This command discovers and builds all provider plugins in the ./providers/ directory,
storing build artifacts in ~/.afe/providers/ and maintaining cache information.`,
	RunE: runBuildProviders,
}

// agentsCmd represents the 'afe build agents' command
var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Build agent plugins",
	Long: `Build agent plugins with intelligent caching.
This command discovers and builds all agent plugins in the ./agents/ directory,
storing build artifacts in ~/.afe/agents/ and maintaining cache information.`,
	RunE: runBuildAgents,
}

// allCmd represents the 'afe build all' command
var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Build all plugins",
	Long: `Build all provider and agent plugins with intelligent caching.
This command discovers and builds all plugins, using the build cache to avoid
unnecessary rebuilds and supporting parallel compilation for faster builds.`,
	RunE: runBuildAll,
}

// runBuildProviders handles the 'afe build providers' command
func runBuildProviders(cmd *cobra.Command, args []string) error {
	return runBuildCommand("provider", args)
}

// runBuildAgents handles the 'afe build agents' command
func runBuildAgents(cmd *cobra.Command, args []string) error {
	return runBuildCommand("agent", args)
}

// runBuildAll handles the 'afe build all' command
func runBuildAll(cmd *cobra.Command, args []string) error {
	if verboseBuild {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		fmt.Println("🏗️  Building all AgentForge plugins...")
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

	// Clean cache if requested
	if cleanBuild {
		if verboseBuild {
			fmt.Println("🧹 Cleaning build cache...")
		}
		// TODO: Implement cache cleaning
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Discover all plugins
	providers, err := discoverPlugins("provider", filepath.Join(cwd, "providers"))
	if err != nil {
		return fmt.Errorf("failed to discover providers: %w", err)
	}

	agents, err := discoverPlugins("agent", filepath.Join(cwd, "agents"))
	if err != nil {
		return fmt.Errorf("failed to discover agents: %w", err)
	}

	if verboseBuild {
		fmt.Printf("📦 Discovered %d providers and %d agents\n", len(providers), len(agents))
	}

	// Create build plan
	buildPlan := &BuildPlan{
		ProvidersToBuild: []string{},
		AgentsToBuild:    []string{},
		ProvidersCached:  []string{},
		AgentsCached:     []string{},
	}

	// Analyze providers
	for _, provider := range providers {
		shouldRebuild, reason, err := cacheManager.ShouldRebuild("provider", provider, filepath.Join(cwd, "providers", provider))
		if err != nil && verboseBuild {
			fmt.Printf("⚠️  Error checking provider %s: %v\n", provider, err)
			shouldRebuild = true
		}

		if forceBuild || cleanBuild || shouldRebuild {
			buildPlan.ProvidersToBuild = append(buildPlan.ProvidersToBuild, provider)
			if verboseBuild {
				fmt.Printf("🔨 Provider %s: %s\n", provider, reason)
			}
		} else {
			buildPlan.ProvidersCached = append(buildPlan.ProvidersCached, provider)
			if verboseBuild {
				fmt.Printf("📦 Provider %s: cached (unchanged)\n", provider)
			}
		}
	}

	// Analyze agents
	for _, agent := range agents {
		shouldRebuild, reason, err := cacheManager.ShouldRebuild("agent", agent, filepath.Join(cwd, "agents", agent))
		if err != nil && verboseBuild {
			fmt.Printf("⚠️  Error checking agent %s: %v\n", agent, err)
			shouldRebuild = true
		}

		if forceBuild || cleanBuild || shouldRebuild {
			buildPlan.AgentsToBuild = append(buildPlan.AgentsToBuild, agent)
			if verboseBuild {
				fmt.Printf("🔨 Agent %s: %s\n", agent, reason)
			}
		} else {
			buildPlan.AgentsCached = append(buildPlan.AgentsCached, agent)
			if verboseBuild {
				fmt.Printf("📦 Agent %s: cached (unchanged)\n", agent)
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
	cacheManager.RecordBuildHistory(
		"afe build all",
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
		if len(buildPlan.ProvidersToBuild)+len(buildPlan.AgentsToBuild) > 0 {
			fmt.Println("🔄 Hot reloading updated plugins...")
			if err := triggerHotReload(buildPlan, userDirs); err != nil {
				fmt.Printf("⚠️  Hot reload failed: %v\n", err)
			} else {
				fmt.Println("✅ Hot reload completed successfully")
			}
		}

		fmt.Println("🎉 System ready with all plugins")
	} else {
		fmt.Printf("❌ Build failed: %d successful, %d failed\n",
			buildResult.SuccessCount, buildResult.FailureCount)
		return fmt.Errorf("build failed")
	}

	return nil
}

func init() {
	buildCmd.AddCommand(providersCmd)
	buildCmd.AddCommand(agentsCmd)
	buildCmd.AddCommand(allCmd)

	// Add name flag to specific commands
	providersCmd.Flags().StringVar(&buildName, "name", "", "Build specific provider by name")
	agentsCmd.Flags().StringVar(&buildName, "name", "", "Build specific agent by name")
}
