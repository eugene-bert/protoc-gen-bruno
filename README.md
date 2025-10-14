# protoc-gen-bruno

Generate Bruno API collections from Protocol Buffer definitions with gRPC-Gateway annotations.

## Features

- ✅ Generates Bruno HTTP requests from `google.api.http` annotations
- ✅ Generates Bruno gRPC requests from proto service definitions
- ✅ Auto-generates example JSON request bodies based on proto message types
- ✅ **Auto-generates query parameters for GET/DELETE requests**
- ✅ Smart field mapping (path params, query params, body)
- ✅ Supports nested messages, repeated fields, enums, and all proto types
- ✅ Configurable generation modes (HTTP only, gRPC only, or both)
- ✅ Custom or auto-generated collection names
- ✅ **Multi-environment support** (Dev, Staging, Production)
- ✅ Multi-module workspace support

## Installation

```bash
go build -o protoc-gen-bruno .
```

## Usage

### Basic Setup

1. Create a `buf.gen.yaml`:

```yaml
version: v2
plugins:
  - local: ./protoc-gen-bruno
    out: bruno/collections
```

2. Run `buf generate`:

```bash
buf generate
```

3. Open the generated collection in Bruno:
   - Open Bruno
   - Click "Open Collection"
   - Select `bruno/collections` folder

### Custom Collection Name

By default, the collection name is auto-generated from your service names or package name. You can override this:

```yaml
version: v2
plugins:
  - local: protoc-gen-bruno
    out: bruno/collections
    opt:
      - collection_name=My Company API  # Custom name
```

If not specified, the plugin generates names like:
- Single service: `UserService API`
- Multiple services: `Example V1 API` (from package `example.v1`)

### Generation Modes

Control what gets generated using the `mode` option:

```yaml
version: v2
plugins:
  # Generate both HTTP and gRPC (default)
  - local: ./protoc-gen-bruno
    out: bruno/collections
    opt: mode=all

  # HTTP requests only
  - local: ./protoc-gen-bruno
    out: bruno/collections
    opt: mode=http

  # gRPC requests only
  - local: ./protoc-gen-bruno
    out: bruno/collections
    opt: mode=grpc
```

### Environment Configuration

Generate multiple environments (Development, Staging, Production) automatically by specifying environment URLs:

```yaml
version: v2
plugins:
  - local: protoc-gen-bruno
    out: bruno/collections
    opt:
      - collection_name=My API
      - dev_url=https://api.dev.example.com/service
      - stg_url=https://api.staging.example.com/service
      - prd_url=https://api.example.com/service
```

This generates:
- `environments/Development.bru` - with dev URLs
- `environments/Staging.bru` - with staging URLs
- `environments/Production.bru` - with production URLs

**Environment file example (Development.bru):**
```
vars {
  base_url: https://api.dev.example.com/service
  grpc_url: api.dev.example.com:443
}
```

The plugin automatically:
- Converts HTTPS URLs to gRPC host:port (e.g., `https://api.dev.example.com` → `api.dev.example.com:443`)
- Uses port 443 for HTTPS, port 80 for HTTP
- Strips paths from URLs for gRPC endpoints

**Custom local environment:**
```yaml
opt:
  - local_url=http://localhost:3000/api
```

**Default behavior:**
- If no environment URLs specified → generates `Local` environment with `localhost:8080` and `localhost:50051`
- If any environment URL specified → only generates those environments (no default Local)

### Multi-Module Workspaces

When working with buf workspaces that have multiple modules, you may see "duplicate generated file" warnings for `bruno.json`. This happens because buf invokes the plugin once per module.

**Don't worry - these warnings are completely harmless!** Buf automatically uses the first `bruno.json` and drops duplicates. Your Bruno collection will work perfectly:

```yaml
version: v2
plugins:
  - local: protoc-gen-bruno
    out: bruno/collections
    opt:
      - collection_name=My API  # Optional
      - mode=all
```

The warnings don't affect functionality - everything works fine!

### Separate Collections Per Package

By default, all services are combined into a single collection. You can generate separate collections per proto package:

```yaml
version: v2
plugins:
  - local: protoc-gen-bruno
    out: bruno/collections
    opt:
      - single_collection=false  # Separate collection per package
```

