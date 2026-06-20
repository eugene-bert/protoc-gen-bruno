package main

import (
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func generateBrunoRequest(gen *protogen.Plugin, service *protogen.Service, method *protogen.Method, prefix string) error {
	opts := method.Desc.Options()
	if !proto.HasExtension(opts, annotations.E_Http) {
		return nil
	}

	httpRule := proto.GetExtension(opts, annotations.E_Http).(*annotations.HttpRule)

	httpMethod, path := extractHTTPRule(httpRule)
	if httpMethod == "" || path == "" {
		return nil
	}

	pathParams := extractPathParams(path)

	serviceFolderName := getServiceFolderName(service.GoName)
	filename := fmt.Sprintf("%s%s/%s.bru", prefix, serviceFolderName, method.GoName)
	g := gen.NewGeneratedFile(filename, "")

	g.P("meta {")
	g.P("  name: ", method.GoName)
	g.P("  type: http")
	g.P("  seq: 1")
	g.P("}")
	g.P("")
	g.P(httpMethod, " {")
	g.P("  url: {{base_url}}", path)
	g.P("  body: none")
	if collectionAuthMode != "" {
		g.P("  auth: inherit")
	}
	g.P("}")

	var queryFields []*protogen.Field
	var bodyFields []*protogen.Field

	for _, field := range method.Input.Fields {
		fieldName := string(field.Desc.Name())

		if isPathParam(fieldName, pathParams) {
			continue
		}

		if httpMethod == "get" || httpMethod == "delete" {
			queryFields = append(queryFields, field)
		} else {
			bodyFieldName := httpRule.Body
			if bodyFieldName == "*" {
				bodyFields = append(bodyFields, field)
			} else if bodyFieldName == fieldName {
				bodyFields = append(bodyFields, field)
			} else if bodyFieldName == "" {
				queryFields = append(queryFields, field)
			} else {
				queryFields = append(queryFields, field)
			}
		}
	}

	if len(queryFields) > 0 {
		g.P("")
		g.P("params:query {")
		for _, field := range queryFields {
			value := generateFieldValue(field, 0)
			value = strings.Trim(value, `"`)
			g.P("  ", field.Desc.JSONName(), ": ", value)
		}
		g.P("}")
	}

	if len(bodyFields) > 0 {
		g.P("")
		g.P("body:json {")

		if httpRule.Body == "*" {
			exampleJSON := generateExampleJSON(method.Input, 1)
			g.P(exampleJSON)
		} else {
			exampleJSON := generateExampleJSON(bodyFields[0].Message, 1)
			g.P(exampleJSON)
		}

		g.P("}")
	}

	return nil
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

func extractPathParams(path string) []string {
	var params []string
	start := -1

	for i, ch := range path {
		if ch == '{' {
			start = i + 1
		} else if ch == '}' && start != -1 {
			param := path[start:i]
			if idx := strings.Index(param, "="); idx != -1 {
				param = param[:idx]
			}
			params = append(params, param)
			start = -1
		}
	}

	return params
}

func isPathParam(fieldName string, pathParams []string) bool {
	for _, param := range pathParams {
		if fieldName == param {
			return true
		}
	}
	return false
}
