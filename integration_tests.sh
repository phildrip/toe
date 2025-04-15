#!/bin/bash
set -e

# Build the toe binary
echo "Building toe..."
go build -o toe

# Create a temporary directory for test files
TEMP_DIR=$(mktemp -d)
echo "Using temporary directory: $TEMP_DIR"

# Clean up on exit
trap "rm -rf $TEMP_DIR; rm -f toe" EXIT

# Create a test file with an interface
cat > $TEMP_DIR/test.go << EOF
package testpkg

type TestInterface interface {
    DoSomething(input string) (string, error)
    DoSomethingElse(count int) error
}
EOF

# Generate a stub
OUTPUT_FILE=$TEMP_DIR/stub.go
echo "Generating stub..."
./toe -o $OUTPUT_FILE $TEMP_DIR TestInterface

# Check if the file was created
if [ ! -f "$OUTPUT_FILE" ]; then
    echo "ERROR: Output file was not created"
    exit 1
fi

echo "Checking generated stub..."
# Check if the file contains expected content
if ! grep -q "type StubTestInterface struct" "$OUTPUT_FILE"; then
    echo "ERROR: Generated stub doesn't contain expected struct definition"
    exit 1
fi

if ! grep -q "func (s \*StubTestInterface) DoSomething(input string) (string, error)" "$OUTPUT_FILE"; then
    echo "ERROR: Generated stub doesn't contain expected method DoSomething"
    exit 1
fi

if ! grep -q "func (s \*StubTestInterface) DoSomethingElse(count int) error" "$OUTPUT_FILE"; then
    echo "ERROR: Generated stub doesn't contain expected method DoSomethingElse"
    exit 1
fi

# Test with -no-fmt option
NO_FMT_OUTPUT=$TEMP_DIR/stub_nofmt.go
echo "Generating stub with -no-fmt option..."
./toe -no-fmt -o $NO_FMT_OUTPUT $TEMP_DIR TestInterface

if [ ! -f "$NO_FMT_OUTPUT" ]; then
    echo "ERROR: Output file with -no-fmt option was not created"
    exit 1
fi

# Test with empty interface
cat > $TEMP_DIR/empty_interface.go << EOF
package testpkg

type EmptyInterface interface {
}
EOF

EMPTY_OUTPUT=$TEMP_DIR/empty_stub.go
echo "Generating stub for empty interface..."
./toe -o $EMPTY_OUTPUT $TEMP_DIR EmptyInterface

if [ ! -f "$EMPTY_OUTPUT" ]; then
    echo "ERROR: Output file for empty interface was not created"
    exit 1
fi

# Test with complex types
cat > $TEMP_DIR/complex_interface.go << EOF
package testpkg

type ComplexInterface interface {
    HandleMap(data map[string]interface{}) map[int]string
    ProcessSlice(items []string, flags []bool) ([]int, error)
    WithPointers(p1 *string, p2 *int) *bool
}
EOF

COMPLEX_OUTPUT=$TEMP_DIR/complex_stub.go
echo "Generating stub for interface with complex types..."
./toe -o $COMPLEX_OUTPUT $TEMP_DIR ComplexInterface

if [ ! -f "$COMPLEX_OUTPUT" ]; then
    echo "ERROR: Output file for complex interface was not created"
    exit 1
fi

# Test output to stdout
echo "Testing output to stdout..."
STDOUT_OUTPUT=$(./toe $TEMP_DIR TestInterface)

if ! echo "$STDOUT_OUTPUT" | grep -q "type StubTestInterface struct"; then
    echo "ERROR: Stdout output doesn't contain expected struct definition"
    exit 1
fi

# Test with non-existent interface
echo "Testing with non-existent interface..."
if ./toe $TEMP_DIR NonExistentInterface > /dev/null 2>&1; then
    echo "ERROR: Command should fail with non-existent interface"
    exit 1
fi

echo "All tests passed!"
