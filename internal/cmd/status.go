package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

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

	// Check if the process is running
	running := isRunning()

	if running {
		fmt.Println("Status: RUNNING ✓")
		fmt.Println("Process: AgentForge Engine is active")

		// Try to get more detailed status
		if verbose {
			fmt.Println("\nDetailed Information:")
			printDetailedStatus()
		}
	} else {
		fmt.Println("Status: STOPPED ✗")
		fmt.Println("Process: No AgentForge Engine instance found")
	}

	return nil
}

func isRunning() bool {
	var command string
	var args []string

	switch runtime.GOOS {
	case "windows":
		command = "tasklist"
		args = []string{"/FI", "IMAGENAME eq agentforge-engine.exe"}
	case "linux", "darwin":
		command = "pgrep"
		args = []string{"-f", "agentforge-engine"}
	default:
		return false
	}

	cmd := exec.Command(command, args...)
	err := cmd.Run()
	return err == nil
}

func printDetailedStatus() {
	// This will be expanded when we implement the actual status endpoints
	// For now, just show placeholder information

	fmt.Println("  - Models: Not connected")
	fmt.Println("  - Agents: 0 loaded")
	fmt.Println("  - Uptime: Unknown")
	fmt.Println("  - Memory: Unknown")

	// Show config file being used
	if cfgFile != "" {
		fmt.Printf("  - Config: %s\n", cfgFile)
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
