# Agent Middleware Design Discussion

## 🤔 **What is Agent Middleware?**

Agent middleware would be agents that intercept, process, and potentially modify requests/responses before they reach the target agent or after they return. Think of it like web middleware but for agent operations.

## 💡 **Compelling Middleware Agent Examples**

### **1. 🔐 Security Middleware Agent**
**Use Case**: Prevent malicious file operations and enforce security policies

**Why it's needed**: Your `ls` and `cat` agents currently accept any path - including `../../../etc/passwd` for directory traversal attacks.

**Example**: A `security-middleware` agent that:
- Validates paths against allowlist/denylist rules
- Prevents directory traversal attacks
- Checks file permissions before access
- Blocks access to sensitive system files (`/etc/`, `/proc/`, etc.)
- Logs security violations for audit trails

**Implementation**: Would wrap file system agents and validate `path` parameters before passing through.

### **2. 📊 Audit Logging Middleware Agent**
**Use Case**: Comprehensive audit trail for compliance and security

**Why it's needed**: Enterprises need to track who ran what, when, and what the results were for compliance (SOX, HIPAA, GDPR).

**Example**: An `audit-middleware` agent that:
- Records all agent operations with timestamps
- Logs user identity and request context
- Captures inputs and outputs (with sensitive data redaction)
- Creates tamper-evident audit logs
- Provides forensic search capabilities

**Implementation**: Intercepts all agent calls and logs structured data to secure storage.

### **3. 🎭 Data Redaction Middleware Agent**
**Use Case**: Protect sensitive information in agent outputs

**Why it's needed**: Agent outputs might contain passwords, API keys, or PII that shouldn't be displayed or logged.

**Example**: A `redact-middleware` agent that:
- Detects and redacts passwords, tokens, API keys
- Masks PII (emails, phone numbers, SSNs)
- Filters confidential business data
- Applies custom redaction rules
- Preserves data structure while protecting content

**Implementation**: Post-processes agent outputs and applies pattern-based redaction.

### **4. ⚡ Rate Limiting Middleware Agent**
**Use Case**: Prevent abuse and resource exhaustion

**Why it's needed**: Without rate limiting, users could abuse expensive operations like large file reads or recursive searches.

**Example**: A `rate-limit-middleware` agent that:
- Implements per-user rate limits
- Controls expensive operation frequency
- Prevents DoS attacks on the system
- Provides fair resource allocation
- Returns clear rate limit exceeded errors

**Implementation**: Tracks request frequency and applies throttling rules.

### **5. 💾 Caching Middleware Agent**
**Use Case**: Improve performance for expensive operations

**Why it's needed**: File system operations like `ls` on large directories or `find` operations are expensive and often repeated.

**Example**: A `cache-middleware` agent that:
- Caches file system listings with TTL
- Stores expensive command outputs
- Implements cache invalidation strategies
- Reduces system load significantly
- Provides cache hit/miss metrics

**Implementation**: Key-based caching with intelligent invalidation.

### **6. 🔍 Input Validation Middleware Agent**
**Use Case**: Ensure data quality and provide better error messages

**Why it's needed**: Currently, agents might fail with cryptic errors when given invalid input.

**Example**: A `validate-middleware` agent that:
- Validates file paths exist before operations
- Checks parameter types and ranges
- Ensures required fields are present
- Provides clear, user-friendly error messages
- Prevents downstream agent failures

**Implementation**: Schema-based validation with custom error messages.

### **7. 🛡️ Error Recovery Middleware Agent**
**Use Case**: Graceful handling of failures and system resilience

**Why it's needed**: Network operations or file access might fail intermittently and benefit from retry logic.

**Example**: An `error-recovery-middleware` agent that:
- Implements retry logic with exponential backoff
- Provides fallback responses for failed operations
- Wraps and enriches error messages
- Maintains system stability during partial failures
- Tracks failure patterns for alerting

**Implementation**: Circuit breaker pattern with retry strategies.

### **8. 📈 Metrics Collection Middleware Agent**
**Use Case**: Monitor system performance and usage patterns

**Why it's needed**: You need observability to understand how agents are used and perform.

**Example**: A `metrics-middleware` agent that:
- Tracks request latency and throughput
- Monitors error rates and success rates
- Records resource usage (CPU, memory, file handles)
- Provides per-agent and per-user metrics
- Feeds monitoring and alerting systems

**Implementation**: Telemetry collection with Prometheus/Graphite integration.

## 🏗️ **Architectural Approaches**

### **Option 1: Wrapper Pattern**
```go
type SecurityMiddleware struct {
    targetAgent interfaces.Agent
    rules      SecurityRules
}

func (sm *SecurityMiddleware) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
    // Validate input
    if err := sm.validateInput(input); err != nil {
        return errorResponse(err), nil
    }
    
    // Call target agent
    return sm.targetAgent.Process(ctx, input)
}
```

### **Option 2: Chain Pattern**
```go
type MiddlewareChain struct {
    agents []interfaces.Agent
}

func (mc *MiddlewareChain) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
    // Process through each middleware in sequence
    for _, agent := range mc.agents {
        output, err := agent.Process(ctx, input)
        if err != nil || !output.Success {
            return output, err
        }
        // Modify input for next agent based on output
        input = mc.modifyInput(input, output)
    }
    return finalOutput, nil
}
```

### **Option 3: Decorator Pattern**
```go
// Dynamically add middleware to existing agents
func WithMiddleware(agent interfaces.Agent, middlewares ...MiddlewareFunc) interfaces.Agent {
    // Wrap agent with middleware chain
}
```

## 🎯 **Most Valuable Use Cases**

I think the most compelling starting points would be:

1. **Security Middleware** - Critical for production safety
2. **Audit Logging Middleware** - Essential for enterprise adoption  
3. **Caching Middleware** - High impact on performance
4. **Input Validation Middleware** - Improves user experience significantly

## 🤔 **Questions for Implementation**

1. **Which of these use cases resonates most with your needs?**
2. **Do you prefer the wrapper, chain, or decorator approach?**
3. **Should middleware be configurable per-agent or globally?**
4. **Should middleware agents be discoverable and manageable like regular agents?**

## 📝 **Implementation Notes**

### Current Architecture Context
- AgentForgeEngine uses `interfaces.Agent` interface
- Agents process `AgentInput` and return `AgentOutput`
- Plugin system supports dynamic loading
- Function response format: `<function_response name="agent">{JSON}</function_response>`

### Integration Points
- Could extend `interfaces.Agent` with middleware support
- Plugin manager could handle middleware registration
- Configuration system could support middleware rules
- Testing framework would need middleware testing capabilities

### File Structure Considerations
```
middleware/
├── security/
│   ├── main.go
│   ├── main_test.go
│   └── rules.go
├── audit/
│   ├── main.go
│   ├── logger.go
│   └── storage.go
├── cache/
│   ├── main.go
│   ├── cache.go
│   └── invalidation.go
└── README.md
```

## 🚀 **Next Steps for Implementation**

1. **Choose architectural pattern** (wrapper vs chain vs decorator)
2. **Define middleware interface** extending current agent interface
3. **Implement priority use cases** (security, audit, caching)
4. **Update plugin system** to support middleware registration
5. **Add configuration support** for middleware rules
6. **Extend testing framework** for middleware validation
7. **Update documentation** with middleware usage examples

---

**Stored on**: 2025-02-05  
**Context**: Design discussion for AgentForgeEngine middleware agents  
**Status**: Conceptual design complete, implementation deferred