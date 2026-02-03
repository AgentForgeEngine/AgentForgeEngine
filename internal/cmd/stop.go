package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

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

	// Try to stop gracefully by finding and killing the process
	var cmdStr string
	var cmdArgs []string

	switch runtime.GOOS {
	case "windows":
		cmdStr = "taskkill"
		cmdArgs = []string{"/F", "/IM", "agentforge-engine.exe"}
	case "linux", "darwin":
		cmdStr = "pkill"
		cmdArgs = []string{"-f", "agentforge-engine"}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	stopCmd := exec.Command(cmdStr, cmdArgs...)
	if err := stopCmd.Run(); err != nil {
		if verbose {
			fmt.Printf("No running AgentForge Engine instances found\n")
		}
		return nil
	}

	fmt.Println("AgentForge Engine stopped")
	return nil
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
