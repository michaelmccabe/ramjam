# Creating the ramjam CLI Project

This document provides a detailed, step-by-step guide on how the ramjam CLI project was created using Go and the Cobra framework.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Project Initialization](#project-initialization)
3. [Setting Up Go Modules and Workspace](#setting-up-go-modules-and-workspace)
4. [Creating Project Structure](#creating-project-structure)
5. [Installing Dependencies](#installing-dependencies)
6. [Implementing CLI Commands](#implementing-cli-commands)
7. [Creating the Makefile](#creating-the-makefile)
8. [Configuring Git Ignore](#configuring-git-ignore)
9. [Writing Documentation](#writing-documentation)
10. [Building and Testing](#building-and-testing)
11. [Installation and Usage](#installation-and-usage)

## Prerequisites

Before starting, ensure you have the following installed:

- **Go 1.20 or higher**: Download from [golang.org](https://golang.org/dl/)
- **Git**: For version control
- **Make**: For build automation (usually pre-installed on Unix systems)

Verify installations:
```bash
go version
git --version
make --version
```

## Project Initialization

### Step 1: Create Project Directory

While the repository may already exist, we'll work within it:

```bash
cd /path/to/ramjam
```

### Step 2: Initialize Git (if not already done)

```bash
git init
git remote add origin https://github.com/michaelmccabe/ramjam.git
```

## Setting Up Go Modules and Workspace

### Step 3: Initialize Go Module

Go modules are the standard way to manage dependencies in Go projects.

```bash
go mod init github.com/michaelmccabe/ramjam
```

This creates a `go.mod` file that tracks your project's dependencies.

**What this does:**
- Creates a `go.mod` file with the module path
- Enables dependency management
- Sets up the project as a Go module

### Step 4: Create Go Workspace

Go workspaces allow you to work with multiple modules simultaneously.

```bash
go work init .
```

This creates a `go.work` file that references your module.

**What this does:**
- Creates a `go.work` file
- Adds the current directory as a workspace module
- Allows for multi-module development

**Note:** The `go.work` and `go.work.sum` files should be in `.gitignore` as they're local development files.

## Creating Project Structure

### Step 5: Create Standard Directory Layout

Following Go's standard project layout, create the following directories:

```bash
mkdir -p cmd/ramjam/cmd    # Main application and commands
mkdir -p pkg/api           # Reusable packages
```

**Directory Structure Explanation:**

```
ramjam/
├── cmd/                    # Command-line applications
│   └── ramjam/            # Main CLI application
│       ├── main.go        # Application entry point
│       └── cmd/           # Cobra command implementations
│           ├── root.go    # Root command
│           ├── get.go     # GET command
│           └── version.go # Version command
├── pkg/                   # Public library code
│   └── api/              # API client package
│       └── client.go     # HTTP client
├── go.mod                # Go module definition
├── go.work               # Go workspace file
├── Makefile              # Build automation
├── .gitignore            # Git ignore patterns
└── README.md             # Project documentation
```

**Why this structure?**
- `cmd/` - Contains main applications for the project
- `pkg/` - Contains library code that can be used by external applications
- Clear separation of concerns

## Installing Dependencies

### Step 6: Install Cobra Framework

Cobra is a powerful library for creating CLI applications in Go.

```bash
go get -u github.com/spf13/cobra@latest
```

**What this does:**
- Downloads and installs the Cobra library
- Updates `go.mod` with the dependency
- Makes Cobra available for import in your code

**Cobra Features:**
- Easy command definition
- Automatic help generation
- Flag parsing (global and local)
- Nested subcommands
- Intelligent suggestions

## Implementing CLI Commands

### Step 7: Create Main Entry Point

Create `cmd/ramjam/main.go`:

```go
package main

import (
	"github.com/michaelmccabe/ramjam/cmd/ramjam/cmd"
)

func main() {
	cmd.Execute()
}
```

**Purpose:**
- Application entry point
- Delegates to Cobra command structure
- Keeps main minimal and focused

### Step 8: Implement Root Command

Create `cmd/ramjam/cmd/root.go`:

This file contains:
- Root command definition
- Global flags
- Application metadata (name, description, version)
- Execute function for running commands

**Key Components:**
- `rootCmd`: The main command that runs when you type `ramjam`
- `Execute()`: Entry point for the Cobra command tree
- `init()`: Setup function for flags and configuration

### Step 9: Implement Subcommands

#### GET Command (`cmd/ramjam/cmd/get.go`)

Purpose: Send HTTP GET requests to APIs

Features:
- URL validation
- Timeout configuration
- Custom headers support
- Verbose output mode
- Response display

#### Version Command (`cmd/ramjam/cmd/version.go`)

Purpose: Display application version

Features:
- Version information display
- Can be updated via build flags

### Step 10: Create Reusable Packages

Create `pkg/api/client.go`:

Purpose: HTTP client abstraction

Features:
- Configurable base URL
- Timeout management
- Verbose logging
- Reusable across commands

**Why separate packages?**
- Promotes code reuse
- Makes testing easier
- Allows external projects to use your code
- Better organization

## Creating the Makefile

### Step 11: Write Build Automation

Create a `Makefile` with the following targets:

**Essential Targets:**

1. **build**: Compile the binary
   ```bash
   make build
   ```
   - Creates `bin/` directory
   - Builds binary with version info
   - Platform-specific output

2. **install**: Install to system
   ```bash
   make install
   ```
   - Installs to `$GOPATH/bin`
   - Makes binary globally accessible
   - Includes version information

3. **clean**: Remove artifacts
   ```bash
   make clean
   ```
   - Removes `bin/` directory
   - Cleans Go cache
   - Prepares for fresh build

4. **test**: Run tests
   ```bash
   make test
   ```
   - Runs all tests
   - Verbose output
   - Quick feedback

5. **test-coverage**: Generate coverage report
   ```bash
   make test-coverage
   ```
   - Runs tests with coverage
   - Generates HTML report
   - Helps identify untested code

6. **tidy**: Clean dependencies
   ```bash
   make tidy
   ```
   - Removes unused dependencies
   - Updates `go.mod` and `go.sum`
   - Keeps dependencies clean

7. **build-all**: Multi-platform build
   ```bash
   make build-all
   ```
   - Builds for Linux, macOS, Windows
   - Multiple architectures
   - Distribution ready

**Makefile Benefits:**
- Consistent build process
- Easy to remember commands
- Cross-platform compatibility
- Automation of common tasks
- Version management via ldflags

## Configuring Git Ignore

### Step 12: Update .gitignore

Add the following patterns:

```gitignore
# Binary output
bin/
ramjam
ramjam-*

# Test binaries
*.test

# Coverage files
*.out
coverage.*
*.coverprofile

# Go workspace (local development)
go.work
go.work.sum

# OS files
.DS_Store
```

**Why ignore these?**
- `bin/` - Build artifacts, not source
- Binary files - Generated, not authored
- `go.work*` - Local development configuration
- Coverage reports - Generated from tests
- Test binaries - Temporary files

## Writing Documentation

### Step 13: Create README.md

A good README includes:

1. **Project Overview**: What is it?
2. **Features**: What can it do?
3. **Installation**: How to install?
4. **Usage**: How to use it?
5. **Development**: How to contribute?
6. **Project Structure**: How is it organized?
7. **Examples**: Real-world usage

### Step 14: Create This Documentation

Document the creation process for:
- Future reference
- Team onboarding
- Learning purposes
- Process improvement

## Building and Testing

### Step 15: Build the Project

```bash
# Build for current platform
make build

# Output: bin/ramjam
```

**What happens during build:**
1. Go compiler reads source files
2. Resolves dependencies from `go.mod`
3. Compiles code to machine code
4. Links into single binary
5. Embeds version information via ldflags
6. Outputs to `bin/` directory

### Step 16: Test the Binary

```bash
# Test help output
./bin/ramjam --help

# Test version
./bin/ramjam version

# Test GET command
./bin/ramjam get https://api.github.com/zen

# Test with verbose flag
./bin/ramjam get https://api.github.com/zen -v
```

## Installation and Usage

### Step 17: Install Globally

```bash
make install
```

**What this does:**
1. Compiles the binary with version info
2. Installs to `$GOPATH/bin/ramjam`
3. Makes `ramjam` available system-wide

**Verify installation:**
```bash
which ramjam
ramjam version
```

### Step 18: Ensure PATH is Configured

If `ramjam` is not found, add Go's bin directory to PATH:

```bash
# For bash (~/.bashrc or ~/.bash_profile)
export PATH=$PATH:$(go env GOPATH)/bin

# For zsh (~/.zshrc)
export PATH=$PATH:$(go env GOPATH)/bin

# Reload shell configuration
source ~/.bashrc  # or ~/.zshrc
```

### Step 19: Use the CLI Tool

```bash
# Get help
ramjam --help

# Check version
ramjam version

# Make a GET request
ramjam get https://api.github.com/users/octocat

# With verbose output
ramjam get https://api.github.com/users/octocat -v

# With custom timeout
ramjam get https://slow-api.com/data --timeout 60
```

## Advanced Topics

### Adding New Commands

To add a new command (e.g., `post`):

1. Create `cmd/ramjam/cmd/post.go`
2. Define the command using Cobra
3. Add command-specific flags
4. Implement the RunE function
5. Register in `init()` with `rootCmd.AddCommand(postCmd)`

Example structure:
```go
var postCmd = &cobra.Command{
    Use:   "post [url]",
    Short: "Send a POST request",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        // Implementation
        return nil
    },
}

func init() {
    rootCmd.AddCommand(postCmd)
    // Add flags
}
```

### Version Management

Version is set during build via ldflags:

```bash
# Build with specific version
go build -ldflags "-X github.com/michaelmccabe/ramjam/cmd/ramjam/cmd.Version=1.0.0" ./cmd/ramjam
```

The Makefile handles this automatically:
```makefile
LDFLAGS=-ldflags "-X github.com/michaelmccabe/ramjam/cmd/ramjam/cmd.Version=$(VERSION)"
```

Build with custom version:
```bash
VERSION=1.2.3 make build
```

### Cross-Platform Building

Build for different platforms:

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o bin/ramjam-linux-amd64 ./cmd/ramjam

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o bin/ramjam-darwin-amd64 ./cmd/ramjam

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o bin/ramjam-darwin-arm64 ./cmd/ramjam

# Windows
GOOS=windows GOARCH=amd64 go build -o bin/ramjam-windows-amd64.exe ./cmd/ramjam
```

Or use the Makefile:
```bash
make build-all
```

## Best Practices

### 1. Command Design
- Keep commands focused and single-purpose
- Use clear, descriptive names
- Provide helpful error messages
- Support common flags (--help, --version)

### 2. Flag Naming
- Use consistent naming conventions
- Short flags for common options (-v, -h)
- Long flags for clarity (--verbose, --timeout)
- Set sensible defaults

### 3. Error Handling
- Always return errors, don't panic
- Provide context in error messages
- Use `fmt.Errorf` with `%w` for wrapping
- Validate input early

### 4. Testing
- Write tests for business logic
- Test error conditions
- Use table-driven tests
- Mock external dependencies

### 5. Documentation
- Keep README up to date
- Document all commands
- Provide examples
- Include troubleshooting section

### 6. Version Control
- Commit early and often
- Write meaningful commit messages
- Tag releases with semantic versioning
- Keep `.gitignore` current

## Troubleshooting

### Command Not Found After Install

**Problem:** `ramjam: command not found`

**Solution:**
```bash
# Check if binary exists
ls $(go env GOPATH)/bin/ramjam

# Add to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Make permanent by adding to ~/.bashrc or ~/.zshrc
```

### Build Failures

**Problem:** Build fails with dependency errors

**Solution:**
```bash
# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify go.mod
cat go.mod
```

### Import Errors

**Problem:** Cannot import package

**Solution:**
- Verify module path in `go.mod`
- Ensure imports match directory structure
- Run `go mod tidy`
- Check for typos in import paths

## Summary

You've successfully created a production-ready CLI tool with:

✅ Go modules for dependency management  
✅ Go workspace for development  
✅ Cobra framework for CLI structure  
✅ Standard project layout  
✅ Comprehensive Makefile  
✅ Proper .gitignore configuration  
✅ Complete documentation  
✅ Global installation capability  

The project follows Go best practices and can be easily extended with new commands and features.

## Next Steps

Consider adding:
- More HTTP methods (POST, PUT, DELETE, PATCH)
- Configuration file support
- Request/response logging
- Authentication mechanisms
- JSON formatting and pretty-printing
- Response validation
- Test suites
- CI/CD pipelines
- Distribution via package managers

## Resources

- [Cobra Documentation](https://github.com/spf13/cobra)
- [Go Modules Reference](https://go.dev/ref/mod)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Go CLI Tutorial](https://golang.org/doc/tutorial/cli)

---

**Created:** 2025  
**Last Updated:** 2025-12-19  
**Author:** ramjam development team
