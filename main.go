package main

import (
	"flag"
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/pluginpb"
)

type generationMode string

const (
	modeAll  generationMode = "all"
	modeHTTP generationMode = "http"
	modeGRPC generationMode = "grpc"
)

var (
	mode = modeAll
)

type environmentConfig struct {
	name    string
	httpURL string
	grpcURL string
}

func main() {
	var flags flag.FlagSet
	var protoFiles []*protogen.File
	var modeFlag string
	var singleCollectionFlag string
	var collectionNameFlag string
	var devURL, stgURL, prdURL, localURL string
	var grpcDevURL, grpcStgURL, grpcPrdURL, grpcLocalURL string
	var protoRootFlag string

	flags.StringVar(&modeFlag, "mode", "all", "Generation mode: all, http, or grpc")
	flags.StringVar(&singleCollectionFlag, "single_collection", "true", "Generate a single collection for all modules")
	flags.StringVar(&collectionNameFlag, "collection_name", "", "Custom collection name (defaults to auto-generated from services)")
	flags.StringVar(&devURL, "dev_url", "", "Development environment base URL (e.g., https://api.dev.example.com/service)")
	flags.StringVar(&stgURL, "stg_url", "", "Staging environment base URL")
	flags.StringVar(&prdURL, "prd_url", "", "Production environment base URL")
	flags.StringVar(&localURL, "local_url", "", "Local environment base URL (defaults to http://localhost:8080)")
	flags.StringVar(&grpcDevURL, "grpc_dev_url", "", "Development gRPC URL (e.g., api.dev.example.com:443) - overrides auto-generated from dev_url")
	flags.StringVar(&grpcStgURL, "grpc_stg_url", "", "Staging gRPC URL - overrides auto-generated from stg_url")
	flags.StringVar(&grpcPrdURL, "grpc_prd_url", "", "Production gRPC URL - overrides auto-generated from prd_url")
	flags.StringVar(&grpcLocalURL, "grpc_local_url", "", "Local gRPC URL (e.g., localhost:50051) - overrides auto-generated from local_url")
	flags.StringVar(&protoRootFlag, "proto_root", "../../proto", "Path to proto files root directory relative to bruno/collections (e.g., ../../api/proto/src)")

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		// Parse and validate mode flag
		switch modeFlag {
		case "all", "http", "grpc":
			mode = generationMode(modeFlag)
		default:
			mode = modeAll
		}

		singleCollection := singleCollectionFlag != "false"

		// Build environment configurations
		var environments []environmentConfig
		if localURL != "" || (devURL == "" && stgURL == "" && prdURL == "") {
			// Add Local environment (default or custom)
			baseURL := "http://localhost:8080"
			grpcURL := "localhost:50051"
			if localURL != "" {
				baseURL = localURL
				grpcURL = urlToGrpcHost(localURL)
			}
			// Override with explicit gRPC URL if provided
			if grpcLocalURL != "" {
				grpcURL = grpcLocalURL
			}
			environments = append(environments, environmentConfig{
				name:    "Local",
				httpURL: baseURL,
				grpcURL: grpcURL,
			})
		}
		if devURL != "" {
			grpcURL := urlToGrpcHost(devURL)
			// Override with explicit gRPC URL if provided
			if grpcDevURL != "" {
				grpcURL = grpcDevURL
			}
			environments = append(environments, environmentConfig{
				name:    "Development",
				httpURL: devURL,
				grpcURL: grpcURL,
			})
		}
		if stgURL != "" {
			grpcURL := urlToGrpcHost(stgURL)
			// Override with explicit gRPC URL if provided
			if grpcStgURL != "" {
				grpcURL = grpcStgURL
			}
			environments = append(environments, environmentConfig{
				name:    "Staging",
				httpURL: stgURL,
				grpcURL: grpcURL,
			})
		}
		if prdURL != "" {
			grpcURL := urlToGrpcHost(prdURL)
			// Override with explicit gRPC URL if provided
			if grpcPrdURL != "" {
				grpcURL = grpcPrdURL
			}
			environments = append(environments, environmentConfig{
				name:    "Production",
				httpURL: prdURL,
				grpcURL: grpcURL,
			})
		}

		// Collect all proto files first
		for _, f := range gen.Files {
			if f.Generate {
				protoFiles = append(protoFiles, f)
			}
		}

		// Generate config - either once for single collection or per module
		configGenerated := make(map[string]bool)

		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			// Determine collection path prefix
			collectionPrefix := ""
			if !singleCollection && len(f.Services) > 0 {
				// Use package name as collection subfolder
				pkg := string(f.Desc.Package())
				if pkg != "" {
					collectionPrefix = strings.ReplaceAll(pkg, ".", "_") + "/"
				}
			}

			// Generate config once per collection
			if len(f.Services) > 0 && !configGenerated[collectionPrefix] {
				generateCollectionConfigWithPrefix(gen, protoFiles, collectionPrefix, collectionNameFlag, environments, protoRootFlag)
				configGenerated[collectionPrefix] = true
			}

			generateBrunoCollectionWithPrefix(gen, f, collectionPrefix)
		}
		return nil
	})
}

