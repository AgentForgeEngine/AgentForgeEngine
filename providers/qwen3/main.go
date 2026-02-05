package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/templates"
)

type Qwen3Provider struct {
	name          string
	endpoint      string
	templatePath  string
	timeout       time.Duration
	client        *http.Client
	templateCache *templates.TemplateCache
}

type Message struct {
	Role         string        `json:"role"`
	Content      string        `json:"content"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type TemplateData struct {
	Messages []Message `json:"messages"`
}

func NewQwen3Provider() *Qwen3Provider {
	return &Qwen3Provider{
		name:          "qwen3",
		timeout:       120 * time.Second,
		templateCache: templates.NewTemplateCache(),
	}
}

func (p *Qwen3Provider) Name() string {
	return p.name
}

func (p *Qwen3Provider) Initialize(config map[string]interface{}) error {
	// Parse configuration
	if endpoint, ok := config["endpoint"].(string); ok {
		p.endpoint = endpoint
	} else {
		// Default to localhost:8080 for qwen3
		p.endpoint = "http://localhost:8080"
	}

	if templatePath, ok := config["template_path"].(string); ok {
		p.templatePath = templatePath
	} else {
		// Default to qwen3.j2
		p.templatePath = "qwen3"
	}

	// Setup HTTP client
	p.client = &http.Client{
		Timeout: p.timeout,
	}

	log.Printf("Qwen3 provider initialized: endpoint=%s, template=%s", p.endpoint, p.templatePath)
	return nil
}

func (p *Qwen3Provider) Generate(ctx context.Context, input interfaces.GenerationRequest) (*interfaces.GenerationResponse, error) {
	// Parse messages from prompt
	messages, err := p.parseMessages(input.Prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse messages: %w", err)
	}

	// Apply template
	renderedPrompt, err := p.applyTemplate(messages)
	if err != nil {
		return nil, fmt.Errorf("failed to apply template: %w", err)
	}

	// Create llama.cpp request payload
	payload := map[string]interface{}{
		"prompt":      renderedPrompt,
		"n_predict":   input.MaxTokens,
		"temperature": input.Temperature,
		"stop":        []string{"<|im_end|>"},
		"stream":      input.Stream,
	}

	// Add JSON system message header if needed
	if p.hasJSONSystemMessage(messages) {
		payload["system"] = p.getJSONSystemMessage(messages)
	}

	// Serialize payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/completion", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Handle streaming response
	if input.Stream {
		return p.handleStreamingResponse(resp)
	}

	// Handle non-streaming response
	return p.handleNonStreamingResponse(resp)
}

func (p *Qwen3Provider) parseMessages(prompt string) ([]Message, error) {
	// Try to parse as JSON first
	var messages []Message
	if err := json.Unmarshal([]byte(prompt), &messages); err == nil {
		return messages, nil
	}

	// If not JSON, create a simple user message
	return []Message{
		{Role: "user", Content: prompt},
	}, nil
}

func (p *Qwen3Provider) applyTemplate(messages []Message) (string, error) {
	// Find template file
	templateFile, err := templates.FindTemplate(p.templatePath)
	if err != nil {
		return "", fmt.Errorf("template not found: %w", err)
	}

	// Read template content
	content, err := os.ReadFile(templateFile)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	// Create qwen3 template processor
	tmpl := templates.NewQwen3Template(string(content))

	// Convert messages to map format
	msgMaps := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		msgMap := map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}
		if msg.FunctionCall != nil {
			msgMap["function_call"] = map[string]interface{}{
				"name":      msg.FunctionCall.Name,
				"arguments": msg.FunctionCall.Arguments,
			}
		}
		msgMaps[i] = msgMap
	}

	// Render template
	return tmpl.Render(msgMaps)
}

func (p *Qwen3Provider) hasJSONSystemMessage(messages []Message) bool {
	for _, msg := range messages {
		if msg.Role == "system" && strings.HasPrefix(msg.Content, "{") {
			return true
		}
	}
	return false
}

func (p *Qwen3Provider) getJSONSystemMessage(messages []Message) string {
	for _, msg := range messages {
		if msg.Role == "system" && strings.HasPrefix(msg.Content, "{") {
			return msg.Content
		}
	}
	return ""
}

func (p *Qwen3Provider) handleStreamingResponse(resp *http.Response) (*interfaces.GenerationResponse, error) {
	var response strings.Builder
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var chunk struct {
				Content string `json:"content"`
				Stopped bool   `json:"stopped"`
			}

			if err := json.Unmarshal([]byte(data), &chunk); err == nil {
				response.WriteString(chunk.Content)
				if chunk.Stopped {
					break
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read streaming response: %w", err)
	}

	return &interfaces.GenerationResponse{
		Text:     response.String(),
		Finished: true,
		Model:    p.name,
	}, nil
}

func (p *Qwen3Provider) handleNonStreamingResponse(resp *http.Response) (*interfaces.GenerationResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response struct {
		Content string `json:"content"`
		Stopped bool   `json:"stopped"`
		Tokens  int    `json:"tokens_predicted"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &interfaces.GenerationResponse{
		Text:     response.Content,
		Tokens:   response.Tokens,
		Finished: response.Stopped,
		Model:    p.name,
	}, nil
}

func (p *Qwen3Provider) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", p.endpoint+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
}

func (p *Qwen3Provider) Shutdown() error {
	// No cleanup needed for HTTP client
	return nil
}

// Export the provider for plugin loading
var Provider interfaces.Provider = NewQwen3Provider()