This creates subdirectories like `example_v1/`, `myapp_v2/` based on the proto package names.

### Available Options

- **collection_name** - Custom collection name (default: auto-generated from services/package)
- **mode** - Generation mode: `all`, `http`, or `grpc` (default: `all`)
- **single_collection** - Combine all services in one collection: `true` or `false` (default: `true`)
- **dev_url** - Development environment base URL (e.g., `https://api.dev.example.com/service`)
- **stg_url** - Staging environment base URL
- **prd_url** - Production environment base URL
- **local_url** - Local environment base URL (default: `http://localhost:8080`)

## Generated Structure

```
bruno/collections/
├── bruno.json                    # Collection config
├── environments/
│   ├── Local.bru                 # Default local environment
│   ├── Development.bru           # If dev_url specified
│   ├── Staging.bru               # If stg_url specified
│   └── Production.bru            # If prd_url specified
├── UserService/                  # HTTP requests (from google.api.http)
│   ├── GetUser.bru
│   ├── CreateUser.bru
│   └── ...
└── UserService-gRPC/             # gRPC requests
    ├── GetUser.bru
    ├── CreateUser.bru
    └── ...
```

## Example Proto

```protobuf
syntax = "proto3";

package example.v1;

import "google/api/annotations.proto";

service UserService {
  rpc GetUser(GetUserRequest) returns (User) {
    option (google.api.http) = {
      get: "/v1/users/{user_id}"
    };
  }

  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse) {
    option (google.api.http) = {
      get: "/v1/users"
    };
  }

  rpc CreateUser(CreateUserRequest) returns (User) {
    option (google.api.http) = {
      post: "/v1/users"
      body: "*"
    };
  }
}

message User {
  string user_id = 1;
  string name = 2;
  string email = 3;
}

message GetUserRequest {
  string user_id = 1;
}

message ListUsersRequest {
  int32 page_size = 1;
  string page_token = 2;
}

message ListUsersResponse {
  repeated User users = 1;
  string next_page_token = 2;
}

message CreateUserRequest {
  string name = 1;
  string email = 2;
}
```

## Generated Output

### HTTP Request with Body (CreateUser.bru)
```
meta {
  name: CreateUser
  type: http
  seq: 1
}

post {
  url: {{base_url}}/v1/users
}

body:json {
  {
    "name": "example_name",
    "email": "example_email"
  }
}
```

### HTTP Request with Query Params (ListUsers.bru)
```
meta {
  name: ListUsers
  type: http
  seq: 1
}

get {
  url: {{base_url}}/v1/users
}

params:query {
  pageSize: 0
  pageToken: example_pageToken
}
```

**How it works:**
- **Path parameters** (like `{user_id}`) stay in the URL
- **GET/DELETE requests**: All non-path fields become query parameters
- **POST/PUT/PATCH with `body: "*"`**: All non-path fields go in the request body
- **POST/PUT/PATCH with specific body field**: That field goes in body, others become query params

### gRPC Request (CreateUser.bru)
```
meta {
  name: CreateUser
  type: grpc
  seq: 1
}

grpc {
  url: {{grpc_url}}
  method: example.v1.UserService/CreateUser
}

metadata {
}

body {
  {
    "name": "example_name",
    "email": "example_email"
  }
}
```

## Environment Variables

### Using Environments in Bruno

1. Open Bruno and load your generated collection
2. Click the **environment dropdown** in the top-right corner
3. Select the environment you want (Local, Development, Staging, Production)
4. All `{{base_url}}` and `{{grpc_url}}` variables will be replaced with the selected environment's values

### Generated Environments

**Default (no environment URLs specified):**
- `Local.bru` with `base_url: http://localhost:8080` and `grpc_url: localhost:50051`

**With environment URLs specified:**
- Each environment gets its own `.bru` file with appropriate URLs
- gRPC URLs are automatically derived from HTTP URLs (HTTPS → port 443, HTTP → port 80)

### Customizing Environments

You can manually edit environment files in `environments/` folder to:
- Add custom variables
- Modify URLs
- Add authentication tokens
- Configure headers
