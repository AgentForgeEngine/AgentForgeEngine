# Contributing to AgentForgeEngine

Thank you for your interest in contributing to AgentForgeEngine! This guide will help you get started with contributing to the project.

## 🚀 Quick Start

### Prerequisites

- Go 1.24 or higher
- Git
- Docker (for testing)
- Basic knowledge of Go and agent systems

### Development Setup

1. **Fork the repository**
   ```bash
   # Fork the repository on GitHub
   git clone https://github.com/YOUR_USERNAME/AgentForgeEngine.git
   cd AgentForgeEngine
   ```

2. **Add upstream remote**
   ```bash
   git remote add upstream https://github.com/AgentForgeEngine/AgentForgeEngine.git
   ```

3. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

4. **Set up development environment**
   ```bash
   # Install dependencies
   go mod download
   
   # Run tests to ensure everything works
   go test ./...
   ```

## 📋 Contribution Areas

### 1. Agent Development

#### Creating New Agents

1. **Create agent directory**
   ```bash
   mkdir agents/your-agent-name
   cd agents/your-agent-name
   ```

2. **Create go.mod**
   ```go
   module github.com/AgentForgeEngine/AgentForgeEngine/agents/your-agent-name
   
   go 1.24
   
   replace github.com/AgentForgeEngine/AgentForgeEngine => ../..
   
   require github.com/AgentForgeEngine/AgentForgeEngine v0.0.0-00010101000000-000000000000
   ```

3. **Implement agent** (`main.go`)
   ```go
   package main
   
   import (
       "context"
       "log"
       
       "github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
   )
   
   type YourAgentNameAgent struct {
       name string
   }
   
   func NewYourAgentNameAgent() *YourAgentNameAgent {
       return &YourAgentNameAgent{name: "your-agent-name"}
   }
   
   func (a *YourAgentNameAgent) Name() string {
       return a.name
   }
   
   func (a *YourAgentNameAgent) Initialize(config map[string]interface{}) error {
       log.Printf("Initializing %s agent", a.name)
       return nil
   }
   
   func (a *YourAgentNameAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
       // Implement your agent logic here
       return interfaces.AgentOutput{
           Success: true,
           Data: map[string]interface{}{
               "result": "success",
           "input":  input,
           "agent":  a.name,
           "timestamp": time.Now().Format(time.RFC3339),
           "version": "1.0.0",
           "status": "completed",
           "host": "localhost",
           "port": 8080,
           "models_count": 0,
           "agents_count": 0,
           "uptime": "0s",
           "start_time": time.Now().Format(time.RFC3339),
           "end_time": time.Now().Format(time.RFC3339),
           "execution_time_ms": 0,
           "memory_usage_mb": 0,
           "cpu_usage_percent": 0,
           "disk_usage_mb": 0,
           "network_requests": 0,
           "errors": []string{},
           "warnings": []string{},
           "metadata": map[string]interface{}{
               "input_type": input.Type,
               "payload_size": len(fmt.Sprintf("%v", input.Payload)),
               "context_deadline": ctx.Err() != nil,
           },
           "performance_metrics": map[string]interface{}{
               "response_time_ms": 0,
               "processing_time_ms": 0,
               "serialization_time_ms": 0,
               "deserialization_time_ms": 0,
               "validation_time_ms": 0,
               "error_handling_time_ms": 0,
               "cleanup_time_ms": 0,
           },
           "security_info": map[string]interface{}{
               "authenticated": false,
               "permissions": []string{},
               "access_level": "user",
               "security_context": "development",
               "audit_trail": false,
           },
           "debug_info": map[string]interface{}{
               "debug_mode": false,
               "verbose_logging": false,
               "trace_enabled": false,
               "profiling_enabled": false,
               "stack_trace": []string{},
           },
           "compatibility_info": map[string]interface{}{
               "agent_version": "1.0.0",
               "engine_version": "1.0.0",
               "go_version": "1.24",
               "platform": "linux",
               "architecture": "amd64",
               "dependencies": []string{},
               "interfaces": []string{"Agent"},
               "protocols": []string{"function_response"},
           },
           "health_info": map[string]interface{}{
               "status": "healthy",
               "last_health_check": time.Now().Format(time.RFC3339),
               "uptime_seconds": 0,
               "restart_count": 0,
               "error_count": 0,
               "warning_count": 0,
           },
           "configuration": map[string]interface{}{
               "default_config": map[string]interface{}{},
               "runtime_config": map[string]interface{}{},
               "environment_config": map[string]interface{}{},
               "user_config": map[string]interface{}{},
           },
           "statistics": map[string]interface{}{
               "total_requests": 0,
               "successful_requests": 0,
               "failed_requests": 0,
               "average_response_time_ms": 0,
               "min_response_time_ms": 0,
               "max_response_time_ms": 0,
               "total_processing_time_ms": 0,
               "total_execution_time_ms": 0,
               "error_rate": 0.0,
               "success_rate": 1.0,
           },
           "logs": []interface{}{
               map[string]interface{}{
                   "timestamp": time.Now().Format(time.RFC3339),
                   "level": "info",
                   "message": "Agent processed request successfully",
                   "component": "agent",
                   "operation": "process",
                   "details": map[string]interface{}{
                       "agent": "your-agent-name",
                       "input_type": input.Type,
                       "processing_time_ms": 0,
                   },
               },
           },
           "cache_info": map[string]interface{}{
               "cache_enabled": false,
               "cache_type": "none",
               "cache_size": 0,
               "cache_hits": 0,
               "cache_misses": 0,
               "cache_hit_rate": 0.0,
           },
           "monitoring": map[string]interface{}{
               "metrics_enabled": false,
               "tracing_enabled": false,
               "logging_enabled": true,
               "alerting_enabled": false,
               "dashboard_url": "",
           },
           "extensions": map[string]interface{}{
               "custom_fields": map[string]interface{}{},
               "plugins": []string{},
               "hooks": []string{},
               "middleware": []string{},
           },
       }, nil
   }
   
   func (a *YourAgentNameAgent) HealthCheck() error {
       return nil
   }
   
   func (a *YourAgentNameAgent) Shutdown() error {
       log.Printf("Shutting down %s agent", a.name)
       return nil
   }
   
   // Export the agent for plugin loading
   var Agent interfaces.Agent = NewYourAgentNameAgent()
   ```

