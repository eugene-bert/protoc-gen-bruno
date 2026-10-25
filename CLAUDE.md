# protoc-gen-bruno

protoc plugin (Go) that generates [Bruno](https://www.usebruno.com/) API collections from Protocol Buffer definitions with gRPC-Gateway HTTP annotations.

## Commands

```bash
go build -v ./...       # build
go vet ./...            # lint
go test -v -race ./...  # test
go install .            # install plugin to $GOPATH/bin
```

## CI/CD

- **CI** (`ci.yml`): build, vet, test across Go 1.25 and 1.26 on push/PR to main
- **Dependabot**: auto-merge patch/minor updates
- No release CI — installed via `go install`

## Usage

```bash
protoc --bruno_out=. --bruno_opt=base_url=http://localhost:8080 your_service.proto
```

## Key files

- `main.go` — protoc plugin entrypoint, reads CodeGeneratorRequest from stdin
- Parses `google.api.http` annotations to extract HTTP methods/paths
- Outputs `.bru` files compatible with Bruno API client
- Supports both HTTP (REST via gRPC-Gateway) and gRPC request types

## Important

- This is a **protoc plugin** — runs as `protoc-gen-bruno` binary invoked by `protoc`
- Requires protobuf + gRPC-Gateway annotations in `.proto` files
- Test matrix: Go 1.25 + 1.26
