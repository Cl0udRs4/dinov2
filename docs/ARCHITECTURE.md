# DinoC2 System Architecture

## Overview

DinoC2 is a modular Command and Control (C2) framework designed for red team operations. It provides a flexible, secure, and extensible platform for managing client connections, executing modules, and maintaining persistent access to target systems. This document outlines the system architecture, component interactions, and design principles.

## System Components

The DinoC2 system consists of three main components:

1. **Server**: Central command and control node that manages listeners, client connections, and module execution
2. **Client**: Lightweight agent that establishes and maintains connection with the server, executing modules as directed
3. **Builder**: Tool for generating customized client binaries with specified protocols and embedded modules

### High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                           SERVER                                │
│                                                                 │
│  ┌───────────┐    ┌───────────┐    ┌───────────────────────┐   │
│  │ Listener  │    │ Protocol  │    │ Module Management     │   │
│  │ Manager   │◄───┤ Layer     │◄───┤ - Registry            │   │
│  │           │    │           │    │ - Loader              │   │
│  └───────────┘    └───────────┘    │ - Execution           │   │
│        ▲                ▲          └───────────────────────┘   │
│        │                │                      ▲                │
│  ┌───────────┐    ┌───────────┐               │                │
│  │ Listener  │    │ Crypto    │    ┌───────────────────────┐   │
│  │ Instances │◄───┤ Module    │    │ Security              │   │
│  │ TCP/DNS/  │    │           │    │ - Authentication      │   │
│  │ ICMP/HTTP │    └───────────┘    │ - Signature Verif.    │   │
│  └───────────┘                     │ - Traffic Obfuscation │   │
│                                    └───────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                           ▲
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                          CLIENT                                 │
│                                                                 │
│  ┌───────────┐    ┌───────────┐    ┌───────────────────────┐   │
│  │ Connection│    │ Protocol  │    │ Module Execution      │   │
│  │ Manager   │◄───┤ Layer     │◄───┤ Environment           │   │
│  │           │    │           │    │                       │   │
│  └───────────┘    └───────────┘    └───────────────────────┘   │
│        ▲                ▲                      ▲                │
│        │                │                      │                │
│  ┌───────────┐    ┌───────────┐    ┌───────────────────────┐   │
│  │ Connection│    │ Crypto    │    │ Security              │   │
│  │ Types     │◄───┤ Module    │    │ - Anti-Debug          │   │
│  │ TCP/DNS/  │    │           │    │ - Anti-Sandbox        │   │
│  │ ICMP/HTTP │    └───────────┘    │ - Memory Protection   │   │
│  └───────────┘                     └───────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                           ▲
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                          BUILDER                                │
│                                                                 │
│  ┌───────────┐    ┌───────────┐    ┌───────────────────────┐   │
│  │ Config    │    │ Protocol  │    │ Module Embedding      │   │
│  │ Parser    │◄───┤ Selection │◄───┤ System                │   │
│  │           │    │           │    │                       │   │
│  └───────────┘    └───────────┘    └───────────────────────┘   │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │ Build System                                              │ │
│  │ - Cross-Platform Compilation                              │ │
│  │ - Binary Hardening                                        │ │
│  │ - Obfuscation                                             │ │
│  └───────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Component Details

### Server Component

The server is the central management component of the DinoC2 system, responsible for:

1. **Listener Management**:
   - Dynamic creation and management of protocol listeners (TCP, DNS, ICMP, HTTP)
   - Status monitoring and control of listener instances
   - Protocol-specific configuration and optimization

2. **Client Management**:
   - Client registration and authentication
   - Connection state tracking
   - Client capability discovery

3. **Module Management**:
   - Module registry for available modules
   - Module loading and execution
   - Module result collection and processing

4. **Protocol Layer**:
   - Unified protocol handling across different transport mechanisms
   - Message encoding and decoding
   - Traffic obfuscation

5. **Crypto Module**:
   - Encryption algorithm management (ChaCha20, AES)
   - Key exchange and rotation (ECDHE)
   - Secure session management

6. **Security Features**:
   - Certificate-based authentication
   - Module signature verification
   - Traffic obfuscation profiles

#### Server Directory Structure

```
pkg/
├── listener/
│   ├── manager.go       # Listener management
│   ├── adapter.go       # Common listener interface
│   ├── factory.go       # Listener creation factory
│   ├── tcp.go           # TCP listener implementation
│   ├── dns/             # DNS listener implementation
│   ├── http/            # HTTP listener implementation
│   └── icmp/            # ICMP listener implementation
├── protocol/
│   ├── packet.go        # Packet structure definitions
│   ├── encoder.go       # Message encoding/decoding
│   ├── handler.go       # Protocol message handling
│   └── obfuscator.go    # Traffic obfuscation
├── crypto/
│   ├── crypto.go        # Crypto interface
│   ├── aes.go           # AES implementation
│   ├── chacha20.go      # ChaCha20 implementation
│   ├── ecdhe.go         # Key exchange
│   └── session.go       # Session management
├── module/
│   ├── module.go        # Module interface
│   ├── registry.go      # Module registration
│   ├── manager/         # Module execution management
│   ├── loader/          # Module loading mechanisms
│   └── [module types]/  # Specific module implementations
├── security/
│   ├── security.go      # Security interface
│   ├── authentication.go # Authentication system
│   ├── signature.go     # Signature verification
│   └── traffic_obfuscation.go # Traffic obfuscation
└── task/
    └── manager.go       # Task scheduling and management
```

