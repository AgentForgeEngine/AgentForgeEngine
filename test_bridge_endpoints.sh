#!/bin/bash

echo "=== Testing Bridge API Endpoints ==="
echo

# Test common endpoints
endpoints=(
    "/api/generate"
    "/generate" 
    "/api/chat"
    "/chat"
    "/rpc"
    "/"
)

for endpoint in "${endpoints[@]}"; do
    echo "Testing endpoint: $endpoint"
    
    # Try POST request with simple JSON payload
    response=$(curl -s -w "\n%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d '{"prompt": "test", "max_tokens": 10}' \
        http://localhost:11435"$endpoint" 2>/dev/null)
    
    # Extract status code (last line)
    status_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "200" ]; then
        echo "✅ $endpoint - SUCCESS"
        echo "📝 Response: $(echo "$response_body" | head -c 100)..."
    elif [ "$status_code" = "404" ]; then
        echo "❌ $endpoint - NOT FOUND"
    elif [ "$status_code" = "400" ]; then
        echo "⚠️  $endpoint - BAD REQUEST"
        echo "📝 Response: $(echo "$response_body" | head -c 100)..."
    else
        echo "❓ $endpoint - HTTP $status_code"
        echo "📝 Response: $(echo "$response_body" | head -c 100)..."
    fi
    echo
done

echo "=== Testing GET request for available endpoints ==="
curl -s http://localhost:11435/ | head -5 || echo "No response from GET /"

echo
echo "=== Testing WebSocket upgrade directly ==="
echo "Note: This might time out - that's expected if bridge accepts WebSocket"
timeout 3 curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
    -H "Sec-WebSocket-Key: test" -H "Sec-WebSocket-Version: 13" \
    http://localhost:11435/ 2>/dev/null | head -10 || echo "WebSocket test completed"

echo
echo "=== Bridge API Testing Complete ==="