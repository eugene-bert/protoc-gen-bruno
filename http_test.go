package main

import (
	"reflect"
	"testing"

	"google.golang.org/genproto/googleapis/api/annotations"
)

func TestExtractHTTPRule(t *testing.T) {
	tests := []struct {
		name       string
		rule       *annotations.HttpRule
		wantMethod string
		wantPath   string
	}{
		{
			name:       "GET",
			rule:       &annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/users/{id}"}},
			wantMethod: "get",
			wantPath:   "/v1/users/{id}",
		},
		{
			name:       "POST",
			rule:       &annotations.HttpRule{Pattern: &annotations.HttpRule_Post{Post: "/v1/users"}},
			wantMethod: "post",
			wantPath:   "/v1/users",
		},
		{
			name:       "PUT",
			rule:       &annotations.HttpRule{Pattern: &annotations.HttpRule_Put{Put: "/v1/users/{id}"}},
			wantMethod: "put",
			wantPath:   "/v1/users/{id}",
		},
		{
			name:       "DELETE",
			rule:       &annotations.HttpRule{Pattern: &annotations.HttpRule_Delete{Delete: "/v1/users/{id}"}},
			wantMethod: "delete",
			wantPath:   "/v1/users/{id}",
		},
		{
			name:       "PATCH",
			rule:       &annotations.HttpRule{Pattern: &annotations.HttpRule_Patch{Patch: "/v1/users/{id}"}},
			wantMethod: "patch",
			wantPath:   "/v1/users/{id}",
		},
		{
			name:       "nil pattern",
			rule:       &annotations.HttpRule{},
			wantMethod: "",
			wantPath:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, path := extractHTTPRule(tt.rule)
			if method != tt.wantMethod {
				t.Errorf("method = %q, want %q", method, tt.wantMethod)
			}
			if path != tt.wantPath {
				t.Errorf("path = %q, want %q", path, tt.wantPath)
			}
		})
	}
}

func TestExtractPathParams(t *testing.T) {
	tests := []struct {
		path string
		want []string
	}{
		{"/v1/users/{user_id}", []string{"user_id"}},
		{"/v1/users/{user_id}/posts/{post_id}", []string{"user_id", "post_id"}},
		{"/v1alpha1/{name=environments/*/contact}", []string{"name"}},
		{"/v1alpha1/{parent=accounts/*}/environments", []string{"parent"}},
		{"/v1/users", nil},
		{"/{a}/{b}/{c}", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := extractPathParams(tt.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractPathParams(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsPathParam(t *testing.T) {
	params := []string{"user_id", "post_id"}

	if !isPathParam("user_id", params) {
		t.Error("expected user_id to be a path param")
	}
	if !isPathParam("post_id", params) {
		t.Error("expected post_id to be a path param")
	}
	if isPathParam("name", params) {
		t.Error("expected name to NOT be a path param")
	}
	if isPathParam("user_id", nil) {
		t.Error("expected false for nil params")
	}
}
