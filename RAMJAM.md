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
  body: # Optional JSON body
    name: "John Doe"
    job: "Developer"
```

### Response Validation (`expect`)

The `expect` block defines assertions on the response.

```yaml
expect:
  status: 201 # Expected HTTP status code
  json_path_match: # List of JSONPath assertions
    - path: "name"
      value: "John Doe"
    - path: "id"
      value: 123
```

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


## Integrating ramjam into your development workflow

I suggest to use Ramjam most effectively in a CI pipeline, you should install `ramjam` to your local machine and add a ramjam folder to your project. 


Add ramjam YAML files to this folder as you proceed, so you can reuse them later on for your CI/CD pipelines.


