# AgentForgeEngine (AFE)

AgentForgeEngine (AFE) is a modular agent framework that sits between offline models (llama.cpp, ollama) and agents written in Go. It provides dynamic loading of **agents and providers** from GitHub repositories, hot reload capabilities, and a unified interface for model interactions.

## 🚀 New: Provider Plugin Architecture

AgentForgeEngine now supports **provider plugins** for model connections, enabling extensible support for different protocols:

- **JSON-RPC Bridge**: WebSocket-based connections (supports ollama-websocket-gateway)
- **HTTP Providers**: REST API connections (supports llama.cpp, OpenAI, etc.)
- **WebSocket Providers**: Native WebSocket protocols (supports Ollama, custom implementations)
- **gRPC Providers**: Binary protocol connections (future support)

### **Provider System Benefits**
- **Hot Reloadable**: `afe reload --provider name` for zero-downtime updates
- **Protocol Flexible**: Easy addition of new connection types
- **Same Pattern**: Uses exact same approach as agent plugins
- **Backward Compatible**: Legacy model configs still supported with warnings

## Features

- **Hybrid Model Connections**: Support for both llama.cpp and ollama via HTTP and WebSocket
- **Dynamic Agent Loading**: Load agents from local paths or remote GitHub repositories
- **Hot Reload**: Automatic recovery and reloading of failed agents
- **Built-in Agents**: Task agent and file agent included
- **CLI Interface**: Management via Cobra-based command line
- **Configuration Management**: Viper-based YAML configuration

## Quick Start

```bash
# Build AgentForgeEngine (AFE)
go build -o afe ./cmd/agentforge

# Run with configuration
./afe start --config configs/agentforge.yaml
```

## Configuration

Edit `configs/agentforge.yaml` to configure models and agents:

```yaml
server:
  port: 8080
  host: "localhost"

models:
  - name: "llamacpp"
    type: "http"
    endpoint: "http://localhost:8081"
  - name: "ollama"
    type: "websocket"
    endpoint: "ws://localhost:11434"

agents:
  local:
    - name: "task-agent"
      path: "./agents/task-agent"
    - name: "file-agent"
      path: "./agents/file-agent"
  remote:
    - name: "code-assistant"
      repo: "github.com/user/agent-code-assistant"
      version: "latest"

recovery:
  hot_reload: true
  max_retries: 3
  backoff_seconds: 5
```

## CLI Commands

```bash
# Start AgentForgeEngine
./afe start [--config path/to/config.yaml]

# Stop running instance
./afe stop

# Check status
./afe status

# Reload agents
./afe reload [--agent agent-name]
```

## Agent Development

Agents implement the Agent interface:

```go
type Agent interface {
    Name() string
    Initialize(config map[string]interface{}) error
    Process(ctx context.Context, input AgentInput) (AgentOutput, error)
    HealthCheck() error
    Shutdown() error
}
```

Export your agent as a plugin:

```go
package main

import "github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"

var Agent interfaces.Agent = &MyAgent{}
```