### Client Component

The client is a lightweight agent designed to establish and maintain connection with the server:

1. **Connection Management**:
   - Protocol selection and switching
   - Connection establishment and maintenance
   - Heartbeat and reconnection logic

2. **Module Execution**:
   - Module loading and execution
   - Result collection and reporting
   - Resource management

3. **Protocol Layer**:
   - Message encoding and decoding
   - Protocol-specific handling

4. **Security Features**:
   - Anti-debugging mechanisms
   - Anti-sandbox techniques
   - Memory protection
   - Traffic obfuscation

#### Client Directory Structure

```
pkg/
├── client/
│   ├── client.go        # Main client implementation
│   ├── connection.go    # Connection interface
│   ├── tcp_connection.go # TCP connection implementation
│   ├── dns_connection.go # DNS connection implementation
│   ├── http_connection.go # HTTP connection implementation
│   ├── icmp_connection.go # ICMP connection implementation
│   └── security.go      # Client-side security features
├── protocol/
│   ├── packet.go        # Packet structure definitions
│   ├── encoder.go       # Message encoding/decoding
│   └── handler.go       # Protocol message handling
├── crypto/
│   ├── crypto.go        # Crypto interface
│   ├── aes.go           # AES implementation
│   ├── chacha20.go      # ChaCha20 implementation
│   └── ecdhe.go         # Key exchange
└── security/
    ├── anti_debug.go    # Anti-debugging mechanisms
    ├── anti_sandbox.go  # Anti-sandbox techniques
    ├── memory_protection.go # Memory protection
    └── traffic_obfuscation.go # Traffic obfuscation
```

### Builder Component

The builder generates customized client binaries:

1. **Configuration Parsing**:
   - Command-line argument parsing
   - Configuration validation

2. **Protocol Selection**:
   - Protocol configuration
   - Protocol switching strategy setup

3. **Module Embedding**:
   - Module selection and embedding
   - Module configuration

4. **Build System**:
   - Cross-platform compilation
   - Binary hardening
   - Obfuscation techniques

#### Builder Directory Structure

```
cmd/
└── builder/
    ├── main.go          # Builder entry point
    └── builder.go       # Builder implementation

pkg/
└── module/
    └── builder/
        └── builder.go   # Module embedding system
```

## Data Flow

### Client Registration Process

1. Client initiates connection to server using configured protocol
2. Server and client perform mutual authentication
3. Client and server perform ECDHE key exchange
4. Client sends registration information (system info, capabilities)
5. Server registers client and assigns unique identifier
6. Server sends acknowledgment to client
7. Client enters main communication loop

### Command Execution Flow

1. Server creates task for client
2. Server encodes and encrypts task
3. Server sends task to client through established protocol
4. Client receives, decrypts, and decodes task
5. Client executes task or loads and executes module
6. Client collects results
7. Client encodes and encrypts results
8. Client sends results to server
9. Server receives, decrypts, and decodes results
10. Server processes and stores results

### Protocol Switching Flow

1. Server decides to switch protocol (or client initiates switch due to timeout)
2. Server sends protocol switch command to client
3. Client acknowledges command
4. Client prepares new connection using specified protocol
5. Client establishes new connection
6. Server and client perform authentication on new connection
7. Server transfers session state to new connection
8. Old connection is terminated

## Module System

The module system is designed to be flexible and extensible:

1. **Module Interface**:
   - Standard interface for all modules
   - Lifecycle methods (Init, Exec, Shutdown)
   - Capability reporting

2. **Module Loading Mechanisms**:
   - Native modules (built-in)
   - Plugin modules (dynamic loading)
   - WebAssembly modules
   - DLL modules (Windows)
   - RPC-based modules

3. **Module Isolation**:
   - Resource limitation
   - Execution timeouts
   - Crash recovery

4. **Module Registry**:
   - Module discovery
   - Version management
   - Dependency resolution

### Standard Modules

1. **Shell Module**: Command execution
2. **File Module**: File system operations
3. **Process Module**: Process management
4. **Screenshot Module**: Screen capture
5. **Keylogger Module**: Keyboard monitoring
6. **Sysinfo Module**: System information gathering

## Security Architecture

The security architecture is designed to protect both the client and server:

1. **Communication Security**:
   - End-to-end encryption
   - Perfect forward secrecy
   - Traffic obfuscation

2. **Authentication**:
   - Certificate-based mutual authentication
   - Pre-shared key support

3. **Client Protection**:
   - Anti-debugging mechanisms
   - Anti-sandbox techniques
   - Memory protection
   - Integrity checking

4. **Server Protection**:
   - Module signature verification
   - Client validation
   - Access control

## Design Principles

1. **Modularity**: Components are designed to be modular and replaceable
2. **Protocol Agnosticism**: Core functionality is independent of transport protocol
3. **Security by Design**: Security is integrated at all levels
4. **Extensibility**: System is designed to be easily extended with new modules and protocols
5. **Resilience**: System can recover from connection failures and protocol blocking

## Conclusion

The DinoC2 architecture provides a flexible, secure, and extensible framework for command and control operations. Its modular design allows for easy customization and extension, while its security features protect against detection and analysis.
