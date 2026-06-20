package main

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

func generateGrpcRequest(gen *protogen.Plugin, service *protogen.Service, method *protogen.Method, file *protogen.File, prefix string) error {
	serviceFolderName := getServiceFolderName(service.GoName)
	filename := fmt.Sprintf("%s%s-gRPC/%s.bru", prefix, serviceFolderName, method.GoName)
	g := gen.NewGeneratedFile(filename, "")

	grpcMethod := fmt.Sprintf("%s.%s/%s", file.Desc.Package(), service.Desc.Name(), method.Desc.Name())
	protoFilePath := file.Desc.Path()

	g.P("meta {")
	g.P("  name: ", method.GoName)
	g.P("  type: grpc")
	g.P("  seq: 1")
	g.P("}")
	g.P("")
	g.P("grpc {")
	g.P("  url: {{grpc_url}}")
	g.P("  method: ", grpcMethod)
	if collectionAuthMode != "" {
		g.P("  auth: inherit")
	}
	g.P("}")
	g.P("")
	g.P("metadata {")
	g.P("}")
	g.P("")
	g.P("body {")
	exampleJSON := generateExampleJSON(method.Input, 1)
	g.P(exampleJSON)
	g.P("}")
	g.P("")
	g.P("script:pre-request {")
	g.P("  // Proto file: ", protoFilePath)
	g.P("}")

	return nil
}
