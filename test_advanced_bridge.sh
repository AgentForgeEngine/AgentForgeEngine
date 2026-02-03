#!/bin/bash

echo "=== Advanced WebSocket Bridge Testing ==="
echo

# Test 1: Different WebSocket paths
echo "1. Testing different WebSocket paths..."
paths=(
    "/"
    "/ws"
    "/rpc"
    "/api"
    "/jsonrpc"
)

for path in "${paths[@]}"; do
    echo "Testing: ws://localhost:11435$path"
    timeout 3 wscat -x ws://localhost:11435$path -c '{"test":"connection"}' 2>/dev/null && \
        echo "✅ Connected to $path" || \
        echo "❌ Failed to connect to $path"
    echo
done

# Test 2: Test with different protocols
echo "2. Testing with different subprotocols..."
echo "Note: If wscat not available, this will skip"

# Test 3: Netcat connection to see if anything is sent
echo "3. Raw connection test to see initial data..."
echo "Connecting to localhost:11435..."
timeout 3 nc localhost 11435 < /dev/null || echo "Raw connection failed"
echo

# Test 4: Check if bridge responds to any HTTP method
echo "4. Testing different HTTP methods..."
methods=("GET" "POST" "PUT" "OPTIONS" "HEAD")

for method in "${methods[@]}"; do
    echo "Testing $method /"
    response=$(curl -s -w "%{http_code}" -X "$method" http://localhost:11435/ 2>/dev/null)
    echo "Response: $response"
done

echo
echo "=== Advanced Testing Complete ==="

# Test 5: Try to run wscat interactive mode test
echo "5. Interactive WebSocket test (press Ctrl+C after testing)..."
echo "Command: wscat -x ws://localhost:11435 -c '{\"jsonrpc\":\"2.0\",\"method\":\"test\",\"id\":1}'"
echo "If wscat is available, you can run this manually to test interactively"