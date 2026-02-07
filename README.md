# AgentForgeEngine 🚀

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/AgentForgeEngine/AgentForgeEngine)
[![Testing Framework](https://img.shields.io/badge/Tests-Comprehensive-orange.svg)](docs/AGENT_TESTING.md)
[![Function Response](https://img.shields.io/badge/Protocol-Standardized-purple.svg)](docs/AGENT_TESTING.md)

A modular, high-performance agent framework that sits between offline models (llama.cpp, ollama) and agents written in Go. Features dynamic loading of agents, hot reload capabilities, and a unified interface for model interactions.

## 🌟 Key Features

- **🏗️ Intelligent Build System**: Smart caching with YAML-based build management
- **🔄 Hot Reload Integration**: Zero-downtime plugin updates
- **🛡️ Secure User Management**: Enterprise-grade authentication with LevelDB
- **📦 Plugin Architecture**: Dynamic loading of providers and agents
- **⚡ High Performance**: Parallel builds and optimized caching
- **🔧 Developer-Friendly**: Comprehensive CLI with clear feedback
- **🌐 Cross-Platform**: Works on Linux, macOS, and Windows
- **📊 Advanced Status Tracking**: Hybrid PID file and Unix socket monitoring
- **🧪 Comprehensive Testing**: Model-independent agent testing framework
- **🤖 Function Response Format**: Standardized agent-model communication protocol

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Installation](#installation)
- [Configuration](#configuration)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

## 📚 Documentation

- **[Documentation Index](docs/README.md)** - Complete documentation overview and navigation

### Core Documentation
- **[Agent Testing Guide](docs/AGENT_TESTING.md)** - Agent testing framework documentation
- **[Build System Guide](docs/BUILD_SYSTEM.md)** - Build system architecture and usage
- **[User Management Guide](docs/USER_MANAGEMENT.md)** - User management system guide

### API & Development
- **[API Reference](docs/API_REFERENCE.md)** - Comprehensive API reference
- **[Plugin Development](#plugin-development)** - Plugin development guide

### Implementation & Deployment
- **[New File System Agents](docs/NEW_FILE_SYSTEM_AGENTS.md)** - New file system agents implementation
- **[Docker & GitHub Action Setup](docs/DOCKER_GITHUB_ACTION_SETUP.md)** - Docker and CI/CD setup

### Architecture & Design
- **[Middleware Agent Design](docs/MIDDLEWARE_AGENT_DESIGN.md)** - Middleware agent design discussion
- **[Implementation Plan](docs/IMPLEMENTATION_PLAN.md)** - Middleware implementation plan

### Project Status
- **[AFE Build Complete](docs/AFE_BUILD_COMPLETE.md)** - Build system completion status
- **[Web Agent Complete](docs/WEB_AGENT_COMPLETE.md)** - Web agent completion status
- **[Status Investigation](docs/SESSION_STATUS_INVESTIGATION.md)** - Status system investigation notes

### Community
- **[Contributing Guide](docs/CONTRIBUTING.md)** - How to contribute to AgentForgeEngine



## 🚀 Quick Start

### Prerequisites

- Go 1.24 or higher
- Git
- Access to offline models (llama.cpp, ollama)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/AgentForgeEngine/AgentForgeEngine.git
   cd AgentForgeEngine
   ```

2. **Initialize user directories**
   ```bash
   go build -o afe ./cmd/agentforge
   ./afe init --migrate
   ```

3. **Build all plugins**
   ```bash
   ./afe build all
   ```

4. **Start the engine**
   ```bash
   ./afe start
   ```

### First-Time Setup

The `afe init` command creates the necessary user directory structure:

```bash
$ ./afe init --verbose
✅ Creating ~/.afe directory structure
✅ Migrating existing plugins
✅ User directories ready
```

### 📚 Need Help?

For detailed documentation and guides, see the [Documentation Index](docs/README.md) which includes:
- Complete setup instructions
- Agent development guides
- Testing framework documentation
- Build system details
- And much more!

## 🏗️ Architecture

### Directory Structure

```
AgentForgeEngine/
├── cmd/                    # CLI commands
├── internal/               # Internal packages
│   ├── cmd/               # Command implementations
│   ├── loader/            # Plugin loading system
│   └── models/            # Model interfaces
├── pkg/                    # Public packages
│   ├── interfaces/        # Core interfaces
│   ├── cache/            # Build cache system
│   ├── hotreload/        # Hot reload manager
│   ├── auth/             # User management
│   ├── status/           # Status tracking system
│   ├── testing/          # Agent testing framework
│   └── userdirs/         # User directory management
├── providers/              # Provider plugins
│   ├── qwen3/
│   └── json-rpc-bridge/
├── agents/                 # Agent plugins
│   ├── ls/                # File listing agent
│   ├── cat/               # File reading agent
│   ├── todo/              # Task management agent
│   ├── pwd/               # Working directory agent
│   ├── whoami/            # User identity agent
│   ├── uname/             # System information agent
│   ├── ps/                # Process listing agent
│   ├── df/                # Disk space agent
│   ├── du/                # Disk usage agent
│   ├── grep/              # Text search agent
│   ├── find/              # File search agent
│   ├── stat/              # File status agent
│   ├── chat/              # Chat interface agent
│   ├── echo/              # Message output agent
│   ├── touch/             # File creation agent
│   ├── mkdir/             # Directory creation agent
│   ├── rm/                # File/directory removal agent
│   ├── cp/                # File/directory copy agent
│   ├── mv/                # File/directory move agent
│   ├── web-agent/         # Web interaction agent
│   ├── file-agent/        # File management agent
│   └── task-agent/        # Task execution agent
├── scripts/                # Utility scripts
│   └── test_agents.sh     # Agent testing runner
├── docs/                   # Documentation
│   └── AGENT_TESTING.md   # Testing framework guide
└── providers/models/       # Shared templates
```

### User Directory Structure

```
~/.afe/
├── accounts/              # Secure user management
│   ├── users/            # LevelDB user database
│   └── api_keys/         # LevelDB API key database
├── providers/             # Built provider plugins
├── agents/                # Built agent plugins
├── cache/                 # Build cache system
├── config/                # User configuration
├── logs/                  # System logs
├── afe.pid               # Process ID file for status tracking
└── afe.sock              # Unix socket for detailed status communication
```

## 🔧 Configuration

Configuration is managed through YAML files with the following priority:

1. **User Config**: `~/.afe/config/build_config.yaml` (highest)
2. **Project Config**: `./agentforge.yaml` (medium)
3. **Default Config**: Built-in defaults (lowest)

### Example Configuration

```yaml
# ~/.afe/config/build_config.yaml
build:
  plugins_dir: "~/.afe/providers"
  agents_dir: "~/.afe/agents"
  cache_dir: "~/.afe/cache"
  go_version_min: "1.24"
  build_flags: ["-ldflags=-s -w"]
  parallel_builds: true
  timeout: 300

cache:
  enabled: true
  max_size_mb: 100
  retention_days: 30
  auto_cleanup: true

logging:
  verbose: false
  max_log_size_mb: 10
```

## 🏗️ Build System

The AgentForgeEngine features an intelligent build system with caching and hot reload.

### Build Commands

```bash
# Build all plugins with intelligent caching
afe build all

# Build specific plugin types
afe build providers
afe build agents

# Build specific plugins
afe build providers --name qwen3
afe build agents --name web-agent

# Force rebuild all plugins
afe build all --force

# Clean and rebuild
afe build all --clean
```

### Build Caching

The build system automatically caches plugins to avoid unnecessary rebuilds:

```bash
$ afe build all --verbose
📦 Discovered 2 providers and 3 agents
📦 Provider qwen3: cached (unchanged)
📦 Provider json-rpc-bridge: cached (unchanged)
🔨 Agent web-agent: REBUILD (source modified)
📊 Build Plan: 1 to rebuild, 4 cached
✅ Build completed: 1 rebuilt, 4 cached in 494ms
```

### Hot Reload

Built-in hot reload automatically updates plugins after successful builds:

```bash
✅ Build completed: 1 rebuilt, 4 cached in 494ms
🔄 Hot reloading updated plugins...
✅ Hot reload completed successfully
🎉 System ready with all plugins
```

### Cache Management

```bash
# View cache statistics
afe cache status

# Clean cache
afe cache clean --force

# Validate cache integrity
afe cache validate
```

## 📊 Status Management

AgentForgeEngine features a hybrid status tracking system using both PID files and Unix sockets for reliable process monitoring.

### Status Commands

```bash
# Check engine status (basic PID file check)
./afe status

# Check detailed status with verbose output
./afe status --verbose

# Start the engine with status tracking
./afe start

# Stop the engine gracefully
./afe stop
```

### Status Output Examples

```bash
# When running with detailed status
$ ./afe status --verbose
AgentForgeEngine Status:
=========================
Status: RUNNING ✓
Process: AgentForgeEngine is active (PID: 12345)

Detailed Information:
  - Version: 1.0.0
  - Uptime: 5m23s
  - Start Time: 2024-01-15 10:30:00
  - Server: localhost:8080
  - Models Loaded: 0
  - Agents Loaded: 0
  - Config: ./configs/agentforge.yaml

# When stopped
$ ./afe status
AgentForgeEngine Status:
=========================
Status: STOPPED ✗
Process: No AgentForgeEngine instance found
```

### Status Tracking Features

- **🔄 Hybrid Monitoring**: PID file for basic detection + Unix socket for detailed status
- **📊 Rich Status Information**: Uptime, version, server details, plugin counts
- **⚡ Real-time Updates**: Live status via Unix socket communication
- **🛡️ Graceful Shutdown**: SIGTERM → SIGKILL fallback with proper cleanup
- **🧹 Automatic Cleanup**: PID and socket files removed on exit

## 🛡️ User Management

AgentForgeEngine includes a secure user management system with LevelDB storage and bcrypt password hashing.

### User Commands

```bash
# Create a new user
afe user create --name "John Doe" --email "john@example.com" --password "secure123"

# Authenticate user
afe user login --email "john@example.com" --password "secure123"

# Create API key
afe user api-key create --name "Production Key" --email "john@example.com"

# List API keys
afe user api-key list --email "john@example.com"
```

### Security Features

- **🔐 bcrypt Password Hashing**: Secure password storage
- **🗄️ LevelDB Storage**: Encrypted database with proper permissions
- **🔑 API Key Management**: Cryptographically secure key generation
- **📊 Audit Trail**: Creation dates, last login, usage tracking
- **🔒 Access Control**: Role-based permissions and scopes
```

## 🧪 Agent Testing

AgentForgeEngine includes a comprehensive testing framework for validating agent functionality and model communication without requiring running models.

### Function Response Format

Agents communicate with models using the standardized format:

```xml
<function_response name="agent_name">{JSON_DATA}</function_response>
```

### Testing Commands

```bash
# Run all integration tests
./scripts/test_agents.sh integration

# Test specific agent
./scripts/test_agents.sh agent ls

# Test all agents with comprehensive validation
./scripts/test_agents.sh all

# Run tests manually
go test -v ./pkg/testing
```

### Testing Features

- **🔍 Model-Independent Testing**: No dependencies on running models
- **📋 Function Response Validation**: XML/JSON format compliance checking
- **🧪 Comprehensive Agent Tests**: Unit, integration, and error handling tests
- **🎭 Mock Model Responses**: Complete simulation of model-agent communication
- **🤖 Interface Compliance**: Ensures proper agent implementation
- **⚡ Automated Test Runner**: Easy testing of all agents

### Test Categories

1. **Unit Tests**: Individual agent functionality and parameter validation
2. **Function Response Tests**: XML format validation and JSON structure checking
3. **Integration Tests**: Model response simulation and end-to-end workflows
4. **Error Handling Tests**: Invalid input handling and edge case validation

### Example Test Output

```bash
✅ Integration tests passed
✅ Function response format test passed for ls
✅ Unit tests passed for ls
✅ Built ls plugin
✅ Model response parsing validated
✅ Round-trip testing completed
```

For detailed testing documentation, see [Agent Testing Guide](docs/AGENT_TESTING.md).

## 📦 Plugin Development

### Creating a Provider

1. **Create provider directory**
   ```bash
   mkdir providers/my-provider
   cd providers/my-provider
   ```

2. **Create go.mod**
   ```go
   module github.com/AgentForgeEngine/AgentForgeEngine/providers/my-provider

   go 1.24

   replace github.com/AgentForgeEngine/AgentForgeEngine => ../..
   ```

3. **Implement provider**
   ```go
   package main

   import "github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"

   type MyProvider struct {
       name string
   }

   func NewMyProvider() *MyProvider {
       return &MyProvider{name: "my-provider"}
   }

   func (p *MyProvider) Name() string {
       return p.name
   }

   func (p *MyProvider) Initialize(config map[string]interface{}) error {
       // Initialize provider
       return nil
   }

   func (p *MyProvider) Generate(ctx context.Context, input interfaces.GenerationRequest) (*interfaces.GenerationResponse, error) {
       // Generate response
       return &interfaces.GenerationResponse{
           Text:     "Hello from my provider!",
           Finished: true,
           Model:    p.name,
       }, nil
   }

   func (p *MyProvider) HealthCheck() error {
       return nil
   }

   func (p *MyProvider) Shutdown() error {
       return nil
   }

   // Export the provider for plugin loading
   var Provider interfaces.Provider = NewMyProvider()
   ```

4. **Build the provider**
   ```bash
   go build -buildmode=plugin -o my-provider.so .
   ```

### Creating an Agent

Follow the same pattern as providers, but implement the `interfaces.Agent` interface instead.

## 📚 API Reference

### Core Interfaces

#### Provider Interface
```go
type Provider interface {
    Name() string
    Initialize(config map[string]interface{}) error
    Generate(ctx context.Context, input GenerationRequest) (*GenerationResponse, error)
    HealthCheck() error
    Shutdown() error
}
```

#### Agent Interface
```go
type Agent interface {
    Name() string
    Initialize(config map[string]interface{}) error
    Process(ctx context.Context, input AgentInput) (AgentOutput, error)
    HealthCheck() error
    Shutdown() error
}
```

### CLI Commands

#### Build Commands
- `afe build all` - Build all plugins
- `afe build providers` - Build provider plugins
- `afe build agents` - Build agent plugins
- `afe cache status` - View cache statistics

#### User Management Commands
- `afe user create` - Create user account
- `afe user login` - Authenticate user
- `afe user api-key create` - Create API key

#### System Commands
- `afe init` - Initialize user directories
- `afe start` - Start the engine with status tracking
- `afe stop` - Stop the engine gracefully
- `afe status` - Check system status (PID file + socket monitoring)

#### Testing Commands
- `./scripts/test_agents.sh integration` - Run integration tests
- `./scripts/test_agents.sh agent <name>` - Test specific agent
- `./scripts/test_agents.sh all` - Test all agents
- `go test -v ./pkg/testing` - Run testing framework

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs/CONTRIBUTING.md) for details.

### Development Setup

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```
3. **Make your changes**
4. **Run tests**
   ```bash
   # Run all tests
   go test ./...
   
   # Run agent integration tests
   ./scripts/test_agents.sh integration
   
   # Test specific agent
   ./scripts/test_agents.sh agent ls
   ```
5. **Commit your changes**
   ```bash
   git commit -m "Add amazing feature"
   ```
6. **Push to the branch**
   ```bash
   git push origin feature/amazing-feature
   ```
7. **Open a Pull Request**

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **llama.cpp** for the excellent offline model server
- **Ollama** for the user-friendly model management
- **Go Community** for the amazing language and tools
- **LevelDB** for the high-performance key-value storage

## 📞 Support

- **Documentation**: [AgentForgeEngine Docs](https://docs.agentforge.engine)
- **Issues**: [GitHub Issues](https://github.com/AgentForgeEngine/AgentForgeEngine/issues)
- **Discussions**: [GitHub Discussions](https://github.com/AgentForgeEngine/AgentForgeEngine/discussions)

---

**Built with ❤️ by the AgentForgeEngine team**