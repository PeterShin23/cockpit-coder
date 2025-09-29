# Cockpit Coder Backend

## Overview

The backend server for Cockpit Coder, providing HTTP and WebSocket APIs for AI-assisted coding tasks.

## Features

- HTTP REST API for session management, tasks, commands, and git operations
- WebSocket endpoints for real-time PTY streaming and event notifications
- JWT-based authentication
- Pluggable agent system with mock implementation
- Policy-based security with repository and command allowlists
- In-memory session management with TTL

## API Endpoints

### Health Check
- `GET /healthz` - Server health status

### Session Management
- `POST /api/session` - Create new session
- `GET /api/session/{id}` - Get session details

### Task Management
- `POST /api/tasks` - Start new task
- `GET /api/tasks/{id}` - Get task status
- `GET /api/tasks/{id}/patches` - Get task patches
- `POST /api/tasks/{id}/apply` - Apply task patches

### Command Execution
- `POST /api/cmd` - Execute command

### Git Operations
- `GET /api/git/diff` - Get git diff

### WebSocket Endpoints
- `GET /ws/pty` - PTY streaming
- `GET /ws/events` - Event notifications

## Environment Variables

```bash
PORT=8080
JWT_SECRET=change_me
SESSION_TTL_SECONDS=86400
REPO_ALLOWLIST=/abs/path/repo1,/abs/path/repo2
CMD_ALLOWLIST="npm test,go test,npm run build,pytest"
CORS_ORIGINS=http://localhost:19006
```

## Development

### Prerequisites

- Go 1.21+

### Build and Run

```bash
# Build the server
make build

# Run the server
make run

# Or use the combined command
make be
```

### Manual Test Sequence

```bash
# Start the server
make be

# Test health endpoint
curl http://localhost:8080/healthz

# Create session
curl -X POST http://localhost:8080/api/session \
  -H "Authorization: Bearer DEV" \
  -d '{"repo":"/abs/your/repo","via":"direct"}'

# Test WebSocket endpoints with wscat
# wscat -c ws://localhost:8080/ws/pty?sessionId=...&token=...
# wscat -c ws://localhost:8080/ws/events?sessionId=...&token=...

# Start task
curl -X POST http://localhost:8080/api/tasks \
  -H "Authorization: Bearer <token>" \
  -d '{"instruction":"test task","agent":"cline"}'

# Get patches
curl -X GET http://localhost:8080/api/tasks/{id}/patches \
  -H "Authorization: Bearer <token>"
```

## Architecture

The backend is structured into the following packages:

- `cmd/server` - Main application entry point
- `internal/agents` - AI agent implementations and factory
- `internal/auth` - Authentication and JWT handling
- `internal/cmdexec` - Command execution with policy enforcement
- `internal/events` - Event bus for pub/sub messaging
- `internal/git` - Git operations (diff, apply)
- `internal/httpserver` - HTTP server and routing
- `internal/policy` - Security policies and validation
- `internal/pty` - PTY management for terminal streaming
- `internal/session` - Session and task management

## Security

- JWT-based authentication for all API endpoints
- Repository path allowlist validation
- Command execution allowlist
- CORS policy enforcement
- Context-based timeouts and cancellation
