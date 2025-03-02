#!/bin/bash

# Test script for builder and client functionality

# Set up variables
SERVER_PORT=8080
SERVER_ADDR="127.0.0.1:$SERVER_PORT"
BUILD_DIR="$(pwd)/bin"
PROTOCOLS=("tcp" "http" "websocket" "dns" "icmp")
ENCRYPTION_ALGS=("aes" "chacha20")

# Create build directory if it doesn't exist
mkdir -p "$BUILD_DIR"

# Build the server
echo "Building server..."
go build -o "$BUILD_DIR/server" ./cmd/server

# Build the builder
echo "Building builder..."
go build -o "$BUILD_DIR/builder" ./cmd/builder

# Function to test a client with specific configuration
test_client() {
    local protocols=$1
    local encryption=$2
    local output_file="$BUILD_DIR/client_${protocols// /_}_${encryption}"
    
    echo "Building client with protocols: $protocols, encryption: $encryption"
    "$BUILD_DIR/builder" -output "$output_file" -protocol "$protocols" -server "$SERVER_ADDR" -encryption "$encryption"
    
    if [ $? -ne 0 ]; then
        echo "Failed to build client with protocols: $protocols, encryption: $encryption"
        return 1
    fi
    
    echo "Successfully built client: $output_file"
    return 0
}

# Test single protocol clients with different encryption algorithms
for proto in "${PROTOCOLS[@]}"; do
    for enc in "${ENCRYPTION_ALGS[@]}"; do
        test_client "$proto" "$enc"
    done
done

# Test multi-protocol clients with different encryption algorithms
for enc in "${ENCRYPTION_ALGS[@]}"; do
    test_client "tcp,http" "$enc"
    test_client "tcp,websocket,dns" "$enc"
    test_client "tcp,http,websocket,dns,icmp" "$enc"
done

echo "All tests completed."