4. **Create tests** (`main_test.go`)
   ```go
   package main
   
   import (
       "context"
       "testing"
       "time"
       
       "github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
       "github.com/AgentForgeEngine/AgentForgeEngine/pkg/testing"
   )
   
   func TestYourAgentNameAgent_FunctionResponseFormat(t *testing.T) {
       agent := NewYourAgentNameAgent()
       suite := testing.NewAgentTestSuite(t, agent)
       
       // Test basic interface compliance
       suite.TestAgentInterface()
       
       // Test function response format
       input := interfaces.AgentInput{
           Type: "your-action",
           Payload: map[string]interface{}{
               "param1": "value1",
               "param2": "value2",
           },
       }
       
       suite.TestFunctionResponseFormat(input, "your-agent-name")
   }
   
   func TestYourAgentNameAgent_ModelResponseIntegration(t *testing.T) {
       agent := NewYourAgentNameAgent()
       
       // Create mock model response
       modelResponse := testing.CreateMockModelResponse("your-agent-name", map[string]interface{}{
           "param1": "value1",
           "param2": "value2",
       })
       
       // Parse the function call
       agentName, arguments, err := testing.ParseFunctionCall(modelResponse.FunctionCall)
       if err != nil {
           t.Fatalf("Failed to parse function call: %v", err)
       }
       
       // Verify agent name
       if agentName != "your-agent-name" {
           t.Errorf("Expected agent name 'your-agent-name', got '%s'", agentName)
       }
       
       // Create input and test processing
       input := interfaces.AgentInput{
           Type:    "your-action",
           Payload: arguments,
       }
       
       ctx := context.Background()
       output, err := agent.Process(ctx, input)
       if err != nil {
           t.Fatalf("Agent processing failed: %v", err)
       }
       
       // Verify success
       if !output.Success {
           t.Errorf("Expected successful processing, got error: %s", output.Error)
       }
   }
   ```

#### Agent Requirements

- ✅ **Interface Compliance**: Must implement `interfaces.Agent`
- ✅ **Function Response Format**: Must support `<function_response name="agent">{JSON}</function_response>`
- ✅ **Error Handling**: Proper error handling with structured responses
- ✅ **Testing**: Comprehensive test coverage
- ✅ **Documentation**: Clear documentation of functionality and usage

