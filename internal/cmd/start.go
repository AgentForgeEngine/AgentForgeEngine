package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/internal/api"
	"github.com/AgentForgeEngine/AgentForgeEngine/internal/config"
	"github.com/AgentForgeEngine/AgentForgeEngine/internal/loader"
	"github.com/AgentForgeEngine/AgentForgeEngine/internal/models"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/status"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/userdirs"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start AgentForgeEngine",
	Long:  "Start the AgentForgeEngine with the specified configuration",
	RunE:  runStart,
}

var serverCtx context.Context
var serverCancel context.CancelFunc
var statusManager *status.Manager
var pluginManager *loader.Manager

// var orchestratorManager *orchestrator.Manager // Disabled for now

func runStart(cmd *cobra.Command, args []string) error {
	// Initialize user directories and status manager
	userDirs, err := userdirs.NewUserDirectories()
	if err != nil {
		return fmt.Errorf("failed to initialize user directories: %w", err)
	}

	statusManager = status.NewManager(userDirs.AFEDir)

	// Write PID file
	if err := statusManager.WritePID(); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Initialize config manager
	configManager := config.NewManager()
	configPath := getConfigPath()

	if err := configManager.Load(configPath); err != nil {
		statusManager.Cleanup()
		return fmt.Errorf("failed to load config: %w", err)
	}

	if verbose {
		fmt.Println("Starting AgentForgeEngine...")
		fmt.Printf("Config loaded from: %s\n", configPath)
		fmt.Printf("PID file: %s\n", statusManager.GetPIDFile())
		fmt.Printf("Socket file: %s\n", statusManager.GetSocketFile())
	}

	// Create context with cancellation
	serverCtx, serverCancel = context.WithCancel(context.Background())

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create status info
	statusInfo := &status.StatusInfo{
		PID:         os.Getpid(),
		StartTime:   time.Now(),
		Version:     "1.0.0",
		Status:      "RUNNING",
		Host:        "localhost",
		Port:        8080,
		ModelsCount: 0,
		AgentsCount: 0,
	}

	// Start socket server for status queries
	if err := statusManager.StartSocketServer(statusInfo); err != nil {
		statusManager.Cleanup()
		return fmt.Errorf("failed to start status socket server: %w", err)
	}

	// Start the server components
	go func() {
		if err := startServerComponents(serverCtx, configManager, statusInfo); err != nil {
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

	// Cleanup status files
	if err := statusManager.Cleanup(); err != nil && verbose {
		log.Printf("Cleanup error: %v", err)
	}

	fmt.Println("AgentForgeEngine stopped")
	return nil
}

func startServerComponents(ctx context.Context, configManager *config.Manager, statusInfo *status.StatusInfo) error {
	serverConfig := configManager.GetServerConfig()
	if verbose {
		fmt.Printf("Server starting on %s:%d\n", serverConfig.Host, serverConfig.Port)
	}

	// Update status info with server config
	statusInfo.Host = serverConfig.Host
	statusInfo.Port = serverConfig.Port

	// Initialize plugin manager
	userDirs, err := userdirs.NewUserDirectories()
	if err != nil {
		return fmt.Errorf("failed to create user directories: %w", err)
	}

	pluginManager = loader.NewManager(userDirs.AgentsDir, userDirs.CacheDir)

	if verbose {
		fmt.Printf("Plugin manager initialized with plugins dir: %s\n", userDirs.AgentsDir)
	}

	// Load available agents
	agentConfigs := configManager.GetAgentConfigs()
	for _, agentConfig := range agentConfigs {
		if agentConfig.Type == "local" {
			err := pluginManager.LoadLocalAgent(agentConfig.Path, agentConfig.Name)
			if err != nil && verbose {
				log.Printf("Failed to load agent %s: %v", agentConfig.Name, err)
			} else if verbose {
				fmt.Printf("Loaded agent: %s\n", agentConfig.Name)
			}
		}

		// This would be called by model via function_call

	}

	// Initialize model manager
	modelManager := models.NewManager()
	modelConfigs := configManager.GetModelConfigs()
	if err := modelManager.InitializeModels(modelConfigs); err != nil {
		log.Printf("Failed to initialize models: %v", err)
	} else if verbose {
		fmt.Printf("Initialized %d models\n", len(modelConfigs))
	}

	// Initialize HTTP API server
	apiServer := api.NewServer(serverConfig.Host, serverConfig.Port)
	apiServer.SetComponents(statusManager, pluginManager, modelManager)

	// Start API server in goroutine
	go func() {
		if err := apiServer.Start(serverCtx); err != nil {
			log.Printf("API server error: %v", err)
			serverCancel()
		}
	}()

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
		"./configs/afe.yaml",
		"./afe.yaml",
		"$HOME/.afe/configs/afe.yaml",
	}

	for _, path := range defaultPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "./configs/afe.yaml" // fallback
}

func init() {
	rootCmd.AddCommand(startCmd)
}
