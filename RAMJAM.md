# How To Use Ramjam

Ramjam is a CLI tool for executing HTTP API workflows defined in YAML files. It allows you to define a series of HTTP requests, validate responses, capture values from responses, and use those captured values in subsequent requests.

## Command Line Usage

The primary command is `run`. You can pass one or more file paths or directory paths.

```bash
# Run a single workflow file
ramjam run my-workflow.yaml

# Run all YAML files in a directory
ramjam run ./tests/integration/

# Run multiple specific files
ramjam run login.yaml create-post.yaml

# Enable verbose output
ramjam run my-workflow.yaml --verbose

# Override variables at runtime (e.g. override base_url and custom variables)
ramjam run my-workflow.yaml --var base_url=http://staging.api.com --var user_id=42
```

## Workflow DSL Reference

A Ramjam workflow file is a YAML file with three main sections: `metadata`, `config`, and `workflow`.

### Structure

```yaml
metadata:
  name: "Workflow Name"
  author: "Author Name"
  description: "Description of what this workflow does"

config:
  base_url: "https://api.example.com" # Optional base URL for requests

workflow:
  - step: "step-id"
    description: "Step description"
    request:
      # ... request details ...
    expect:
      # ... validation rules ...
    capture:
      # ... variable capture rules ...
    output:
      # ... output messages ...
```

### Request Definition

The `request` block defines the HTTP request to be made.

```yaml
request:
  method: "POST" # GET, POST, PUT, DELETE, PATCH, etc.
  url: "${base_url}/users" # Supports variable substitution
  content_type: "application/json" # Optional (defaults to application/json). Supports application/x-www-form-urlencoded and multipart/form-data.
  body: # Optional request payload
    name: "John Doe"
    job: "Developer"
```

#### Request Encodings & Payload Types
- **JSON (Default)**: Use `content_type: "application/json"` (or leave blank). The `body` map is serialized as JSON.
- **Form URL-Encoded**: Use `content_type: "application/x-www-form-urlencoded"`. The `body` map keys/values are serialized as standard query form parameters.
- **Multipart Form-Data**: Use `content_type: "multipart/form-data"`. The `body` keys are written as form parts.
  - **File Uploads**: To upload a file in a multipart form request, prefix the value of a field with `@` followed by the path (relative to the workflow YAML file or absolute). For example:
    ```yaml
    request:
      method: "POST"
      url: "${base_url}/upload"
      content_type: "multipart/form-data"
      body:
        description: "User profile picture"
        avatar: "@/path/to/profile.png"
    ```

### Response Validation (`expect`)

The `expect` block defines assertions on the response.

```yaml
expect:
  status: 201 # Expected HTTP status code
  json_path_match: # List of JSONPath assertions
    - path: "name"
      operator: "eq" # Optional operator: eq, ne, gt, gte, lt, lte, contains
      value: "John Doe"
    - path: "id"
      value: 123 # Operator defaults to "eq" if omitted
```

#### Comparison Operators for JSONPath Matches
You can specify the `operator` field to run checks beyond simple equality:
* `eq` (Default): Checks if actual value equals the expected value.
* `ne`: Checks if actual value is not equal to the expected value.
* `gt` / `gte`: Checks if actual value is greater than (or equal to) the expected value. (Converts values to floats for comparison if numeric, otherwise falls back to string comparison).
* `lt` / `lte`: Checks if actual value is less than (or equal to) the expected value.
* `contains`: Checks if a string contains a substring, or if a JSON array contains the specified element.

### Capturing Variables (`capture`)

The `capture` block allows you to extract values from the response and store them as variables for use in later steps.

```yaml
capture:
  - json_path: "id" # Extract value using JSONPath
    as: "user_id"   # Variable name (usage: ${user_id})
  
  - header: "Authorization" # Extract from response header
    as: "auth_token"

  - regex: "Token: (.*)" # Extract using Regex (from body)
    as: "token_string"
```

