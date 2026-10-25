# Copilot Instructions

Go protoc plugin: generates Bruno API collections from .proto files with gRPC-Gateway annotations.

## Build & Test

```bash
go build -v ./...
go vet ./...
go test -v -race ./...
go install .
```

## Architecture

- `main.go` — protoc plugin, reads CodeGeneratorRequest from stdin
- Parses `google.api.http` annotations → outputs `.bru` files
- Supports HTTP (REST) and gRPC request types
- Test matrix: Go 1.25 + 1.26
