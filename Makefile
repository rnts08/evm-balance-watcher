# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt ./...
GOVET=$(GOCMD) vet ./...
GORUN=$(GOCMD) run

# Binary Name
BINARY_NAME=evmbal

# Build flags
LDFLAGS = -ldflags="-s -w -buildid= -X main.Version=$(VERSION)"
BUILDFLAGS = -trimpath $(LDFLAGS)

# Release version - can be overridden e.g. `make release VERSION=v1.0.0`
VERSION ?= $(shell cat VERSION | tr -d '[:space:]')

.PHONY: all build run clean test unittest fmt vet lint help cross-compile release bump mobile-apk mobile-ipa mobile-setup

all: build

# Build the binary for the current system
build:
	@echo "Building $(BINARY_NAME) for $(shell go env GOOS)/$(shell go env GOARCH)..."
	@CGO_ENABLED=0 $(GOBUILD) $(BUILDFLAGS) -o $(BINARY_NAME) .

# Build binaries for multiple platforms
cross-compile:
	@echo "Cross-compiling for Linux, Windows, and macOS..."
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 $(GOBUILD) $(BUILDFLAGS) -o dist/$(BINARY_NAME)-windows-arm64.exe .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	@echo "Cross-compilation complete. Binaries are in dist/"

# Create compressed release archives
release: cross-compile
	@echo "Creating release archives for version $(VERSION)..."
	cd dist && \
		tar -czf $(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
		tar -czf $(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
		zip -q $(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe && \
		zip -q $(BINARY_NAME)-$(VERSION)-windows-arm64.zip $(BINARY_NAME)-windows-arm64.exe && \
		tar -czf $(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
		tar -czf $(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	@echo "Cleaning up raw binaries..."
	@rm dist/$(BINARY_NAME)-linux-amd64 dist/$(BINARY_NAME)-linux-arm64 dist/$(BINARY_NAME)-windows-amd64.exe dist/$(BINARY_NAME)-windows-arm64.exe dist/$(BINARY_NAME)-darwin-amd64 dist/$(BINARY_NAME)-darwin-arm64
	@echo "Release archives for version $(VERSION) are in dist/"

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Run the configuration test
test:
	@echo "Running configuration test..."
	@$(GORUN) . -test -config test_config.json

# Run unit tests
unittest:
	@echo "Running unit tests..."
	@$(GOTEST) ./...

# Bump version (usage: make bump part=patch)
bump:
	@chmod +x bump_version.sh
	@./bump_version.sh $(part)

# Clean the binary and dist folder
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -f $(BINARY_NAME)
	@rm -rf dist
	@rm -f *.bak
	@echo "Cleaned."

# Format the code
fmt:
	@echo "Formatting code..."
	@$(GOFMT)

# Vet the code
vet:
	@echo "Vetting code..."
	@$(GOVET)

# Lint the code
lint:
	@echo "Linting code..."
	@golangci-lint run

# Mobile builds (requires EAS CLI: npm install -g eas-cli)
mobile-setup:
	@echo "Installing mobile dependencies..."
	cd mobile-app && npm install
	@echo "Ensuring EAS CLI is available..."
	@npx eas-cli --version || (echo "EAS CLI not found. Please install it with: npm install -g eas-cli" && exit 1)

mobile-apk:
	@echo "Building Android APK locally for sideloading..."
	cd mobile-app && npx eas-cli build --platform android --local --profile preview

mobile-ipa:
	@echo "Building iOS IPA locally for sideloading..."
	@echo "Note: This requires a Mac with Xcode and appropriate Apple Developer credentials."
	cd mobile-app && npx eas-cli build --platform ios --local --profile preview

# Help
help:
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build          Build an optimized binary for the current system"
	@echo "  cross-compile  Build binaries for Linux, Windows, and macOS into ./dist"
	@echo "  release        Build and create compressed release archives in ./dist"
	@echo "  run            Build and run the application"
	@echo "  test           Run the configuration test mode"
	@echo "  unittest       Run unit tests"
	@echo "  bump part=...  Bump version (major, minor, patch), commit, and push tag"
	@echo "  clean          Remove the built binary and dist directory"
	@echo "  fmt            Format the source code"
	@echo "  vet            Run go vet to check for issues"
	@echo "  lint           Run golangci-lint to find issues"
	@echo "  help           Show this help message"
	@echo ""
