#!/bin/bash

echo "Starting HTTP protocol test..."

# Start the server with HTTP listener
cd "$(dirname "$0")"
CONFIG_FILE="config.json"

# Ensure the server is not running
pkill -f "server -config"

# Start the server in the background
echo "Starting server with HTTP listener..."
../cmd/server/server -config $CONFIG_FILE &
SERVER_PID=$!

# Give the server time to start
sleep 2

# Run an HTTP client test
echo "Running HTTP client test..."
curl -v http://127.0.0.1:8443/

# Test POST request with command header
echo "Testing command POST request..."
echo "test data" | curl -v -X POST -H "X-Command: test" -d @- http://127.0.0.1:8443/

# Check if the server received the request
sleep 1

# Clean up
echo "Cleaning up..."
kill $SERVER_PID

echo "HTTP protocol test completed."
