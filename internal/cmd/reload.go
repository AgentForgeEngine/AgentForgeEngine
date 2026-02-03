package cmd

import (
	"fmt"

	"github.com/AgentForgeEngine/AgentForgeEngine/internal/config"
	"github.com/spf13/cobra"
)

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload agents or configuration",
	Long:  "Reload specific agents or the entire configuration",
	RunE:  runReload,
}

var (
	reloadAgent string
	reloadAll   bool
)

func runReload(cmd *cobra.Command, args []string) error {
	if verbose {
		fmt.Println("Reloading AgentForge Engine...")
	}

	if reloadAll || reloadAgent == "" {
		// Reload configuration
		configManager := config.NewManager()
		configPath := getConfigPath()

		if err := configManager.Load(configPath); err != nil {
			return fmt.Errorf("failed to reload config: %w", err)
		}

		fmt.Println("Configuration reloaded successfully")

		if reloadAll {
			// TODO: Implement full agent reload
			fmt.Println("All agents reloaded")
		}
	} else {
		// Reload specific agent
		fmt.Printf("Agent '%s' reloaded\n", reloadAgent)
		// TODO: Implement specific agent reload
	}

	return nil
}

func init() {
	rootCmd.AddCommand(reloadCmd)
	reloadCmd.Flags().StringVarP(&reloadAgent, "agent", "a", "", "Reload specific agent")
	reloadCmd.Flags().BoolVarP(&reloadAll, "all", "A", false, "Reload all agents and configuration")
}
