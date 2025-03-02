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

# Rebuild the clients to ensure they're up to date
echo "Building clients..."
cd cmd/builder
go build -o builder

# Build clients in the temporary directory to avoid permission issues
echo "Building TCP client with AES encryption..."
./builder -server "127.0.0.1:8080" -protocol "tcp" -encryption "aes" -output "$TEST_DIR/client_tcp_aes" -verbose

echo "Building TCP client with ChaCha20 encryption..."
./builder -server "127.0.0.1:8080" -protocol "tcp" -encryption "chacha20" -output "$TEST_DIR/client_tcp_chacha20" -verbose

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
    kill $CLIENT_PID
    echo "TCP client with AES encryption test completed."
else
    echo "Client executable not found: client_tcp_aes"
fi

echo "Testing TCP client with ChaCha20 encryption..."
if [ -f "./client_tcp_chacha20" ]; then
    ./client_tcp_chacha20 &
    CLIENT_PID=$!
    sleep 5
    kill $CLIENT_PID
    echo "TCP client with ChaCha20 encryption test completed."
else
    echo "Client executable not found: client_tcp_chacha20"
fi

echo "Testing multi-protocol client with AES encryption..."
if [ -f "./client_multi_aes" ]; then
    ./client_multi_aes &
    CLIENT_PID=$!
    sleep 5
    kill $CLIENT_PID
    echo "Multi-protocol client with AES encryption test completed."
else
    echo "Client executable not found: client_multi_aes"
fi

echo "Testing multi-protocol client with ChaCha20 encryption..."
if [ -f "./client_multi_chacha20" ]; then
    ./client_multi_chacha20 &
    CLIENT_PID=$!
    sleep 5
    kill $CLIENT_PID
    echo "Multi-protocol client with ChaCha20 encryption test completed."
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
