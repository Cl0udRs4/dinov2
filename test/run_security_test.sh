#!/bin/bash

# Run security features test
echo "Running security features test..."

# Set up environment
export GO111MODULE=on

# Run test
cd "$(dirname "$0")"
go test -v -run TestSecurityFeatures

echo "Test completed"
