# ramjam Architecture and Workflow Guide

This document provides a comprehensive overview of the ramjam architecture, its workflow engine, and how to extend the project with new capabilities.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Core Components](#core-components)
3. [Workflow Execution Flow](#workflow-execution-flow)
4. [YAML Workflow Format](#yaml-workflow-format)
5. [Variable Substitution System](#variable-substitution-system)
6. [Extending ramjam](#extending-ramjam)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

## Architecture Overview

ramjam follows a clean, layered architecture focused on executing HTTP API workflows defined in YAML files.

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Layer                             │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐                     │
│  │  root   │  │   run   │  │ version │                     │
│  └────┬────┘  └────┬────┘  └─────────┘                     │
└───────┼────────────┼──────────────────────────────────────┘
        │            │
        └────────────┼──────────────────────────────────────┐
                     │                                       │
┌────────────────────▼────────────────────────────────────┐ │
│                 Runner Layer                            │ │
│  ┌──────────────────────────────────────────────────┐  │ │
│  │  Workflow Parser & Executor                      │  │ │
│  │  - Load YAML                                     │  │ │
│  │  - Variable substitution                         │  │ │
│  │  - HTTP request execution                        │  │ │
│  │  - Response validation                           │  │ │
│  │  - Value capturing                               │  │ │
│  └──────────────────────────────────────────────────┘  │ │
└─────────────────────────────────────────────────────────┘ │
                                                             │
┌────────────────────────────────────────────────────────────▼┐
│                    Config Layer                              │
│  ┌──────────────┐  ┌──────────────┐                         │
│  │ YAML Loader  │  │ Command Text │                         │
│  └──────────────┘  └──────────────┘                         │
└──────────────────────────────────────────────────────────────┘
```

### Design Principles

1. **Declarative Over Imperative**: All HTTP interactions are defined in YAML, not code
2. **Single Responsibility**: Each package has a clear, focused purpose
3. **Testability**: Mock-friendly architecture with dependency injection
4. **Extensibility**: Easy to add new commands, validators, and features
5. **Zero Code for Users**: Users define workflows without writing Go code

## Core Components

### 1. CLI Layer (`cmd/ramjam/cmd/`)

The CLI layer handles command-line interface concerns using the Cobra framework.

#### Root Command (`root.go`)

```go
// Responsibilities:
// - Application entry point
// - Global flags (verbose, etc.)
// - Version information
// - Help text coordination
```

**Key Features**:
- Global `--verbose` flag for debug output
- Version display via `--version`
- Automatic help generation
- Subcommand registration

#### Run Command (`run.go`)

```go
// Responsibilities:
// - Accept workflow file/directory paths
// - Validate input paths
// - Delegate to runner for execution
// - Handle verbose flag
```

**Key Features**:
- Accepts single file or directory
- Sorts workflow files alphabetically
- Passes verbose flag to runner
- Error handling and user feedback

#### Version Command (`version.go`)

```go
// Responsibilities:
// - Display version information
// - Format version output
```

**Key Features**:
- Version set at build time via ldflags
- Simple, clean output format

### 2. Runner Layer (`pkg/runner/`)

The runner is the heart of ramjam, executing YAML-defined workflows.

#### Workflow Structure

```go
type Workflow struct {
    Name     string    // Workflow name for logging
    BaseURL  string    // Base URL for all requests
    Timeout  int       // Request timeout in seconds
    Variables map[string]string // Workflow-level variables
    Steps    []Step    // Ordered list of steps
}

type Step struct {
    Name    string            // Step name for logging
    Method  string            // HTTP method (GET, POST, etc.)
    Path    string            // URL path (appended to baseURL)
    Headers map[string]string // Request headers
    Body    string            // Request body
    Expect  *Expect          // Response expectations
    Capture []Capture        // Values to capture from response
    Output  []string         // Messages to print
}

type Expect struct {
    Status int              // Expected HTTP status code
    Body   []BodyExpectation // JSONPath expectations
}

type BodyExpectation struct {
    Path  string // JSONPath expression
    Value string // Expected value
}

type Capture struct {
    Name string // Variable name
    Path string // JSONPath to value
}
```

#### Execution Flow

```
1. Load YAML workflow file
   ↓
2. Parse into Workflow struct
   ↓
3. Initialize HTTP client with timeout
   ↓
4. For each step:
   ├─ Substitute variables in all fields
   ├─ Build HTTP request
   ├─ Execute request
   ├─ Validate status code
   ├─ Validate JSONPath expectations
   ├─ Capture values from response
   ├─ Print output messages
   └─ Add captures to variable context
   ↓
5. Report success/failure
```

#### Variable Substitution

Variables are substituted using `${varName}` syntax:

```yaml
variables:
  userId: "123"
  
steps:
  - name: "Get user ${userId}"
    path: "/users/${userId}"
    output:
      - "Fetched user ${userId}"
```

**Variable Sources** (in order of precedence):
1. Captured values from previous steps
2. Workflow-level variables
3. Environment variables (`${env:VAR_NAME}`)

#### Response Validation

Two types of validation:

**Status Code Validation**:
```yaml
expect:
  status: 200  # Must match exactly
```

**JSONPath Validation**:
```yaml
expect:
  body:
    - path: "$.user.id"
      value: "123"
    - path: "$.user.active"
      value: "true"
```

#### Value Capturing

Extract values from responses for later use:

```yaml
capture:
  - name: userId
    path: "$.data.id"
  - name: token
    path: "$.auth.token"
```

Captured values become available as `${userId}` and `${token}` in subsequent steps.

### 3. Config Layer (`pkg/config/`)

Handles all YAML loading and configuration management.

#### Generic YAML Loader (`loader.go`)

```go
// Reusable YAML loading functions
func LoadFile(path string, target interface{}) error
func Parse(data []byte, target interface{}) error
```

**Use Cases**:
- Loading workflow files
- Loading command text configuration
- Future: instruction files, settings files

#### Command Text Config (`commands.go`)

```go
type CommandsConfig struct {
    Root    CommandText `yaml:"root"`
    Run     CommandText `yaml:"run"`
    Version CommandText `yaml:"version"`
}
```

Externalizes command descriptions to `resources/commands.yaml`, making them easy to update without code changes.

## Workflow Execution Flow

### Detailed Step-by-Step Execution

```
┌─────────────────────────────────────────────────────────────┐
│ 1. User runs: ramjam run workflow.yaml                      │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Run Command validates file path                          │
│    - Check file/directory exists                            │
│    - Collect all .yaml/.yml files                           │
│    - Sort files alphabetically                              │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. For each workflow file:                                  │
│    Runner.Execute(workflow)                                 │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Parse YAML into Workflow struct                          │
│    - Validate structure                                     │
│    - Initialize variable context                            │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. Create HTTP client with timeout                          │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 6. Execute each step sequentially                           │
└───────────────────────────┬─────────────────────────────────┘
                            │
        ┌───────────────────┴───────────────────┐
        │                                       │
        ▼                                       ▼
┌──────────────────┐                  ┌──────────────────┐
│ Step Execution   │                  │ Error Handling   │
│ (see below)      │                  │ - Log error      │
│                  │                  │ - Stop workflow  │
│                  │                  │ - Return failure │
└──────────────────┘                  └──────────────────┘
```

### Individual Step Execution

```
┌─────────────────────────────────────────────────────────────┐
│ Step Execution Flow                                          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 1. Substitute Variables                                      │
│    - Replace ${var} in all string fields                    │
│    - URL, headers, body, output messages                    │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Build HTTP Request                                        │
│    - Construct URL (baseURL + path)                         │
│    - Set method (GET, POST, etc.)                           │
│    - Add headers                                            │
│    - Add body if present                                    │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. Execute HTTP Request                                      │
│    - Send request                                           │
│    - Wait for response (with timeout)                       │
│    - Read response body                                     │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Validate Status Code                                      │
│    - Compare actual vs expected                             │
│    - Fail if mismatch                                       │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. Parse Response JSON                                       │
│    - Unmarshal JSON                                         │
│    - Prepare for JSONPath queries                           │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 6. Validate Body Expectations                                │
│    - For each expectation:                                  │
│      * Query JSONPath                                       │
│      * Compare value                                        │
│      * Fail if mismatch                                     │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 7. Capture Values                                            │
│    - For each capture:                                      │
│      * Query JSONPath                                       │
│      * Store in variable context                            │
│      * Available for subsequent steps                       │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 8. Print Output Messages                                     │
│    - Substitute variables in messages                       │
│    - Print to stdout                                        │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 9. Log Success (if verbose)                                 │
│    - Step name                                              │
│    - Status code                                            │
│    - Response time                                          │
└─────────────────────────────────────────────────────────────┘
```

## YAML Workflow Format

### Complete Example

```yaml
# Workflow metadata
name: "User Management API Tests"
baseURL: "https://api.example.com"
timeout: 30  # seconds

# Workflow-level variables
variables:
  apiVersion: "v1"
  environment: "staging"

# Sequential steps
steps:
  # Step 1: Create a user
  - name: "Create new user"
    method: POST
    path: "/api/${apiVersion}/users"
    headers:
      Content-Type: "application/json"
      X-Environment: "${environment}"
    body: |
      {
        "email": "test@example.com",
        "name": "Test User"
      }
    expect:
      status: 201
      body:
        - path: "$.success"
          value: "true"
    capture:
      - name: newUserId
        path: "$.data.id"
      - name: createdAt
        path: "$.data.createdAt"
    output:
      - "✓ Created user with ID: ${newUserId}"
      - "  Created at: ${createdAt}"

  # Step 2: Fetch the created user
  - name: "Fetch user ${newUserId}"
    method: GET
    path: "/api/${apiVersion}/users/${newUserId}"
    headers:
      Accept: "application/json"
    expect:
      status: 200
      body:
        - path: "$.data.id"
          value: "${newUserId}"
        - path: "$.data.email"
          value: "test@example.com"
    output:
      - "✓ Successfully fetched user ${newUserId}"

  # Step 3: Update the user
  - name: "Update user ${newUserId}"
    method: PUT
    path: "/api/${apiVersion}/users/${newUserId}"
    headers:
      Content-Type: "application/json"
    body: |
      {
        "name": "Updated Test User"
      }
    expect:
      status: 200
      body:
        - path: "$.data.name"
          value: "Updated Test User"
    output:
      - "✓ Updated user ${newUserId}"

  # Step 4: Delete the user
  - name: "Delete user ${newUserId}"
    method: DELETE
    path: "/api/${apiVersion}/users/${newUserId}"
    expect:
      status: 204
    output:
      - "✓ Deleted user ${newUserId}"
```

### Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Workflow name for logging |
| `baseURL` | string | Yes | Base URL for all requests |
| `timeout` | int | No | Request timeout in seconds (default: 30) |
| `variables` | map | No | Workflow-level variables |
| `steps` | array | Yes | List of steps to execute |

#### Step Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Step name for logging |
| `method` | string | Yes | HTTP method (GET, POST, PUT, DELETE, PATCH) |
| `path` | string | Yes | URL path (appended to baseURL) |
| `headers` | map | No | Request headers |
| `body` | string | No | Request body (usually JSON) |
| `expect` | object | No | Response expectations |
| `capture` | array | No | Values to capture from response |
| `output` | array | No | Messages to print after step |

#### Expect Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | int | No | Expected HTTP status code |
| `body` | array | No | JSONPath expectations |

#### Body Expectation Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `path` | string | Yes | JSONPath expression |
| `value` | string | Yes | Expected value (as string) |

#### Capture Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Variable name for captured value |
| `path` | string | Yes | JSONPath to value in response |

## Variable Substitution System

### Variable Syntax

All variables use the syntax: `${variableName}`

### Variable Sources

1. **Captured Values** (highest precedence)
   ```yaml
   capture:
     - name: userId
       path: "$.id"
   # Later: ${userId}
   ```

2. **Workflow Variables**
   ```yaml
   variables:
     apiKey: "secret123"
   # Use: ${apiKey}
   ```

3. **Environment Variables**
   ```yaml
   # Use: ${env:HOME}
   # Use: ${env:API_KEY}
   ```

### Where Variables Work

Variables can be used in:
- Request paths: `path: "/users/${userId}"`
- Request headers: `Authorization: "Bearer ${token}"`
- Request body: `"userId": "${userId}"`
- Output messages: `"User ${userId} created"`
- Expected values: `value: "${expectedId}"`

### Variable Scope

```
┌─────────────────────────────────────────┐
│ Variable Context Lifecycle              │
└─────────────────────────────────────────┘

Workflow Start
│
├─ Load workflow variables
│  (available to all steps)
│
└─ For each step:
   │
   ├─ Inherit previous captures
   ├─ Substitute variables in step
   ├─ Execute request
   ├─ Capture new values
   └─ New captures available to next step
```

## Extending ramjam

### Adding a New Command

#### 1. Create Command File

```bash
# Create new command file
touch cmd/ramjam/cmd/newcommand.go
```

#### 2. Implement Command

```go
// filepath: cmd/ramjam/cmd/newcommand.go
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [args]",
	Short: "Short description",
	Long:  "Long description with examples",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation here
		verbose, _ := cmd.Flags().GetBool("verbose")
		
		if verbose {
			fmt.Println("[DEBUG] Executing new command")
		}
		
		// Your logic
		return nil
	},
}

func init() {
	// Register with root command
	rootCmd.AddCommand(newCmd)
	
	// Add command-specific flags
	newCmd.Flags().StringP("option", "o", "", "Option description")
}
```

#### 3. Add Tests

```go
// filepath: cmd/ramjam/cmd/newcommand_test.go
package cmd

import (
	"testing"
)

func TestNewCmd(t *testing.T) {
	// Test implementation
}
```

#### 4. Update Command Text (Optional)

```yaml
# filepath: resources/commands.yaml
new:
  use: "new [args]"
  short: "Short description"
  long: |
    Long description with examples
```

### Adding a New HTTP Method

Currently supported: GET, POST, PUT, DELETE, PATCH

To add a new method (e.g., HEAD, OPTIONS):

#### 1. Update Runner

```go
// filepath: pkg/runner/runner.go

// In buildRequest function, add case:
switch strings.ToUpper(step.Method) {
// ...existing cases...
case "HEAD":
	req, err = http.NewRequest("HEAD", url, nil)
case "OPTIONS":
	req, err = http.NewRequest("OPTIONS", url, nil)
// ...
}
```

#### 2. Add Tests

```go
// filepath: pkg/runner/runner_test.go

func TestRunnerHEADRequest(t *testing.T) {
	// Test implementation
}
```

### Adding New Validation Types

Current validation: status code and JSONPath

To add new validation (e.g., response headers, schema validation):

#### 1. Extend Expect Struct

```go
// filepath: pkg/runner/runner.go

type Expect struct {
	Status  int                `yaml:"status"`
	Body    []BodyExpectation  `yaml:"body"`
	Headers map[string]string  `yaml:"headers"` // NEW
}
```

#### 2. Implement Validation

```go
// Add validation function
func validateHeaders(actual http.Header, expected map[string]string) error {
	for key, expectedValue := range expected {
		actualValue := actual.Get(key)
		if actualValue != expectedValue {
			return fmt.Errorf("header %s: expected %s, got %s", 
				key, expectedValue, actualValue)
		}
	}
	return nil
}

// Call in executeStep
if step.Expect != nil && step.Expect.Headers != nil {
	if err := validateHeaders(resp.Header, step.Expect.Headers); err != nil {
		return fmt.Errorf("header validation failed: %w", err)
	}
}
```

#### 3. Update Documentation

```yaml
# Example usage
expect:
  status: 200
  headers:
    Content-Type: "application/json"
    X-Custom-Header: "expected-value"
```

### Adding Authentication Support

To add built-in authentication (e.g., Bearer token, API key):

#### 1. Extend Workflow Struct

```go
// filepath: pkg/runner/runner.go

type Workflow struct {
	// ...existing fields...
	Auth *Auth `yaml:"auth"`
}

type Auth struct {
	Type   string `yaml:"type"`   // "bearer", "apikey", "basic"
	Token  string `yaml:"token"`  // For bearer
	Key    string `yaml:"key"`    // For API key
	Header string `yaml:"header"` // Header name for API key
	User   string `yaml:"user"`   // For basic auth
	Pass   string `yaml:"pass"`   // For basic auth
}
```

#### 2. Implement Auth Injection

```go
func (r *Runner) applyAuth(req *http.Request, auth *Auth) error {
	if auth == nil {
		return nil
	}
	
	switch auth.Type {
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+auth.Token)
	case "apikey":
		req.Header.Set(auth.Header, auth.Key)
	case "basic":
		req.SetBasicAuth(auth.User, auth.Pass)
	default:
		return fmt.Errorf("unsupported auth type: %s", auth.Type)
	}
	
	return nil
}

// In executeStep:
if err := r.applyAuth(req, workflow.Auth); err != nil {
	return err
}
```

#### 3. Document Usage

```yaml
# Bearer token example
auth:
  type: bearer
  token: "${env:API_TOKEN}"

# API key example
auth:
  type: apikey
  header: "X-API-Key"
  key: "${env:API_KEY}"
```

### Adding Conditional Step Execution

To add conditional logic (if/then):

#### 1. Extend Step Struct

```go
type Step struct {
	// ...existing fields...
	If *Condition `yaml:"if"`
}

type Condition struct {
	Variable string `yaml:"variable"`
	Operator string `yaml:"operator"` // "equals", "contains", "exists"
	Value    string `yaml:"value"`
}
```

#### 2. Implement Condition Evaluation

```go
func (r *Runner) evaluateCondition(cond *Condition, vars map[string]string) bool {
	if cond == nil {
		return true // No condition = always execute
	}
	
	varValue, exists := vars[cond.Variable]
	
	switch cond.Operator {
	case "exists":
		return exists
	case "equals":
		return exists && varValue == cond.Value
	case "contains":
		return exists && strings.Contains(varValue, cond.Value)
	default:
		return false
	}
}

// In executeStep:
if !r.evaluateCondition(step.If, r.variables) {
	r.debugLog("Skipping step: condition not met")
	return nil
}
```

#### 3. Document Usage

```yaml
steps:
  - name: "Only if user exists"
    if:
      variable: "userId"
      operator: "exists"
    method: GET
    path: "/users/${userId}"
```

## Best Practices

### Workflow Design

1. **Use Descriptive Names**
   ```yaml
   name: "User Registration Flow - Happy Path"
   ```

2. **Group Related Steps**
   ```yaml
   # Group 1: Setup
   - name: "Setup: Create test data"
   # Group 2: Main flow
   - name: "Main: Execute workflow"
   # Group 3: Cleanup
   - name: "Cleanup: Remove test data"
   ```

3. **Capture Reusable Values**
   ```yaml
   capture:
     - name: authToken
       path: "$.token"
   # Use in subsequent steps
   ```

4. **Use Environment Variables for Secrets**
   ```yaml
   headers:
     Authorization: "Bearer ${env:API_TOKEN}"
   ```

5. **Add Helpful Output Messages**
   ```yaml
   output:
     - "✓ User ${userId} created successfully"
     - "  Email: ${email}"
     - "  Role: ${role}"
   ```

### Code Organization

1. **Keep Packages Focused**
   - `cmd/` - CLI concerns only
   - `pkg/runner/` - Workflow execution only
   - `pkg/config/` - Configuration loading only

2. **Write Tests First**
   - Test happy paths
   - Test error conditions
   - Use table-driven tests

3. **Use Interfaces for Testability**
   ```go
   type HTTPClient interface {
       Do(*http.Request) (*http.Response, error)
   }
   ```

4. **Handle Errors Properly**
   ```go
   if err != nil {
       return fmt.Errorf("failed to X: %w", err)
   }
   ```

### Testing Strategies

1. **Unit Tests for Logic**
   ```go
   func TestVariableSubstitution(t *testing.T) {
       // Test pure functions
   }
   ```

2. **Integration Tests with Mock Servers**
   ```go
   server := httptest.NewServer(handler)
   defer server.Close()
   ```

3. **End-to-End Tests with Real Workflows**
   ```go
   func TestRunWorkflowFile(t *testing.T) {
       // Test actual YAML workflows
   }
   ```

## Troubleshooting

### Common Issues

#### 1. Variable Not Substituted

**Problem**: `${userId}` appears literally in output

**Causes**:
- Variable not captured
- Typo in variable name
- Variable scope issue

**Solution**:
```bash
# Run with verbose flag
ramjam run workflow.yaml -v

# Check capture section
capture:
  - name: userId  # Must match exactly
    path: "$.data.id"
```

#### 2. JSONPath Not Matching

**Problem**: Expectation fails even though value looks correct

**Causes**:
- Type mismatch (number vs string)
- Whitespace differences
- Wrong path

**Solution**:
```yaml
# Ensure values are strings in expectations
expect:
  body:
    - path: "$.id"
      value: "123"  # String, not number

# Test JSONPath separately
# Use verbose mode to see actual response
```

#### 3. Request Timeout

**Problem**: Request times out

**Causes**:
- Default timeout too short
- Slow API
- Network issues

**Solution**:
```yaml
# Increase timeout
timeout: 60  # seconds

# Or per-step (future enhancement)
```

#### 4. Workflow File Not Found

**Problem**: `ramjam run workflow.yaml` fails

**Causes**:
- Wrong path
- File extension not .yaml or .yml
- Permission issues

**Solution**:
```bash
# Use absolute path
ramjam run /full/path/to/workflow.yaml

# Check file exists
ls -la workflow.yaml

# Check permissions
chmod 644 workflow.yaml
```

### Debug Techniques

1. **Use Verbose Flag**
   ```bash
   ramjam run workflow.yaml -v
   ```

2. **Add Debug Output**
   ```yaml
   output:
     - "DEBUG: userId = ${userId}"
     - "DEBUG: token = ${token}"
   ```

3. **Test Individual Steps**
   - Create minimal workflow with single step
   - Verify each step works independently

4. **Check HTTP Responses**
   - Use verbose mode to see actual responses
   - Compare with expectations

5. **Validate YAML Syntax**
   ```bash
   # Use online YAML validator
   # Or use yamllint
   yamllint workflow.yaml
   ```

## Summary

ramjam's architecture is designed for:

- **Simplicity**: Users write YAML, not code
- **Flexibility**: Easy to extend with new features
- **Testability**: All components are testable
- **Maintainability**: Clear separation of concerns

Key extension points:
- New commands (CLI layer)
- New HTTP methods (Runner layer)
- New validation types (Runner layer)
- New authentication methods (Runner layer)
- Conditional execution (Runner layer)

The workflow-driven approach ensures that:
- API tests are version controlled
- Tests are reproducible
- No programming knowledge required
- Easy to share across teams

---

**Last Updated**: 2025-12-19  
**Author**: ramjam development team
