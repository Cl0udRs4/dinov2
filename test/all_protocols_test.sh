#!/bin/bash

# Test script for encryption algorithm detection across all protocols

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

# Function to create a listener
create_listener() {
    local id=$1
    local type=$2
    local port=$3
    
    echo "Creating $type listener on port $port..."
    
    local response=$(curl -s -X POST "$API_URL/listeners/create" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -d "{\"id\":\"$id\",\"type\":\"$type\",\"address\":\"0.0.0.0\",\"port\":$port,\"options\":{}}")
    
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

# Function to build a client with specific encryption algorithm and protocols
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

# Function to check if a client is registered with specific encryption and protocol
check_client_registration() {
    local encryption=$1
    local protocol=$2
    
    echo "Checking if client with encryption=$encryption and protocol=$protocol is registered..."
    
    # Wait for client to register
    for i in {1..10}; do
        if list_clients | grep -q "\"encryption_algorithm\":\"$encryption\"" && list_clients | grep -q "\"protocol\":\"$protocol\""; then
            echo "Client with encryption=$encryption and protocol=$protocol is registered"
            return 0
        fi
        
        echo "Waiting for client to register... ($i/10)"
        sleep 2
    done
    
    echo "Client with encryption=$encryption and protocol=$protocol is not registered"
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
        
        curl -s -X POST "$API_URL/listeners/delete" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            -d "{\"id\":\"http1\"}"
        
        curl -s -X POST "$API_URL/listeners/delete" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            -d "{\"id\":\"ws1\"}"
        
        curl -s -X POST "$API_URL/listeners/delete" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            -d "{\"id\":\"dns1\"}"
        
        curl -s -X POST "$API_URL/listeners/delete" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            -d "{\"id\":\"icmp1\"}"
    fi
}

# Main test sequence
echo "Starting encryption algorithm detection tests for all protocols..."

# Create output directories
mkdir -p "$CLIENT_DIR"

# Get authentication token
get_auth_token

# Create and start listeners for all protocols
create_listener "tcp1" "tcp" "$SERVER_PORT"
start_listener "tcp1"

create_listener "http1" "http" "8081"
start_listener "http1"

create_listener "ws1" "websocket" "8082"
start_listener "ws1"

create_listener "dns1" "dns" "8053"
start_listener "dns1"

# ICMP requires root privileges, so we'll only test it if we're running as root
if [ $(id -u) -eq 0 ]; then
    create_listener "icmp1" "icmp" "0"
    start_listener "icmp1"
fi

# Build clients with different encryption algorithms and protocols
build_client "$CLIENT_DIR" "aes" "tcp"
build_client "$CLIENT_DIR" "chacha20" "tcp"
build_client "$CLIENT_DIR" "aes" "http"
build_client "$CLIENT_DIR" "chacha20" "http"
build_client "$CLIENT_DIR" "aes" "websocket"
build_client "$CLIENT_DIR" "chacha20" "websocket"
build_client "$CLIENT_DIR" "aes" "dns"
build_client "$CLIENT_DIR" "chacha20" "dns"

# Only build ICMP clients if we're running as root
if [ $(id -u) -eq 0 ]; then
    build_client "$CLIENT_DIR" "aes" "icmp"
    build_client "$CLIENT_DIR" "chacha20" "icmp"
fi

# Run clients one by one and check if they're registered
echo "Testing TCP clients..."
tcp_aes_pid=$(run_client "$CLIENT_DIR/client_aes_tcp" "$CLIENT_DIR/client_aes_tcp.log")
sleep 5
check_client_registration "aes" "tcp"
tcp_aes_registered=$?

pkill -f "client_aes_tcp"
sleep 2

tcp_chacha20_pid=$(run_client "$CLIENT_DIR/client_chacha20_tcp" "$CLIENT_DIR/client_chacha20_tcp.log")
sleep 5
check_client_registration "chacha20" "tcp"
tcp_chacha20_registered=$?

pkill -f "client_chacha20_tcp"
sleep 2

echo "Testing HTTP clients..."
http_aes_pid=$(run_client "$CLIENT_DIR/client_aes_http" "$CLIENT_DIR/client_aes_http.log")
sleep 5
check_client_registration "aes" "http"
http_aes_registered=$?

pkill -f "client_aes_http"
sleep 2

http_chacha20_pid=$(run_client "$CLIENT_DIR/client_chacha20_http" "$CLIENT_DIR/client_chacha20_http.log")
sleep 5
check_client_registration "chacha20" "http"
http_chacha20_registered=$?

pkill -f "client_chacha20_http"
sleep 2

echo "Testing WebSocket clients..."
ws_aes_pid=$(run_client "$CLIENT_DIR/client_aes_websocket" "$CLIENT_DIR/client_aes_websocket.log")
sleep 5
check_client_registration "aes" "websocket"
ws_aes_registered=$?

pkill -f "client_aes_websocket"
sleep 2

ws_chacha20_pid=$(run_client "$CLIENT_DIR/client_chacha20_websocket" "$CLIENT_DIR/client_chacha20_websocket.log")
sleep 5
check_client_registration "chacha20" "websocket"
ws_chacha20_registered=$?

pkill -f "client_chacha20_websocket"
sleep 2

echo "Testing DNS clients..."
dns_aes_pid=$(run_client "$CLIENT_DIR/client_aes_dns" "$CLIENT_DIR/client_aes_dns.log")
sleep 5
check_client_registration "aes" "dns"
dns_aes_registered=$?

pkill -f "client_aes_dns"
sleep 2

dns_chacha20_pid=$(run_client "$CLIENT_DIR/client_chacha20_dns" "$CLIENT_DIR/client_chacha20_dns.log")
sleep 5
check_client_registration "chacha20" "dns"
dns_chacha20_registered=$?

pkill -f "client_chacha20_dns"
sleep 2

# Only test ICMP clients if we're running as root
if [ $(id -u) -eq 0 ]; then
    echo "Testing ICMP clients..."
    icmp_aes_pid=$(run_client "$CLIENT_DIR/client_aes_icmp" "$CLIENT_DIR/client_aes_icmp.log")
    sleep 5
    check_client_registration "aes" "icmp"
    icmp_aes_registered=$?

    pkill -f "client_aes_icmp"
    sleep 2

    icmp_chacha20_pid=$(run_client "$CLIENT_DIR/client_chacha20_icmp" "$CLIENT_DIR/client_chacha20_icmp.log")
    sleep 5
    check_client_registration "chacha20" "icmp"
    icmp_chacha20_registered=$?

    pkill -f "client_chacha20_icmp"
    sleep 2
fi

# Clean up
cleanup

# Report results
echo "Test results:"
echo "TCP AES client registration: $([ $tcp_aes_registered -eq 0 ] && echo "SUCCESS" || echo "FAILED")"
echo "TCP ChaCha20 client registration: $([ $tcp_chacha20_registered -eq 0 ] && echo "SUCCESS" || echo "FAILED")"
echo "HTTP AES client registration: $([ $http_aes_registered -eq 0 ] && echo "SUCCESS" || echo "FAILED")"
echo "HTTP ChaCha20 client registration: $([ $http_chacha20_registered -eq 0 ] && echo "SUCCESS" || echo "FAILED")"
echo "WebSocket AES client registration: $([ $ws_aes_registered -eq 0 ] && echo "SUCCESS" || echo "FAILED")"
echo "WebSocket ChaCha20 client registration: $([ $ws_chacha20_registered -eq 0 ] && echo "SUCCESS" || echo "FAILED")"
echo "DNS AES client registration: $([ $dns_aes_registered -eq 0 ] && echo "SUCCESS" || echo "FAILED")"
echo "DNS ChaCha20 client registration: $([ $dns_chacha20_registered -eq 0 ] && echo "SUCCESS" || echo "FAILED")"

if [ $(id -u) -eq 0 ]; then
    echo "ICMP AES client registration: $([ $icmp_aes_registered -eq 0 ] && echo "SUCCESS" || echo "FAILED")"
    echo "ICMP ChaCha20 client registration: $([ $icmp_chacha20_registered -eq 0 ] && echo "SUCCESS" || echo "FAILED")"
fi

# Calculate overall success
if [ $(id -u) -eq 0 ]; then
    # With ICMP
    if [ $tcp_aes_registered -eq 0 ] && [ $tcp_chacha20_registered -eq 0 ] && \
       [ $http_aes_registered -eq 0 ] && [ $http_chacha20_registered -eq 0 ] && \
       [ $ws_aes_registered -eq 0 ] && [ $ws_chacha20_registered -eq 0 ] && \
       [ $dns_aes_registered -eq 0 ] && [ $dns_chacha20_registered -eq 0 ] && \
       [ $icmp_aes_registered -eq 0 ] && [ $icmp_chacha20_registered -eq 0 ]; then
        echo "All tests passed!"
        exit 0
    else
        echo "Some tests failed!"
        exit 1
    fi
else
    # Without ICMP
    if [ $tcp_aes_registered -eq 0 ] && [ $tcp_chacha20_registered -eq 0 ] && \
       [ $http_aes_registered -eq 0 ] && [ $http_chacha20_registered -eq 0 ] && \
       [ $ws_aes_registered -eq 0 ] && [ $ws_chacha20_registered -eq 0 ] && \
       [ $dns_aes_registered -eq 0 ] && [ $dns_chacha20_registered -eq 0 ]; then
        echo "All tests passed!"
        exit 0
    else
        echo "Some tests failed!"
        exit 1
    fi
fi
