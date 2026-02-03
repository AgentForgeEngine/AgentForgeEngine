package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

func (wa *WebAgent) fetchURL(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract URL from payload
	urlStr, ok := input.Payload["url"].(string)
	if !ok {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "url not specified in payload",
		}, nil
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("invalid URL: %w", err),
		}, nil
	}

	// Check domain restrictions
	if !wa.isAllowedDomain(parsedURL.Hostname()) {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("domain not allowed: %s", parsedURL.Hostname()),
		}, nil
	}

	// Get max tokens for this request
	maxTokens := wa.getMaxTokens(input.Payload)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("request creation failed: %w", err),
		}, nil
	}

	req.Header.Set("User-Agent", wa.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	// Make request
	resp, err := wa.httpClient.Do(req)
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("request failed: %w", err),
		}, nil
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
		}, nil
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !wa.isAllowedContentType(contentType) {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("content type not allowed: %s", contentType),
		}, nil
	}

	// Read content with size limit
	content, err := wa.readContent(resp.Body, 10*1024*1024) // 10MB max
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("content reading failed: %w", err),
		}, nil
	}

	// Extract and process content
	result := wa.extractAndOptimizeContent(content, urlStr, maxTokens)

	return interfaces.AgentOutput{
		Success: true,
		Data:    result,
	}, nil
}

func (wa *WebAgent) validateURL(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	urlStr, ok := input.Payload["url"].(string)
	if !ok {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "url not specified in payload",
		}, nil
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return interfaces.AgentOutput{
			Success: true,
			Data: map[string]interface{}{
				"url":   urlStr,
				"valid": false,
				"error": fmt.Sprintf("invalid URL: %w", err),
			},
		}, nil
	}

	// Check domain restrictions
	domainAllowed := wa.isAllowedDomain(parsedURL.Hostname())

	// Do a HEAD request to check accessibility
	req, err := http.NewRequestWithContext(ctx, "HEAD", urlStr, nil)
	if err != nil {
		return interfaces.AgentOutput{
			Success: true,
			Data: map[string]interface{}{
				"url":            urlStr,
				"valid":          false,
				"domain_allowed": domainAllowed,
				"error":          fmt.Sprintf("request creation failed: %w", err),
			},
		}, nil
	}

	req.Header.Set("User-Agent", wa.userAgent)

	resp, err := wa.httpClient.Do(req)
	if err != nil {
		return interfaces.AgentOutput{
			Success: true,
			Data: map[string]interface{}{
				"url":            urlStr,
				"valid":          false,
				"domain_allowed": domainAllowed,
				"error":          fmt.Sprintf("request failed: %w", err),
			},
		}, nil
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	contentTypeAllowed := wa.isAllowedContentType(contentType)

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"url":            urlStr,
			"valid":          resp.StatusCode < 400,
			"status_code":    resp.StatusCode,
			"domain_allowed": domainAllowed,
			"content_type":   contentType,
			"type_allowed":   contentTypeAllowed,
		},
	}, nil
}

func (wa *WebAgent) extractContent(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// For now, this is the same as fetchURL - can be enhanced later
	return wa.fetchURL(ctx, input)
}

func (wa *WebAgent) readContent(body io.ReadCloser, maxSize int64) (string, error) {
	limiter := io.LimitReader(body, maxSize)
	content, err := io.ReadAll(limiter)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (wa *WebAgent) isAllowedDomain(hostname string) bool {
	if hostname == "" {
		return false
	}

	hostname = strings.ToLower(hostname)

	// Check blocked domains first
	for _, blocked := range wa.blockedDomains {
		if strings.HasSuffix(hostname, strings.TrimPrefix(blocked, ".")) {
			return false
		}
	}

	// If allowed domains is not set, allow all (except blocked)
	if len(wa.allowedDomains) == 0 {
		return true
	}

	// Check allowed domains
	for _, allowed := range wa.allowedDomains {
		if allowed == "*" || strings.HasSuffix(hostname, strings.TrimPrefix(allowed, ".")) {
			return true
		}
	}

	return false
}

func (wa *WebAgent) isAllowedContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)

	for _, allowed := range wa.allowedContentTypes {
		if strings.Contains(contentType, allowed) {
			return true
		}
	}

	return false
}

func (wa *WebAgent) getMaxTokens(payload map[string]interface{}) int {
	if maxTokens, ok := payload["max_tokens"].(int); ok {
		// Enforce bounds
		if maxTokens < wa.minAllowedTokens {
			return wa.minAllowedTokens
		}
		if maxTokens > wa.maxAllowedTokens {
			return wa.maxAllowedTokens
		}
		return maxTokens
	}
	return wa.defaultMaxTokens
}
