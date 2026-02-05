package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/internal/loader"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/hotreload"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/userdirs"
)

// triggerHotReload triggers hot reload for updated plugins
func triggerHotReload(buildPlan *BuildPlan, userDirs *userdirs.UserDirectories) error {
	// Create plugin manager for hot reload
	pluginManager := loader.NewManager("", "")

	// Create hot reload manager
	hotReloadManager := hotreload.NewManager(pluginManager, userDirs)

	// Start hot reload manager
	if err := hotReloadManager.Start(); err != nil {
		return fmt.Errorf("failed to start hot reload manager: %w", err)
	}
	defer hotReloadManager.Stop()

	// Queue hot reload for updated providers
	for _, provider := range buildPlan.ProvidersToBuild {
		if err := hotReloadManager.ReloadPlugin("provider", provider, false); err != nil {
			log.Printf("⚠️  Failed to queue hot reload for provider %s: %v", provider, err)
		}
	}

	// Queue hot reload for updated agents
	for _, agent := range buildPlan.AgentsToBuild {
		if err := hotReloadManager.ReloadPlugin("agent", agent, false); err != nil {
			log.Printf("⚠️  Failed to queue hot reload for agent %s: %v", agent, err)
		}
	}

	// Wait a moment for hot reload to complete
	time.Sleep(100 * time.Millisecond)

	return nil
}
