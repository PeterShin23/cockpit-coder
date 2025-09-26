.PHONY: build run clean test deps

# Build the server
build:
	go build -o bin/server ./cmd/server

# Run the server
run: build
	./bin/server

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test ./...

# Download dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Install dependencies
install-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Development setup
setup: install-deps deps

# Build for different platforms
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/server-linux ./cmd/server

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o bin/server-darwin ./cmd/server

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/server.exe ./cmd/server

# Run with hot reload (requires air)
dev:
	air -c .air.toml

# Generate documentation
docs:
	godoc -http=:6060

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Profile CPU
profile-cpu:
	go test -cpuprofile=cpu.prof -bench=. ./...
	go tool pprof cpu.prof

# Profile memory
profile-mem:
	go test -memprofile=mem.prof -bench=. ./...
	go tool pprof mem.prof

# Clean dependencies
clean-deps:
	rm -rf go.sum
	go mod tidy

# Update dependencies
update-deps:
	go get -u ./...
	go mod tidy

# Security audit
audit:
	gosec ./...

# Generate coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Run all checks
check: fmt lint test
