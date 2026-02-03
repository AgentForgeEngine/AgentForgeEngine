package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/internal/config"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start AgentForge Engine",
	Long:  "Start the AgentForge Engine with the specified configuration",
	RunE:  runStart,
}

var serverCtx context.Context
var serverCancel context.CancelFunc

func runStart(cmd *cobra.Command, args []string) error {
	// Initialize config manager
	configManager := config.NewManager()
	configPath := getConfigPath()

	if err := configManager.Load(configPath); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if verbose {
		fmt.Println("Starting AgentForge Engine...")
		fmt.Printf("Config loaded from: %s\n", configPath)
	}

	// Create context with cancellation
	serverCtx, serverCancel = context.WithCancel(context.Background())

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the server components
	go func() {
		if err := startServerComponents(serverCtx, configManager); err != nil {
			log.Printf("Server error: %v", err)
			serverCancel()
		}
	}()

	// Wait for signals
	select {
	case sig := <-sigChan:
		if verbose {
			fmt.Printf("\nReceived signal: %v\n", sig)
		}
	case <-serverCtx.Done():
		if verbose {
			fmt.Println("Server context cancelled")
		}
	}

	// Graceful shutdown
	if verbose {
		fmt.Println("Shutting down gracefully...")
	}

	serverCancel()
	time.Sleep(2 * time.Second) // Give components time to cleanup

	fmt.Println("AgentForge Engine stopped")
	return nil
}

func startServerComponents(ctx context.Context, configManager *config.Manager) error {
	// This will be implemented when we add the actual server components
	// For now, just demonstrate the structure

	serverConfig := configManager.GetServerConfig()
	if verbose {
		fmt.Printf("Server starting on %s:%d\n", serverConfig.Host, serverConfig.Port)
	}

	// TODO: Initialize model manager
	// TODO: Initialize plugin manager
	// TODO: Initialize recovery system
	// TODO: Start HTTP server

	// Keep the server running
	<-ctx.Done()
	return nil
}

func getConfigPath() string {
	if cfgFile != "" {
		return cfgFile
	}

	// Try default locations
	defaultPaths := []string{
		"./configs/agentforge.yaml",
		"./agentforge.yaml",
		"$HOME/.agentforge.yaml",
	}

	for _, path := range defaultPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "./configs/agentforge.yaml" // fallback
}

func init() {
	rootCmd.AddCommand(startCmd)
}
