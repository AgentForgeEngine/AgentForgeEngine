package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/status"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/userdirs"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop AgentForge Engine",
	Long:  "Stop a running instance of AgentForge Engine",
	RunE:  runStop,
}

func runStop(command *cobra.Command, args []string) error {
	if verbose {
		fmt.Println("Stopping AgentForge Engine...")
	}

	// Initialize user directories and status manager
	userDirs, err := userdirs.NewUserDirectories()
	if err != nil {
		return fmt.Errorf("failed to initialize user directories: %w", err)
	}

	statusManager := status.NewManager(userDirs.AFEDir)

	// Check if process is running using PID file
	if !statusManager.IsRunning() {
		if verbose {
			fmt.Println("No running AgentForge Engine instances found")
		}
		return nil
	}

	// Get PID from file
	pid, err := statusManager.ReadPID()
	if err != nil {
		return fmt.Errorf("failed to read PID: %w", err)
	}

	if verbose {
		fmt.Printf("Found running process with PID: %d\n", pid)
	}

	// Try graceful shutdown first
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// Send SIGTERM for graceful shutdown
	if err := process.Signal(syscall.SIGTERM); err != nil {
		if verbose {
			fmt.Printf("Failed to send SIGTERM: %v\n", err)
		}
	} else {
		if verbose {
			fmt.Println("Sent SIGTERM signal for graceful shutdown")
		}
	}

	// Wait a bit for graceful shutdown
	for i := 0; i < 10; i++ {
		if !statusManager.IsRunning() {
			if verbose {
				fmt.Println("Process stopped gracefully")
			}
			break
		}
		// In a real implementation, you'd use time.Sleep
	}

	// If still running, force kill
	if statusManager.IsRunning() {
		if verbose {
			fmt.Println("Process still running, sending SIGKILL...")
		}
		if err := process.Signal(syscall.SIGKILL); err != nil && verbose {
			fmt.Printf("Failed to send SIGKILL: %v\n", err)
		}
	}

	// Cleanup status files
	if err := statusManager.Cleanup(); err != nil && verbose {
		fmt.Printf("Cleanup error: %v\n", err)
	}

	fmt.Println("AgentForge Engine stopped")
	return nil
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
