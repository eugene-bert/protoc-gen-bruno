package main

import (
	"fmt"
	"strings"

	openapiv2 "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const maxDepth = 3

func generateExampleJSON(msg *protogen.Message, indent int) string {
	if indent > maxDepth {
		return "{}"
	}

	var lines []string
	indentStr := strings.Repeat("  ", indent)

	lines = append(lines, "{")

	for i, field := range msg.Fields {
		fieldIndent := strings.Repeat("  ", indent+1)
		jsonName := field.Desc.JSONName()

		var value string
		if field.Desc.IsList() {
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

func generateFieldValue(field *protogen.Field, indent int) string {
	kind := field.Desc.Kind()

	switch kind {
	case protoreflect.StringKind:
		if example := getOpenAPIExample(field); example != "" {
			return example
		}
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
		enum := field.Enum
		if enum != nil && len(enum.Values) > 0 {
			return fmt.Sprintf(`"%s"`, enum.Values[0].Desc.Name())
		}
		return `"ENUM_VALUE"`
	case protoreflect.MessageKind:
		if field.Message != nil {
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
				return generateExampleJSON(field.Message, indent)
			}
		}
		return "{}"
	default:
		return `"unknown"`
	}
}

func getOpenAPIExample(field *protogen.Field) string {
	opts := field.Desc.Options()
	if opts == nil {
		return ""
	}

	if !proto.HasExtension(opts, openapiv2.E_Openapiv2Field) {
		return ""
	}

	fieldOpts := proto.GetExtension(opts, openapiv2.E_Openapiv2Field).(*openapiv2.JSONSchema)
	if fieldOpts == nil {
		return ""
	}

	if fieldOpts.Example != "" {
		return fieldOpts.Example
	}

	return ""
}
