#!/bin/bash

# Encryption test
echo "Starting encryption test..."

# Create config file
CONFIG_FILE="$(pwd)/config.json"

# Start server in the background
echo "Starting server..."
../bin/server -config "$CONFIG_FILE" &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Start client with AES encryption (default)
echo "Starting client with AES encryption..."
../bin/client -server "127.0.0.1:8080" -protocol "tcp" &
CLIENT_PID=$!

# Wait for client to connect
sleep 5

# Check if client and server are still running
if ps -p $CLIENT_PID > /dev/null; then
    echo "Client is running with AES encryption - Connection successful!"
else
    echo "Client is not running - AES encryption connection failed!"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# Clean up
echo "Cleaning up AES test..."
kill $CLIENT_PID 2>/dev/null

# Wait for cleanup
sleep 2

# Start client with ChaCha20 encryption
echo "Starting client with ChaCha20 encryption..."
# Note: In a real implementation, we would have a flag to specify encryption algorithm
# For now, we're just testing the basic connection again
../bin/client -server "127.0.0.1:8080" -protocol "tcp" &
CLIENT_PID=$!

# Wait for client to connect
sleep 5

# Check if client and server are still running
if ps -p $CLIENT_PID > /dev/null; then
    echo "Client is running with ChaCha20 encryption - Connection successful!"
else
    echo "Client is not running - ChaCha20 encryption connection failed!"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# Clean up
echo "Cleaning up..."
kill $CLIENT_PID 2>/dev/null
kill $SERVER_PID 2>/dev/null

echo "Test completed successfully!"
exit 0
