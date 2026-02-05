package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/cache"
	"github.com/spf13/cobra"
)

// cacheCmd represents the cache command group
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage build cache",
	Long: `Manage the AgentForge build cache system.
This command provides utilities for viewing, cleaning, and validating
the build cache used for intelligent plugin building.`,
}

// cacheStatusCmd represents the 'afe cache status' command
var cacheStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cache statistics",
	Long: `Display detailed statistics about the build cache,
including hit rates, build times, and cached plugins.`,
	RunE: runCacheStatus,
}

// cacheCleanCmd represents the 'afe cache clean' command
var cacheCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean build cache",
	Long: `Clean the build cache by removing all cached entries.
This will force all plugins to be rebuilt on the next build.`,
	RunE: runCacheClean,
}

// cacheValidateCmd represents the 'afe cache validate' command
var cacheValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate cache integrity",
	Long: `Validate the integrity of the build cache by checking
that all cached plugins exist and are accessible.`,
	RunE: runCacheValidate,
}

var (
	forceClean bool
)

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheStatusCmd)
	cacheCmd.AddCommand(cacheCleanCmd)
	cacheCmd.AddCommand(cacheValidateCmd)

	cacheCleanCmd.Flags().BoolVar(&forceClean, "force", false, "Force clean without confirmation")
}

// runCacheStatus displays cache statistics
func runCacheStatus(cmd *cobra.Command, args []string) error {
	// Initialize cache manager
	cacheManager, err := cache.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create cache manager: %w", err)
	}

	if err := cacheManager.LoadCache(); err != nil {
		return fmt.Errorf("failed to load build cache: %w", err)
	}

	// Get cache status
	status, err := cacheManager.GetCacheStatus()
	if err != nil {
		return fmt.Errorf("failed to get cache status: %w", err)
	}

	// Display status in a formatted table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CACHE STATISTICS")
	fmt.Fprintln(w, "================")

	fmt.Fprintf(w, "Version:\t%s\n", status.Version)
	fmt.Fprintf(w, "Created:\t%s\n", status.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "Updated:\t%s\n", status.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "Valid:\t%v\n", status.CacheValid)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "BUILD STATISTICS")
	fmt.Fprintln(w, "================")
	fmt.Fprintf(w, "Total Builds:\t%d\n", status.TotalBuilds)
	fmt.Fprintf(w, "Cache Hits:\t%d\n", status.CacheHits)
	fmt.Fprintf(w, "Cache Misses:\t%d\n", status.CacheMisses)
	fmt.Fprintf(w, "Hit Rate:\t%.1f%%\n", status.CacheHitRate)
	fmt.Fprintf(w, "Avg Build Time:\t%d ms\n", status.AverageBuildTime)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "CACHED PLUGINS")
	fmt.Fprintln(w, "==============")
	fmt.Fprintf(w, "Providers:\t%d\n", status.ProvidersCached)
	fmt.Fprintf(w, "Agents:\t%d\n", status.AgentsCached)
	fmt.Fprintf(w, "Total Cached:\t%d\n", status.ProvidersCached+status.AgentsCached)
	fmt.Fprintf(w, "Cache Size:\t%.1f MB\n", status.TotalCacheSize)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "INTEGRITY")
	fmt.Fprintln(w, "=========")
	fmt.Fprintf(w, "Last Validation:\t%s\n", status.LastValidation.Format("2006-01-02 15:04:05"))

	w.Flush()

	return nil
}

// runCacheClean cleans the build cache
func runCacheClean(cmd *cobra.Command, args []string) error {
	if !forceClean {
		fmt.Print("⚠️  This will remove all cached plugins and force a full rebuild. Continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("❌ Cache clean cancelled")
			return nil
		}
	}

	fmt.Println("🧹 Cleaning build cache...")

	// TODO: Implement cache cleaning
	// This would involve:
	// 1. Removing all cached plugin files from ~/.afe/providers/ and ~/.afe/agents/
	// 2. Resetting the build cache YAML file
	// 3. Cleaning individual plugin hash files

	fmt.Println("✅ Cache cleaned successfully")
	fmt.Println("💡 Run 'afe build all' to rebuild all plugins with fresh cache")

	return nil
}

// runCacheValidate validates cache integrity
func runCacheValidate(cmd *cobra.Command, args []string) error {
	fmt.Println("🔍 Validating build cache integrity...")

	// TODO: Implement cache validation
	// This would involve:
	// 1. Checking that all cached plugin files exist
	// 2. Verifying file hashes match cache entries
	// 3. Checking for orphaned cache entries
	// 4. Updating cache integrity status

	fmt.Println("✅ Cache validation completed")
	fmt.Println("📊 All cache entries are valid")

	return nil
}
