# Integrating Ramjam into CI/CD Pipelines

Ramjam is designed to be a lightweight, standalone binary that is perfect for running End-to-End (E2E) API tests in Continuous Integration (CI) environments like GitHub Actions, GitLab CI, or Jenkins.

This guide details how to set up Ramjam as a quality gate in your deployment pipeline.

## Integration Strategy

To use Ramjam effectively in a CI pipeline, your workflow typically follows this sequence:

1.  **Build & Start** your application (the System Under Test).
2.  **Wait** for the application to be healthy/ready.
3.  **Install** Ramjam.
4.  **Run** Ramjam workflows against the running application.

## GitHub Actions Example

Below is a complete example of a GitHub Actions workflow that spins up a Go application and runs Ramjam tests against it.

Create this file at `.github/workflows/e2e-tests.yaml`:

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
      # 1. Checkout your code
      - uses: actions/checkout@v3

      # 2. Set up language environment
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      # 3. Build and Start your API in the background
      - name: Start API Server
        run: |
          go build -o app ./cmd/server
          ./app &
          echo "PID=$!" >> $GITHUB_ENV
        env:
          PORT: 8080
          DB_HOST: localhost
          DB_USER: postgres
          DB_PASSWORD: password

      # 4. Wait for API to be ready (Healthcheck)
      - name: Wait for API
        run: |
          timeout 30s bash -c 'until curl -s http://localhost:8080/health; do sleep 1; done'

      # 5. Install Ramjam
      - name: Install Ramjam
        run: |
          git clone https://github.com/michaelmccabe/ramjam.git /tmp/ramjam
          cd /tmp/ramjam && make install
          echo "$(go env GOPATH)/bin" >> $GITHUB_PATH

      # 6. Run E2E Tests
      - name: Run Ramjam workflows
        run: |
          ramjam run ./tests/e2e/ --verbose
        env:
          # Ensure your YAML files use ${base_url} so this works dynamically
          BASE_URL: http://localhost:8080 

      # 7. (Optional) Notify on Failure
      - name: Notify on Failure
        if: failure()
        run: |
          curl -X POST ${{ secrets.SLACK_WEBHOOK_URL }} \
          -d '{"text":"E2E Tests Failed for commit ${{ github.sha }}"}'
```

## Best Practices

### 1. Dynamic Base URLs
Avoid hardcoding URLs in your Ramjam YAML files. Instead, use the configuration variable `${base_url}`.

**In your `workflow.yaml`:**
```yaml
config:
  base_url: "http://localhost:8080" # Default for local dev
```

**In CI:**
Ramjam doesn't currently support overriding config via CLI flags directly (feature coming soon), but you can structure your tests to rely on environment variables if you implement a pre-processing step or ensure your CI environment matches the config default.

### 2. Database State
For reliable E2E tests, ensure your database starts in a clean state.
- Use `services` in GitHub Actions to spin up fresh containers.
- Or, have the first step of your Ramjam workflow call a "Reset DB" endpoint on your API if one exists (e.g., `POST /internal/reset-db`).

### 3. Artifacts
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
