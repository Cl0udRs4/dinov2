#!/bin/bash

# Create a temporary directory for testing
TEST_DIR=$(mktemp -d)
echo "Using temporary directory: $TEST_DIR"

# Build the server
cd cmd/server
go build -o server

# Start the server with the test configuration
./server -config ../../test_server_config.json &
SERVER_PID=$!

# Wait for the server to start
sleep 2

# Go back to the root directory
cd ../../

# Function to check if clients are registered via API
check_clients() {
    echo "Checking registered clients via API..."
    # Get authentication token
    TOKEN=$(curl -s -X POST -H "Content-Type: application/json" -d '{"username":"admin","password":"admin"}' http://127.0.0.1:8443/api/auth/login | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    
    if [ -z "$TOKEN" ]; then
        echo "Failed to get authentication token"
        return 1
    fi
    
    # Get client list
    CLIENT_LIST=$(curl -s -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8443/api/clients)
    echo "Client list: $CLIENT_LIST"
    
    # Check if client list is empty
    if [[ "$CLIENT_LIST" == *"\"clients\":[]"* ]] || [[ "$CLIENT_LIST" == "[]" ]]; then
        echo "No clients registered!"
        return 1
    else
        echo "Clients successfully registered!"
        return 0
    fi
}

# Rebuild the clients to ensure they're up to date
echo "Building clients..."
cd cmd/builder
go build -o builder

# Build clients in the temporary directory to avoid permission issues
echo "Building TCP client with AES encryption..."
./builder -server "127.0.0.1:8080" -protocol "tcp" -encryption "aes" -output "$TEST_DIR/client_tcp_aes" -verbose

echo "Building TCP client with ChaCha20 encryption..."
./builder -server "127.0.0.1:8080" -protocol "tcp" -encryption "chacha20" -output "$TEST_DIR/client_tcp_chacha20" -verbose

echo "Building HTTP client with AES encryption..."
./builder -server "127.0.0.1:8000" -protocol "http" -encryption "aes" -output "$TEST_DIR/client_http_aes" -verbose

echo "Building WebSocket client with AES encryption..."
./builder -server "127.0.0.1:8001" -protocol "websocket" -encryption "aes" -output "$TEST_DIR/client_ws_aes" -verbose

echo "Building multi-protocol client with AES encryption..."
./builder -server "127.0.0.1:8080" -protocol "tcp,http,websocket" -encryption "aes" -output "$TEST_DIR/client_multi_aes" -verbose

echo "Building multi-protocol client with ChaCha20 encryption..."
./builder -server "127.0.0.1:8080" -protocol "tcp,http,websocket" -encryption "chacha20" -output "$TEST_DIR/client_multi_chacha20" -verbose

# Go to the temporary directory
cd "$TEST_DIR"

# Test client-server communication
echo "Testing TCP client with AES encryption..."
if [ -f "./client_tcp_aes" ]; then
    ./client_tcp_aes &
    CLIENT_PID=$!
    sleep 5
    check_clients
    CLIENT_STATUS=$?
    kill $CLIENT_PID
    if [ $CLIENT_STATUS -eq 0 ]; then
        echo "TCP client with AES encryption test PASSED."
    else
        echo "TCP client with AES encryption test FAILED - client not registered."
    fi
else
    echo "Client executable not found: client_tcp_aes"
fi

echo "Testing TCP client with ChaCha20 encryption..."
if [ -f "./client_tcp_chacha20" ]; then
    ./client_tcp_chacha20 &
    CLIENT_PID=$!
    sleep 5
    check_clients
    CLIENT_STATUS=$?
    kill $CLIENT_PID
    if [ $CLIENT_STATUS -eq 0 ]; then
        echo "TCP client with ChaCha20 encryption test PASSED."
    else
        echo "TCP client with ChaCha20 encryption test FAILED - client not registered."
    fi
else
    echo "Client executable not found: client_tcp_chacha20"
fi

echo "Testing HTTP client with AES encryption..."
if [ -f "./client_http_aes" ]; then
    ./client_http_aes &
    CLIENT_PID=$!
    sleep 5
    check_clients
    CLIENT_STATUS=$?
    kill $CLIENT_PID
    if [ $CLIENT_STATUS -eq 0 ]; then
        echo "HTTP client with AES encryption test PASSED."
    else
        echo "HTTP client with AES encryption test FAILED - client not registered."
    fi
else
    echo "Client executable not found: client_http_aes"
fi

echo "Testing WebSocket client with AES encryption..."
if [ -f "./client_ws_aes" ]; then
    ./client_ws_aes &
    CLIENT_PID=$!
    sleep 5
    check_clients
    CLIENT_STATUS=$?
    kill $CLIENT_PID
    if [ $CLIENT_STATUS -eq 0 ]; then
        echo "WebSocket client with AES encryption test PASSED."
    else
        echo "WebSocket client with AES encryption test FAILED - client not registered."
    fi
else
    echo "Client executable not found: client_ws_aes"
fi

echo "Testing multi-protocol client with AES encryption..."
if [ -f "./client_multi_aes" ]; then
    ./client_multi_aes &
    CLIENT_PID=$!
    sleep 5
    check_clients
    CLIENT_STATUS=$?
    kill $CLIENT_PID
    if [ $CLIENT_STATUS -eq 0 ]; then
        echo "Multi-protocol client with AES encryption test PASSED."
    else
        echo "Multi-protocol client with AES encryption test FAILED - client not registered."
    fi
else
    echo "Client executable not found: client_multi_aes"
fi

echo "Testing multi-protocol client with ChaCha20 encryption..."
if [ -f "./client_multi_chacha20" ]; then
    ./client_multi_chacha20 &
    CLIENT_PID=$!
    sleep 5
    check_clients
    CLIENT_STATUS=$?
    kill $CLIENT_PID
    if [ $CLIENT_STATUS -eq 0 ]; then
        echo "Multi-protocol client with ChaCha20 encryption test PASSED."
    else
        echo "Multi-protocol client with ChaCha20 encryption test FAILED - client not registered."
    fi
else
    echo "Client executable not found: client_multi_chacha20"
fi

# Stop the server
if [ -n "$SERVER_PID" ]; then
    kill $SERVER_PID
    echo "Server stopped."
else
    echo "Server PID not found."
fi

# Clean up
cd -
rm -rf "$TEST_DIR"

echo "Server-client communication tests completed."
