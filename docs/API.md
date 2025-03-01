# DinoC2 API Documentation

## Overview

The DinoC2 API provides a RESTful interface for managing the C2 server, including listeners, clients, tasks, and modules. This document provides detailed information about the available endpoints, request/response formats, and authentication.

## Authentication

The API uses JWT (JSON Web Token) for authentication. To access protected endpoints, you need to include a valid JWT token in the Authorization header:

```
Authorization: Bearer <token>
```

### Obtaining a Token

To obtain a token, send a POST request to the `/api/auth/login` endpoint with your credentials:

```
POST /api/auth/login
Content-Type: application/json

{
  "username": "your_username",
  "password": "your_password"
}
```

The response will include a JWT token:

```
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Refreshing a Token

To refresh an existing token, send a POST request to the `/api/auth/refresh` endpoint with your current token in the Authorization header:

```
POST /api/auth/refresh
Authorization: Bearer <current_token>
```

The response will include a new JWT token:

```
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## API Endpoints

### Listeners

#### List Listeners

```
GET /api/listeners
```

Returns a list of all listeners.

#### Create Listener

```
POST /api/listeners/create
Content-Type: application/json

{
  "id": "http1",
  "type": "http",
  "address": "0.0.0.0",
  "port": 8000,
  "options": {
    "use_http2": true,
    "enable_api": true
  }
}
```

Creates a new listener.

#### Delete Listener

```
DELETE /api/listeners/delete
Content-Type: application/json

{
  "id": "http1"
}
```

Deletes a listener.

#### Get Listener Status

```
GET /api/listeners/status?id=http1
```

Returns the status of a listener.

### Tasks

#### List Tasks

```
GET /api/tasks
```

Returns a list of all tasks.

#### Create Task

```
POST /api/tasks/create
Content-Type: application/json

{
  "client_id": "client1",
  "type": "shell",
  "data": "ls -la"
}
```

Creates a new task.

#### Get Task Status

```
GET /api/tasks/status?id=task1
```

Returns the status of a task.

### Modules

#### List Modules

```
GET /api/modules
```

Returns a list of all modules.

#### Load Module

```
POST /api/modules/load
Content-Type: application/json

{
  "name": "shell",
  "path": "/path/to/module"
}
```

Loads a module.

#### Execute Module

```
POST /api/modules/exec
Content-Type: application/json

{
  "name": "shell",
  "params": {
    "command": "ls -la"
  }
}
```

Executes a module.

### Clients

#### List Clients

```
GET /api/clients
```

Returns a list of all clients.

#### Get Client Tasks

```
GET /api/clients/tasks?id=client1
```

Returns the tasks for a client.

### Protocol

#### Switch Protocol

```
POST /api/protocol/switch
Content-Type: application/json

{
  "client_id": "client1",
  "protocol": "http"
}
```

Switches the protocol for a client.

## Configuration

The API can be configured in the server configuration file:

```json
{
  "api": {
    "enabled": true,
    "address": "127.0.0.1",
    "port": 8443,
    "tls_enabled": true,
    "tls_cert_file": "/path/to/cert.pem",
    "tls_key_file": "/path/to/key.pem",
    "auth_enabled": true,
    "jwt_secret": "your_secret_key",
    "token_expiry": 60
  }
}
```

- `enabled`: Whether the API is enabled
- `address`: The address to bind the API server to
- `port`: The port to listen on
- `tls_enabled`: Whether to use TLS
- `tls_cert_file`: The path to the TLS certificate file
- `tls_key_file`: The path to the TLS key file
- `auth_enabled`: Whether authentication is enabled
- `jwt_secret`: The secret key for JWT token generation
- `token_expiry`: The token expiry time in minutes
