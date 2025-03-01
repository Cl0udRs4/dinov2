# DinoC2 Makefile

# Variables
BINARY_DIR = bin
SERVER_BINARY = $(BINARY_DIR)/server
CLIENT_BINARY = $(BINARY_DIR)/client
BUILDER_BINARY = $(BINARY_DIR)/builder

# Go build flags
GOFLAGS = -ldflags="-s -w"

.PHONY: all build clean test server client builder run-server run-client

all: build

build: server client builder

# Create binary directory
$(BINARY_DIR):
	mkdir -p $(BINARY_DIR)

# Build server
server: $(BINARY_DIR)
	go build $(GOFLAGS) -o $(SERVER_BINARY) ./cmd/server

# Build client
client: $(BINARY_DIR)
	go build $(GOFLAGS) -o $(CLIENT_BINARY) ./cmd/client

# Build builder
builder: $(BINARY_DIR)
	go build $(GOFLAGS) -o $(BUILDER_BINARY) ./cmd/builder

# Run server
run-server: server
	$(SERVER_BINARY) -protocol tcp -address 127.0.0.1:8080

# Run client
run-client: client
	$(CLIENT_BINARY) -protocol tcp -server 127.0.0.1:8080

# Run tests
test:
	go test -v ./...

# Run integration tests
integration-test: build
	cd test/integration && go test -v

# Clean
clean:
	rm -rf $(BINARY_DIR)
	find . -name "*.test" -delete