### Output

The `output` block allows printing custom messages to the console.

```yaml
output:
  print: "Created user with ID: ${user_id}"
```

## Variable Substitution

Variables can be used in `url`, `body`, and `output` fields using the `${variable_name}` syntax.

* `${base_url}` is available if defined in `config`.
* Variables captured in previous steps are available by their `as` name.
* **CLI Overrides**: You can pass runtime variable values using the `--var` flag (e.g. `--var key=value`). These will take precedence over workflow configuration defaults (like `base_url`).

## Authentication Example

This example demonstrates a common pattern: logging in to get a JWT, and then using that token in the header of a subsequent request.

```yaml
metadata:
  name: "Auth Flow"
  author: "DevOps"
  description: "Login and access protected resource"

config:
  base_url: "https://api.example.com"

workflow:
  - step: "login"
    description: "Login with username and password"
    request:
      method: "POST"
      url: "${base_url}/login"
      body:
        username: "admin"
        password: "secret_password"
    expect:
      status: 200
    capture:
      # Assuming response is like: {"token": "eyJhbGci..."}
      - json_path: "token"
        as: "jwt_token"

  - step: "access-protected"
    description: "Access a protected resource using the JWT"
    request:
      method: "GET"
      url: "${base_url}/protected/resource"
      headers:
        Authorization: "Bearer ${jwt_token}"
    expect:
      status: 200
```

## Full Example

Here is a complete example showing a workflow that creates a user, verifies the creation, and then fetches the user's details.

```yaml
metadata:
  name: "User Lifecycle"
  author: "QA Team"
  description: "Creates a user and verifies retrieval"

config:
  base_url: "https://reqres.in/api"

workflow:
  - step: "create-user"
    description: "Create a new user"
    request:
      method: "POST"
      url: "${base_url}/users"
      body:
        name: "Morpheus"
        job: "Leader"
    expect:
      status: 201
      json_path_match:
        - path: "name"
          value: "Morpheus"
    capture:
      - json_path: "id"
        as: "new_user_id"
    output:
      print: "User created with ID: ${new_user_id}"

  - step: "get-user"
    description: "Retrieve the created user"
    request:
      method: "GET"
      url: "${base_url}/users/${new_user_id}"
    expect:
      status: 200
      json_path_match:
        - path: "data.id"
          value: ${new_user_id} # Validates against the captured variable
```


## Global Configuration Defaults

You can configure global defaults for all workflows in your local environment. Ramjam automatically checks for configuration files named `.ramjam.yaml` or `.ramjam.yml` in the following locations (with local workspace files taking precedence):
1. The current working directory (e.g. `./.ramjam.yaml`)
2. The user's home directory (e.g. `~/.ramjam.yaml`)

### Global Config Options
```yaml
defaults:
  base_url: "https://api.staging.example.com"
  timeout: "15s" # Supports duration strings (s, ms, m, h)
  headers:
    Authorization: "Bearer default_token_value"
    Accept: "application/json"
```

---

## Error Categories

When a workflow fails, the runner classifies the failure into one of several distinct error categories to make debugging easier:
* **Validation Error (`ValidationError`)**: Returned when status code check fails, request headers are missing, or a JSONPath assertion fails.
* **Network Error (`NetworkError`)**: Returned when the HTTP client is unable to establish a connection, resolves a bad hostname, or times out.
* **Parsing Error (`ParsingError`)**: Returned when a local request body payload cannot be parsed as JSON, or the server response is not in valid JSON format.
* **Resolution Error (`ResolutionError`)**: Returned when external resource files (like a referenced `body_file` or a multipart file upload) are missing or inaccessible.

---

## Integrating ramjam into your development workflow

I suggest to use Ramjam most effectively in a CI pipeline, you should install `ramjam` to your local machine and add a ramjam folder to your project. 


Add ramjam YAML files to this folder as you proceed, so you can reuse them later on for your CI/CD pipelines.