// urlToGrpcHost converts an HTTP(S) URL to a gRPC host:port
// Examples:
//
//	https://api.dev.example.com/service -> api.dev.example.com:443
//	http://localhost:8080 -> localhost:8080
func urlToGrpcHost(httpURL string) string {
	// Remove protocol
	url := strings.TrimPrefix(httpURL, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Remove path if present
	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}

	// Add default port if not present
	if !strings.Contains(url, ":") {
		if strings.HasPrefix(httpURL, "https://") {
			url += ":443"
		} else {
			url += ":80"
		}
	}

	return url
}

func generateCollectionConfigWithPrefix(gen *protogen.Plugin, protoFiles []*protogen.File, prefix string, customName string, environments []environmentConfig, protoRoot string) {
	generateCollectionConfig(gen, protoFiles, prefix, customName, environments, protoRoot)
}

func generateCollectionConfig(gen *protogen.Plugin, protoFiles []*protogen.File, prefix string, customName string, environments []environmentConfig, protoRoot string) {
	// Use custom name if provided, otherwise auto-generate
	collectionName := "API Collection"

	if customName != "" {
		collectionName = customName
	} else {
		// Build collection name from services
		var serviceNames []string

		for _, f := range protoFiles {
			for _, service := range f.Services {
				serviceNames = append(serviceNames, service.GoName)
			}
		}

		if len(serviceNames) > 0 {
			if len(serviceNames) == 1 {
				collectionName = serviceNames[0] + " API"
			} else {
				// Use the package name or first service for multiple services
				if len(protoFiles) > 0 {
					pkg := string(protoFiles[0].Desc.Package())
					if pkg != "" {
						// Convert package name like "example.v1" to "Example V1 API"
						collectionName = formatPackageName(pkg) + " API"
					} else {
						collectionName = strings.Join(serviceNames, " & ") + " APIs"
					}
				}
			}
		}
	}

	// Generate bruno.json with proto paths
	brunoConfig := gen.NewGeneratedFile(prefix+"bruno.json", "")
	brunoConfig.P("{")
	brunoConfig.P(`  "version": "1",`)
	brunoConfig.P(`  "name": "`, collectionName, `",`)

	// Add comma after "type" if we're adding grpc config
	if mode == modeAll || mode == modeGRPC {
		brunoConfig.P(`  "type": "collection",`)
		brunoConfig.P(`  "grpc": {`)
		brunoConfig.P(`    "proto": {`)
		brunoConfig.P(`      "root": "`, protoRoot, `"`)
		brunoConfig.P(`    }`)
		brunoConfig.P(`  }`)
	} else {
		brunoConfig.P(`  "type": "collection"`)
	}

	brunoConfig.P("}")

	// Generate environment files for each configured environment
	for _, env := range environments {
		envFile := gen.NewGeneratedFile(prefix+"_environments/"+env.name+".bru", "")
		envFile.P("vars {")

		// Add relevant environment variables based on mode
		if mode == modeAll || mode == modeHTTP {
			envFile.P("  base_url: ", env.httpURL)
		}
		if mode == modeAll || mode == modeGRPC {
			envFile.P("  grpc_url: ", env.grpcURL)
		}

		envFile.P("}")
	}
}

