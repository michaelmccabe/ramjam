# ramjam

A  command-line tool for testing HTTP APIs, built with Go and using the Cobra framework.

## Overview

`ramjam` is a CLI tool designed to simplify HTTP API testing and interaction. It provides an intuitive interface for making HTTP requests, inspecting responses, and validating API behavior directly from your terminal.

## Features

* Simple and intuitive command-line interface
* Built with the Cobra CLI framework
* Configurable request timeouts
* Verbose mode for detailed request/response information
* Easy installation as a local binary

## Prerequisites

* Go 1.20 or higher
* Make (for using Makefile commands)

## Installation

### Quick Install

Install directly to your `$GOPATH/bin` or `$GOBIN`:

```bash
make install
```

This will compile and install the `ramjam` binary to your Go bin directory (typically `~/go/bin`). Make sure this directory is in your `PATH`:

```bash
# Add to your ~/.bashrc, ~/.zshrc, or equivalent
export PATH=$PATH:$(go env GOPATH)/bin
```

### Manual Installation




1. Clone the repository:

```bash
git clone https://github.com/michaelmccabe/ramjam.git
cd ramjam
```


2\. Build the binary:

```bash
make build
```


3\. (Optional) Move the binary to a location in your PATH:

```bash
sudo mv bin/ramjam /usr/local/bin/
# or
cp bin/ramjam ~/bin/  # if ~/bin is in your PATH
```

### Building from Source

```bash
# Build for current platform
make build

# Build for all platforms (Linux, macOS, Windows)
make build-all
```

The binary will be created in the `bin/` directory.

## Usage

For full details for how to use, see [How To Use Ramjam](./RAMJAM.md).

For details on integrating with CI/CD pipelines, see [CI/CD Integration](./INTEGRATE.md).

### Basic Commands

Display help and available commands:

```bash
ramjam --help
```

Check version:

```bash
ramjam version
```

### Making HTTP Requests

`ramjam` makes HTTP requests by running the workflows defined in the YAML files fed into the tool via the command line.

### Running YAML Workflows

Execute one or more workflow files or a directory of workflows:

```bash
ramjam run test-get.yaml
ramjam run ./tests/integration/
ramjam run login.yaml signup.yaml profile.yaml
```


You can try this out quickly yourself with the test files included

```bash
❯ ramjam run ./resources/testdata                                        
[simpleGetTests.yaml] Running workflow file: resources/testdata/simpleGetTests.yaml
[User Cross-Reference Validation] Successfully verified Clementine Bauch lives in McKenziehaven
[User Cross-Reference Validation] The first post title for user 3 is: asperiores ea ipsam voluptatibus modi minima quia sint
[FailingGetTests.yaml] Running workflow file: resources/testdata/FailingGetTests.yaml
[patchInputTest.yaml] Running workflow file: resources/testdata/patchInputTest.yaml
[postInpuTest.yaml] Running workflow file: resources/testdata/postInpuTest.yaml
[putInputTest.yaml] Running workflow file: resources/testdata/putInputTest.yaml
Failed step: get-specific-user
Failed step: validate-user-in-list
Failed step: fetch-user-posts
Error: workflow failed with 3 errors
```

### Global Flags

* `-v, --verbose`: Enable verbose output for detailed request/response information
* `-h, --help`: Display help information

## Development

### Project Structure

```
ramjam/
├── cmd/
│   └── ramjam/           # Main application entry point
│       ├── main.go       # Application entry
│       └── cmd/          # Cobra command definitions
│           ├── root.go   # Root command
│           ├── run.go    # Run command (executes workflows)
│           └── version.go # Version command
├── pkg/
│   ├── config/           # Configuration loading
│   └── runner/           # Workflow execution logic
├── resources/            # Test resources and examples
├── Makefile              # Build automation
├── go.mod                # Go module definition
├── INTEGRATE.md          # CI/CD integration guide
├── RAMJAM.md             # Usage documentation
└── README.md             # This file
```

### Building

```bash
# Build the project
make build

# Clean build artifacts
make clean

# Run tests
make test

# Run tests with coverage
make test-coverage

# Tidy dependencies
make tidy
```

### Running in Development

Run without building:

```bash
make run

# Or directly with go
go run ./cmd/ramjam
```

### Testing

Run all tests:

```bash
make test
```

Run tests with coverage report:

```bash
make test-coverage
```

## Makefile Targets

* `make build` - Build the binary for the current platform
* `make install` - Install the binary to `$GOPATH/bin`
* `make clean` - Remove build artifacts
* `make test` - Run all tests
* `make test-coverage` - Run tests with coverage report
* `make tidy` - Tidy Go module dependencies
* `make deps` - Download dependencies
* `make run` - Run the application without building
* `make build-all` - Build for multiple platforms (Linux, macOS, Windows)
* `make help` - Display available targets

## Configuration

Currently, `ramjam` uses command-line flags for configuration. Future versions may include support for configuration files.

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/michaelmccabe/ramjam).
