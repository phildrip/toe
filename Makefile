.PHONY: build test test-integration clean

# Build the toe binary
build:
	go build -o toe

# Run all unit tests
test:
	go test ./...

# Run the integration test script
test-integration: build
	bash integration_test.sh

# Run all tests including integration tests
test-all: test test-integration

# Clean up build artifacts
clean:
	rm -f toe
	go clean
