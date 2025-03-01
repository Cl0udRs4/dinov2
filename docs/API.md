# DinoC2 HTTP API Documentation

This document describes the HTTP API endpoints provided by the DinoC2 server.

## Authentication

All API endpoints require authentication using a Bearer token in the Authorization header:

```
Authorization: Bearer <token>
```

## Endpoints

### Listeners

#### GET /api/v1/listeners

Returns a list of all listeners.

**Response:**

```json
{
  "success": true,
  "message": "Listeners retrieved successfully",
  "data": [
    {
      "id": "listener1",
      "type": "tcp",
      "address": "0.0.0.0",
      "port": 8080,
      "enabled": true,
      "status": "running"
    }
  ],
  "time": "2023-01-01T00:00:00Z"
}
```

#### POST /api/v1/listeners

Creates a new listener.

**Request:**

```json
{
  "id": "listener1",
  "type": "tcp",
  "address": "0.0.0.0",
  "port": 8080,
  "enabled": true
}
```

**Response:**

```json
{
  "success": true,
  "message": "Listener created successfully",
  "data": {
    "id": "listener1",
    "type": "tcp",
    "address": "0.0.0.0",
    "port": 8080,
    "enabled": true,
    "status": "running"
  },
  "time": "2023-01-01T00:00:00Z"
}
```

#### GET /api/v1/listeners/{id}

Returns a specific listener.

**Response:**

```json
{
  "success": true,
  "message": "Listener retrieved successfully",
  "data": {
    "id": "listener1",
    "type": "tcp",
    "address": "0.0.0.0",
    "port": 8080,
    "enabled": true,
    "status": "running"
  },
  "time": "2023-01-01T00:00:00Z"
}
```

#### PUT /api/v1/listeners/{id}

Updates a specific listener.

**Request:**

```json
{
  "id": "listener1",
  "type": "tcp",
  "address": "0.0.0.0",
  "port": 8080,
  "enabled": true
}
```

**Response:**

```json
{
  "success": true,
  "message": "Listener updated successfully",
  "data": {
    "id": "listener1",
    "type": "tcp",
    "address": "0.0.0.0",
    "port": 8080,
    "enabled": true,
    "status": "running"
  },
  "time": "2023-01-01T00:00:00Z"
}
```

#### DELETE /api/v1/listeners/{id}

Deletes a specific listener.

**Response:**

```json
{
  "success": true,
  "message": "Listener deleted successfully",
  "time": "2023-01-01T00:00:00Z"
}
```

### Clients

#### GET /api/v1/clients

Returns a list of all connected clients.

**Response:**

```json
{
  "success": true,
  "message": "Clients retrieved successfully",
  "data": [
    {
      "id": "client1",
      "ip": "192.168.1.100",
      "hostname": "workstation1",
      "os": "Windows 10",
      "connected": true
    }
  ],
  "time": "2023-01-01T00:00:00Z"
}
```

#### GET /api/v1/clients/{id}

Returns a specific client.

**Response:**

```json
{
  "success": true,
  "message": "Client retrieved successfully",
  "data": {
    "id": "client1",
    "ip": "192.168.1.100",
    "hostname": "workstation1",
    "os": "Windows 10",
    "connected": true
  },
  "time": "2023-01-01T00:00:00Z"
}
```

#### DELETE /api/v1/clients/{id}

Disconnects a specific client.

**Response:**

```json
{
  "success": true,
  "message": "Client disconnected successfully",
  "time": "2023-01-01T00:00:00Z"
}
```

#### POST /api/v1/clients/{id}/tasks

Sends a task to a specific client.

**Request:**

```json
{
  "module": "shell",
  "command": "whoami"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Task sent successfully",
  "data": {
    "task_id": "task1"
  },
  "time": "2023-01-01T00:00:00Z"
}
```

### Modules

#### GET /api/v1/modules

Returns a list of all available modules.

**Response:**

```json
{
  "success": true,
  "message": "Modules retrieved successfully",
  "data": [
    {
      "name": "shell",
      "description": "Provides shell access to the client",
      "version": "1.0.0"
    },
    {
      "name": "file",
      "description": "Provides file system access",
      "version": "1.0.0"
    }
  ],
  "time": "2023-01-01T00:00:00Z"
}
```

#### POST /api/v1/modules

Uploads a new module.

**Request:**

Multipart form with a file field named "module".

**Response:**

```json
{
  "success": true,
  "message": "Module uploaded successfully",
  "time": "2023-01-01T00:00:00Z"
}
```

#### GET /api/v1/modules/{name}

Returns a specific module.

**Response:**

```json
{
  "success": true,
  "message": "Module retrieved successfully",
  "data": {
    "name": "shell",
    "description": "Provides shell access to the client",
    "version": "1.0.0"
  },
  "time": "2023-01-01T00:00:00Z"
}
```

#### DELETE /api/v1/modules/{name}

Deletes a specific module.

**Response:**

```json
{
  "success": true,
  "message": "Module deleted successfully",
  "time": "2023-01-01T00:00:00Z"
}
```
