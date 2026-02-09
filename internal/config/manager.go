package config

import (
	"fmt"
	"log"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Manager struct {
	config *Config
	v      *viper.Viper
}

type Config struct {
	Server       interfaces.ServerConfig   `yaml:"server"`
	Models       []interfaces.ModelConfig  `yaml:"models"`
	Agents       AgentsConfig              `yaml:"agents"`
	Recovery     interfaces.RecoveryConfig `yaml:"recovery"`
	Orchestrator OrchestratorConfig        `yaml:"orchestrator"`
}

type OrchestratorConfig struct {
	Enabled            bool   `yaml:"enabled"`
	MaxConcurrentTasks int    `yaml:"max_concurrent_tasks"`
	TaskTimeout        string `yaml:"task_timeout"`
	RetryAttempts      int    `yaml:"retry_attempts"`
	TaskQueueSize      int    `yaml:"task_queue_size"`
}

type AgentsConfig struct {
	Local  []interfaces.AgentConfig `yaml:"local"`
	Remote []interfaces.AgentConfig `yaml:"remote"`
}

func NewManager() *Manager {
	return &Manager{
		v: viper.New(),
	}
}

func (m *Manager) Load(path string) error {
	// Set config file path and name
	m.v.SetConfigFile(path)

	// Set default values
	m.setDefaults()

	// Read config file
	if err := m.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("Config file not found, using defaults: %s", err)
			// Continue with defaults - don't return error
		} else {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal config
	var config Config
	if err := m.v.Unmarshal(&config); err != nil {
		return fmt.Errorf("unable to decode config: %w", err)
	}

	m.config = &config
	return nil
}

func (m *Manager) setDefaults() {
	// Server defaults
	m.v.SetDefault("server.host", "localhost")
	m.v.SetDefault("server.port", 8080)

	// Recovery defaults
	m.v.SetDefault("recovery.hot_reload", true)
	m.v.SetDefault("recovery.max_retries", 3)
	m.v.SetDefault("recovery.backoff_sec", 5)
	m.v.SetDefault("recovery.health_check", 30)
	m.v.SetDefault("recovery.backoff_seconds", 5)
	m.v.SetDefault("recovery.health_check_interval", 30)
}

func (m *Manager) GetModelConfigs() []interfaces.ModelConfig {
	if m.config == nil {
		return []interfaces.ModelConfig{}
	}
	return m.config.Models
}

func (m *Manager) GetAgentConfigs() []interfaces.AgentConfig {
	if m.config == nil {
		return []interfaces.AgentConfig{}
	}

	var allAgents []interfaces.AgentConfig

	// Add local agents
	for _, agent := range m.config.Agents.Local {
		agent.Type = "local"
		allAgents = append(allAgents, agent)
	}

	// Add remote agents
	for _, agent := range m.config.Agents.Remote {
		agent.Type = "remote"
		allAgents = append(allAgents, agent)
	}

	return allAgents
}

func (m *Manager) GetServerConfig() interfaces.ServerConfig {
	if m.config == nil {
		return interfaces.ServerConfig{
			Host: "localhost",
			Port: 8080,
		}
	}
	return m.config.Server
}

func (m *Manager) GetRecoveryConfig() interfaces.RecoveryConfig {
	return m.config.Recovery
}

func (m *Manager) GetOrchestratorConfig() OrchestratorConfig {
	return m.config.Orchestrator
}

func (m *Manager) Watch(callback func()) error {
	m.v.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("Config file changed: %s", e.Name)
		if err := m.Load(m.v.ConfigFileUsed()); err != nil {
			log.Printf("Error reloading config: %v", err)
			return
		}
		callback()
	})
	m.v.WatchConfig()
	return nil
}
