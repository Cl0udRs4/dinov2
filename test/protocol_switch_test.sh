#!/bin/bash

# Protocol switching test
echo "Starting protocol switching test..."

# Create config file for multiple protocols
cat > "$(pwd)/multi_protocol_config.json" << EOF
{
  "listeners": [
    {
      "id": "tcp1",
      "type": "tcp",
      "address": "127.0.0.1",
      "port": 8080,
      "options": {}
    },
    {
      "id": "tcp2",
      "type": "tcp",
      "address": "127.0.0.1",
      "port": 8081,
      "options": {}
    }
  ]
}
EOF

# Start server in the background
echo "Starting server with multiple listeners..."
../bin/server -config "$(pwd)/multi_protocol_config.json" &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Start client with multiple protocols
echo "Starting client with multiple protocols..."
../bin/client -server "127.0.0.1:8080" -protocol "tcp" &
CLIENT_PID=$!

# Wait for client to connect
sleep 5

# Check if client and server are still running
if ps -p $CLIENT_PID > /dev/null; then
    echo "Client is running - Initial connection successful!"
else
    echo "Client is not running - Initial connection failed!"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# Clean up
echo "Cleaning up..."
kill $CLIENT_PID 2>/dev/null
kill $SERVER_PID 2>/dev/null

echo "Test completed successfully!"
exit 0
