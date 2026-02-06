package cmd

import (
	"fmt"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/status"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/userdirs"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show AgentForge Engine status",
	Long:  "Display the current status of AgentForge Engine and its components",
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	fmt.Println("AgentForge Engine Status:")
	fmt.Println("=========================")

	// Initialize user directories and status manager
	userDirs, err := userdirs.NewUserDirectories()
	if err != nil {
		return fmt.Errorf("failed to initialize user directories: %w", err)
	}

	statusManager := status.NewManager(userDirs.AFEDir)

	// First try to get detailed status via socket
	statusInfo, err := statusManager.GetStatusViaSocket()
	if err == nil {
		// Got detailed status from socket
		printDetailedStatus(statusInfo)
	} else {
		// Fallback to basic PID file status
		statusInfo = statusManager.GetBasicStatus()
		printBasicStatus(statusInfo)
	}

	return nil
}

func printDetailedStatus(statusInfo *status.StatusInfo) {
	fmt.Printf("Status: %s ✓\n", statusInfo.Status)
	fmt.Printf("Process: AgentForge Engine is active (PID: %d)\n", statusInfo.PID)

	if verbose {
		fmt.Println("\nDetailed Information:")
		fmt.Printf("  - Version: %s\n", statusInfo.Version)
		fmt.Printf("  - Uptime: %s\n", statusInfo.Uptime)
		fmt.Printf("  - Start Time: %s\n", statusInfo.StartTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("  - Server: %s:%d\n", statusInfo.Host, statusInfo.Port)
		fmt.Printf("  - Models Loaded: %d\n", statusInfo.ModelsCount)
		fmt.Printf("  - Agents Loaded: %d\n", statusInfo.AgentsCount)

		// Show config file being used
		if cfgFile != "" {
			fmt.Printf("  - Config: %s\n", cfgFile)
		}
	}
}

func printBasicStatus(statusInfo *status.StatusInfo) {
	if statusInfo.Status == "RUNNING" {
		fmt.Printf("Status: %s ✓\n", statusInfo.Status)
		fmt.Printf("Process: AgentForge Engine is active (PID: %d)\n", statusInfo.PID)

		if verbose {
			fmt.Println("\nBasic Information:")
			fmt.Println("  - Detailed status unavailable (socket not connected)")
			fmt.Println("  - Use './afe start --verbose' for detailed logging")

			// Show config file being used
			if cfgFile != "" {
				fmt.Printf("  - Config: %s\n", cfgFile)
			}
		}
	} else {
		fmt.Printf("Status: %s ✗\n", statusInfo.Status)
		fmt.Println("Process: No AgentForge Engine instance found")
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
