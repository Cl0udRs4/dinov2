#!/bin/bash

# API test script
echo "Running API tests..."

# Set up environment
API_URL="http://localhost:8081"
AUTH_TOKEN="test-token"

# Start server in the background
echo "Starting server..."
cd "$(dirname "$0")/.."
./bin/server -config test/config.json -api-port 8081 &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Test GET /api/v1/listeners
echo "Testing GET /api/v1/listeners..."
RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $AUTH_TOKEN" $API_URL/api/v1/listeners)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
    echo "GET /api/v1/listeners: Success"
else
    echo "GET /api/v1/listeners: Failed with HTTP code $HTTP_CODE"
    echo "Response: $BODY"
    kill $SERVER_PID
    exit 1
fi

# Test POST /api/v1/listeners
echo "Testing POST /api/v1/listeners..."
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST -H "Authorization: Bearer $AUTH_TOKEN" -H "Content-Type: application/json" -d '{"id":"test-listener","type":"tcp","address":"127.0.0.1","port":8082,"enabled":true}' $API_URL/api/v1/listeners)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 201 ]; then
    echo "POST /api/v1/listeners: Success"
    # Extract the listener ID from the response
    LISTENER_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    echo "Created listener with ID: $LISTENER_ID"
else
    echo "POST /api/v1/listeners: Failed with HTTP code $HTTP_CODE"
    echo "Response: $BODY"
    kill $SERVER_PID
    exit 1
fi

# Test GET /api/v1/listeners/{id}
echo "Testing GET /api/v1/listeners/$LISTENER_ID..."
RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $AUTH_TOKEN" $API_URL/api/v1/listeners/$LISTENER_ID)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
    echo "GET /api/v1/listeners/$LISTENER_ID: Success"
else
    echo "GET /api/v1/listeners/$LISTENER_ID: Failed with HTTP code $HTTP_CODE"
    echo "Response: $BODY"
    kill $SERVER_PID
    exit 1
fi

# Test DELETE /api/v1/listeners/{id}
echo "Testing DELETE /api/v1/listeners/$LISTENER_ID..."
RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE -H "Authorization: Bearer $AUTH_TOKEN" $API_URL/api/v1/listeners/$LISTENER_ID)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
    echo "DELETE /api/v1/listeners/$LISTENER_ID: Success"
else
    echo "DELETE /api/v1/listeners/$LISTENER_ID: Failed with HTTP code $HTTP_CODE"
    echo "Response: $BODY"
    kill $SERVER_PID
    exit 1
fi

# Test GET /api/v1/clients
echo "Testing GET /api/v1/clients..."
RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $AUTH_TOKEN" $API_URL/api/v1/clients)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
    echo "GET /api/v1/clients: Success"
else
    echo "GET /api/v1/clients: Failed with HTTP code $HTTP_CODE"
    echo "Response: $BODY"
    kill $SERVER_PID
    exit 1
fi

# Test GET /api/v1/modules
echo "Testing GET /api/v1/modules..."
RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $AUTH_TOKEN" $API_URL/api/v1/modules)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
    echo "GET /api/v1/modules: Success"
else
    echo "GET /api/v1/modules: Failed with HTTP code $HTTP_CODE"
    echo "Response: $BODY"
    kill $SERVER_PID
    exit 1
fi

# Clean up
echo "Cleaning up..."
kill $SERVER_PID

echo "All tests passed!"
exit 0
