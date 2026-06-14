.PHONY: build test lint clean install release build-all

# Version — override with: make release VERSION=0.2.0
VERSION ?= 0.1.0
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -Iseconds 2>/dev/null || date +%Y-%m-%dT%H:%M:%S%z)

LDFLAGS = -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

# Build the shifter binary (with version info)
build:
	go build -ldflags="$(LDFLAGS)" -o shifter .

# Run all tests
test:
	go test ./... -v

# Run tests with coverage
test-cover:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

# Run tests with race detection
test-race:
	go test ./... -race -v

# Run E2E tests
test-e2e:
	go test ./test/e2e/... -v

# Lint the codebase
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -f shifter
	go clean

# Install to $GOPATH/bin and ~/.local/bin
install: build
	go install -ldflags="$(LDFLAGS)" .
	@if [ -d "$(HOME)/.local/bin" ]; then \
		cp -f shifter "$(HOME)/.local/bin/shifter" && \
		echo "✓ Copied to ~/.local/bin/shifter"; \
	fi
	@echo "✓ Installed shifter v$(VERSION)"

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Build for all platforms
build-all:
	@mkdir -p dist
	GOOS=linux   GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/shifter-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/shifter-linux-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/shifter-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/shifter-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/shifter-windows-amd64.exe
	@echo "✓ Built all platforms in dist/"

# Create release archives
release: build-all
	@echo "Creating release archives for v$(VERSION)..."
	@for binary in dist/shifter-*; do \
		platform=$$(echo $$binary | sed 's/dist\/shifter-//'); \
		if echo $$platform | grep -q '.exe$$'; then \
			zip "dist/shifter_$(VERSION)_$${platform%.exe}.zip" "$$binary" && \
			echo "  ✓ dist/shifter_$(VERSION)_$${platform%.exe}.zip"; \
		else \
			tar -czf "dist/shifter_$(VERSION)_$$platform.tar.gz" -C dist $$(basename $$binary) && \
			echo "  ✓ dist/shifter_$(VERSION)_$$platform.tar.gz"; \
		fi; \
	done
	@echo ""
	@echo "Release v$(VERSION) ready in dist/"

# Show current version
version:
	@echo "v$(VERSION) (commit: $(COMMIT), date: $(DATE))"

# Full CI check
check: fmt vet test build
	@echo "All checks passed!"
