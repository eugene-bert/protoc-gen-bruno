package main

import (
	"fmt"
	"os"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func generateCollectionConfig(gen *protogen.Plugin, protoFiles []*protogen.File, prefix string, customName string, environments []environmentConfig, protoRoot string, preRequestScriptPath string, postRequestScriptPath string, authMode string, authTokenVar string) {
	collectionName := "API Collection"

	if customName != "" {
		collectionName = customName
	} else {
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
				if len(protoFiles) > 0 {
					pkg := string(protoFiles[0].Desc.Package())
					if pkg != "" {
						collectionName = formatPackageName(pkg) + " API"
					} else {
						collectionName = strings.Join(serviceNames, " & ") + " APIs"
					}
				}
			}
		}
	}

	var preRequestScript string
	if preRequestScriptPath != "" {
		scriptBytes, err := os.ReadFile(preRequestScriptPath)
		if err != nil {
			gen.Error(fmt.Errorf("warning: could not read pre-request script file %s: %v", preRequestScriptPath, err))
		} else {
			preRequestScript = string(scriptBytes)
		}
	}

	var postRequestScript string
	if postRequestScriptPath != "" {
		scriptBytes, err := os.ReadFile(postRequestScriptPath)
		if err != nil {
			gen.Error(fmt.Errorf("warning: could not read post-request script file %s: %v", postRequestScriptPath, err))
		} else {
			postRequestScript = string(scriptBytes)
		}
	}

	brunoConfig := gen.NewGeneratedFile(prefix+"bruno.json", "")
	brunoConfig.P("{")
	brunoConfig.P(`  "version": "1",`)
	brunoConfig.P(`  "name": "`, collectionName, `",`)

	hasScripts := preRequestScript != "" || postRequestScript != ""
	needsComma := (mode == modeAll || mode == modeGRPC)

	if needsComma {
		brunoConfig.P(`  "type": "collection",`)
	} else {
		brunoConfig.P(`  "type": "collection"`)
	}

	if mode == modeAll || mode == modeGRPC {
		brunoConfig.P(`  "protobuf": {`)
		brunoConfig.P(`    "proto": {`)
		brunoConfig.P(`      "root": "`, protoRoot, `"`)
		brunoConfig.P(`    }`)
		brunoConfig.P(`  }`)
	}

	brunoConfig.P("}")

	if hasScripts || authMode != "" {
		collectionBru := gen.NewGeneratedFile(prefix+"collection.bru", "")

		if authMode != "" {
			collectionBru.P("auth {")
			collectionBru.P("  mode: ", authMode)
			collectionBru.P("}")
			collectionBru.P("")

			switch authMode {
			case "bearer":
				collectionBru.P("auth:bearer {")
				collectionBru.P("  token: {{", authTokenVar, "}}")
				collectionBru.P("}")
			case "basic":
				collectionBru.P("auth:basic {")
				collectionBru.P("  username: {{username}}")
				collectionBru.P("  password: {{password}}")
				collectionBru.P("}")
			case "apikey":
				collectionBru.P("auth:apikey {")
				collectionBru.P("  key: {{api_key}}")
				collectionBru.P("  value: {{api_key_value}}")
				collectionBru.P("  placement: header")
				collectionBru.P("}")
			case "awsv4":
				collectionBru.P("auth:awsv4 {")
				collectionBru.P("  accessKeyId: {{aws_access_key_id}}")
				collectionBru.P("  secretAccessKey: {{aws_secret_access_key}}")
				collectionBru.P("  sessionToken: {{aws_session_token}}")
				collectionBru.P("  service: {{aws_service}}")
				collectionBru.P("  region: {{aws_region}}")
				collectionBru.P("}")
			}

			if hasScripts {
				collectionBru.P("")
			}
		}

		if preRequestScript != "" {
			collectionBru.P("script:pre-request {")
			for _, line := range strings.Split(preRequestScript, "\n") {
				collectionBru.P("  " + line)
			}
			collectionBru.P("}")
			if postRequestScript != "" {
				collectionBru.P("")
			}
		}

		if postRequestScript != "" {
			collectionBru.P("script:post-response {")
			for _, line := range strings.Split(postRequestScript, "\n") {
				collectionBru.P("  " + line)
			}
			collectionBru.P("}")
		}
	}

	for _, env := range environments {
		envFile := gen.NewGeneratedFile(prefix+"environments/"+env.name+".bru", "")
		envFile.P("vars {")

		if mode == modeAll || mode == modeHTTP {
			envFile.P("  base_url: ", env.httpURL)
		}
		if mode == modeAll || mode == modeGRPC {
			envFile.P("  grpc_url: ", env.grpcURL)
		}

		envFile.P("}")
	}
}

func formatPackageName(pkg string) string {
	parts := strings.Split(pkg, ".")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

func getServiceFolderName(serviceName string) string {
	if strings.ToLower(serviceName) == "environments" {
		return serviceName + "Service"
	}
	return serviceName
}
