#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "Testing client functionality..."

# Test TCP protocol
echo -e "Testing TCP protocol..."
./bin/client -server localhost:8080 -protocol tcp -heartbeat 5 &
CLIENT_PID=$!
sleep 10
kill $CLIENT_PID
echo -e "✓ TCP protocol test completed"

# Test HTTP protocol
echo -e "Testing HTTP protocol..."
./bin/client -server localhost:8000 -protocol http -heartbeat 5 &
CLIENT_PID=$!
sleep 10
kill $CLIENT_PID
echo -e "✓ HTTP protocol test completed"

# Test WebSocket protocol
echo -e "Testing WebSocket protocol..."
./bin/client -server localhost:8001 -protocol websocket -heartbeat 5 &
CLIENT_PID=$!
sleep 10
kill $CLIENT_PID
echo -e "✓ WebSocket protocol test completed"

# Test DNS protocol
echo -e "Testing DNS protocol..."
./bin/client -server localhost:5353 -protocol dns -heartbeat 5 &
CLIENT_PID=$!
sleep 10
kill $CLIENT_PID
echo -e "✓ DNS protocol test completed"

echo -e "Client testing completed"
