# Cockpit Coder

A mobile-first coding terminal that connects to your development environment and coding assistant through QR code scanning, providing a terminal interface, code diff viewing, and command execution capabilities.

## Features

- **📱 Mobile-First Interface**: Native mobile app built with React Native and Expo
- **🔗 QR Code Connection**: Secure connection setup via QR code scanning
- **💻 Terminal Interface**: Full terminal access with xterm.js integration
- **📊 Code Diff Viewer**: Visualize and apply code changes
- **⚡ Command Execution**: Allow-listed command execution with security controls
- **🔄 Auto-Reconnection**: Automatic reconnection with session persistence
- **🔐 Security**: JWT authentication, CORS protection, and command allowlisting

## Architecture

```
Cockpit Coder/
├── frontend/           # React Native/Expo mobile app
│   ├── app/           # App screens and components
│   ├── components/    # Reusable UI components
│   ├── lib/          # Utilities and API clients
│   └── assets/       # Static assets (HTML, CSS, JS)
├── backend/           # Go server
│   ├── cmd/server/   # Application entry point
│   ├── internal/     # Core business logic
│   └── go.mod        # Go module definition
└── README.md         # This file
```

## Quick Start

### Prerequisites

- Node.js 18+ and npm
- Go 1.21+
- Expo CLI (`npm install -g expo-cli`)
- React Native CLI (`npm install -g react-native-cli`)
- Xcode (for iOS) or Android Studio (for Android)
- Physical device or emulator/simulator

### 1. Setup Frontend

```bash
# Clone the repository
git clone <repository-url>
cd cockpit-coder

# Install dependencies
cd frontend
npm install

# Install Expo Go app on your mobile device
# Search for "Expo Go" on App Store or Google Play
```

### 2. Setup Backend

```bash
# Open new terminal and navigate to backend
cd backend

# Install Go dependencies
make deps

# Build the server
make build

# Start the server
make run
```

### 3. Run Frontend Development Server

```bash
# In frontend directory
npx expo start

# Scan QR code with Expo Go app on your mobile device
# Or press 'a' to open in Android emulator or 'i' for iOS simulator
```

### 4. Connect to Backend

1. Open the app on your mobile device
2. Click "Connect to Coding Agent"
3. Scan the QR code displayed in your terminal (if running in development mode)
4. Or use the connection information provided by the backend server

## Development

### Frontend Development

```bash
cd frontend

# Start development server
npx expo start

# Start with specific platform
npx expo start --ios
npx expo start --android

# Build for production
eas build --platform ios
eas build --platform android
```

### Backend Development

```bash
cd backend

# Run tests
make test

# Format code
make fmt

# Run linter
make lint

# Build for production
make build

# Run with hot reload (requires air)
make dev
```

### Environment Variables

#### Frontend (.env)

```env
# API Configuration
API_BASE_URL=http://localhost:8080
WS_URL=ws://localhost:8080

# Development
EXPO_DEV_SERVER_URL=http://localhost:19006
```

#### Backend (.env)

```env
# Server Configuration
PORT=8080
JWT_SECRET=your_secure_jwt_secret_here

# Security
CORS_ORIGINS=http://localhost:19006,exp://localhost:19000
REPO_ALLOWLIST=/path/to/your/repo
CMD_ALLOWLIST=npm test,go test,git status,ls,cd,pwd

# URLs
WS_URL=ws://localhost:8080
API_BASE=http://localhost:8080
```

## Project Structure

### Frontend Structure

```
frontend/
├── app/                    # App screens and navigation
│   ├── _layout.tsx        # Root layout
│   ├── index.tsx          # Home screen
│   ├── screens/           # Feature screens
│   │   ├── QRConnectScreen.tsx
│   │   ├── TerminalScreen.tsx
│   │   ├── DiffsScreen.tsx
│   │   └── CommandsScreen.tsx
├── components/            # Reusable UI components
│   ├── ui/               # Base UI components
│   │   ├── Button.tsx
│   │   ├── Card.tsx
│   │   ├── Input.tsx
│   │   └── Tabs.tsx
│   ├── DiffList.tsx      # Code diff viewer
│   └── Terminal.tsx      # Terminal wrapper
├── lib/                  # Utilities and API clients
│   ├── api.ts           # HTTP API client
│   ├── storage.ts       # Secure storage
│   ├── wsClient.ts      # WebSocket client
│   └── utils.ts         # Helper functions
└── assets/              # Static assets
    └── terminal/        # xterm.js files
        ├── index.html
        ├── xterm.js
        └── xterm.css
```

### Backend Structure

```
backend/
├── cmd/server/          # Application entry point
│   └── main.go
├── internal/
│   ├── auth/           # JWT authentication
│   │   └── jwt.go
│   ├── httpserver/     # HTTP server and routes
│   │   └── server.go
│   ├── session/        # Session and task management
│   │   └── manager.go
│   └── pty/           # PTY operations
│       └── pty.go
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## API Reference

### Authentication

All API requests require a JWT token in the Authorization header:

```
Authorization: Bearer <token>
```

### Sessions

#### Create Session
```http
POST /api/session
Content-Type: application/json

