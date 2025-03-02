#!/bin/bash

# Test script for client registration and encryption algorithm detection

# Set up variables
SERVER_PORT=8080
API_PORT=8443
API_URL="http://127.0.0.1:$API_PORT/api"
AUTH_TOKEN=""
CLIENT_DIR="./bin/clients"
SERVER_LOG="./bin/server.log"

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

# Function to create a TCP listener
create_tcp_listener() {
    local id=$1
    local port=$2
    
    echo "Creating TCP listener on port $port..."
    
    local response=$(curl -s -X POST "$API_URL/listeners/create" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -d "{\"id\":\"$id\",\"type\":\"tcp\",\"address\":\"0.0.0.0\",\"port\":$port,\"options\":{}}")
    
    echo "$response"
}

# Function to start a listener
start_listener() {
    local id=$1
    
    echo "Starting listener $id..."
    
    local response=$(curl -s -X POST "$API_URL/listeners/start" \
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
    
    # Check if there are any clients
    if [[ "$response" == *"\"clients\":[]"* ]]; then
        echo "No clients found"
        return 1
    else
        echo "Clients found"
        return 0
    fi
}

# Function to build a client with specific encryption algorithm
build_client() {
    local output_dir=$1
    local encryption=$2
    local protocols=$3
    
    echo "Building client with encryption=$encryption and protocols=$protocols..."
    
    mkdir -p "$output_dir"
    
    go run cmd/builder/main.go \
        --output "$output_dir/client_${encryption}_${protocols// /_}" \
        --server-address "127.0.0.1:$SERVER_PORT" \
        --encryption "$encryption" \
        --protocols "$protocols"
    
    if [ $? -ne 0 ]; then
        echo "Failed to build client"
        return 1
    fi
    
    echo "Client built successfully"
    return 0
}

# Function to run a client
run_client() {
    local client_path=$1
    local log_file=$2
    
    echo "Running client $client_path..."
    
    # Run the client in the background and redirect output to log file
    "$client_path" > "$log_file" 2>&1 &
    
    # Save the PID
    local pid=$!
    
    echo "Client started with PID $pid"
    echo "$pid"
}

# Function to check if a client is registered
check_client_registration() {
    local encryption=$1
    
    echo "Checking if client with encryption=$encryption is registered..."
    
    # Wait for client to register
    for i in {1..10}; do
        if list_clients | grep -q "\"encryption_algorithm\":\"$encryption\""; then
            echo "Client with encryption=$encryption is registered"
            return 0
        fi
        
        echo "Waiting for client to register... ($i/10)"
        sleep 2
    done
    
    echo "Client with encryption=$encryption is not registered"
    return 1
}

# Function to clean up
cleanup() {
    echo "Cleaning up..."
    
    # Kill all client processes
    pkill -f "client_aes"
    pkill -f "client_chacha20"
    
    # Delete listeners
    if [ -n "$AUTH_TOKEN" ]; then
        curl -s -X POST "$API_URL/listeners/delete" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            -d "{\"id\":\"tcp1\"}"
    fi
}

# Main test sequence
echo "Starting client registration and encryption algorithm detection tests..."

# Create output directories
mkdir -p "$CLIENT_DIR"

# Get authentication token
get_auth_token

# Create and start TCP listener
create_tcp_listener "tcp1" "$SERVER_PORT"
start_listener "tcp1"

# Build clients with different encryption algorithms
build_client "$CLIENT_DIR" "aes" "tcp"
build_client "$CLIENT_DIR" "chacha20" "tcp"

# Run clients
aes_pid=$(run_client "$CLIENT_DIR/client_aes_tcp" "$CLIENT_DIR/client_aes.log")
sleep 5
chacha20_pid=$(run_client "$CLIENT_DIR/client_chacha20_tcp" "$CLIENT_DIR/client_chacha20.log")

# Check if clients are registered
check_client_registration "aes"
aes_registered=$?

check_client_registration "chacha20"
chacha20_registered=$?

# List clients
list_clients

# Clean up
cleanup

# Report results
echo "Test results:"
echo "AES client registration: $([ $aes_registered -eq 0 ] && echo "SUCCESS" || echo "FAILED")"
echo "ChaCha20 client registration: $([ $chacha20_registered -eq 0 ] && echo "SUCCESS" || echo "FAILED")"

if [ $aes_registered -eq 0 ] && [ $chacha20_registered -eq 0 ]; then
    echo "All tests passed!"
    exit 0
else
    echo "Some tests failed!"
    exit 1
fi
