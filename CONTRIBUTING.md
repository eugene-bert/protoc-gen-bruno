# Contributing

## Setup

```bash
go build -v ./...
go test -v -race ./...
```

## Stack

- Go protoc plugin
- Reads CodeGeneratorRequest (protobuf stdin), outputs .bru files
- Requires protobuf compiler (`protoc`) for integration testing

## CI

Tests run on Go 1.25 and 1.26. Both must pass.

## PR Flow

1. Fork & branch from `main`
2. `go vet ./...` and `go test -v -race ./...` must pass
3. CI runs automatically on PR
