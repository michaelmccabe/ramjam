# Integrating Ramjam into CI/CD Pipelines

Ramjam is designed to be a lightweight, standalone binary that is perfect for running End-to-End (E2E) API tests in Continuous Integration (CI) environments like GitHub Actions, GitLab CI, or Jenkins.

This guide details how to set up Ramjam as a quality gate in your deployment pipeline.

## Integration Strategy

To use Ramjam effectively in a CI pipeline, your workflow typically looks something like the following:

1. **Build & Start** your application (the System Under Test)
2. **Wait** for the application to be healthy/ready
3. **Install** Ramjam from GitHub releases
4. **Run** Ramjam workflows against the running application

## Installing Ramjam in CI

The recommended way to install Ramjam in CI is to download the pre-built binary from GitHub releases:

```bash
curl -L -o ramjam https://github.com/michaelmccabe/ramjam/releases/download/v1.0.0-beta.1/ramjam-linux-amd64
chmod +x ramjam
sudo mv ramjam /usr/local/bin/
```

This is faster and more reliable than building from source, and doesn't require Go to be installed.

## Real-World Example: jimjam

The [jimjam](https://github.com/michaelmccabe/jimjam) project uses Ramjam for integration testing. Here's their actual workflow:

**File:** `.github/workflows/integration-tests.yml`

```yaml
name: Integration Tests

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Install Rust
        uses: dtolnay/rust-toolchain@stable

      - name: Cache cargo registry
        uses: actions/cache@v4
        with:
          path: |
            ~/.cargo/registry
            ~/.cargo/git
            target
          key: ${{ runner.os }}-cargo-${{ hashFiles('**/Cargo.lock') }}
          restore-keys: |
            ${{ runner.os }}-cargo-

      - name: Build jimjam
        run: cargo build --release

      - name: Install ramjam
        run: |
          curl -L -o ramjam https://github.com/michaelmccabe/ramjam/releases/download/v1.0.0-beta.1/ramjam-linux-amd64
          chmod +x ramjam
          sudo mv ramjam /usr/local/bin/

      - name: Start jimjam server
        run: |
          ./target/release/jimjam-http &
          echo "Waiting for server to start..."
          sleep 3
          # Verify server is running
          curl --retry 5 --retry-delay 1 --retry-connrefused http://127.0.0.1:8080/api/health

      - name: Run ramjam integration tests
        run: ramjam run ramjam-test/

      - name: Stop jimjam server
        if: always()
        run: pkill jimjam-http || true
```

### Key Patterns from jimjam

1. **Download pre-built binary** - No need to install Go or build from source
2. **Health check with retries** - Uses `curl --retry` to wait for the server
3. **Test directory structure** - All Ramjam YAML files live in `ramjam-test/`
4. **Cleanup step** - Uses `if: always()` to ensure server is stopped even on failure

## GitHub Actions Example (Go Application)

Here's a complete example for a Go application with database dependencies:

```yaml
name: End-to-End API Tests

on:
  push:
    branches: [ "main" ]
  pull_request:
    branches: [ "main" ]

jobs:
  e2e-test:
    runs-on: ubuntu-latest
    
    # Service containers for dependencies (e.g., Postgres, Redis)
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: password
          POSTGRES_DB: testdb
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Build and Start API Server
        run: |
          go build -o app ./cmd/server
          ./app &
          echo "PID=$!" >> $GITHUB_ENV
        env:
          PORT: 8080
          DB_HOST: localhost
          DB_USER: postgres
          DB_PASSWORD: password

      - name: Wait for API
        run: |
          curl --retry 10 --retry-delay 2 --retry-connrefused http://localhost:8080/health

      - name: Install Ramjam
        run: |
          curl -L -o ramjam https://github.com/michaelmccabe/ramjam/releases/latest/download/ramjam-linux-amd64
          chmod +x ramjam
          sudo mv ramjam /usr/local/bin/

      - name: Run E2E Tests
        run: ramjam run ./tests/e2e/ --verbose

      - name: Stop API Server
        if: always()
        run: kill $PID || true
```

## Best Practices

### Dynamic Base URLs

Avoid hardcoding URLs in your Ramjam YAML files. Instead, use the configuration variable `${base_url}`.

**In your** `workflow.yaml`

```yaml
config:
  base_url: "http://localhost:8080" # Default for local dev
```


**In CI**

`ramjam` doesn't currently support overriding config via CLI flags directly, but you can structure your tests to rely on environment variables if you implement a pre-processing step or ensure your CI environment matches the config default.

### Database State

For reliable E2E tests, ensure your database starts in a clean state.

* Use `services` in GitHub Actions to spin up fresh containers.
* Or, have the first step of your Ramjam workflow call a "Reset DB" endpoint on your API if one exists (e.g., `POST /internal/reset-db`).

### Artifacts

If tests fail, you might want to see the logs. You can capture the output of the Ramjam command and upload it as an artifact.


```yaml
      - name: Run Ramjam
        run: ramjam run ./tests/e2e/ > test-results.log 2>&1

      - name: Upload Test Results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: ramjam-results
          path: test-results.log
```


