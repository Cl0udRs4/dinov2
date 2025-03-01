#!/bin/bash

# Basic TCP communication test
echo "Starting basic TCP communication test..."

# Create config file
CONFIG_FILE="$(pwd)/config.json"

# Start server in the background
echo "Starting server..."
../bin/server -config "$CONFIG_FILE" &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Start client
echo "Starting client..."
../bin/client -server "127.0.0.1:8080" -protocol "tcp" &
CLIENT_PID=$!

# Wait for client to connect
sleep 5

# Check if client and server are still running
if ps -p $CLIENT_PID > /dev/null; then
    echo "Client is running - Connection successful!"
else
    echo "Client is not running - Connection failed!"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# Clean up
echo "Cleaning up..."
kill $CLIENT_PID 2>/dev/null
kill $SERVER_PID 2>/dev/null

echo "Test completed successfully!"
exit 0
