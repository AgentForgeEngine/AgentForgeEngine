package interfaces

import "context"

// Agent represents an agent that can process requests
type Agent interface {
	Name() string
	Initialize(config map[string]interface{}) error
	Process(ctx context.Context, input AgentInput) (AgentOutput, error)
	HealthCheck() error
	Shutdown() error
}

// AgentInput represents input to an agent
type AgentInput struct {
	Type     string                 `json:"type"`
	Payload  map[string]interface{} `json:"payload"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AgentOutput represents output from an agent
type AgentOutput struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// Model represents a language model interface
type Model interface {
	Name() string
	Type() ModelType
	Initialize(config ModelConfig) error
	Generate(ctx context.Context, req GenerationRequest) (*GenerationResponse, error)
	HealthCheck() error
	Shutdown() error
}

// ModelType represents the type of model connection
type ModelType string

const (
	ModelTypeHTTP      ModelType = "http"
	ModelTypeWebSocket ModelType = "websocket"
)

// ModelConfig represents configuration for a model
type ModelConfig struct {
	Name     string                 `json:"name"`
	Type     ModelType              `json:"type"`
	Endpoint string                 `json:"endpoint"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// GenerationRequest represents a request to generate text
type GenerationRequest struct {
	Prompt      string                 `json:"prompt"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	StopTokens  []string               `json:"stop_tokens,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

// GenerationResponse represents the response from text generation
type GenerationResponse struct {
	Text     string `json:"text"`
	Tokens   int    `json:"tokens,omitempty"`
	Finished bool   `json:"finished"`
	Model    string `json:"model"`
	Error    string `json:"error,omitempty"`
}

// PluginManager handles dynamic loading of agents
type PluginManager interface {
	LoadLocalAgent(path, name string) error
	LoadRemoteAgent(repo, version, entrypoint string) error
	GetAgent(name string) (Agent, bool)
	ListAgents() []string
	UnloadAgent(name string) error
	ReloadAgent(name string) error
}

// ConfigManager handles configuration loading and management
type ConfigManager interface {
	Load(path string) error
	GetModelConfigs() []ModelConfig
	GetAgentConfigs() []AgentConfig
	GetServerConfig() ServerConfig
	GetRecoveryConfig() RecoveryConfig
	Watch(callback func()) error
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// AgentConfig represents agent configuration
type AgentConfig struct {
	Name       string                 `yaml:"name"`
	Type       string                 `yaml:"type"` // "local" or "remote"
	Path       string                 `yaml:"path,omitempty"`
	Repo       string                 `yaml:"repo,omitempty"`
	Version    string                 `yaml:"version,omitempty"`
	Entrypoint string                 `yaml:"entrypoint,omitempty"`
	Config     map[string]interface{} `yaml:"config,omitempty"`
}

// RecoveryConfig represents recovery configuration
type RecoveryConfig struct {
	HotReload   bool `yaml:"hot_reload"`
	MaxRetries  int  `yaml:"max_retries"`
	BackoffSec  int  `yaml:"backoff_seconds"`
	HealthCheck int  `yaml:"health_check_interval"`
}
