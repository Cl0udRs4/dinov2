#!/bin/bash

echo "Starting DNS protocol test..."

# Start the server with DNS listener
cd "$(dirname "$0")"
CONFIG_FILE="config.json"

# Ensure the server is not running
pkill -f "server -config"

# Start the server in the background
echo "Starting server with DNS listener..."
../cmd/server/server -config $CONFIG_FILE &
SERVER_PID=$!

# Give the server time to start
sleep 2

# Run a DNS client test
echo "Running DNS client test..."
dig @127.0.0.1 -p 53 test.c2.example.com TXT

# Check if the server received the request
sleep 1

# Clean up
echo "Cleaning up..."
kill $SERVER_PID

echo "DNS protocol test completed."
