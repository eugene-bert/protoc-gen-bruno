package main

import "testing"

func TestFormatPackageName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"example.v1", "Example V1"},
		{"myapp.api.v2", "Myapp Api V2"},
		{"single", "Single"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := formatPackageName(tt.input)
			if got != tt.want {
				t.Errorf("formatPackageName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetServiceFolderName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"UserService", "UserService"},
		{"OrderService", "OrderService"},
		{"Environments", "EnvironmentsService"},
		{"environments", "environmentsService"},
		{"ENVIRONMENTS", "ENVIRONMENTSService"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := getServiceFolderName(tt.input)
			if got != tt.want {
				t.Errorf("getServiceFolderName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestUrlToGrpcHost(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://api.example.com", "api.example.com:443"},
		{"https://api.example.com/service", "api.example.com:443"},
		{"http://localhost:8080", "localhost:8080"},
		{"http://localhost", "localhost:80"},
		{"https://api.dev.example.com:9443", "api.dev.example.com:9443"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := urlToGrpcHost(tt.input)
			if got != tt.want {
				t.Errorf("urlToGrpcHost(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildEnvironments(t *testing.T) {
	t.Run("defaults to local", func(t *testing.T) {
		envs := buildEnvironments("", "", "", "", "", "", "", "")
		if len(envs) != 1 {
			t.Fatalf("got %d envs, want 1", len(envs))
		}
		if envs[0].name != "Local" {
			t.Errorf("name = %q, want Local", envs[0].name)
		}
		if envs[0].httpURL != "http://localhost:8080" {
			t.Errorf("httpURL = %q, want http://localhost:8080", envs[0].httpURL)
		}
	})

	t.Run("custom local", func(t *testing.T) {
		envs := buildEnvironments("http://localhost:3000", "", "", "", "", "", "", "")
		if envs[0].httpURL != "http://localhost:3000" {
			t.Errorf("httpURL = %q, want http://localhost:3000", envs[0].httpURL)
		}
	})

	t.Run("all environments", func(t *testing.T) {
		envs := buildEnvironments("", "https://dev.api.com", "https://stg.api.com", "https://prd.api.com", "", "", "", "")
		if len(envs) != 3 {
			t.Fatalf("got %d envs, want 3", len(envs))
		}
		if envs[0].name != "Development" {
			t.Errorf("envs[0].name = %q, want Development", envs[0].name)
		}
	})

	t.Run("grpc overrides", func(t *testing.T) {
		envs := buildEnvironments("http://localhost:8080", "", "", "", "localhost:50051", "", "", "")
		if envs[0].grpcURL != "localhost:50051" {
			t.Errorf("grpcURL = %q, want localhost:50051", envs[0].grpcURL)
		}
	})
}
