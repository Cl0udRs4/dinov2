#!/bin/bash

echo "Starting WebSocket protocol test..."

# Start the server with WebSocket listener
cd "$(dirname "$0")"
CONFIG_FILE="config.json"

# Ensure the server is not running
pkill -f "server -config"

# Start the server in the background
echo "Starting server with WebSocket listener..."
../cmd/server/server -config $CONFIG_FILE &
SERVER_PID=$!

# Give the server time to start
sleep 2

# Run a WebSocket client test using websocat if available
if command -v websocat &> /dev/null; then
  echo "Running WebSocket client test..."
  echo "test message" | websocat ws://127.0.0.1:8444/ws
else
  echo "websocat not found, skipping WebSocket client test"
  echo "You can install websocat with: cargo install websocat"
fi

# Check if the server received the request
sleep 1

# Clean up
echo "Cleaning up..."
kill $SERVER_PID

echo "WebSocket protocol test completed."
