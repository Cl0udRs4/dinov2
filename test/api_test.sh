#!/bin/bash

# Test script for HTTP API operations

# Set up variables
API_URL="http://127.0.0.1:8443/api"
AUTH_TOKEN=""

# Function to get authentication token
get_auth_token() {
    local response=$(curl -s -X POST "$API_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"change_this_in_production"}')
    
    AUTH_TOKEN=$(echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    
    if [ -z "$AUTH_TOKEN" ]; then
        echo "Failed to get authentication token"
        exit 1
    fi
    
    echo "Authentication successful"
}

# Function to create a listener
create_listener() {
    local id=$1
    local type=$2
    local address=$3
    local port=$4
    
    echo "Creating $type listener on $address:$port..."
    
    local response=$(curl -s -X POST "$API_URL/listeners/create" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -d "{\"id\":\"$id\",\"type\":\"$type\",\"address\":\"$address\",\"port\":$port,\"options\":{}}")
    
    echo "$response"
}

# Function to list listeners
list_listeners() {
    echo "Listing listeners..."
    
    local response=$(curl -s -X GET "$API_URL/listeners" \
        -H "Authorization: Bearer $AUTH_TOKEN")
    
    echo "$response"
}

# Function to get listener status
get_listener_status() {
    local id=$1
    
    echo "Getting status for listener $id..."
    
    local response=$(curl -s -X GET "$API_URL/listeners/status?id=$id" \
        -H "Authorization: Bearer $AUTH_TOKEN")
    
    echo "$response"
}

# Function to delete a listener
delete_listener() {
    local id=$1
    
    echo "Deleting listener $id..."
    
    local response=$(curl -s -X POST "$API_URL/listeners/delete" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -d "{\"id\":\"$id\"}")
    
    echo "$response"
}

# Function to list clients
list_clients() {
    echo "Listing clients..."
    
    local response=$(curl -s -X GET "$API_URL/clients" \
        -H "Authorization: Bearer $AUTH_TOKEN")
    
    echo "$response"
}

# Function to switch client protocol
switch_protocol() {
    local client_id=$1
    local protocol=$2
    
    echo "Switching client $client_id to protocol $protocol..."
    
    local response=$(curl -s -X POST "$API_URL/protocol/switch" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -d "{\"client_id\":\"$client_id\",\"protocol\":\"$protocol\"}")
    
    echo "$response"
}

# Main test sequence
echo "Starting API tests..."

# Get authentication token
get_auth_token

# Create listeners
create_listener "tcp1" "tcp" "0.0.0.0" 8080
create_listener "http1" "http" "0.0.0.0" 8081
create_listener "ws1" "websocket" "0.0.0.0" 8082

# List listeners
list_listeners

# Get listener status
get_listener_status "tcp1"

# List clients (should be empty initially)
list_clients

# Wait for clients to connect (manual step)
echo "Please start client(s) now and press Enter when ready..."
read

# List clients again (should show connected clients)
list_clients

# Switch protocol for a client (replace CLIENT_ID with actual client ID)
echo "Enter client ID to switch protocol:"
read CLIENT_ID
switch_protocol "$CLIENT_ID" "http"

# Delete listeners
delete_listener "tcp1"
delete_listener "http1"
delete_listener "ws1"

echo "API tests completed."
