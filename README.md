# protoc-gen-bruno

Generate Bruno API collections from Protocol Buffer definitions with gRPC-Gateway annotations.

## Features

- ✅ Generates Bruno HTTP requests from `google.api.http` annotations
- ✅ Generates Bruno gRPC requests from proto service definitions
- ✅ Auto-generates example JSON request bodies based on proto message types
- ✅ Supports nested messages, repeated fields, enums, and all proto types
- ✅ Configurable generation modes (HTTP only, gRPC only, or both)
- ✅ Custom or auto-generated collection names
- ✅ Environment configuration with variables
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

## Generated Structure

```
bruno/collections/
├── bruno.json                    # Collection config
├── environments/
│   └── Local.bru                 # Environment variables
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

message CreateUserRequest {
  string name = 1;
  string email = 2;
}
```

## Generated Output

### HTTP Request (CreateUser.bru)
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

The generator creates a `Local` environment with:

- `base_url`: `http://localhost:8080` (for HTTP requests)
- `grpc_url`: `localhost:50051` (for gRPC requests)

You can add more environments (Dev, Staging, Prod) by creating additional `.bru` files in the `environments/` folder.
