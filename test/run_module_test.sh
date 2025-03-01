#!/bin/bash

# Run module system test
echo "Running module system test..."

# Set up environment
export GO111MODULE=on

# Run test
cd "$(dirname "$0")"
go test -v -run TestModuleSystem

echo "Test completed"
