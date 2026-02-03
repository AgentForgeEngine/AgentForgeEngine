# AgentForge Engine

AgentForge Engine is a modular agent framework that sits between offline models (llama.cpp, ollama) and agents written in Go. It provides dynamic loading of agents from GitHub repositories, hot reload capabilities, and a unified interface for model interactions.

## Features

- **Hybrid Model Connections**: Support for both llama.cpp and ollama via HTTP and WebSocket
- **Dynamic Agent Loading**: Load agents from local paths or remote GitHub repositories
- **Hot Reload**: Automatic recovery and reloading of failed agents
- **Built-in Agents**: Task agent and file agent included
- **CLI Interface**: Management via Cobra-based command line
- **Configuration Management**: Viper-based YAML configuration

## Quick Start

```bash
# Build AgentForge Engine
go build -o agentforge-engine ./cmd/agentforge

# Run with configuration
./agentforge-engine start --config configs/agentforge.yaml
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
# Start AgentForge Engine
./agentforge-engine start [--config path/to/config.yaml]

# Stop running instance
./agentforge-engine stop

# Check status
./agentforge-engine status

# Reload agents
./agentforge-engine reload [--agent agent-name]
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