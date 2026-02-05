# AgentForge Engine 🚀

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/AgentForgeEngine/AgentForgeEngine)

A modular, high-performance agent framework that sits between offline models (llama.cpp, ollama) and agents written in Go. Features dynamic loading of agents, hot reload capabilities, and a unified interface for model interactions.

## 🌟 Key Features

- **🏗️ Intelligent Build System**: Smart caching with YAML-based build management
- **🔄 Hot Reload Integration**: Zero-downtime plugin updates
- **🛡️ Secure User Management**: Enterprise-grade authentication with LevelDB
- **📦 Plugin Architecture**: Dynamic loading of providers and agents
- **⚡ High Performance**: Parallel builds and optimized caching
- **🔧 Developer-Friendly**: Comprehensive CLI with clear feedback
- **🌐 Cross-Platform**: Works on Linux, macOS, and Windows

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Installation](#installation)
- [Configuration](#configuration)
- [Build System](#build-system)
- [User Management](#user-management)
- [Plugin Development](#plugin-development)
- [API Reference](#api-reference)
- [Contributing](#contributing)
- [License](#license)

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
│   └── userdirs/         # User directory management
├── providers/              # Provider plugins
│   ├── qwen3/
│   └── json-rpc-bridge/
├── agents/                 # Agent plugins
│   ├── web-agent/
│   ├── file-agent/
│   └── task-agent/
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
└── logs/                  # System logs
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

The AgentForge Engine features an intelligent build system with caching and hot reload.

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

## 🛡️ User Management

AgentForge Engine includes a secure user management system with LevelDB storage and bcrypt password hashing.

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
- `afe start` - Start the engine
- `afe stop` - Stop the engine
- `afe status` - Check system status

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```
3. **Make your changes**
4. **Run tests**
   ```bash
   go test ./...
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

- **Documentation**: [AgentForge Engine Docs](https://docs.agentforge.engine)
- **Issues**: [GitHub Issues](https://github.com/AgentForgeEngine/AgentForgeEngine/issues)
- **Discussions**: [GitHub Discussions](https://github.com/AgentForgeEngine/AgentForgeEngine/discussions)

---

**Built with ❤️ by the AgentForge Engine team**