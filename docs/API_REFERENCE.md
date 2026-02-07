# API Reference

This document provides comprehensive API reference for the AgentForgeEngine framework.

## Table of Contents

- [Core Interfaces](#core-interfaces)
  - [Agent Interface](#agent-interface)
  - [Model Interface](#model-interface)
  - [PluginManager Interface](#pluginmanager-interface)
  - [ConfigManager Interface](#configmanager-interface)
- [Data Structures](#data-structures)
  - [AgentInput/Output](#agentinputoutput)
  - [GenerationRequest/Response](#generationrequestresponse)
  - [ModelConfig](#modelconfig)
  - [ServerConfig](#serverconfig)
  - [AgentConfig](#agentconfig)
  - [RecoveryConfig](#recoveryconfig)
- [CLI Commands](#cli-commands)
  - [Build Commands](#build-commands)
  - [User Management Commands](#user-management-commands)
  - [System Commands](#system-commands)
  - [Cache Commands](#cache-commands)
  - [Testing Commands](#testing-commands)
- [Package APIs](#package-apis)
  - [Authentication Package](#authentication-package)
  - [Cache Package](#cache-package)
  - [Status Package](#status-package)
  - [Hot Reload Package](#hot-reload-package)
  - [User Directories Package](#user-directories-package)

## Core Interfaces

### Agent Interface

The `Agent` interface defines the contract that all agents must implement.

```go
type Agent interface {
    Name() string
    Initialize(config map[string]interface{}) error
    Process(ctx context.Context, input AgentInput) (AgentOutput, error)
    HealthCheck() error
    Shutdown() error
}
```

#### Methods

- **Name() string**: Returns the unique name of the agent
- **Initialize(config map[string]interface{}) error**: Initializes the agent with configuration
- **Process(ctx context.Context, input AgentInput) (AgentOutput, error)**: Processes requests and returns results
- **HealthCheck() error**: Performs a health check to verify the agent is functioning
- **Shutdown() error**: Gracefully shuts down the agent and releases resources

#### Example Implementation

```go
type MyAgent struct {
    name string
    config map[string]interface{}
}

func NewMyAgent() *MyAgent {
    return &MyAgent{name: "my-agent"}
}

func (a *MyAgent) Name() string {
    return a.name
}

func (a *MyAgent) Initialize(config map[string]interface{}) error {
    a.config = config
    return nil
}

func (a *MyAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
    // Process the input and return output
    return interfaces.AgentOutput{
        Success: true,
        Data: map[string]interface{}{
            "result": "processed successfully",
        },
    }, nil
}

func (a *MyAgent) HealthCheck() error {
    return nil // Return error if unhealthy
}

func (a *MyAgent) Shutdown() error {
    // Cleanup resources
    return nil
}

// Export for plugin loading
var Agent interfaces.Agent = NewMyAgent()
```

### Model Interface

The `Model` interface defines the contract for language model providers.

```go
type Model interface {
    Name() string
    Type() ModelType
    Initialize(config ModelConfig) error
    Generate(ctx context.Context, req GenerationRequest) (*GenerationResponse, error)
    HealthCheck() error
    Shutdown() error
}
```

#### Model Types

```go
type ModelType string

const (
    ModelTypeHTTP      ModelType = "http"
    ModelTypeWebSocket ModelType = "websocket"
)
```

#### Methods

- **Name() string**: Returns the model name
- **Type() ModelType**: Returns the connection type (HTTP or WebSocket)
- **Initialize(config ModelConfig) error**: Initializes the model with configuration
- **Generate(ctx context.Context, req GenerationRequest) (*GenerationResponse, error)**: Generates text from prompts
- **HealthCheck() error**: Checks model availability
- **Shutdown() error**: Gracefully shuts down the model connection

### PluginManager Interface

The `PluginManager` interface handles dynamic loading and management of agents.

```go
type PluginManager interface {
    LoadLocalAgent(path, name string) error
    LoadRemoteAgent(repo, version, entrypoint string) error
    GetAgent(name string) (Agent, bool)
    ListAgents() []string
    UnloadAgent(name string) error
    ReloadAgent(name string) error
}
```

#### Methods

- **LoadLocalAgent(path, name string) error**: Loads an agent from local path
- **LoadRemoteAgent(repo, version, entrypoint string) error**: Loads an agent from remote repository
- **GetAgent(name string) (Agent, bool)**: Retrieves an agent by name
- **ListAgents() []string**: Returns list of loaded agent names
- **UnloadAgent(name string) error**: Unloads an agent
- **ReloadAgent(name string) error**: Reloads an agent (hot reload)

### ConfigManager Interface

The `ConfigManager` interface handles configuration loading and management.

```go
type ConfigManager interface {
    Load(path string) error
    GetModelConfigs() []ModelConfig
    GetAgentConfigs() []AgentConfig
    GetServerConfig() ServerConfig
    GetRecoveryConfig() RecoveryConfig
    Watch(callback func()) error
}
```

#### Methods

- **Load(path string) error**: Loads configuration from file
- **GetModelConfigs() []ModelConfig**: Returns model configurations
- **GetAgentConfigs() []AgentConfig**: Returns agent configurations
- **GetServerConfig() ServerConfig**: Returns server configuration
- **GetRecoveryConfig() RecoveryConfig**: Returns recovery configuration
- **Watch(callback func()) error**: Watches for configuration changes

## Data Structures

### AgentInput/Output

#### AgentInput

```go
type AgentInput struct {
    Type     string                 `json:"type"`
    Payload  map[string]interface{} `json:"payload"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

**Fields:**
- **Type**: The operation type (e.g., "fetch", "validate", "extract")
- **Payload**: The main data payload for the operation
- **Metadata**: Optional metadata for the operation

#### AgentOutput

```go
type AgentOutput struct {
    Success bool                   `json:"success"`
    Data    map[string]interface{} `json:"data,omitempty"`
    Error   string                 `json:"error,omitempty"`
}
```

**Fields:**
- **Success**: Indicates if the operation was successful
- **Data**: Result data (only present on success)
- **Error**: Error message (only present on failure)

### GenerationRequest/Response

#### GenerationRequest

```go
type GenerationRequest struct {
    Prompt      string                 `json:"prompt"`
    MaxTokens   int                    `json:"max_tokens,omitempty"`
    Temperature float64                `json:"temperature,omitempty"`
    StopTokens  []string               `json:"stop_tokens,omitempty"`
    Stream      bool                   `json:"stream,omitempty"`
    Options     map[string]interface{} `json:"options,omitempty"`
}
```

**Fields:**
- **Prompt**: The text prompt to generate from
- **MaxTokens**: Maximum tokens to generate (optional)
- **Temperature**: Sampling temperature (optional)
- **StopTokens**: Tokens that stop generation (optional)
- **Stream**: Whether to stream the response (optional)
- **Options**: Additional model-specific options (optional)

#### GenerationResponse

```go
type GenerationResponse struct {
    Text     string `json:"text"`
    Tokens   int    `json:"tokens,omitempty"`
    Finished bool   `json:"finished"`
    Model    string `json:"model"`
    Error    string `json:"error,omitempty"`
}
```

**Fields:**
- **Text**: Generated text
- **Tokens**: Number of tokens generated (optional)
- **Finished**: Whether generation is complete
- **Model**: Model name that generated the response
- **Error**: Error message (optional)

### ModelConfig

```go
type ModelConfig struct {
    Name     string                 `json:"name"`
    Type     ModelType              `json:"type"`
    Endpoint string                 `json:"endpoint"`
    Options  map[string]interface{} `json:"options,omitempty"`
}
```

**Fields:**
- **Name**: Model name/identifier
- **Type**: Connection type (HTTP/WebSocket)
- **Endpoint**: Model endpoint URL
- **Options**: Model-specific configuration options

### ServerConfig

```go
type ServerConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}
```

**Fields:**
- **Host**: Server host address
- **Port**: Server port number

### AgentConfig

```go
type AgentConfig struct {
    Name       string                 `yaml:"name"`
    Type       string                 `yaml:"type"` // "local" or "remote"
    Path       string                 `yaml:"path,omitempty"`
    Repo       string                 `yaml:"repo,omitempty"`
    Version    string                 `yaml:"version,omitempty"`
    Entrypoint string                 `yaml:"entrypoint,omitempty"`
    Config     map[string]interface{} `yaml:"config,omitempty"`
}
```

**Fields:**
- **Name**: Agent name
- **Type**: Agent type ("local" or "remote")
- **Path**: Local path (for local agents)
- **Repo**: Repository URL (for remote agents)
- **Version**: Version (for remote agents)
- **Entrypoint**: Entry point (for remote agents)
- **Config**: Agent-specific configuration

### RecoveryConfig

```go
type RecoveryConfig struct {
    HotReload   bool `yaml:"hot_reload"`
    MaxRetries  int  `yaml:"max_retries"`
    BackoffSec  int  `yaml:"backoff_seconds"`
    HealthCheck int  `yaml:"health_check_interval"`
}
```

**Fields:**
- **HotReload**: Enable hot reload functionality
- **MaxRetries**: Maximum retry attempts
- **BackoffSec**: Backoff delay in seconds
- **HealthCheck**: Health check interval in seconds

## CLI Commands

### Build Commands

#### `afe build all`
Builds all providers and agents with intelligent caching.

**Flags:**
- `--parallel, -p`: Build plugins in parallel (default: true)
- `--force`: Force rebuild of all plugins
- `--clean`: Clean cache and rebuild all plugins
- `--verbose, -v`: Verbose build output

**Example:**
```bash
afe build all --verbose --parallel
```

#### `afe build providers`
Builds only provider plugins.

**Arguments:**
- `[name...]`: Optional specific provider names to build

**Example:**
```bash
afe build providers qwen3 json-rpc-bridge
```

#### `afe build agents`
Builds only agent plugins.

**Arguments:**
- `[name...]`: Optional specific agent names to build

**Example:**
```bash
afe build agents ls cat web-agent
```

### User Management Commands

#### `afe user create`
Creates a new user account.

**Flags:**
- `--name`: User's full name
- `--email`: User's email address
- `--password`: User's password
- `--phone`: Optional phone number

**Example:**
```bash
afe user create --name "John Doe" --email "john@example.com" --password "secure123"
```

#### `afe user login`
Authenticates a user.

**Flags:**
- `--email`: User's email address
- `--password`: User's password

**Example:**
```bash
afe user login --email "john@example.com" --password "secure123"
```

#### `afe user api-key create`
Creates an API key for a user.

**Flags:**
- `--name`: API key name/description
- `--email`: User's email address
- `--expires`: Optional expiration duration (e.g., "30d", "24h")
- `--scopes`: Optional comma-separated list of scopes

**Example:**
```bash
afe user api-key create --name "Production Key" --email "john@example.com" --expires "30d"
```

#### `afe user api-key list`
Lists API keys for a user.

**Flags:**
- `--email`: User's email address

**Example:**
```bash
afe user api-key list --email "john@example.com"
```

### System Commands

#### `afe init`
Initializes user directories and migrates existing plugins.

**Flags:**
- `--migrate`: Migrate existing plugins from project directory
- `--verbose, -v`: Verbose output

**Example:**
```bash
afe init --migrate --verbose
```

#### `afe start`
Starts the AgentForgeEngine with status tracking.

**Flags:**
- `--config, -c`: Path to configuration file
- `--daemon, -d`: Run in daemon mode

**Example:**
```bash
afe start --config ./config.yaml --daemon
```

#### `afe stop`
Stops the AgentForgeEngine gracefully.

**Flags:**
- `--force, -f`: Force stop (SIGKILL)

**Example:**
```bash
afe stop
```

#### `afe status`
Checks the engine status.

**Flags:**
- `--verbose, -v`: Show detailed status information

**Example:**
```bash
afe status --verbose
```

### Cache Commands

#### `afe cache status`
Shows build cache statistics.

**Example:**
```bash
afe cache status
```

#### `afe cache clean`
Cleans the build cache.

**Flags:**
- `--force, -f`: Force clean without confirmation

**Example:**
```bash
afe cache clean --force
```

#### `afe cache validate`
Validates cache integrity.

**Example:**
```bash
afe cache validate
```

### Testing Commands

#### `./scripts/test_agents.sh integration`
Runs integration tests for all agents.

**Example:**
```bash
./scripts/test_agents.sh integration
```

#### `./scripts/test_agents.sh agent <name>`
Tests a specific agent.

**Example:**
```bash
./scripts/test_agents.sh agent ls
```

#### `./scripts/test_agents.sh all`
Runs comprehensive tests for all agents.

**Example:**
```bash
./scripts/test_agents.sh all
```

## Package APIs

### Authentication Package

#### UserManager

The `UserManager` handles secure user management with LevelDB storage.

```go
type UserManager struct {
    // Internal fields
}

func NewUserManager(accountsDir string) (*UserManager, error)
func (um *UserManager) Close() error
func (um *UserManager) CreateUser(name, email, password string, phoneNumber *string) (*User, error)
func (um *UserManager) AuthenticateUser(email, password string) (*User, error)
func (um *UserManager) GetUserByEmail(email string) (*User, error)
func (um *UserManager) GetUserByUID(uid string) (*User, error)
func (um *UserManager) UpdateUser(uid string, updates map[string]interface{}) (*User, error)
func (um *UserManager) DeleteUser(uid string) error
func (um *UserManager) CreateAPIKey(uid, name string, expiresAt *time.Time, scopes []string) (*APIKey, string, error)
func (um *UserManager) ValidateAPIKey(apiKey string) (*User, *APIKey, error)
```

#### User Structure

```go
type User struct {
    UID          string     `json:"uid"`
    Name         string     `json:"name"`
    Email        string     `json:"email"`
    PhoneNumber  string     `json:"phone_number,omitempty"`
    PasswordHash string     `json:"password_hash"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
    LastLogin    *time.Time `json:"last_login,omitempty"`
    IsActive     bool       `json:"is_active"`
    Roles        []string   `json:"roles,omitempty"`
}
```

#### APIKey Structure

```go
type APIKey struct {
    UID       string     `json:"uid"`
    KeyID     string     `json:"key_id"`
    KeyHash   string     `json:"key_hash"`
    Name      string     `json:"name"`
    CreatedAt time.Time  `json:"created_at"`
    ExpiresAt *time.Time `json:"expires_at,omitempty"`
    LastUsed  *time.Time `json:"last_used,omitempty"`
    IsActive  bool       `json:"is_active"`
    Scopes    []string   `json:"scopes,omitempty"`
}
```

### Cache Package

#### BuildCache Manager

The build cache manager handles intelligent caching of plugins.

```go
type Manager struct {
    // Internal fields
}

func NewManager() (*Manager, error)
func (m *Manager) LoadCache() error
func (m *Manager) SaveCache() error
func (m *Manager) ShouldRebuild(pluginType, pluginName, pluginPath string) (bool, string, error)
func (m *Manager) UpdatePlugin(pluginType, pluginName, pluginPath string, buildDurationMs int, pluginSizeBytes int64) error
func (m *Manager) RecordBuildHistory(command string, pluginsBuilt, pluginsCached []string, totalDurationMs int, success bool)
func (m *Manager) GetCacheStatus() (*CacheStatus, error)
```

#### CacheStatus Structure

```go
type CacheStatus struct {
    Version          string    `yaml:"version"`
    CreatedAt        time.Time `yaml:"created_at"`
    UpdatedAt        time.Time `yaml:"updated_at"`
    TotalBuilds      int       `yaml:"total_builds"`
    CacheHits        int       `yaml:"cache_hits"`
    CacheMisses      int       `yaml:"cache_misses"`
    CacheHitRate     float64   `yaml:"cache_hit_rate"`
    AverageBuildTime int       `yaml:"average_build_time_ms"`
    TotalCacheSize   float64   `yaml:"total_cache_size_mb"`
    CacheValid       bool      `yaml:"cache_valid"`
    LastValidation   time.Time `yaml:"last_validation"`
    ProvidersCached  int       `yaml:"providers_cached"`
    AgentsCached     int       `yaml:"agents_cached"`
}
```

### Status Package

#### Status Manager

The status manager handles PID file and Unix socket for status tracking.

```go
type Manager struct {
    // Internal fields
}

func NewManager(afeDir string) *Manager
func (m *Manager) WritePID() error
func (m *Manager) ReadPID() (int, error)
func (m *Manager) IsRunning() bool
func (m *Manager) Cleanup() error
func (m *Manager) StartSocketServer(statusInfo *StatusInfo) error
func (m *Manager) GetStatusViaSocket() (*StatusInfo, error)
func (m *Manager) GetBasicStatus() *StatusInfo
```

#### StatusInfo Structure

```go
type StatusInfo struct {
    PID         int       `json:"pid"`
    StartTime   time.Time `json:"start_time"`
    Version     string    `json:"version"`
    Uptime      string    `json:"uptime"`
    Status      string    `json:"status"`
    Host        string    `json:"host"`
    Port        int       `json:"port"`
    ModelsCount int       `json:"models_count"`
    AgentsCount int       `json:"agents_count"`
}
```

### Hot Reload Package

#### Hot Reload Manager

The hot reload manager handles automatic plugin reloading.

```go
type Manager struct {
    // Internal fields
}

func NewManager() *Manager
func (m *Manager) StartWatching(paths []string, callback func(string)) error
func (m *Manager) StopWatching() error
func (m *Manager) TriggerReload(pluginName string) error
```

### User Directories Package

#### UserDirectories

The user directories manager manages the AFE user directory structure.

```go
type UserDirectories struct {
    // Internal fields
}

func NewUserDirectories() (*UserDirectories, error)
func (ud *UserDirectories) Exists() bool
func (ud *UserDirectories) Create() error
func (ud *UserDirectories) GetPluginOutputPath(pluginType, pluginName string) string
func (ud *UserDirectories) GetCachePath() string
func (ud *UserDirectories) GetAccountsPath() string
func (ud *UserDirectories) GetLogsPath() string
func (ud *UserDirectories) GetConfigPath() string
```

## Error Handling

All API methods return errors following Go conventions. Common error types include:

- **ValidationError**: Invalid input parameters
- **AuthenticationError**: Authentication/authorization failures
- **NotFoundError**: Resource not found
- **ConfigurationError**: Configuration-related issues
- **NetworkError**: Network connectivity problems
- **PluginError**: Plugin-related issues

## Version Information

Current API version: v1.0

For backward compatibility, all API calls include version headers and semantic versioning for major changes.