package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type WebAgent struct {
	name                string
	httpClient          *http.Client
	defaultMaxTokens    int
	maxAllowedTokens    int
	minAllowedTokens    int
	timeout             time.Duration
	userAgent           string
	allowedDomains      []string
	blockedDomains      []string
	allowedContentTypes []string
	includeLinks        bool
	includeMetadata     bool
}

func NewWebAgent() *WebAgent {
	return &WebAgent{
		name:             "web-agent",
		defaultMaxTokens: 8000,
		maxAllowedTokens: 15000,
		minAllowedTokens: 500,
		timeout:          15 * time.Second,
		userAgent:        "AgentForgeEngine-WebAgent/1.0",
		allowedContentTypes: []string{
			"text/html",
			"text/plain",
			"application/json",
			"application/xml",
			"text/xml",
		},
		includeLinks:    true,
		includeMetadata: true,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (wa *WebAgent) Name() string {
	return wa.name
}

func (wa *WebAgent) Initialize(config map[string]interface{}) error {
	// Set default max tokens
	if maxTokens, ok := config["default_max_tokens"].(int); ok {
		wa.defaultMaxTokens = maxTokens
	}

	if maxAllowed, ok := config["max_allowed_tokens"].(int); ok {
		wa.maxAllowedTokens = maxAllowed
	}

	if minAllowed, ok := config["min_allowed_tokens"].(int); ok {
		wa.minAllowedTokens = minAllowed
	}

	// Set timeout
	if timeout, ok := config["timeout"].(int); ok {
		wa.timeout = time.Duration(timeout) * time.Second
		wa.httpClient.Timeout = wa.timeout
	}

	// Set user agent
	if userAgent, ok := config["user_agent"].(string); ok {
		wa.userAgent = userAgent
	}

	// Set domains
	if allowed, ok := config["allowed_domains"].([]interface{}); ok {
		var domains []string
		for _, domain := range allowed {
			if domainStr, ok := domain.(string); ok {
				domains = append(domains, strings.ToLower(domainStr))
			}
		}
		wa.allowedDomains = domains
	}

	if blocked, ok := config["blocked_domains"].([]interface{}); ok {
		var domains []string
		for _, domain := range blocked {
			if domainStr, ok := domain.(string); ok {
				domains = append(domains, strings.ToLower(domainStr))
			}
		}
		wa.blockedDomains = domains
	}

	// Set content types
	if contentTypes, ok := config["content_types"].([]interface{}); ok {
		var types []string
		for _, ct := range contentTypes {
			if ctStr, ok := ct.(string); ok {
				types = append(types, strings.ToLower(ctStr))
			}
		}
		wa.allowedContentTypes = types
	}

	// Set feature flags
	if includeLinks, ok := config["include_links"].(bool); ok {
		wa.includeLinks = includeLinks
	}

	if includeMetadata, ok := config["include_metadata"].(bool); ok {
		wa.includeMetadata = includeMetadata
	}

	// Test with a simple request to verify connectivity
	if err := wa.HealthCheck(); err != nil {
		return fmt.Errorf("web-agent initialization failed: %w", err)
	}

	log.Printf("WebAgent initialized: max_tokens=%d, timeout=%v", wa.defaultMaxTokens, wa.timeout)
	return nil
}

func (wa *WebAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	switch input.Type {
	case "fetch":
		return wa.fetchURL(ctx, input)
	case "validate":
		return wa.validateURL(ctx, input)
	case "extract":
		return wa.extractContent(ctx, input)
	default:
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("unknown operation: %s", input.Type),
		}, nil
	}
}

func (wa *WebAgent) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a simple test request to httpbin.org (reliable testing endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/get", nil)
	if err != nil {
		return fmt.Errorf("health check request creation failed: %w", err)
	}

	req.Header.Set("User-Agent", wa.userAgent)

	resp, err := wa.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (wa *WebAgent) Shutdown() error {
	// No cleanup needed for HTTP client
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewWebAgent()
