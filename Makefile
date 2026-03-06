APP_NAME := morphr
VERSION  := 0.1.0

.PHONY: build install clean run test bench lint fmt release-local tag-release

# Build the binary
build:
	go build -ldflags "-s -w" -o $(APP_NAME) .

# Install globally
install:
	go install .

# Clean build artifacts
clean:
	rm -f $(APP_NAME)
	rm -rf bin/ dist/

# Run with sample input
run:
	go run . convert --help

# Run all tests
test:
	go test ./... -v -timeout 120s

# Run benchmarks
bench:
	go test ./internal/encoder/ -bench=. -benchtime=3x -timeout 120s
	go test ./internal/processor/ -bench=. -benchtime=3x -timeout 120s

# Lint
lint:
	golangci-lint run ./...

# Format
fmt:
	gofmt -s -w .
	goimports -w .

# Test GoReleaser config locally (no publish)
release-local:
	goreleaser release --snapshot --clean

# Tag and push a release (triggers GitHub Actions)
# Usage: make tag-release V=0.1.0
tag-release:
	@if [ -z "$(V)" ]; then echo "Usage: make tag-release V=0.1.0"; exit 1; fi
	git tag -a v$(V) -m "Release v$(V)"
	git push origin v$(V)