{
  "repo": "/path/to/repository",
  "label": "My Session",
  "via": "mobile"
}
```

#### Get Session
```http
GET /api/session/{sessionId}
Authorization: Bearer <token>
```

### Tasks

#### Create Task
```http
POST /api/tasks
Authorization: Bearer <token>
Content-Type: application/json

{
  "instruction": "Fix the bug in user authentication",
  "branch": "main",
  "context": {
    "files": ["auth.js", "user.js"],
    "hints": "Check the login validation logic"
  },
  "agent": "cline"
}
```

#### Get Task
```http
GET /api/tasks/{taskId}
Authorization: Bearer <token>
```

#### Get Task Patches
```http
GET /api/tasks/{taskId}/patches
Authorization: Bearer <token>
```

#### Apply Task Patches
```http
POST /api/tasks/{taskId}/apply
Authorization: Bearer <token>
Content-Type: application/json

{
  "select": [
    {
      "file": "src/auth.js",
      "hunks": [0, 2]
    }
  ],
  "commitMessage": "Fix authentication bug"
}
```

### Commands

#### Execute Command
```http
POST /api/cmd
Authorization: Bearer <token>
Content-Type: application/json

{
  "cmd": "npm test",
  "cwd": "/path/to/repository",
  "timeoutMs": 300000
}
```

### WebSocket Endpoints

#### PTY Terminal
```
ws://localhost:8080/ws/pty?sessionId=<id>&token=<token>
```

#### Events
```
ws://localhost:8080/ws/events?sessionId=<id>&token=<token>
```

## Security Considerations

### Command Allowlisting

Only commands specified in `CMD_ALLOWLIST` environment variable can be executed. Examples:

```env
CMD_ALLOWLIST=npm test,go test,git status,ls,cd,pwd,cat,echo,mkdir,rmdir
```

### Repository Allowlisting

Only repositories specified in `REPO_ALLOWLIST` environment variable can be accessed:

```env
REPO_ALLOWLIST=/home/user/projects/myapp,/var/www/production
```

### JWT Security

- Use a strong, randomly generated JWT secret
- Set appropriate token expiration times
- Store secrets securely (not in code)

### CORS Configuration

Configure CORS origins appropriately for your deployment:

```env
CORS_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
```

## Deployment

### Frontend Deployment

#### Expo EAS Build

```bash
# Configure EAS
eas build:configure

# Build for iOS
eas build --platform ios

# Build for Android
eas build --platform android

# Submit to stores
eas submit --platform ios
eas submit --platform android
```

### Backend Deployment

#### Docker Deployment

```dockerfile
# Dockerfile
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

```bash
# Build and run
docker build -t cockpit-coder-backend .
docker run -p 8080:8080 cockpit-coder-backend
```

#### Production Environment Variables

```env
PORT=8080
JWT_SECRET=your_production_jwt_secret
CORS_ORIGINS=https://yourapp.com
REPO_ALLOWLIST=/opt/app,/var/www
CMD_ALLOWLIST=npm test,git status,ls
WS_URL=wss://yourapp.com
API_BASE=https://yourapp.com
```

## Troubleshooting

### Common Issues

#### Frontend Issues

1. **Metro bundler errors**: Clear cache with `npx expo start --clear`
2. **Missing dependencies**: Run `npm install` in frontend directory
3. **Expo Go connection issues**: Ensure device and computer are on same network

#### Backend Issues

1. **Port already in use**: Change `PORT` environment variable
2. **Import errors**: Run `make deps` to download dependencies
3. **Permission denied**: Ensure server has execute permissions

#### Connection Issues

1. **QR code not scanning**: Ensure backend is running and accessible
2. **WebSocket connection failed**: Check firewall settings
3. **Authentication errors**: Verify JWT secret and token validity

### Debug Mode

#### Frontend Debug

```bash
# Enable React Native debugging
npx expo start --dev-client

# Enable Metro bundler debugging
npx expo start --debug
```

#### Backend Debug

```bash
# Enable debug logging
DEBUG=true ./bin/server

# Enable verbose logging
LOG_LEVEL=debug ./bin/server
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`make test` for backend, `npm test` for frontend)
6. Run code quality checks (`make check` for backend, `npm run lint` for frontend)
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

### Development Guidelines

- Follow the existing code style
- Write comprehensive tests
- Update documentation for new features
- Use semantic versioning
- Include changelog entries for significant changes

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- 📧 Email: support@cockpitcoder.com
- 🐛 Issues: [GitHub Issues](https://github.com/your-repo/cockpit-coder/issues)
- 📖 Documentation: [Wiki](https://github.com/your-repo/cockpit-coder/wiki)

## Acknowledgments

- [Expo](https://expo.dev/) for the amazing React Native framework
- [xterm.js](https://xtermjs.org/) for the terminal emulator
- [React Native](https://reactnative.dev/) for the cross-platform mobile development
- [Go](https://golang.org/) for the backend server
