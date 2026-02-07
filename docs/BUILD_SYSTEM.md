# AgentForgeEngine Build System 🏗️

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org/)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/AgentForgeEngine/AgentForgeEngine)

A comprehensive, intelligent build system for AgentForgeEngine plugins with YAML-based caching, parallel compilation, and hot reload integration.

## 🌟 Features

- **🧠 Intelligent Caching**: YAML-based build cache with smart invalidation
- **⚡ Parallel Builds**: Multi-threaded compilation for maximum performance
- **🔄 Hot Reload Integration**: Zero-downtime plugin updates
- **📊 Build Analytics**: Detailed statistics and performance metrics
- **🛠️ Developer-Friendly**: Comprehensive CLI with clear feedback
- **🗂️ Cross-Platform**: Works on Linux, macOS, and Windows

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Build Commands](#build-commands)
- [Cache System](#cache-system)
- [Configuration](#configuration)
- [Hot Reload](#hot-reload)
- [Build Analytics](#build-analytics)
- [Troubleshooting](#troubleshooting)
- [Build Cache API](#build-cache-api)

## 🚀 Quick Start

### First-Time Setup

```bash
# Initialize user directories
./afe init --migrate

# Build all plugins
./afe build all

# Start the engine
./afe start
```

### Basic Usage

```bash
# Build all plugins (intelligent caching)
afe build all

# Build specific plugin types
afe build providers
afe build agents

# Build specific plugins
afe build providers --name qwen3
afe build agents --name web-agent
```

## 🔧 Build Commands

### Core Build Commands

#### `afe build all`
Build all plugins with intelligent caching and hot reload:

```bash
afe build all [--verbose] [--parallel] [--force] [--clean]
```

**Options:**
- `--verbose, -v`: Detailed build output
- `--parallel, -p`: Build plugins concurrently (default: true)
- `--force`: Force rebuild of all plugins
- `--clean`: Clean cache before building

#### `afe build providers`
Build only provider plugins:

```bash
afe build providers [--name qwen3] [--verbose] [--parallel] [--force]
```

#### `afe build agents`
Build only agent plugins:

```bash
afe build agents [--name web-agent] [--verbose] [--parallel] [--force]
```

### Cache Management Commands

#### `afe cache status`
Display build cache statistics and information:

```bash
afe cache status
```

**Output:**
```
CACHE STATISTICS
================
Version:  1.0
Created:  2024-02-05 14:52:57
Updated:  2024-02-05 15:05:40
Valid:    true

BUILD STATISTICS
================
Total Builds:    5
Cache Hits:      2
Cache Misses:    24
Hit Rate:        7.7%
Avg Build Time:  1321 ms

CACHED PLUGINS
==============
Providers:     2
Agents:        3
Total Cached:  5
Cache Size:    0.0 MB
```

#### `afe cache clean`
Clean the build cache:

```bash
afe cache clean [--force]
```

#### `afe cache validate`
Validate cache integrity:

```bash
afe cache validate
```

## 🧠 Cache System

### Cache Architecture

The build system uses a sophisticated caching mechanism based on:

- **Source File Changes**: Monitors `.go` files for modifications
- **Go Module Changes**: Tracks `go.mod` and `go.sum` updates
- **Build Configuration**: Detects build flag changes
- **Output Validation**: Ensures built plugins exist and are accessible

### Cache Storage

```
~/.afe/cache/
├── build_cache.yaml          # Main cache database
├── plugin_hashes/
│   ├── providers/
│   │   ├── qwen3.yaml
│   │   └── json-rpc-bridge.yaml
│   └── agents/
│       ├── web-agent.yaml
│       ├── file-agent.yaml
│       └── task-agent.yaml
└── build_metadata/
    ├── last_build_all.yaml
    └── build_stats.yaml
```

### Cache Decision Logic

A plugin will be rebuilt if ANY of these conditions are met:

1. **Source Changes**: Any `.go` file modified since last build
2. **Module Changes**: `go.mod` or `go.sum` modified
3. **Config Changes**: Build flags or Go version changed
4. **Output Missing**: Built plugin file no longer exists
5. **Force Flag**: `--force` flag specified
6. **Cache Corruption**: Cache data invalid or inconsistent

### Cache Performance

```bash
$ afe build all --verbose
📦 Discovered 2 providers and 3 agents
📦 Provider qwen3: cached (unchanged)
📦 Provider json-rpc-bridge: cached (unchanged)
🔨 Agent web-agent: REBUILD (source modified)
📊 Build Plan: 1 to rebuild, 4 cached
✅ Build completed: 1 rebuilt, 4 cached in 494ms
```

## ⚙️ Configuration

### Build Configuration

User-specific build configuration in `~/.afe/config/build_config.yaml`:

```yaml
# AgentForgeEngine User Configuration
build:
  plugins_dir: "~/.afe/providers"
  agents_dir: "~/.afe/agents"
  cache_dir: "~/.afe/cache"
  go_version_min: "1.24"
  build_flags: ["-ldflags=-s -w"]
  parallel_builds: true
  clean_before_build: false
  timeout: 300
  max_workers: 0  # 0 = auto-detect CPU cores

cache:
  enabled: true
  max_size_mb: 100
  retention_days: 30
  auto_cleanup: true
  compression_enabled: false
  validation_interval: "24h"

logging:
  build_log: "~/.afe/logs/build.log"
  cache_log: "~/.afe/logs/cache.log"
  verbose: false
  max_log_size_mb: 10
```

### Configuration Priority

1. **User Config**: `~/.afe/config/build_config.yaml` (highest priority)
2. **Project Config**: `./agentforge.yaml` (medium priority)
3. **Default Config**: Built-in defaults (lowest priority)

## 🔄 Hot Reload Integration

### Hot Reload Process

The build system automatically triggers hot reload after successful builds:

```bash
✅ Build completed: 3 rebuilt, 2 cached in 494ms
🔄 Hot reloading updated plugins...
2026/02/05 15:17:53 manager.go:75: 🔄 Hot reload manager started
2026/02/05 15:17:53 manager.go:137: 🔄 Hot reload worker 1 started
2026/02/05 15:17:53 manager.go:118: 🔄 Queued hot reload for provider json-rpc-bridge
2026/02/05 15:17:53 manager.go:160: 🔄 Processing hot reload for provider json-rpc-bridge
2026/02/05 15:17:53 manager.go:118: 🔄 Queued hot reload for provider qwen3
2026/02/05 15:17:53 manager.go:160: 🔄 Processing hot reload for provider qwen3
✅ Hot reload completed successfully
🎉 System ready with all plugins
```

### Hot Reload Features

- **🔄 Automatic Detection**: Triggers after successful builds
- **⚡ Parallel Processing**: Multiple workers for concurrent reloads
- **🛡️ Error Handling**: Graceful failure recovery
- **📊 Status Reporting**: Detailed reload status and metrics
- **🔧 Plugin Validation**: Ensures plugins are properly loaded

## 📊 Build Analytics

### Performance Metrics

The build system tracks comprehensive performance metrics:

```bash
$ afe cache status
CACHE STATISTICS
================
Total Builds:    47
Cache Hits:      32
Cache Misses:    15
Cache Hit Rate:  68.1%
Total Build Time: 125430ms
Average Build Time: 2668ms
Total Cache Size: 23.4MB
```

### Build History

Each build operation is logged with detailed information:

```yaml
build_history:
  - build_id: "build_20240205_154522"
    timestamp: "2024-02-05T15:45:22Z"
    command: "afe build all"
    plugins_built: ["json-rpc-bridge"]
    plugins_cached: ["qwen3", "web-agent"]
    total_duration_ms: 4500
    success: true
    cache_hit_rate: 80.0
```

## 🐛 Troubleshooting

### Common Issues

#### Build Failures

```bash
$ afe build providers --name qwen3
❌ Build failed: build failed: exit status 1, output: main module does not contain package
```

**Solution**: Check that the plugin directory has a proper `go.mod` file and implements the required interface.

#### Cache Issues

```bash
$ afe build all
❌ Cache validation failed: cache entry invalid
```

**Solution**: Clean the cache and rebuild:
```bash
afe cache clean --force
afe build all
```

#### Hot Reload Issues

```bash
⚠️ Hot reload failed: provider qwen3 not found
```

**Solution**: This is normal if the plugin isn't currently loaded. The hot reload will still work for newly built plugins.

### Debug Mode

Enable verbose output for detailed debugging:

```bash
afe build all --verbose
```

This provides detailed information about:
- Plugin discovery process
- Cache validation results
- Build command execution
- Hot reload operations

### Performance Optimization

#### Parallel Build Tuning

Adjust the number of parallel workers:

```yaml
# ~/.afe/config/build_config.yaml
build:
  max_workers: 8  # Limit to 8 workers for optimal performance
```

#### Cache Size Management

Configure cache size limits:

```yaml
cache:
  max_size_mb: 200  # Increase cache size for better hit rates
  retention_days: 60  # Keep cache longer
```

## 📚 API Reference

### Build Cache Schema

The main cache file (`~/.afe/cache/build_cache.yaml`) contains:

```yaml
version: "1.0"
created_at: "2024-02-05T14:30:00Z"
updated_at: "2024-02-05T15:45:22Z"
afe_version: "v0.1.0"
go_version: "1.24.6"

statistics:
  total_builds: 47
  cache_hits: 32
  cache_misses: 15
  cache_hit_rate: 68.1
  total_build_time_ms: 125430
  average_build_time_ms: 2668
  total_cache_size_mb: 23.4

plugins:
  providers:
    qwen3:
      build_info:
        source_hash: "sha256:abc123..."
        output_path: "~/.afe/providers/qwen3.so"
        build_time: "2024-02-05T14:30:00Z"
        build_duration_ms: 2340
        plugin_size_bytes: 1234567
        needs_rebuild: false
        cache_valid: true
  agents:
    web-agent:
      build_info:
        source_hash: "sha256:def456..."
        output_path: "~/.afe/agents/web-agent.so"
        build_time: "2024-02-05T14:35:00Z"
        build_duration_ms: 3120
        plugin_size_bytes: 2345678
        needs_rebuild: false
        cache_valid: true
```

### CLI Command Reference

#### Build Commands

| Command | Description | Options |
|---------|-------------|---------|
| `afe build all` | Build all plugins | `--verbose`, `--parallel`, `--force`, `--clean` |
| `afe build providers` | Build provider plugins | `--name`, `--verbose`, `--parallel`, `--force` |
| `afe build agents` | Build agent plugins | `--name`, `--verbose`, `--parallel`, `--force` |

#### Cache Commands

| Command | Description | Options |
|---------|-------------|---------|
| `afe cache status` | Show cache statistics | None |
| `afe cache clean` | Clean build cache | `--force` |
| `afe cache validate` | Validate cache integrity | None |

---

**Built with ❤️ by the AgentForgeEngine team**