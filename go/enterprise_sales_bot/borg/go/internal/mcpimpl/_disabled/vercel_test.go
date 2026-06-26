package mcpimpl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandleVercelTools(t *testing.T) {
	// Start mock Vercel API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-vercel-token" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"code": "not_authorized", "message": "Not authorized"}}`))
			return
		}

		path := r.URL.Path
		switch path {
		case "/v9/projects":
			w.Write([]byte(`{"projects": [{"id": "p1", "name": "project-one"}]}`))
		case "/v9/projects/p1":
			w.Write([]byte(`{"id": "p1", "name": "project-one"}`))
		case "/v6/deployments":
			w.Write([]byte(`{"deployments": [{"uid": "d1", "name": "deploy-one"}]}`))
		case "/v13/deployments/d1":
			w.Write([]byte(`{"id": "d1", "name": "deploy-one"}`))
		case "/v12/deployments/d1/cancel":
			w.Write([]byte(`{"id": "d1", "status": "CANCELED"}`))
		case "/v9/projects/p1/env":
			if r.Method == "GET" {
				w.Write([]byte(`{"envs": [{"id": "env1", "key": "API_KEY"}]}`))
			}
		case "/v9/projects/p1/env/env1":
			if r.Method == "DELETE" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"id": "env1"}`))
			}
		case "/v10/projects/p1/env":
			if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"id": "env1", "key": "API_KEY"}`))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	os.Setenv("VERCEL_TOKEN", "test-vercel-token")
	os.Setenv("VERCEL_API_URL", server.URL)
	defer os.Unsetenv("VERCEL_TOKEN")
	defer os.Unsetenv("VERCEL_API_URL")

	ctx := context.Background()

	// Test 1: HandleVercelListProjects
	resp, err := HandleVercelListProjects(ctx, map[string]interface{}{"limit": 10.0})
	if err != nil {
		t.Fatalf("HandleVercelListProjects failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "project-one") {
		t.Errorf("Expected project list response, got: %s", resp.Content[0].Text)
	}

	// Test 2: HandleVercelGetProject
	resp, err = HandleVercelGetProject(ctx, map[string]interface{}{"projectId": "p1"})
	if err != nil {
		t.Fatalf("HandleVercelGetProject failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "project-one") {
		t.Errorf("Expected project details, got: %s", resp.Content[0].Text)
	}

	// Test 3: HandleVercelListDeployments
	resp, err = HandleVercelListDeployments(ctx, map[string]interface{}{"projectId": "p1"})
	if err != nil {
		t.Fatalf("HandleVercelListDeployments failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "deploy-one") {
		t.Errorf("Expected deployments list, got: %s", resp.Content[0].Text)
	}

	// Test 4: HandleVercelGetDeployment
	resp, err = HandleVercelGetDeployment(ctx, map[string]interface{}{"deploymentId": "d1"})
	if err != nil {
		t.Fatalf("HandleVercelGetDeployment failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "deploy-one") {
		t.Errorf("Expected deployment details, got: %s", resp.Content[0].Text)
	}

	// Test 5: HandleVercelCancelDeployment
	resp, err = HandleVercelCancelDeployment(ctx, map[string]interface{}{"deploymentId": "d1"})
	if err != nil {
		t.Fatalf("HandleVercelCancelDeployment failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "CANCELED") {
		t.Errorf("Expected deployment canceled status, got: %s", resp.Content[0].Text)
	}

	// Test 6: HandleVercelListEnvVars
	resp, err = HandleVercelListEnvVars(ctx, map[string]interface{}{"projectId": "p1"})
	if err != nil {
		t.Fatalf("HandleVercelListEnvVars failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "API_KEY") {
		t.Errorf("Expected env vars list, got: %s", resp.Content[0].Text)
	}

	// Test 7: HandleVercelCreateEnvVar
	resp, err = HandleVercelCreateEnvVar(ctx, map[string]interface{}{
		"projectId": "p1",
		"key":       "API_KEY",
		"value":     "secret",
		"target":    []interface{}{"production"},
	})
	if err != nil {
		t.Fatalf("HandleVercelCreateEnvVar failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "env1") {
		t.Errorf("Expected env var created details, got: %s", resp.Content[0].Text)
	}

	// Test 8: HandleVercelDeleteEnvVar
	resp, err = HandleVercelDeleteEnvVar(ctx, map[string]interface{}{
		"projectId": "p1",
		"envVarId":  "env1",
	})
	if err != nil {
		t.Fatalf("HandleVercelDeleteEnvVar failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "env1") {
		t.Errorf("Expected env var deleted details, got: %s", resp.Content[0].Text)
	}
}
