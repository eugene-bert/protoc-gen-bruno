package main

import (
	"flag"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

type generationMode string

const (
	modeAll  generationMode = "all"
	modeHTTP generationMode = "http"
	modeGRPC generationMode = "grpc"
)

var (
	mode               = modeAll
	collectionAuthMode = ""
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
	var preRequestScriptPath string
	var postRequestScriptPath string
	var authMode string
	var authTokenVar string

	flags.StringVar(&modeFlag, "mode", "all", "Generation mode: all, http, or grpc")
	flags.StringVar(&singleCollectionFlag, "single_collection", "true", "Generate a single collection for all modules")
	flags.StringVar(&collectionNameFlag, "collection_name", "", "Custom collection name (defaults to auto-generated from services)")
	flags.StringVar(&devURL, "dev_url", "", "Development environment base URL")
	flags.StringVar(&stgURL, "stg_url", "", "Staging environment base URL")
	flags.StringVar(&prdURL, "prd_url", "", "Production environment base URL")
	flags.StringVar(&localURL, "local_url", "", "Local environment base URL (defaults to http://localhost:8080)")
	flags.StringVar(&grpcDevURL, "grpc_dev_url", "", "Development gRPC URL override")
	flags.StringVar(&grpcStgURL, "grpc_stg_url", "", "Staging gRPC URL override")
	flags.StringVar(&grpcPrdURL, "grpc_prd_url", "", "Production gRPC URL override")
	flags.StringVar(&grpcLocalURL, "grpc_local_url", "", "Local gRPC URL override")
	flags.StringVar(&protoRootFlag, "proto_root", "../../proto", "Path to proto files root directory relative to bruno/collections")
	flags.StringVar(&preRequestScriptPath, "pre_request_script", "", "Path to pre-request script file")
	flags.StringVar(&postRequestScriptPath, "post_request_script", "", "Path to post-request script file")
	flags.StringVar(&authMode, "auth_mode", "", "Authentication mode: bearer, basic, apikey, or awsv4")
	flags.StringVar(&authTokenVar, "auth_token_var", "bearer_token", "Variable name for bearer token")

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		switch modeFlag {
		case "all", "http", "grpc":
			mode = generationMode(modeFlag)
		default:
			mode = modeAll
		}

		collectionAuthMode = authMode
		singleCollection := singleCollectionFlag != "false"

		environments := buildEnvironments(localURL, devURL, stgURL, prdURL, grpcLocalURL, grpcDevURL, grpcStgURL, grpcPrdURL)

		for _, f := range gen.Files {
			if f.Generate {
				protoFiles = append(protoFiles, f)
			}
		}

		configGenerated := make(map[string]bool)

		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			collectionPrefix := ""
			if !singleCollection && len(f.Services) > 0 {
				pkg := string(f.Desc.Package())
				if pkg != "" {
					collectionPrefix = strings.ReplaceAll(pkg, ".", "_") + "/"
				}
			}

			if len(f.Services) > 0 && !configGenerated[collectionPrefix] {
				generateCollectionConfig(gen, protoFiles, collectionPrefix, collectionNameFlag, environments, protoRootFlag, preRequestScriptPath, postRequestScriptPath, authMode, authTokenVar)
				configGenerated[collectionPrefix] = true
			}

			generateBrunoCollection(gen, f, collectionPrefix)
		}
		return nil
	})
}

func buildEnvironments(localURL, devURL, stgURL, prdURL, grpcLocalURL, grpcDevURL, grpcStgURL, grpcPrdURL string) []environmentConfig {
	var environments []environmentConfig

	if localURL != "" || (devURL == "" && stgURL == "" && prdURL == "") {
		baseURL := "http://localhost:8080"
		grpcURL := "localhost:50051"
		if localURL != "" {
			baseURL = localURL
			grpcURL = urlToGrpcHost(localURL)
		}
		if grpcLocalURL != "" {
			grpcURL = grpcLocalURL
		}
		environments = append(environments, environmentConfig{name: "Local", httpURL: baseURL, grpcURL: grpcURL})
	}

	type envDef struct {
		url, grpcOverride, name string
	}
	for _, e := range []envDef{
		{devURL, grpcDevURL, "Development"},
		{stgURL, grpcStgURL, "Staging"},
		{prdURL, grpcPrdURL, "Production"},
	} {
		if e.url != "" {
			grpcURL := urlToGrpcHost(e.url)
			if e.grpcOverride != "" {
				grpcURL = e.grpcOverride
			}
			environments = append(environments, environmentConfig{name: e.name, httpURL: e.url, grpcURL: grpcURL})
		}
	}

	return environments
}

func urlToGrpcHost(httpURL string) string {
	url := strings.TrimPrefix(httpURL, "https://")
	url = strings.TrimPrefix(url, "http://")

	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}

	if !strings.Contains(url, ":") {
		if strings.HasPrefix(httpURL, "https://") {
			url += ":443"
		} else {
			url += ":80"
		}
	}

	return url
}
