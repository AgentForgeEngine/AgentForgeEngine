package main

import (
	"os"

	"github.com/AgentForgeEngine/AgentForgeEngine/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
