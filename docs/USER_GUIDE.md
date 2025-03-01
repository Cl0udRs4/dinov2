# DinoC2 User Guide

This guide provides instructions for using the DinoC2 system, including server operation, client management, and module execution.

## Table of Contents

1. [Server Operation](#server-operation)
2. [Client Management](#client-management)
3. [Listener Management](#listener-management)
4. [Module Management](#module-management)
5. [Protocol Switching](#protocol-switching)
6. [Security Features](#security-features)
7. [Advanced Usage](#advanced-usage)

## Server Operation

### Starting the Server

Start the server with a configuration file:

```bash
dinoc2-server -config /path/to/config.json
```

Command-line options:

- `-config`: Path to configuration file
- `-verbose`: Enable verbose logging
- `-console`: Enable interactive console
- `-port`: Override listener port (overrides config file)

### Server Console

The server console provides an interactive interface for managing the server:

```
DinoC2 Server Console
Type 'help' for a list of commands

server>
```

Common console commands:

- `help`: Display help information
- `clients`: List connected clients
- `listeners`: List active listeners
- `modules`: List available modules
- `exit`: Exit the server

### Client Interaction

To interact with a connected client:

```
server> clients
ID        | IP Address    | Hostname    | OS          | Last Seen
--------------------------------------------------------------
c1a2b3d4  | 192.168.1.100 | WORKSTATION | Windows 10  | 2023-01-01 12:34:56
e5f6g7h8  | 10.0.0.5      | LAPTOP      | Ubuntu 22.04 | 2023-01-01 12:35:12

server> use c1a2b3d4
[client:c1a2b3d4]> help
```

Client commands:

- `info`: Display detailed client information
- `shell`: Execute shell commands
- `upload`: Upload a file to the client
- `download`: Download a file from the client
- `screenshot`: Capture a screenshot
- `processes`: List running processes
- `modules`: List available modules for this client
- `run`: Run a module
- `back`: Return to server console

## Client Management

### Client Information

View detailed information about a client:

```
[client:c1a2b3d4]> info
Client ID: c1a2b3d4
IP Address: 192.168.1.100
Hostname: WORKSTATION
OS: Windows 10 Pro (Build 19044)
Architecture: amd64
Username: Administrator
Domain: WORKGROUP
Uptime: 3 days, 5 hours
Connection Time: 2023-01-01 12:34:56
Last Seen: 2023-01-01 12:36:45
Active Protocols: TCP, HTTP
Available Modules: shell, file, sysinfo, screenshot, keylogger, process
```

### Shell Access

Execute shell commands on the client:

```
[client:c1a2b3d4]> shell
[shell:c1a2b3d4]> whoami
WORKSTATION\Administrator

[shell:c1a2b3d4]> dir
 Volume in drive C has no label.
 Volume Serial Number is 1234-5678

 Directory of C:\Users\Administrator

01/01/2023  12:00 PM    <DIR>          .
01/01/2023  12:00 PM    <DIR>          ..
01/01/2023  12:00 PM    <DIR>          Desktop
01/01/2023  12:00 PM    <DIR>          Documents
01/01/2023  12:00 PM    <DIR>          Downloads
...

[shell:c1a2b3d4]> exit
[client:c1a2b3d4]>
```

### File Operations

Upload and download files:

```
[client:c1a2b3d4]> upload /path/to/local/file.txt C:\temp\file.txt
[+] Uploading file.txt to C:\temp\file.txt
[+] Upload complete: 1024 bytes transferred

[client:c1a2b3d4]> download C:\Windows\System32\drivers\etc\hosts /tmp/hosts
[+] Downloading C:\Windows\System32\drivers\etc\hosts to /tmp/hosts
[+] Download complete: 824 bytes transferred
```

### Process Management

View and manage processes:

```
[client:c1a2b3d4]> processes
PID     | Name           | User          | Memory
------------------------------------------------------
4       | System         | SYSTEM        | 148 KB
504     | svchost.exe    | SYSTEM        | 14.2 MB
1234    | explorer.exe   | Administrator | 76.8 MB
...

[client:c1a2b3d4]> kill 1234
[+] Process 1234 (explorer.exe) terminated successfully
```

## Listener Management

### Listing Listeners

List active listeners:

```
server> listeners
ID      | Type    | Address       | Status  | Clients
------------------------------------------------------
tcp1    | TCP     | 0.0.0.0:8080  | Running | 2
http1   | HTTP    | 0.0.0.0:8443  | Running | 1
dns1    | DNS     | 0.0.0.0:53    | Stopped | 0
icmp1   | ICMP    | 0.0.0.0       | Stopped | 0
```

### Managing Listeners

Create, start, stop, and remove listeners:

```
server> listener create tcp 0.0.0.0:9090
[+] Created TCP listener 'tcp2' on 0.0.0.0:9090

server> listener start tcp2
[+] Started listener 'tcp2'

server> listener stop http1
[+] Stopped listener 'http1'

server> listener remove dns1
[+] Removed listener 'dns1'
```

### Listener Configuration

Configure listener options:

```
server> listener config http1 set tls true
[+] Set 'tls' to 'true' for listener 'http1'

server> listener config http1 set cert_file /path/to/cert.pem
[+] Set 'cert_file' to '/path/to/cert.pem' for listener 'http1'

server> listener config http1 set key_file /path/to/key.pem
[+] Set 'key_file' to '/path/to/key.pem' for listener 'http1'

server> listener config http1 show
Type: HTTP
Address: 0.0.0.0:8443
Status: Stopped
Options:
  - tls: true
  - cert_file: /path/to/cert.pem
  - key_file: /path/to/key.pem
```

## Module Management

### Listing Modules

List available modules:

```
server> modules
Name        | Description                       | Version
----------------------------------------------------------------
shell       | Execute shell commands            | 1.0.0
file        | File system operations            | 1.0.0
sysinfo     | System information gathering      | 1.0.0
screenshot  | Screen capture                    | 1.0.0
keylogger   | Keyboard monitoring               | 1.0.0
process     | Process management                | 1.0.0
```

### Running Modules

Run a module on a client:

```
[client:c1a2b3d4]> run sysinfo
[+] Running module 'sysinfo' on client 'c1a2b3d4'
[+] Module execution complete

System Information:
- Hostname: WORKSTATION
- OS: Windows 10 Pro (Build 19044)
- Architecture: amd64
- CPU: Intel(R) Core(TM) i7-10700K @ 3.80GHz (8 cores, 16 threads)
- Memory: 32.0 GB (16.2 GB available)
- Disk: C:\ 500.0 GB (325.8 GB available)
- Network Interfaces:
  - Ethernet: 192.168.1.100/24
  - Loopback: 127.0.0.1/8
```

### Module Options

Set module options:

```
[client:c1a2b3d4]> run screenshot -h
Module: screenshot
Description: Screen capture
Options:
  - quality: Image quality (1-100, default: 80)
  - format: Image format (png, jpeg, default: png)
  - display: Display number (default: 0)

[client:c1a2b3d4]> run screenshot quality=90 format=jpeg
[+] Running module 'screenshot' on client 'c1a2b3d4'
[+] Module execution complete
[+] Screenshot saved to /tmp/screenshot_c1a2b3d4_20230101123845.jpeg
```

## Protocol Switching

### Manual Protocol Switching

Switch a client's communication protocol:

```
[client:c1a2b3d4]> protocols
Active Protocol: TCP
Available Protocols: TCP, HTTP, DNS

[client:c1a2b3d4]> switch http
[+] Switching client 'c1a2b3d4' to HTTP protocol
[+] Protocol switch successful
```

### Automatic Protocol Switching

Configure automatic protocol switching:

```
[client:c1a2b3d4]> autoswitch enable
[+] Enabled automatic protocol switching for client 'c1a2b3d4'

[client:c1a2b3d4]> autoswitch add dns 300
[+] Added DNS protocol with 300 second timeout to autoswitch rotation

[client:c1a2b3d4]> autoswitch list
Automatic Protocol Switching: Enabled
Rotation:
  1. TCP (current) - 600 seconds
  2. HTTP - 600 seconds
  3. DNS - 300 seconds

[client:c1a2b3d4]> autoswitch remove http
[+] Removed HTTP protocol from autoswitch rotation
```

## Security Features

### Traffic Obfuscation

Configure traffic obfuscation:

```
server> security obfuscation profiles
Available Profiles:
  - http
  - dns
  - tls

server> security obfuscation set http
[+] Set traffic obfuscation profile to 'http'

server> security obfuscation config http
Profile: http
Options:
  - jitter_enabled: true
  - jitter_min: 100ms
  - jitter_max: 500ms
  - padding_enabled: true
  - padding_min: 16
  - padding_max: 64

server> security obfuscation config http set jitter_max 1000
[+] Set 'jitter_max' to '1000' for profile 'http'
```

### Authentication

Manage authentication:

```
server> security auth list
Client Certificates:
  - client1 (expires: 2024-01-01)
  - client2 (expires: 2024-01-01)

server> security auth generate client3
[+] Generated certificate for 'client3'
[+] Certificate saved to ~/.dinoc2/certs/client3.crt
[+] Private key saved to ~/.dinoc2/certs/client3.key

server> security auth revoke client2
[+] Revoked certificate for 'client2'
```

### Integrity Checking

Configure integrity checking:

```
server> security integrity status
Integrity Checking: Enabled
Critical Files: 5
Last Check: 2023-01-01 12:30:45
Violations: 0

server> security integrity add /path/to/critical/file
[+] Added '/path/to/critical/file' to integrity database

server> security integrity check
[+] Running integrity check
[+] Integrity check complete: No violations detected
```

## Advanced Usage

### Scripting

Create and run scripts:

```
server> script create reconnect.txt
Enter script (end with EOF on a new line):
clients
sleep 5
use c1a2b3d4
switch http
sleep 2
switch tcp
EOF
[+] Script saved to ~/.dinoc2/scripts/reconnect.txt

server> script run reconnect.txt
[+] Running script 'reconnect.txt'
...
[+] Script execution complete
```

### Logging

Configure logging:

```
server> logging level set debug
[+] Set logging level to 'debug'

server> logging file set /path/to/logfile.log
[+] Set log file to '/path/to/logfile.log'

server> logging status
Level: debug
File: /path/to/logfile.log
Console: enabled
```

### Exporting Data

Export client data:

```
server> export clients /path/to/clients.json
[+] Exported client data to '/path/to/clients.json'

server> export results c1a2b3d4 /path/to/results.json
[+] Exported results for client 'c1a2b3d4' to '/path/to/results.json'
```

### Batch Commands

Execute batch commands:

```
server> batch c1a2b3d4,e5f6g7h8 run sysinfo
[+] Running module 'sysinfo' on 2 clients
[+] Batch execution complete
```

## Conclusion

This guide covers the basic usage of the DinoC2 system. For more detailed information, refer to the following documentation:

- [Installation Guide](INSTALLATION.md): Instructions for installing and setting up DinoC2
- [Architecture](ARCHITECTURE.md): Overview of the system architecture
- [Security](SECURITY.md): Details on security features
- [Module Development](MODULE_DEVELOPMENT.md): Guide for developing custom modules
