#!/bin/bash

# Build the builder
cd cmd/builder
go build -o builder

# Test AES encryption with TCP protocol
./builder -server "127.0.0.1:8080" -protocol "tcp" -encryption "aes" -output "client_tcp_aes"

# Test ChaCha20 encryption with TCP protocol
./builder -server "127.0.0.1:8080" -protocol "tcp" -encryption "chacha20" -output "client_tcp_chacha20"

# Test AES encryption with multiple protocols
./builder -server "127.0.0.1:8080" -protocol "tcp,http,websocket" -encryption "aes" -output "client_multi_aes"

# Test ChaCha20 encryption with multiple protocols
./builder -server "127.0.0.1:8080" -protocol "tcp,http,websocket" -encryption "chacha20" -output "client_multi_chacha20"

echo "Client generation tests completed."
