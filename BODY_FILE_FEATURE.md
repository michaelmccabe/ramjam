# Body File Feature

## Overview

The `body_file` feature allows you to load request bodies from external JSON files instead of defining them inline in your YAML test files. This is useful for:

- **Reusability**: Share the same request body across multiple tests
- **Maintainability**: Keep large JSON payloads separate from test logic
- **Clarity**: Keep test YAML files clean and focused on test flow
- **Version Control**: Track changes to request bodies separately

## Usage

### Basic Usage

Instead of defining the body inline:

```yaml
request:
  method: "POST"
  url: "/api/posts"
  body:
    title: "My Post"
    content: "Post content"
    userId: 1
```

You can reference an external JSON file:

```yaml
request:
  method: "POST"
  url: "/api/posts"
  body_file: "post-body.json"
```

### File Path Resolution

- **Relative paths**: Resolved relative to the YAML test file's directory
- **Absolute paths**: Used as-is

Example directory structure:
```
testdata/
  ├── test.yaml
  ├── post-body.json          # Same directory: "post-body.json"
  └── payloads/
      └── user.json           # Subdirectory: "payloads/user.json"
```

### Variable Substitution

Variables work in JSON files loaded via `body_file`, just like inline bodies:

**body.json**:
```json
{
  "userId": "${user_id}",
  "action": "create",
  "timestamp": "${timestamp}"
}
```

**test.yaml**:
```yaml
workflow:
- step: "capture-user-id"
  request:
    method: "GET"
    url: "/current-user"
  capture:
  - json_path: "id"
    as: "user_id"

- step: "perform-action"
  request:
    method: "POST"
    url: "/actions"
    body_file: "body.json"  # ${user_id} will be substituted
```

## Examples

### Example 1: Simple POST with External Body

**post-request.json**:
```json
{
  "title": "New Blog Post",
  "body": "This is the content of the post",
  "userId": 1,
  "tags": ["tutorial", "api"]
}
```

**test.yaml**:
```yaml
metadata:
  name: "Blog Post Test"

config:
  base_url: "https://jsonplaceholder.typicode.com"

workflow:
- step: "create-post"
  description: "Create a new blog post"
  request:
    method: "POST"
    url: "${base_url}/posts"
    body_file: "post-request.json"
  expect:
    status: 201
    json_path_match:
    - path: "title"
      value: "New Blog Post"
```

### Example 2: Using Variables with Body File

**user-profile.json**:
```json
{
  "name": "${user_name}",
  "email": "${user_email}",
  "role": "admin"
}
```

**test.yaml**:
```yaml
workflow:
- step: "get-user-info"
  request:
    method: "GET"
    url: "/user/current"
  capture:
  - json_path: "name"
    as: "user_name"
  - json_path: "email"
    as: "user_email"

- step: "update-profile"
  request:
    method: "PUT"
    url: "/profile"
    body_file: "user-profile.json"
  expect:
    status: 200
```

### Example 3: Complex Nested JSON

**complex-payload.json**:
```json
{
  "user": {
    "name": "Test User",
    "email": "test@example.com",
    "preferences": {
      "notifications": true,
      "theme": "dark"
    }
  },
  "metadata": {
    "source": "api-test",
    "version": "1.0",
    "tags": ["test", "automation"]
  }
}
```

**test.yaml**:
```yaml
workflow:
- step: "submit-complex-data"
  request:
    method: "POST"
    url: "/api/data"
    body_file: "complex-payload.json"
  expect:
    status: 201
```

## Notes

- **Priority**: If both `body` and `body_file` are specified, `body_file` takes precedence
- **Format**: Only JSON format is supported for body files
- **Validation**: The JSON file must be valid JSON or the test will fail
- **Content-Type**: Automatically set to `application/json` when a body is present
- **Verbose Mode**: Use `-v` flag to see which file was used for the body

## Error Handling

Common errors and solutions:

| Error | Cause | Solution |
|-------|-------|----------|
| `read body file: no such file or directory` | File not found | Check the path is correct and relative to YAML file |
| `parse body file: invalid character` | Invalid JSON | Validate JSON syntax in the file |
| `resolve body file: permission denied` | No read permissions | Check file permissions |

## Testing

Run the included example:
```bash
./bin/ramjam run resources/testdata/success/postInpuTest.yaml -v
```

This will demonstrate both inline body and body_file usage.
