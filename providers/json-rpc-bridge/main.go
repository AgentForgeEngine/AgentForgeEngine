package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
	"github.com/gorilla/websocket"
)

type JSONRPCBridgeProvider struct {
	name      string
	endpoint  string
	modelName string
	timeout   time.Duration
	client    *websocket.Conn
}

func NewJSONRPCBridgeProvider() *JSONRPCBridgeProvider {
	return &JSONRPCBridgeProvider{
		name:    "json-rpc-bridge",
		timeout: 60 * time.Second,
	}
}

func (p *JSONRPCBridgeProvider) Name() string {
	return p.name
}

func (p *JSONRPCBridgeProvider) Initialize(config map[string]interface{}) error {
	// Parse configuration
	if endpoint, ok := config["endpoint"].(string); ok {
		p.endpoint = endpoint
	} else {
		return fmt.Errorf("endpoint not specified in config")
	}

	if modelName, ok := config["model_name"].(string); ok {
		p.modelName = modelName
	} else {
		return fmt.Errorf("model_name not specified in config")
	}

	// Ensure endpoint has /ws path
	if !strings.HasSuffix(p.endpoint, "/ws") {
		p.endpoint += "/ws"
	}

	log.Printf("JSON-RPC Bridge initialized: endpoint=%s, model=%s", p.endpoint, p.modelName)
	return nil
}

func (p *JSONRPCBridgeProvider) Generate(ctx context.Context, input interfaces.GenerationRequest) (*interfaces.GenerationResponse, error) {
	// Connect to WebSocket
	dialer := websocket.Dialer{}
	c, _, err := dialer.Dial(p.endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("WebSocket dial failed: %w", err)
	}
	defer c.Close()
	p.client = c

	// Set deadlines
	if err := c.SetWriteDeadline(time.Now().Add(p.timeout)); err != nil {
		return nil, fmt.Errorf("failed to set write deadline: %w", err)
	}
	if err := c.SetReadDeadline(time.Now().Add(120 * time.Second)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Create request message
	request := map[string]interface{}{
		"model":  p.modelName,
		"prompt": input.Prompt,
	}

	// Send request
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	err = c.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Read streaming response
	var response strings.Builder
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) {
				break // Normal closure
			}
			return nil, fmt.Errorf("failed to read message: %w", err)
		}

		msgStr := string(message)
		if msgStr == "[DONE]" {
			break
		}
		response.WriteString(msgStr)
	}

	return &interfaces.GenerationResponse{
		Text:     response.String(),
		Finished: true,
		Model:    p.modelName,
	}, nil
}

func (p *JSONRPCBridgeProvider) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to connect and test
	dialer := websocket.Dialer{}
	c, _, err := dialer.DialContext(ctx, p.endpoint, nil)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer c.Close()

	log.Printf("JSON-RPC Bridge health check passed")
	return nil
}

func (p *JSONRPCBridgeProvider) Shutdown() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// Export the provider for plugin loading
var Provider interfaces.Provider = NewJSONRPCBridgeProvider()
