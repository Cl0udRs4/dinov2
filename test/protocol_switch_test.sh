#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "Testing protocol switching functionality..."

# Start the server
echo -e "Starting server..."
./bin/server -config test/server_config.json &
SERVER_PID=$!
sleep 2

# Start the client with TCP protocol
echo -e "Starting client with TCP protocol..."
./bin/client -server localhost:8080 -protocol tcp,http,websocket -heartbeat 5 &
CLIENT_PID=$!
sleep 5

# Test protocol switching via API
echo -e "Testing protocol switching via API..."
curl -s -X POST -H "Content-Type: application/json" -d '{"client_id":"test_client","protocol":"http"}' http://localhost:8000/api/protocol/switch
sleep 5

# Check if client is still connected
if ps -p $CLIENT_PID > /dev/null; then
    echo -e "✓ Client still running after protocol switch request"
else
    echo -e "✗ Client terminated after protocol switch request"
fi

# Clean up
echo -e "Cleaning up..."
kill $CLIENT_PID
kill $SERVER_PID
wait

echo -e "Protocol switching test completed"