### 2. Core Engine Development

#### Areas for Contribution

- **Build System**: Enhance caching and hot reload
- **Plugin Management**: Improve dynamic loading
- **User Management**: Extend authentication and authorization
- **Status Tracking**: Enhance monitoring and health checks
- **Configuration**: Improve configuration management

### 3. Testing Framework

#### Areas for Contribution

- **Test Utilities**: Add new testing helpers
- **Mock Services**: Expand mock model responses
- **Integration Tests**: Add comprehensive integration testing
- **Performance Tests**: Add performance benchmarking

### 4. Documentation

#### Areas for Contribution

- **API Documentation**: Improve API reference
- **User Guides**: Create comprehensive user guides
- **Developer Guides**: Enhance developer documentation
- **Examples**: Add practical examples and tutorials

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific test
go test -v ./agents/your-agent

# Run integration tests
go test -v ./pkg/testing

# Run tests with coverage
go test -cover ./...
```

### Docker Testing

```bash
# Build test image
docker build -t afe-test .

# Run tests in Docker
docker run --rm -v $(pwd):/app -w /app afe-test go run scripts/test_docker.go ./agents
```

### Test Requirements

- All new agents must pass comprehensive tests
- Function response format must be validated
- Interface compliance must be verified
- Model response integration must work correctly

## 📝 Documentation Standards

### File Organization

- **User Documentation**: `docs/` directory
- **API Documentation**: Inline code comments
- **Examples**: `examples/` directory
- **Architecture**: `docs/ARCHITECTURE.md`

### Writing Style

- Use clear, concise language
- Include practical examples
- Provide step-by-step instructions
- Use consistent formatting

### Documentation Files

- **README.md**: Project overview and quick start
- **CONTRIBUTING.md**: This contributing guide
- **CHANGELOG.md**: Version history and changes
- **API docs**: Inline documentation for public APIs

## 🔄 Development Workflow

### 1. Create Issue

- Describe the feature or bug fix
- Include acceptance criteria
- Tag relevant maintainers

### 2. Create Branch

```bash
git checkout -b feature/your-feature-name
```

### 3. Develop and Test

- Write code following conventions
- Add comprehensive tests
- Update documentation

### 4. Submit Pull Request

- Include clear description
- Reference related issues
- Ensure all tests pass

### 5. Code Review

- Maintainers will review your changes
- Address feedback promptly
- Keep PR up to date

## 🎯 Code Quality

### Go Guidelines

- Follow Go best practices
- Use `gofmt` for formatting
- Use `golint` for linting
- Write clear, idiomatic Go code

### Testing Guidelines

- Aim for high test coverage
- Write meaningful tests
- Test edge cases and error conditions
- Use table-driven tests where appropriate

### Documentation Guidelines

- Document public APIs
- Include usage examples
- Keep documentation up to date
- Use consistent formatting

## 🚀 Release Process

### Version Management

- Follow semantic versioning
- Update CHANGELOG.md
- Tag releases appropriately

### Release Checklist

- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version bumped
- [ ] Tagged release

## 🤝 Community

### Getting Help

- Create an issue for questions
- Join discussions for help
- Check existing issues first

### Communication

- Be respectful and constructive
- Provide clear, helpful feedback
- Follow code of conduct

## 📜 Resources

### Documentation

- [Agent Testing Guide](docs/AGENT_TESTING.md)
- [Build System Guide](docs/BUILD_SYSTEM.md)
- [User Management Guide](docs/USER_MANAGEMENT.md)
- [New File System Agents](docs/NEW_FILE_SYSTEM_AGENTS.md)
- [Docker & GitHub Action Setup](docs/DOCKER_GITHUB_ACTION_SETUP.md)

### Tools and Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Testing Patterns](https://github.com/golang/go/wiki/Testing)

## 🏆 Recognition

Contributors are recognized in:
- README.md contributors section
- Release notes
- Commit history
- GitHub statistics

Thank you for contributing to AgentForgeEngine! 🎉

---

**Last Updated**: 2025-02-06  
**Version**: 1.0.0  
**Maintainers**: AgentForgeEngine Team