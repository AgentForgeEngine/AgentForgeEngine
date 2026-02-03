package recovery

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/internal/loader"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type Manager struct {
	config          interfaces.RecoveryConfig
	pluginManager   *loader.Manager
	agentConfigs    []interfaces.AgentConfig
	stopChan        chan struct{}
	retryCounts     map[string]int
	lastFailureTime map[string]time.Time
}

func NewManager(config interfaces.RecoveryConfig, pluginManager *loader.Manager) *Manager {
	return &Manager{
		config:          config,
		pluginManager:   pluginManager,
		stopChan:        make(chan struct{}),
		retryCounts:     make(map[string]int),
		lastFailureTime: make(map[string]time.Time),
	}
}

func (rm *Manager) Start(ctx context.Context, agentConfigs []interfaces.AgentConfig) error {
	rm.agentConfigs = agentConfigs

	if !rm.config.HotReload {
		log.Println("Hot reload disabled")
		return nil
	}

	log.Printf("Starting recovery system with health check interval: %d seconds", rm.config.HealthCheck)

	// Start health monitoring
	go rm.healthMonitor(ctx)

	return nil
}

func (rm *Manager) Stop() {
	close(rm.stopChan)
}

func (rm *Manager) healthMonitor(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(rm.config.HealthCheck) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rm.stopChan:
			return
		case <-ticker.C:
			rm.performHealthCheck(ctx)
		}
	}
}

func (rm *Manager) performHealthCheck(ctx context.Context) {
	// Get health status of all agents
	results := rm.pluginManager.HealthCheckAll(ctx)

	for agentName, err := range results {
		if err != nil {
			rm.handleAgentFailure(agentName, err)
		} else {
			// Reset retry count on successful health check
			if count, exists := rm.retryCounts[agentName]; exists && count > 0 {
				log.Printf("Agent %s is healthy again, resetting retry count", agentName)
				rm.retryCounts[agentName] = 0
				delete(rm.lastFailureTime, agentName)
			}
		}
	}
}

func (rm *Manager) handleAgentFailure(agentName string, err error) {
	log.Printf("Agent %s health check failed: %v", agentName, err)

	// Check if we should retry
	if !rm.shouldRetry(agentName) {
		log.Printf("Agent %s max retries exceeded, giving up", agentName)
		return
	}

	// Calculate backoff delay
	backoff := rm.calculateBackoff(agentName)
	if backoff > 0 {
		log.Printf("Waiting %v before retrying agent %s", backoff, agentName)
		time.Sleep(backoff)
	}

	// Attempt recovery
	if err := rm.reloadAgent(agentName); err != nil {
		log.Printf("Failed to reload agent %s: %v", agentName, err)
		rm.retryCounts[agentName]++
		rm.lastFailureTime[agentName] = time.Now()
	} else {
		log.Printf("Successfully reloaded agent %s", agentName)
		rm.retryCounts[agentName] = 0
		delete(rm.lastFailureTime, agentName)
	}
}

func (rm *Manager) shouldRetry(agentName string) bool {
	retryCount := rm.retryCounts[agentName]
	return retryCount < rm.config.MaxRetries
}

func (rm *Manager) calculateBackoff(agentName string) time.Duration {
	retryCount := rm.retryCounts[agentName]

	// Exponential backoff: 2^retryCount * base backoff
	backoffSeconds := rm.config.BackoffSec * (1 << retryCount)

	// Cap at 5 minutes
	if backoffSeconds > 300 {
		backoffSeconds = 300
	}

	return time.Duration(backoffSeconds) * time.Second
}

func (rm *Manager) reloadAgent(agentName string) error {
	// Find the agent configuration
	var agentConfig *interfaces.AgentConfig
	for _, config := range rm.agentConfigs {
		if config.Name == agentName {
			agentConfig = &config
			break
		}
	}

	if agentConfig == nil {
		return fmt.Errorf("agent configuration not found for %s", agentName)
	}

	// Unload the failed agent
	if err := rm.pluginManager.UnloadAgent(agentName); err != nil {
		log.Printf("Failed to unload agent %s: %v", agentName, err)
		// Continue anyway and try to load fresh
	}

	// Reload the agent based on type
	switch agentConfig.Type {
	case "local":
		if agentConfig.Path == "" {
			return fmt.Errorf("local agent %s missing path", agentName)
		}
		return rm.pluginManager.LoadLocalAgent(agentConfig.Path, agentName)

	case "remote":
		if agentConfig.Repo == "" {
			return fmt.Errorf("remote agent %s missing repo", agentName)
		}
		version := agentConfig.Version
		if version == "" {
			version = "latest"
		}
		entrypoint := agentConfig.Entrypoint
		if entrypoint == "" {
			entrypoint = "main.go"
		}
		return rm.pluginManager.LoadRemoteAgent(agentConfig.Repo, version, entrypoint)

	default:
		return fmt.Errorf("unknown agent type: %s", agentConfig.Type)
	}
}

func (rm *Manager) GetStatus() map[string]interface{} {
	status := make(map[string]interface{})

	agentsStatus := make(map[string]interface{})
	for agentName := range rm.retryCounts {
		agentsStatus[agentName] = map[string]interface{}{
			"retry_count":   rm.retryCounts[agentName],
			"last_failure":  rm.lastFailureTime[agentName],
			"should_retry":  rm.shouldRetry(agentName),
			"next_retry_in": rm.calculateBackoff(agentName),
		}
	}

	status["enabled"] = rm.config.HotReload
	status["max_retries"] = rm.config.MaxRetries
	status["backoff_seconds"] = rm.config.BackoffSec
	status["health_check_interval"] = rm.config.HealthCheck
	status["agents"] = agentsStatus

	return status
}

// Manual reload trigger for individual agents
func (rm *Manager) ReloadAgent(agentName string) error {
	return rm.reloadAgent(agentName)
}

// Manual reload trigger for all agents
func (rm *Manager) ReloadAllAgents(ctx context.Context) error {
	var lastErr error

	for _, config := range rm.agentConfigs {
		if err := rm.reloadAgent(config.Name); err != nil {
			log.Printf("Failed to reload agent %s: %v", config.Name, err)
			lastErr = err
		}
	}

	return lastErr
}