// formatPackageName converts "example.v1" to "Example V1"
func formatPackageName(pkg string) string {
	parts := strings.Split(pkg, ".")
	for i, part := range parts {
		// Capitalize first letter
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

func generateBrunoCollectionWithPrefix(gen *protogen.Plugin, file *protogen.File, prefix string) error {
	return generateBrunoCollection(gen, file, prefix)
}

func generateBrunoCollection(gen *protogen.Plugin, file *protogen.File, prefix string) error {
	// We'll iterate through services and their methods
	for _, service := range file.Services {
		// For each service, create a Bruno collection folder
		// and generate .bru files for each RPC method
		for _, method := range service.Methods {
			// Generate HTTP request (if mode allows and it has HTTP annotations)
			if mode == modeAll || mode == modeHTTP {
				if err := generateBrunoRequest(gen, service, method, prefix); err != nil {
					return err
				}
			}
			// Generate gRPC request (if mode allows)
			if mode == modeAll || mode == modeGRPC {
				if err := generateGrpcRequest(gen, service, method, file, prefix); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func generateBrunoRequest(gen *protogen.Plugin, service *protogen.Service, method *protogen.Method, prefix string) error {
	// Extract HTTP annotation from method options
	opts := method.Desc.Options()
	if !proto.HasExtension(opts, annotations.E_Http) {
		// Skip methods without HTTP annotations
		return nil
	}

	httpRule := proto.GetExtension(opts, annotations.E_Http).(*annotations.HttpRule)

	httpMethod, path := extractHTTPRule(httpRule)
	if httpMethod == "" || path == "" {
		// Skip if we can't determine HTTP method or path
		return nil
	}

	// Extract path parameters from URL (e.g., {user_id}, {name})
	pathParams := extractPathParams(path)

	filename := fmt.Sprintf("%s%s/%s.bru", prefix, service.GoName, method.GoName)
	g := gen.NewGeneratedFile(filename, "")

	// Generate Bruno file format
	g.P("meta {")
	g.P("  name: ", method.GoName)
	g.P("  type: http")
	g.P("  seq: 1")
	g.P("}")
	g.P("")
	g.P(httpMethod, " {")
	g.P("  url: {{base_url}}", path)
	g.P("}")

	// Determine which fields should be query params vs body
	var queryFields []*protogen.Field
	var bodyFields []*protogen.Field

	for _, field := range method.Input.Fields {
		fieldName := string(field.Desc.Name())

		// Skip path parameters
		if isPathParam(fieldName, pathParams) {
			continue
		}

		// For GET/DELETE, all non-path fields become query params
		if httpMethod == "get" || httpMethod == "delete" {
			queryFields = append(queryFields, field)
		} else {
			// For POST/PUT/PATCH, check the body field
			bodyFieldName := httpRule.Body
			if bodyFieldName == "*" {
				// All non-path fields go in body
				bodyFields = append(bodyFields, field)
			} else if bodyFieldName == fieldName {
				// This specific field goes in body
				bodyFields = append(bodyFields, field)
			} else if bodyFieldName == "" {
				// No body specified, treat like GET (query params)
				queryFields = append(queryFields, field)
			} else {
				// Other fields become query params
				queryFields = append(queryFields, field)
			}
		}
	}

	// Generate query parameters section
	if len(queryFields) > 0 {
		g.P("")
		g.P("params:query {")
		for _, field := range queryFields {
			value := generateFieldValue(field, 0)
			// Remove quotes from string values for query params
			value = strings.Trim(value, `"`)
			g.P("  ", field.Desc.JSONName(), ": ", value)
		}
		g.P("}")
	}

	// Add request body if needed
	if len(bodyFields) > 0 {
		g.P("")
		g.P("body:json {")

		// Generate JSON for body fields only
		if httpRule.Body == "*" {
			// All fields in body
			exampleJSON := generateExampleJSON(method.Input, 1)
			g.P(exampleJSON)
		} else {
			// Specific field in body
			exampleJSON := generateExampleJSON(bodyFields[0].Message, 1)
			g.P(exampleJSON)
		}

		g.P("}")
	}

	return nil
}

// extractPathParams extracts parameter names from a URL path
// Examples:
//
//	"/v1/users/{user_id}/posts/{post_id}" -> ["user_id", "post_id"]
//	"/v1alpha1/{name=environments/*/contact}" -> ["name"]
//	"/v1alpha1/{parent=accounts/*}/environments" -> ["parent"]
func extractPathParams(path string) []string {
	var params []string
	start := -1

	for i, ch := range path {
		if ch == '{' {
			start = i + 1
		} else if ch == '}' && start != -1 {
			param := path[start:i]
			// Handle resource patterns like {name=environments/*/contact}
			// Extract just the parameter name before the '='
			if idx := strings.Index(param, "="); idx != -1 {
				param = param[:idx]
			}
			params = append(params, param)
			start = -1
		}
	}

	return params
}

// isPathParam checks if a field name matches any path parameter
func isPathParam(fieldName string, pathParams []string) bool {
	for _, param := range pathParams {
		if fieldName == param {
			return true
		}
	}
	return false
}

func extractHTTPRule(rule *annotations.HttpRule) (method, path string) {
	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		return "get", pattern.Get
	case *annotations.HttpRule_Post:
		return "post", pattern.Post
	case *annotations.HttpRule_Put:
		return "put", pattern.Put
	case *annotations.HttpRule_Delete:
		return "delete", pattern.Delete
	case *annotations.HttpRule_Patch:
		return "patch", pattern.Patch
	}
	return "", ""
}

func generateGrpcRequest(gen *protogen.Plugin, service *protogen.Service, method *protogen.Method, file *protogen.File, prefix string) error {
	// Generate gRPC .bru file in a gRPC subfolder
	filename := fmt.Sprintf("%s%s-gRPC/%s.bru", prefix, service.GoName, method.GoName)
	g := gen.NewGeneratedFile(filename, "")

	// Construct the full gRPC method name: package.Service/Method
	grpcMethod := fmt.Sprintf("%s.%s/%s", file.Desc.Package(), service.Desc.Name(), method.Desc.Name())

	// Get proto file path relative to workspace
	protoFilePath := file.Desc.Path()

	// Generate Bruno gRPC file format
	g.P("meta {")
	g.P("  name: ", method.GoName)
	g.P("  type: grpc")
	g.P("  seq: 1")
	g.P("}")
	g.P("")
	g.P("grpc {")
	g.P("  url: {{grpc_url}}")
	g.P("  method: ", grpcMethod)
	g.P("}")
	g.P("")
	g.P("metadata {")
	g.P("}")
	g.P("")
	g.P("body {")
	// Generate example JSON from the request message
	exampleJSON := generateExampleJSON(method.Input, 1)
	g.P(exampleJSON)
	g.P("}")
	g.P("")
	g.P("script:pre-request {")
	g.P("  // Proto file: ", protoFilePath)
	g.P("}")

	return nil
}

// generateExampleJSON creates example JSON for a proto message
// Maximum nesting depth to prevent infinite recursion
const maxDepth = 3

func generateExampleJSON(msg *protogen.Message, indent int) string {
	// Prevent infinite recursion by limiting depth
	if indent > maxDepth {
		return "{}"
	}

	var lines []string
	indentStr := strings.Repeat("  ", indent)

	lines = append(lines, "{")

	for i, field := range msg.Fields {
		fieldIndent := strings.Repeat("  ", indent+1)
		jsonName := field.Desc.JSONName()

		// Generate value based on field type
		var value string
		if field.Desc.IsList() {
			// Handle repeated fields (arrays)
			value = "[" + generateFieldValue(field, indent+1) + "]"
		} else {
			value = generateFieldValue(field, indent+1)
		}

		line := fmt.Sprintf(`%s"%s": %s`, fieldIndent, jsonName, value)
		if i < len(msg.Fields)-1 {
			line += ","
		}
		lines = append(lines, line)
	}

	lines = append(lines, indentStr+"}")
	return strings.Join(lines, "\n")
}

// generateFieldValue generates an example value for a field
func generateFieldValue(field *protogen.Field, indent int) string {
	kind := field.Desc.Kind()

	switch kind {
	case protoreflect.StringKind:
		// Use field name as example value
		return fmt.Sprintf(`"example_%s"`, field.Desc.JSONName())
	case protoreflect.Int32Kind, protoreflect.Int64Kind,
		protoreflect.Uint32Kind, protoreflect.Uint64Kind,
		protoreflect.Sint32Kind, protoreflect.Sint64Kind,
		protoreflect.Fixed32Kind, protoreflect.Fixed64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind:
		return "0"
	case protoreflect.BoolKind:
		return "false"
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return "0.0"
	case protoreflect.BytesKind:
		return `"base64_encoded_data"`
	case protoreflect.EnumKind:
		// Get first enum value
		enum := field.Enum
		if enum != nil && len(enum.Values) > 0 {
			return fmt.Sprintf(`"%s"`, enum.Values[0].Desc.Name())
		}
		return `"ENUM_VALUE"`
	case protoreflect.MessageKind:
		// Handle well-known types with special JSON representations
		if field.Message != nil {
			// Check for well-known types that have special JSON serialization
			fullName := string(field.Message.Desc.FullName())
			switch fullName {
			case "google.protobuf.Timestamp":
				return `"2024-01-01T00:00:00Z"`
			case "google.protobuf.Duration":
				return `"1.5s"`
			case "google.protobuf.Any":
				return `{"@type": "type.googleapis.com/example.Type", "value": "..."}`
			case "google.protobuf.FieldMask":
				return `"field1,field2.subfield"`
			case "google.protobuf.Struct":
				return `{}`
			case "google.protobuf.Value":
				return `null`
			case "google.protobuf.ListValue":
				return `[]`
			case "google.protobuf.Empty":
				return `{}`
			default:
				// For other message types, recursively generate JSON
				return generateExampleJSON(field.Message, indent)
			}
		}
		return "{}"
	default:
		return `"unknown"`
	}
}
