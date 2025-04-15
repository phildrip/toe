# toe

#### trip-free test stubs

toe is a Go tool that automatically generates stub implementations for Go interfaces. It's useful
for creating test doubles in unit tests - but only stubs, not mocks or fakes.

## Features

- Generates stub implementations for any Go interface
- Records method calls and their parameters
- Allows setting up return values for stubbed methods
- Thread-safe

## Installation

To install toe, use the following command:

```bash
go get github.com/phildrip/toe
```

Or for the latest version:

```bash
go install github.com/phildrip/toe@latest
```

## Usage

```bash
toe -o <output.go> <input_directory> <interface> 
```

- `<input_directory>`: The directory containing the Go file with the interface definition
- `<interface>`: The name of the interface you want to generate a stub for
- `-o <output.go>`: (Optional) The output file name. If not provided, the stub code will be printed
  to stdout
- `-no-fmt`: (Optional) Disable formatting of the output

### Example

```bash
toe . Thinger -o stub_thinger.go
```

This command will generate a stub implementation for the Thinger interface defined in the current
directory and save it to stub_thinger.go.

## Generated Stub Structure

The generated stub includes:

- A struct to record method calls
- A main struct implementing the interface
- Methods to record calls and their parameters
- ThenReturn methods to set up return values
- Methods to retrieve recorded calls for assertions in tests

## Example Usage in Tests

```golang
stub := &StubThinger{}
stub.On().ThingThenReturn(nil)
stub.Thing()
stub.ThingWithParam(42)

// Assert on calls
calls := stub.ThingCalls()
if len(calls) != 1 {
t.Errorf("Expected 1 call to Thing(), got %d", len(calls))
}

paramCalls := stub.ThingWithParamCalls()
if len(paramCalls) != 1 || paramCalls[0].arg1 != 42 {
t.Errorf("Expected 1 call to ThingWithParam(42), got %+v", paramCalls)
}
```

## Why another generator?

toe keeps things super-simple. It doesn't try to support all the features of mocking libraries
like [gomock](https://github.com/golang/mock) or [pegomock](https://github.com/petergtz/pegomock).
It's just a simple tool that generates a stub implementation for a given interface.

By staying simple, toe can fulfill 95% of the use cases I've needed in unit and integration tests -
but makes those 95% of cases easy. toe doesn't allow you to specify different behaviour for
different method calls. It also doesn't support method chaining or chaining multiple calls together.
But it does make stubbing - the ability to define a return value given a method call, and recording
those calls - easy and quick.

If you need more complex behaviour, you can use a mocking library like gomock or pegomock.

## Building from Source

To build toe from source:

1. Clone the repository:

```bash
git clone https://github.com/phildrip/toe.git
```

2. Navigate to the project directory:

```bash
cd toe
```

3. Build the project:

```bash
go build
```

## Testing

To run the unit tests:

```bash
go test ./...
```

To run the integration tests:

```bash
# First build the binary
go build -o toe

# Then run the integration test script
./integration_test.sh
```

Or use the Makefile:

```bash
# Run all unit tests
make test

# Run integration tests
make test-integration

# Run all tests
make test-all
```

## Contributing

Contributions are welcome! Here are some ways you can contribute:

1. Report bugs by opening an issue
2. Suggest new features or improvements
3. Submit pull requests with bug fixes or new features
4. Improve documentation

Please make sure to run tests before submitting a pull request.

## License

MIT License

## Author

Phil Richards (https://github.com/phildrip)
