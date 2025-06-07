# The Go entry point files
MAIN_FILE := main.go
SERVER_FILE := server.go
CLIENT_FILE := client.go

# Default port for the server
PORT := 4433

# Default config file
CONFIG := config.json

.PHONY: all run-server run-client build clean

# Default target: show usage
all:
	@echo "Usage:"
	@echo "  make run-server    # Run the QUIC server"
	@echo "  make run-client    # Run the QUIC client"
	@echo "  make build         # (Optional) Compile all source files"
	@echo "  make clean         # Clean up build artifacts (if any)"

GO_FILES := main.go server.go client.go utils.go message.go state.go cert.go config.go

run-server:
	@echo "Starting server on localhost:4433..."
	RUN_MODE=server go run $(GO_FILES)

run-client:
	@echo "Starting client..."
	RUN_MODE=client go run $(GO_FILES)

build:
	@echo "Building project..."
	go build -o qtgp $(GO_FILES)

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f qtgp