# 🏋🏽ramjam

A  command-line tool for testing HTTP APIs, built with Go and using the Cobra framework.

## Overview

`ramjam` is a CLI tool designed to simplify HTTP API testing and interaction. It provides an intuitive interface for making HTTP requests, inspecting responses, and validating API behavior directly from your terminal.

## Latest Releases

The latest binary releases for `ramjam` can be found [here](https://github.com/michaelmccabe/ramjam/releases)

## Documentation

| Document | Description |
|----|----|
| [How To Use Ramjam](./RAMJAM.md) | Complete workflow DSL reference, variable substitution, authentication patterns, and full examples |
| [CI/CD Integration](./INTEGRATE.md) | Guide for integrating ramjam into GitHub Actions, GitLab CI, and other CI/CD pipelines |
| [Body File Feature](./BODY_FILE_FEATURE.md) | Using external JSON files for request bodies |

## Features

* Simple and intuitive command-line interface
* Built with the Cobra CLI framework
* Configurable request timeouts
* Verbose mode for detailed request/response information
* Easy installation as a local binary
* Load request bodies from external JSON files via `body_file`

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

For details on using JSON files for the body of requests see  [body file feature](./BODY_FILE_FEASTURE.md).

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

### Loading Request Bodies from JSON Files

Payloads can be kept in standalone JSON files and referenced with the `body_file` keyword (see [Body File Feature](./BODY_FILE_FEATURE.md)).

### Running YAML Workflows

Execute one or more workflow files or a directory of workflows:

```bash
ramjam run test-get.yaml
ramjam run ./tests/integration/
ramjam run login.yaml signup.yaml profile.yaml
```


You can try this out quickly yourself with the test files included

```bash
❯ ramjam run resources/testdata/success             
[patchInputTest.yaml] Running workflow file: resources/testdata/success/patchInputTest.yaml
[postInpuTest.yaml] Running workflow file: resources/testdata/success/postInpuTest.yaml
[Complex POST Integration] Successfully created post from external JSON file
[putInputTest.yaml] Running workflow file: resources/testdata/success/putInputTest.yaml
[bodyFileDemo.yaml] Running workflow file: resources/testdata/success/bodyFileDemo.yaml
[Body File Feature Demo] ✓ Created post using inline body
[Body File Feature Demo] ✓ Created post using external JSON file
[Body File Feature Demo] ✓ Captured user: Leanne Graham (Sincere@april.biz)
[Body File Feature Demo] ✓ Updated user profile using JSON file with variables
[simpleGetTests.yaml] Running workflow file: resources/testdata/success/simpleGetTests.yaml
[User Cross-Reference Validation] Successfully verified Clementine Bauch lives in McKenziehaven with cache max-age 43200
[User Cross-Reference Validation] The first post title for user 3 is: asperiores ea ipsam voluptatibus modi minima quia sint
All steps were run successfully


❯ ramjam run resources/testdata/fail   
[FailingGetTests.yaml] Running workflow file: resources/testdata/fail/FailingGetTests.yaml
Failed step: get-specific-user
Failed step: validate-user-in-list
Failed step: fetch-user-posts
Error: workflow failed with 3 errors
Usage:
  ramjam run <files-or-folders...> [flags]

Flags:
  -h, --help   help for run

Global Flags:
  -v, --verbose   Enable verbose output

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

## Creating Releases

Releases are automated via GitHub Actions. When you push a version tag, the workflow builds binaries for multiple platforms and creates a GitHub release with all assets attached.

### Creating a New Release



1. **Update version** (optional): Edit the default version in `cmd/ramjam/cmd/root.go` if desired
2. **Commit your changes**:

   ```bash
   git add .
   git commit -m "Prepare release v1.0.0"
   ```
3. **Create and push a version tag**:

   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
4. The GitHub Actions workflow will automatically:
   * Run all tests
   * Build binaries for Linux, macOS, and Windows (both AMD64 and ARM64)
   * Generate SHA256 checksums
   * Create a GitHub release with all binaries attached

### Release Assets

Each release includes:

* `ramjam-linux-amd64` - Linux (Intel/AMD)
* `ramjam-linux-arm64` - Linux (ARM)
* `ramjam-darwin-amd64` - macOS (Intel)
* `ramjam-darwin-arm64` - macOS (Apple Silicon)
* `ramjam-windows-amd64.exe` - Windows (Intel/AMD)
* `ramjam-windows-arm64.exe` - Windows (ARM)
* `checksums.txt` - SHA256 checksums for verification

### Pre-release Versions

Tags containing `-alpha`, `-beta`, or `-rc` are automatically marked as pre-releases:

```bash
git tag v1.0.0-beta.1
git push origin v1.0.0-beta.1
```

### Version Information

The version displayed by `ramjam version` is determined at build time:

* **Release builds**: Version comes from the git tag (e.g., `v1.0.0`)
* **Local builds via Make**: Version comes from `git describe` (e.g., `v1.0.0-5-g2a3b4c5`)
* **Direct** `go build`: Uses the default value in `root.go`

To build locally with a specific version:

```bash
VERSION=1.0.0 make build
```

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/michaelmccabe/ramjam).
