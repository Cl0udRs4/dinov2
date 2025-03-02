#!/bin/bash

# Test script for module operations

# Set up variables
API_URL="http://127.0.0.1:8443/api"
AUTH_TOKEN=""
CLIENT_ID=""

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

# Function to list clients
list_clients() {
    echo "Listing clients..."
    
    local response=$(curl -s -X GET "$API_URL/clients" \
        -H "Authorization: Bearer $AUTH_TOKEN")
    
    echo "$response"
    
    # Extract the first client ID if available
    CLIENT_ID=$(echo "$response" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
    
    if [ -n "$CLIENT_ID" ]; then
        echo "Found client ID: $CLIENT_ID"
    else
        echo "No clients found"
    fi
}

# Function to list modules
list_modules() {
    echo "Listing modules..."
    
    local response=$(curl -s -X GET "$API_URL/modules" \
        -H "Authorization: Bearer $AUTH_TOKEN")
    
    echo "$response"
}

# Function to load a module
load_module() {
    local name=$1
    local path=$2
    local loader_type=$3
    
    echo "Loading module $name..."
    
    local response=$(curl -s -X POST "$API_URL/modules/load" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -d "{\"name\":\"$name\",\"path\":\"$path\",\"loader_type\":\"$loader_type\"}")
    
    echo "$response"
}

# Function to deploy a module to a client
deploy_module() {
    local client_id=$1
    local module_name=$2
    local module_path=$3
    local loader_type=$4
    
    echo "Deploying module $module_name to client $client_id..."
    
    local response=$(curl -s -X POST "$API_URL/clients/modules/deploy" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -d "{\"client_id\":\"$client_id\",\"module_name\":\"$module_name\",\"module_path\":\"$module_path\",\"loader_type\":\"$loader_type\"}")
    
    echo "$response"
}

# Function to execute a module command
exec_module() {
    local name=$1
    local command=$2
    shift 2
    local args=("$@")
    
    # Convert args array to JSON array
    local args_json="["
    for arg in "${args[@]}"; do
        args_json+="\"$arg\","
    done
    args_json=${args_json%,}"]"
    
    echo "Executing command $command on module $name..."
    
    local response=$(curl -s -X POST "$API_URL/modules/exec" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -d "{\"name\":\"$name\",\"command\":\"$command\",\"args\":$args_json}")
    
    echo "$response"
}

# Function to execute a module command on a client
exec_client_module() {
    local client_id=$1
    local module_name=$2
    local command=$3
    shift 3
    local args=("$@")
    
    # Convert args array to JSON array
    local args_json="["
    for arg in "${args[@]}"; do
        args_json+="\"$arg\","
    done
    args_json=${args_json%,}"]"
    
    echo "Executing command $command on module $module_name for client $client_id..."
    
    local response=$(curl -s -X POST "$API_URL/clients/modules/exec" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -d "{\"client_id\":\"$client_id\",\"module_name\":\"$module_name\",\"command\":\"$command\",\"args\":$args_json}")
    
    echo "$response"
}

# Function to get client tasks
get_client_tasks() {
    local client_id=$1
    
    echo "Getting tasks for client $client_id..."
    
    local response=$(curl -s -X GET "$API_URL/clients/tasks?client_id=$client_id" \
        -H "Authorization: Bearer $AUTH_TOKEN")
    
    echo "$response"
}

# Function to unload a module
unload_module() {
    local name=$1
    
    echo "Unloading module $name..."
    
    local response=$(curl -s -X POST "$API_URL/modules/unload" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -d "{\"name\":\"$name\"}")
    
    echo "$response"
}

# Function to unload a module from a client
unload_client_module() {
    local client_id=$1
    local module_name=$2
    
    echo "Unloading module $module_name from client $client_id..."
    
    local response=$(curl -s -X POST "$API_URL/clients/modules/unload" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -d "{\"client_id\":\"$client_id\",\"module_name\":\"$module_name\"}")
    
    echo "$response"
}

# Main test sequence
echo "Starting module tests..."

# Get authentication token
get_auth_token

# List clients
list_clients

# Check if we have a client to work with
if [ -z "$CLIENT_ID" ]; then
    echo "No clients available. Please start a client and try again."
    exit 1
fi

# List modules
list_modules

# Test 1: Load a module on the server
echo "=== Test 1: Load a module on the server ==="
load_module "shell" "shell" "native"

# Verify module was loaded
list_modules

# Test 2: Deploy a module to a client
echo "=== Test 2: Deploy a module to a client ==="
deploy_module "$CLIENT_ID" "shell" "shell" "native"

# Test 3: Execute a module command on the server
echo "=== Test 3: Execute a module command on the server ==="
exec_module "shell" "execute" "ls" "-la"

# Test 4: Execute a module command on a client
echo "=== Test 4: Execute a module command on a client ==="
exec_client_module "$CLIENT_ID" "shell" "execute" "ls" "-la"

# Get client tasks to verify execution
get_client_tasks "$CLIENT_ID"

# Test 5: Unload a module from a client
echo "=== Test 5: Unload a module from a client ==="
unload_client_module "$CLIENT_ID" "shell"

# Test 6: Unload a module from the server
echo "=== Test 6: Unload a module from the server ==="
unload_module "shell"

# Verify module was unloaded
list_modules

echo "Module tests completed."
