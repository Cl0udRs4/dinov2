# DinoC2 Installation Guide

This guide provides instructions for installing and setting up the DinoC2 system.

## Prerequisites

Before installing DinoC2, ensure your system meets the following requirements:

- Go 1.23.5 or higher
- Git
- Make
- GCC or equivalent C compiler (for CGO support)
- OpenSSL development libraries

### Installing Prerequisites on Ubuntu/Debian

```bash
sudo apt update
sudo apt install -y golang-1.23 git make gcc libssl-dev
```

### Installing Prerequisites on CentOS/RHEL

```bash
sudo yum install -y golang git make gcc openssl-devel
```

### Installing Prerequisites on macOS

```bash
brew install go git openssl
```

### Installing Prerequisites on Windows

1. Install Go from [golang.org](https://golang.org/dl/)
2. Install Git from [git-scm.com](https://git-scm.com/download/win)
3. Install MinGW-w64 for GCC support
4. Install OpenSSL for Windows

## Getting the Source Code

Clone the DinoC2 repository:

```bash
git clone https://github.com/Cl0udRs4/dinov2.git
cd dinov2
```

## Building from Source

DinoC2 uses a Makefile to simplify the build process:

### Building All Components

```bash
make all
```

This will build the server, client, and builder components.

### Building Individual Components

```bash
# Build server only
make server

# Build client only
make client

# Build builder only
make builder
```

### Cross-Compilation

To build for different platforms:

```bash
# Build Windows client
make client-windows

# Build Linux client
make client-linux

# Build macOS client
make client-macos
```

## Installation

After building, the binaries will be available in the `bin` directory:

```bash
# Install to /usr/local/bin (Linux/macOS)
sudo make install

# Or manually copy binaries
sudo cp bin/server /usr/local/bin/dinoc2-server
sudo cp bin/client /usr/local/bin/dinoc2-client
sudo cp bin/builder /usr/local/bin/dinoc2-builder
```

## Configuration

### Server Configuration

Create a server configuration file:

```bash
mkdir -p ~/.dinoc2
cp config/server.json.example ~/.dinoc2/server.json
```

Edit the configuration file to suit your needs:

```json
{
  "listeners": [
    {
      "type": "tcp",
      "address": "0.0.0.0:8080",
      "enabled": true
    },
    {
      "type": "http",
      "address": "0.0.0.0:8443",
      "enabled": true,
      "options": {
        "tls": true,
        "cert_file": "~/.dinoc2/certs/server.crt",
        "key_file": "~/.dinoc2/certs/server.key"
      }
    },
    {
      "type": "dns",
      "address": "0.0.0.0:53",
      "enabled": false,
      "options": {
        "domain": "c2.example.com"
      }
    },
    {
      "type": "icmp",
      "enabled": false
    }
  ],
  "encryption": {
    "default_algorithm": "chacha20",
    "key_rotation_interval": 3600
  },
  "modules": {
    "directory": "~/.dinoc2/modules",
    "signature_verification": true
  },
  "security": {
    "enable_authentication": true,
    "enable_signature_verification": true,
    "enable_traffic_obfuscation": true,
    "ca_cert": "~/.dinoc2/certs/ca.crt",
    "ca_key": "~/.dinoc2/certs/ca.key"
  },
  "logging": {
    "level": "info",
    "file": "~/.dinoc2/logs/server.log"
  }
}
```

### Certificate Generation

Generate certificates for secure communication:

```bash
mkdir -p ~/.dinoc2/certs
cd ~/.dinoc2/certs

# Generate CA certificate
openssl genrsa -out ca.key 4096
openssl req -new -x509 -key ca.key -out ca.crt -days 365 -subj "/CN=DinoC2 CA"

# Generate server certificate
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj "/CN=DinoC2 Server"
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365
```

## Running DinoC2

### Starting the Server

```bash
dinoc2-server -config ~/.dinoc2/server.json
```

Or if running from the source directory:

```bash
./bin/server -config ~/.dinoc2/server.json
```

### Building a Client

Use the builder to create a customized client:

```bash
dinoc2-builder -output client.exe -protocol tcp,http -mod shell,file,sysinfo
```

Or if running from the source directory:

```bash
./bin/builder -output client.exe -protocol tcp,http -mod shell,file,sysinfo
```

### Running a Client

```bash
./client.exe -server example.com:8080 -protocol tcp
```

## Verifying Installation

To verify that the installation is working correctly:

1. Start the server
2. Build and run a client
3. Check the server logs for client connection
4. Use the server console to interact with the client

## Troubleshooting

### Common Issues

1. **Connection Refused**:
   - Ensure the server is running
   - Check firewall settings
   - Verify the client is using the correct server address and port

2. **TLS Certificate Errors**:
   - Ensure certificates are properly generated
   - Check certificate paths in configuration
   - Verify certificate validity dates

3. **Module Loading Errors**:
   - Check module directory path
   - Ensure modules have correct permissions
   - Verify module signatures if signature verification is enabled

4. **Build Errors**:
   - Ensure Go version is 1.23.5 or higher
   - Check for missing dependencies
   - Verify CGO is enabled if required

### Getting Help

If you encounter issues not covered in this guide:

1. Check the logs for detailed error messages
2. Consult the troubleshooting section in the documentation
3. Open an issue on the GitHub repository

## Next Steps

After installation, refer to the following documentation:

- [User Guide](USER_GUIDE.md): Learn how to use DinoC2
- [Architecture](ARCHITECTURE.md): Understand the system architecture
- [Security](SECURITY.md): Learn about security features
- [Module Development](MODULE_DEVELOPMENT.md): Create custom modules
