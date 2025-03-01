#!/bin/bash

echo "Starting ICMP protocol test..."

# Start the server with ICMP listener
cd "$(dirname "$0")"
CONFIG_FILE="config.json"

# Ensure the server is not running
pkill -f "server -config"

# Start the server in the background with sudo (ICMP requires root)
echo "Starting server with ICMP listener..."
sudo ../cmd/server/server -config $CONFIG_FILE &
SERVER_PID=$!

# Give the server time to start
sleep 2

# Run an ICMP client test
echo "Running ICMP client test..."
ping -c 3 127.0.0.1

# Check if the server received the request
sleep 1

# Clean up
echo "Cleaning up..."
sudo kill $SERVER_PID

echo "ICMP protocol test completed."
