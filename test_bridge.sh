#!/bin/bash

echo "=== Testing WebSocket Bridge Connection ==="
echo

# Test 1: Check if bridge is reachable
echo "1. Testing bridge reachability on ws://localhost:11435..."
timeout 5 bash -c "</dev/tcp/localhost/11435" && echo "✅ Bridge reachable on port 11435" || echo "❌ Bridge not reachable on port 11435"

echo

# Test 2: Test WebSocket upgrade request
echo "2. Testing WebSocket handshake..."
curl -i -N \
     -H "Connection: Upgrade" \
     -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Key: SGVsbG8gV29ybGQ=" \
     -H "Sec-WebSocket-Version: 13" \
     http://localhost:11435 2>/dev/null | head -5 || echo "❌ WebSocket handshake failed"

echo

# Test 3: Check AFE configuration
echo "3. Current AFE model configuration:"
grep -A 5 "qwen3-coder" configs/afe.yaml || echo "❌ qwen3-coder not found in config"

echo

# Test 4: Try to load and test the WebSocket model directly
echo "4. Testing WebSocket model loading..."
cat > /tmp/test_websocket.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
    "github.com/AgentForgeEngine/AgentForgeEngine/internal/models"
)

func main() {
    config := interfaces.ModelConfig{
        Name:     "qwen3-coder",
        Type:     interfaces.ModelTypeWebSocket,
        Endpoint:  "ws://localhost:11435",
        Options: map[string]interface{}{
            "timeout":     60,
            "max_retries": 0,
        },
    }

    model := models.NewWebSocketModel(config)
    
    fmt.Printf("Created WebSocket model: %s\n", model.Name())
    fmt.Printf("Model type: %s\n", model.Type())
    fmt.Printf("Endpoint: %s\n", config.Endpoint)
    
    fmt.Println("\nInitializing model connection...")
    if err := model.Initialize(config); err != nil {
        log.Printf("❌ Model initialization failed: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("✅ Model initialized successfully!")
    
    fmt.Println("\nPerforming health check...")
    if err := model.HealthCheck(); err != nil {
        log.Printf("❌ Health check failed: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("✅ Health check passed!")
    fmt.Println("\nWebSocket bridge connection successful!")
}
EOF

cd /tmp && go run test_websocket.go
rm -f /tmp/test_websocket.go

echo
echo "=== WebSocket Bridge Test Complete ==="