# 🏋🏽 ramjam

A lightweight, declarative CLI tool for validating, testing, and automating HTTP API workflows.

> 🤖 **Working with AI Coding Agents?** Check out the [Agentic Testing Guide](./AGENT_README.md) for prompt templates and best practices to let LLMs write and execute tests.

---

## 🎯 Purpose & Philosophy

`ramjam` is designed to simplify E2E API integration testing. Instead of writing verbose script files (using Python, Node.js, etc.) or clicking around inside heavy desktop GUIs (like Postman or Insomnia), `ramjam` lets you define sequential HTTP workflows, validation assertions, variable captures, and console outputs in clean, source-controlled YAML files.

It is ideal for:
1. **Developer API Prototyping**: Run local request sequences and verify behavior directly from your terminal.
2. **CI/CD Quality Gates**: Run verification tests inside your pipelines (GitHub Actions, GitLab CI, etc.) to ensure APIs are correct before release.
3. **Automated Runbooks**: Script multi-step setups, logins, or migrations without writing code.

---

## 🔌 How to Integrate ramjam into Your Project

Integrating `ramjam` takes less than 5 minutes:

### 1. Structure Your Tests
Create a test directory in your repository (e.g. `tests/integration/` or `ramjam/`) to store your test workflows and JSON payloads:

```
my-project/
├── .github/workflows/ci.yml
├── src/
└── tests/integration/
    ├── .ramjam.yaml             # Optional workspace-specific defaults
    ├── auth_flow.yaml           # Login and token validation workflow
    ├── create_user.yaml         # User creation workflow
    └── payloads/
        └── new_user.json        # External request payload JSON file
```

### 2. Define a Test Workflow
Create a YAML file (e.g. `tests/integration/lifecycle.yaml`):

```yaml
metadata:
  name: "User Lifecycle E2E"
  description: "Creates, retrieves, and validates a user profile."

config:
  base_url: "https://api.example.com" # Default fallback URL

workflow:
  - step: "create-user"
    description: "Submit new profile"
    request:
      method: "POST"
      url: "/users"
      body:
        name: "Alice Smith"
        role: "developer"
    expect:
      status: 201
    capture:
      - json_path: "id"
        as: "new_user_id"

  - step: "verify-user"
    description: "Verify details on GET request"
    request:
      method: "GET"
      url: "/users/${new_user_id}"
    expect:
      status: 200
      json_path_match:
        - path: "name"
          value: "Alice Smith"
        - path: "role"
          value: "developer"
```

### 3. Run Locally or in CI
Run the workflow:
```bash
# Execute local tests
ramjam run ./tests/integration/

# Execute staging E2E tests by overriding base URL via CLI flags
ramjam run ./tests/integration/ --var base_url=https://api.staging.example.com
```

---

## 📖 Documentation Directory

| Document | Description |
|---|---|
| 📖 **[How To Use Ramjam](./RAMJAM.md)** | Complete workflow DSL reference, variable substitution, operators, and global config defaults |
| 🤖 **[Agentic Testing Guide](./AGENT_README.md)** | Best practices and prompt templates for instructing AI coding agents to test APIs with Ramjam |
| 🚀 **[CI/CD Integration](./INTEGRATE.md)** | Guide for running ramjam in GitHub Actions, GitLab CI, and other build servers |
| 📦 **[Body File Feature](./BODY_FILE_FEATURE.md)** | Loading request body payloads dynamically from external files |

---

## ✨ Features

* **Lightweight CLI**: Built with Cobra for clean execution commands.
* **Flexible Execution Paths**: Run single workflows, specific list inputs, or folders of test workflows.
* **Structured Logs**: Concurrent real-time logging streamed using Go's standard `slog` library.
* **Dynamic Scope Variables**: Parameterize your endpoints with workflow configuration variables, capture-mapped response fields, and CLI environment values (`--var`).
* **Rich JSONPath Assertions**: Validate responses via `AsaiYusuke/jsonpath` supporting comparison operators (`eq`, `ne`, `gt`, `gte`, `lt`, `lte`, `contains`).
* **Payload Encodings**: Built-in support for standard JSON, Form URL-Encoded (`application/x-www-form-urlencoded`), and Multipart Form-Data (`multipart/form-data`) with file uploading (`@file`).
* **Configuration Defaults**: Zero-setup workspace or system defaults configuration files (`.ramjam.yaml`).
* **Debugging Diagnostics**: Contextual, categorized failures mapping (`ValidationError`, `NetworkError`, `ParsingError`, `ResolutionError`).

---

## 🛠️ Installation

### Quick Install (Go Installed)
Install to `$GOPATH/bin` or `$GOBIN`:
```bash
make install
```
Make sure `~/go/bin` is in your environment shell `PATH`:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Clone and Build from Source
```bash
git clone https://github.com/michaelmccabe/ramjam.git
cd ramjam

# Build current binary
make build

# Build binaries for Linux, macOS, and Windows
make build-all
```
The compiled binaries will be outputted under the local `bin/` folder.

---

## 💻 Development

### Project Structure
```
ramjam/
├── cmd/
│   └── ramjam/           # CLI Entrypoint
│       ├── main.go
│       └── cmd/          # Cobra Commands
├── pkg/
│   ├── config/           # Spec & Loader
│   └── runner/           # Workflow Evaluator Engine
├── resources/            # Test data & sample suites
├── Makefile              # Compile automation
└── go.mod
```

### Development commands
```bash
# Clean build artifacts
make clean

# Run test suites
make test

# Run tests with coverage reporting
make test-coverage

# Run local development build
go run ./cmd/ramjam run resources/testdata/success
```

---

## 🚀 Creating Releases

Releases are fully automated via GitHub Actions. Pushing a version tag builds binaries for Linux, macOS, and Windows (AMD64 & ARM64) with SHA256 verification checksums:

```bash
git tag v1.0.0
git push origin v1.0.0
```

---

## 💬 Support & Contribution

For issues, feature requests, or contributions, please visit the [GitHub repository](https://github.com/michaelmccabe/ramjam).
