.PHONY: build test lint clean install

# Build the shifter binary
build:
	go build -ldflags="-s -w" -o shifter .

# Run all tests
test:
	go test ./... -v

# Run tests with race detection
test-race:
	go test ./... -race -v

# Lint the codebase
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -f shifter
	go clean

# Install to $GOPATH/bin
install:
	go install .

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/shifter-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/shifter-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/shifter-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/shifter-windows-amd64.exe

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Full CI check
check: fmt vet lint test test-race build
	@echo "All checks passed!"
