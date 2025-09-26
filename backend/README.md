# Cockpit Coder Backend

A Go backend server for the Cockpit Coder application, providing session management, PTY handling, and WebSocket support for real-time terminal operations.

## Features

- **Session Management**: JWT-based authentication and session handling
- **PTY Operations**: Pseudo-terminal support for interactive shell sessions
- **WebSocket Support**: Real-time bidirectional communication
- **Task Management**: Create and manage coding tasks
- **Command Execution**: Allow-listed command execution with security controls
- **CORS Support**: Configurable cross-origin resource sharing

## Project Structure

```
backend/
├── cmd/server/          # Application entry point
├── internal/
│   ├── auth/           # JWT authentication
│   ├── httpserver/     # HTTP server and routes
│   ├── session/        # Session and task management
│   └── pty/           # PTY operations
├── go.mod             # Go module definition
├── go.sum             # Dependency checksums
├── Makefile           # Build automation
└── README.md          # This file
```

## Setup

### Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd backend
```

2. Install dependencies:
```bash
make deps
```

3. Build the application:
```bash
make build
```

### Configuration

The application can be configured through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `JWT_SECRET` | JWT signing secret | `change_me` |
| `CORS_ORIGINS` | Comma-separated allowed origins | `http://localhost:19006` |
| `REPO_ALLOWLIST` | Comma-separated allowed repository paths | `` |
| `CMD_ALLOWLIST` | Comma-separated allowed commands | `` |
| `WS_URL` | WebSocket URL for client connections | `ws://localhost:8080` |
| `API_BASE` | Base URL for API endpoints | `http://localhost:8080` |

### Running

#### Using Makefile (recommended)

```bash
# Build and run
make run

# Run development mode with hot reload
make dev

# Run tests
make test

# Format code
make fmt

# Run linter
make lint
```

#### Using Go directly

```bash
# Build
go build -o bin/server ./cmd/server

# Run
./bin/server

# Run with environment variables
PORT=8080 JWT_SECRET=your_secret ./bin/server
```

## API Endpoints

### Sessions

- `POST /api/session` - Create a new session
- `GET /api/session/{id}` - Get session information

### Tasks

- `POST /api/tasks` - Create a new task
- `GET /api/tasks/{id}` - Get task information
- `GET /api/tasks/{id}/patches` - Get task patches
- `POST /api/tasks/{id}/apply` - Apply task patches

### Commands

- `POST /api/cmd` - Execute a command

### WebSocket

- `ws://localhost:8080/ws/pty` - PTY terminal connection
- `ws://localhost:8080/ws/events` - Event notifications

## Security

### Authentication

All API endpoints require authentication via JWT tokens in the `Authorization` header:

```
Authorization: Bearer <token>
```

### Command Allowlisting

Only commands specified in the `CMD_ALLOWLIST` environment variable can be executed. If not set, no commands are allowed.

### Repository Allowlisting

Only repositories specified in the `REPO_ALLOWLIST` environment variable can be accessed. If not set, no repositories are allowed.

### CORS

Cross-origin requests are only allowed from origins specified in the `CORS_ORIGINS` environment variable.

## Development

### Adding New Endpoints

1. Add the route in `internal/httpserver/server.go`
2. Implement the handler function
3. Add authentication checks where needed
4. Update API documentation

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run specific test
go test ./internal/session -v
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run security audit
make audit
```

## Deployment

### Building for Production

```bash
# Build for current platform
make build

# Build for multiple platforms
make build-all
```

### Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

### Environment Configuration

Production deployments should set the following environment variables:

```bash
PORT=8080
JWT_SECRET=<secure-secret>
CORS_ORIGINS=https://yourdomain.com
REPO_ALLOWLIST=/path/to/repo1,/path/to/repo2
CMD_ALLOWLIST=npm test,go test,git status
WS_URL=wss://yourdomain.com
API_BASE=https://yourdomain.com
```

## Troubleshooting

### Common Issues

1. **Port already in use**: Change the `PORT` environment variable
2. **Permission denied**: Ensure the server has execute permissions
3. **CORS errors**: Verify `CORS_ORIGINS` includes your frontend domain
4. **Authentication failures**: Check JWT secret and token expiration

### Debug Mode

Set `DEBUG=true` environment variable for verbose logging:

```bash
DEBUG=true ./bin/server
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make check` to ensure code quality
6. Submit a pull request

## License

This project is licensed under the MIT License.
