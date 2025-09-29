NAME = server
BINARY = $(NAME)
MAIN = ./cmd/server/main.go

# Build the application
build:
	go build -o $(BINARY) $(MAIN)

# Run the application
run:
	go run $(MAIN)

# Install dependencies
install:
	go mod tidy

# Clean build files
clean:
	rm -f $(BINARY)

# Test the application
test:
	go test ./...

# Run with hot reload (requires air)
dev:
	air

# Build and run
be: build run

.PHONY: build run install clean test dev be
