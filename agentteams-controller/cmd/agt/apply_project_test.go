package main

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	v1beta1 "github.com/agentscope-ai/AgentTeams/agentteams-controller/api/v1beta1"
	authpkg "github.com/agentscope-ai/AgentTeams/agentteams-controller/internal/auth"
	"github.com/agentscope-ai/AgentTeams/agentteams-controller/internal/server"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestApplyFromFiles_ProjectEndToEnd wires the real controller mux (with the
// Project REST routes) behind an httptest.Server and exercises
// `agt apply -f` against a fixture Project YAML.
func TestApplyFromFiles_ProjectEndToEnd(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1beta1.AddToScheme(scheme); err != nil {
		t.Fatalf("register scheme: %v", err)
	}
	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&v1beta1.Project{}).
		Build()

	authMw := authpkg.NewMiddleware(nil, nil, authpkg.NewAuthorizer(), nil, "default")

	httpServer := server.NewHTTPServer("ignored:0", server.ServerDeps{
		Client:    k8sClient,
		AuthMw:    authMw,
		Namespace: "default",
	})

	ts := httptest.NewServer(httpServer.Mux)
	defer ts.Close()

	t.Setenv("AGENTTEAMS_CONTROLLER_URL", ts.URL)
	os.Unsetenv("AGENTTEAMS_AUTH_TOKEN")
	os.Unsetenv("AGENTTEAMS_AUTH_TOKEN_FILE")

	content := `apiVersion: agentteams.io/v1beta1
kind: Project
metadata:
  name: proj-e2e
spec:
  team: alpha-team
  repos:
    - url: https://git.pawcommit.com/org/repo.git
      access: rw
`
	tmpFile, err := os.CreateTemp("", "agentteams-project-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	tmpFile.Close()

	if err := applyFromFiles([]string{tmpFile.Name()}); err != nil {
		t.Fatalf("first apply (create) failed: %v", err)
	}

	var created v1beta1.Project
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-e2e", Namespace: "default"}, &created); err != nil {
		t.Fatalf("project not created: %v", err)
	}
	if created.Spec.Team != "alpha-team" {
		t.Fatalf("team = %q, want alpha-team", created.Spec.Team)
	}
	if len(created.Spec.Repos) != 1 || created.Spec.Repos[0].Access != "rw" {
		t.Fatalf("repos = %+v", created.Spec.Repos)
	}

	if err := applyFromFiles([]string{tmpFile.Name()}); err != nil {
		t.Fatalf("second apply (update) failed: %v", err)
	}
}
