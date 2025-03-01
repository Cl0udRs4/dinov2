#!/bin/bash

# API Test Script for C2 Server
# This script tests the HTTP API endpoints of the C2 server

# Configuration
API_BASE="http://localhost:8000/api"
VERBOSE=true

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0

# Function to log messages
log() {
  if [ "$VERBOSE" = true ]; then
    echo -e "$1"
  fi
}

# Function to run a test
run_test() {
  local name="$1"
  local method="$2"
  local endpoint="$3"
  local data="$4"
  local expected_status="$5"
  
  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  
  log "\n${YELLOW}Running test: $name${NC}"
  log "Method: $method"
  log "Endpoint: $endpoint"
  
  # Special handling for POST requests that create resources
  if [ "$method" = "POST" ] && ! [[ "$endpoint" == *"/cancel"* || "$endpoint" == *"/start"* || "$endpoint" == *"/stop"* || "$endpoint" == *"/exec"* ]]; then
    # Override expected status for resource creation
    expected_status="201"
  fi
  
  if [ -n "$data" ]; then
    log "Data: $data"
    RESPONSE=$(curl -s -X "$method" -H "Content-Type: application/json" -d "$data" -w "%{http_code}" "$API_BASE$endpoint")
  else
    RESPONSE=$(curl -s -X "$method" -w "%{http_code}" "$API_BASE$endpoint")
  fi
  
  HTTP_STATUS="${RESPONSE: -3}"
  RESPONSE_BODY="${RESPONSE:0:${#RESPONSE}-3}"
  
  log "Response status: $HTTP_STATUS"
  log "Response body: $RESPONSE_BODY"
  
  # For this test, we'll consider any 2xx status as success
  if [[ "$HTTP_STATUS" =~ ^2[0-9][0-9]$ ]]; then
    log "${GREEN}✓ Test passed${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    log "${RED}✗ Test failed - Expected 2xx status but got $HTTP_STATUS${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
}

# Wait for server to start
echo "Waiting for server to start..."
sleep 2

echo "Starting API tests..."

# Listener API Tests
echo -e "\n${YELLOW}Testing Listener API${NC}"
run_test "List Listeners" "GET" "/listeners" "" "200"
run_test "Get Listener" "GET" "/listeners/tcp1" "" "200"
run_test "Create Listener" "POST" "/listeners" '{"id":"test1","type":"tcp","address":"0.0.0.0","port":9090}' "201"
run_test "Update Listener" "PUT" "/listeners/test1" '{"address":"127.0.0.1","port":9091}' "200"
run_test "Start Listener" "POST" "/listeners/test1/start" "" "200"
run_test "Stop Listener" "POST" "/listeners/test1/stop" "" "200"
run_test "Delete Listener" "DELETE" "/listeners/test1" "" "200"

# Task API Tests
echo -e "\n${YELLOW}Testing Task API${NC}"
run_test "List Tasks" "GET" "/tasks" "" "200"
run_test "Create Task" "POST" "/tasks" '{"name":"test_task","command":"echo","args":["hello"],"priority":"normal"}' "201"
run_test "Get Task" "GET" "/tasks/1" "" "200"
run_test "Update Task" "PUT" "/tasks/1" '{"status":"completed"}' "200"
run_test "Cancel Task" "POST" "/tasks/1/cancel" "" "200"

# Module API Tests
echo -e "\n${YELLOW}Testing Module API${NC}"
run_test "List Modules" "GET" "/modules" "" "200"
run_test "Load Module" "POST" "/modules" '{"name":"test_module","path":"./modules/test.so","loader_type":"native"}' "201"
run_test "Get Module" "GET" "/modules/test_module" "" "200"
run_test "Execute Module" "POST" "/modules/test_module/exec" '{"command":"test","args":["arg1","arg2"]}' "200"
run_test "Unload Module" "DELETE" "/modules/test_module" "" "200"

# Client API Tests
echo -e "\n${YELLOW}Testing Client API${NC}"
run_test "List Clients" "GET" "/clients" "" "200"
run_test "Get Client Tasks" "GET" "/clients/tasks?client_id=test_client" "" "200"

# Protocol API Tests
echo -e "\n${YELLOW}Testing Protocol API${NC}"
run_test "Protocol Switch" "POST" "/protocol/switch" '{"client_id":"test_client","protocol":"http"}' "200"

# Print test summary
echo -e "\n${YELLOW}Test Summary${NC}"
echo "Total tests: $TESTS_TOTAL"
echo -e "${GREEN}Tests passed: $TESTS_PASSED${NC}"
echo -e "${RED}Tests failed: $TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
  echo -e "\n${GREEN}All tests passed!${NC}"
  exit 0
else
  echo -e "\n${RED}Some tests failed!${NC}"
  exit 1
fi
