package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v1beta1 "github.com/agentscope-ai/AgentTeams/agentteams-controller/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// --- Projects ---

func TestCreateProject_HappyPath(t *testing.T) {
	scheme := newServerTestScheme(t)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	handler := NewResourceHandler(k8sClient, "default", nil, "")

	body := []byte(`{"name":"proj-1","team":"alpha-team","repos":[{"url":"https://git.pawcommit.com/org/repo.git","access":"rw"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.CreateProject(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}
	var stored v1beta1.Project
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-1", Namespace: "default"}, &stored); err != nil {
		t.Fatalf("get project: %v", err)
	}
	if stored.Spec.Team != "alpha-team" {
		t.Fatalf("team = %q, want alpha-team", stored.Spec.Team)
	}
	if len(stored.Spec.Repos) != 1 || stored.Spec.Repos[0].Access != "rw" {
		t.Fatalf("repos = %+v", stored.Spec.Repos)
	}

	var resp ProjectResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Phase != "Pending" {
		t.Fatalf("response phase = %q, want Pending (default)", resp.Phase)
	}
}

func TestCreateProject_WithDependsOn(t *testing.T) {
	scheme := newServerTestScheme(t)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	handler := NewResourceHandler(k8sClient, "default", nil, "")

	body := []byte(`{"name":"proj-2","team":"alpha-team","dependsOn":["upstream-a","upstream-b"],"repos":[{"url":"https://git.pawcommit.com/org/repo.git","access":"rw"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.CreateProject(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}
	var stored v1beta1.Project
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-2", Namespace: "default"}, &stored); err != nil {
		t.Fatalf("get project: %v", err)
	}
	if len(stored.Spec.DependsOn) != 2 || stored.Spec.DependsOn[0] != "upstream-a" {
		t.Fatalf("dependsOn = %+v", stored.Spec.DependsOn)
	}

	var resp ProjectResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.DependsOn) != 2 {
		t.Fatalf("response dependsOn = %+v", resp.DependsOn)
	}
}

func TestCreateProject_MissingName(t *testing.T) {
	scheme := newServerTestScheme(t)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	handler := NewResourceHandler(k8sClient, "default", nil, "")

	body := []byte(`{"team":"alpha-team","repos":[{"url":"https://git.pawcommit.com/org/repo.git","access":"rw"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.CreateProject(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestCreateProject_MissingTeam(t *testing.T) {
	scheme := newServerTestScheme(t)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	handler := NewResourceHandler(k8sClient, "default", nil, "")

	body := []byte(`{"name":"proj-2","repos":[{"url":"https://git.pawcommit.com/org/repo.git","access":"rw"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.CreateProject(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestCreateProject_MissingRepos(t *testing.T) {
	scheme := newServerTestScheme(t)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	handler := NewResourceHandler(k8sClient, "default", nil, "")

	body := []byte(`{"name":"proj-3","team":"alpha-team"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.CreateProject(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestCreateProject_InvalidAccessEnum(t *testing.T) {
	scheme := newServerTestScheme(t)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	handler := NewResourceHandler(k8sClient, "default", nil, "")

	body := []byte(`{"name":"proj-4","team":"alpha-team","repos":[{"url":"https://git.pawcommit.com/org/repo.git","access":"readwrite"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.CreateProject(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestGetProject_NotFound(t *testing.T) {
	scheme := newServerTestScheme(t)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	handler := NewResourceHandler(k8sClient, "default", nil, "")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/missing", nil)
	req.SetPathValue("name", "missing")
	rec := httptest.NewRecorder()
	handler.GetProject(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d: %s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestGetProject_HappyPath(t *testing.T) {
	scheme := newServerTestScheme(t)
	proj := &v1beta1.Project{}
	proj.Name = "proj-5"
	proj.Namespace = "default"
	proj.Spec.Team = "alpha-team"
	proj.Spec.Repos = []v1beta1.ProjectRepo{{URL: "https://git.pawcommit.com/org/repo.git", Access: "rw"}}
	proj.Status.Phase = "Ready"
	proj.Status.RepoCount = 1
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(proj).Build()
	handler := NewResourceHandler(k8sClient, "default", nil, "")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/proj-5", nil)
	req.SetPathValue("name", "proj-5")
	rec := httptest.NewRecorder()
	handler.GetProject(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var resp ProjectResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Phase != "Ready" || resp.RepoCount != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestListProjects_FiltersByTeam(t *testing.T) {
	scheme := newServerTestScheme(t)
	p1 := &v1beta1.Project{}
	p1.Name = "proj-a"
	p1.Namespace = "default"
	p1.Spec.Team = "alpha-team"
	p1.Spec.Repos = []v1beta1.ProjectRepo{{URL: "https://x/a.git", Access: "rw"}}

	p2 := &v1beta1.Project{}
	p2.Name = "proj-b"
	p2.Namespace = "default"
	p2.Spec.Team = "beta-team"
	p2.Spec.Repos = []v1beta1.ProjectRepo{{URL: "https://x/b.git", Access: "rw"}}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(p1, p2).Build()
	handler := NewResourceHandler(k8sClient, "default", nil, "")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects?team=alpha-team", nil)
	rec := httptest.NewRecorder()
	handler.ListProjects(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var list ProjectListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if list.Total != 1 || list.Projects[0].Name != "proj-a" {
		t.Fatalf("unexpected filtered list: %+v", list)
	}
}

func TestUpdateProject_UpdatesSpecFields(t *testing.T) {
	scheme := newServerTestScheme(t)
	proj := &v1beta1.Project{}
	proj.Name = "proj-6"
	proj.Namespace = "default"
	proj.Spec.Team = "alpha-team"
	proj.Spec.Repos = []v1beta1.ProjectRepo{{URL: "https://x/a.git", Access: "rw"}}
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(proj).WithStatusSubresource(&v1beta1.Project{}).Build()
	handler := NewResourceHandler(k8sClient, "default", nil, "")

	body := []byte(`{"description":"updated description"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/proj-6", bytes.NewReader(body))
	req.SetPathValue("name", "proj-6")
	rec := httptest.NewRecorder()
	handler.UpdateProject(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var updated v1beta1.Project
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-6", Namespace: "default"}, &updated); err != nil {
		t.Fatalf("get updated project: %v", err)
	}
	if updated.Spec.Description != "updated description" {
		t.Fatalf("description = %q, want %q", updated.Spec.Description, "updated description")
	}
}

func TestUpdateProject_SetsCompletedPhase(t *testing.T) {
	scheme := newServerTestScheme(t)
	proj := &v1beta1.Project{}
	proj.Name = "proj-7"
	proj.Namespace = "default"
	proj.Spec.Team = "alpha-team"
	proj.Spec.Repos = []v1beta1.ProjectRepo{{URL: "https://x/a.git", Access: "rw"}}
	proj.Status.Phase = "Ready"
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(proj).WithStatusSubresource(&v1beta1.Project{}).Build()
	handler := NewResourceHandler(k8sClient, "default", nil, "")

	body := []byte(`{"phase":"Completed"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/proj-7", bytes.NewReader(body))
	req.SetPathValue("name", "proj-7")
	rec := httptest.NewRecorder()
	handler.UpdateProject(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var updated v1beta1.Project
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-7", Namespace: "default"}, &updated); err != nil {
		t.Fatalf("get updated project: %v", err)
	}
	if updated.Status.Phase != "Completed" {
		t.Fatalf("phase = %q, want Completed", updated.Status.Phase)
	}
}

func TestUpdateProject_RejectsInvalidPhase(t *testing.T) {
	scheme := newServerTestScheme(t)
	proj := &v1beta1.Project{}
	proj.Name = "proj-8"
	proj.Namespace = "default"
	proj.Spec.Team = "alpha-team"
	proj.Spec.Repos = []v1beta1.ProjectRepo{{URL: "https://x/a.git", Access: "rw"}}
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(proj).WithStatusSubresource(&v1beta1.Project{}).Build()
	handler := NewResourceHandler(k8sClient, "default", nil, "")

	body := []byte(`{"phase":"Ready"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/proj-8", bytes.NewReader(body))
	req.SetPathValue("name", "proj-8")
	rec := httptest.NewRecorder()
	handler.UpdateProject(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestUpdateProject_NotFound(t *testing.T) {
	scheme := newServerTestScheme(t)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	handler := NewResourceHandler(k8sClient, "default", nil, "")

	body := []byte(`{"description":"x"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/missing", bytes.NewReader(body))
	req.SetPathValue("name", "missing")
	rec := httptest.NewRecorder()
	handler.UpdateProject(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d: %s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestDeleteProject_HappyPath(t *testing.T) {
	scheme := newServerTestScheme(t)
	proj := &v1beta1.Project{}
	proj.Name = "proj-9"
	proj.Namespace = "default"
	proj.Spec.Team = "alpha-team"
	proj.Spec.Repos = []v1beta1.ProjectRepo{{URL: "https://x/a.git", Access: "rw"}}
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(proj).Build()
	handler := NewResourceHandler(k8sClient, "default", nil, "")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/proj-9", nil)
	req.SetPathValue("name", "proj-9")
	rec := httptest.NewRecorder()
	handler.DeleteProject(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d: %s", http.StatusNoContent, rec.Code, rec.Body.String())
	}
	var after v1beta1.Project
	err := k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-9", Namespace: "default"}, &after)
	if err == nil {
		t.Fatalf("expected project to be gone after delete")
	}
}

func TestDeleteProject_NotFound(t *testing.T) {
	scheme := newServerTestScheme(t)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	handler := NewResourceHandler(k8sClient, "default", nil, "")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/missing", nil)
	req.SetPathValue("name", "missing")
	rec := httptest.NewRecorder()
	handler.DeleteProject(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d: %s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestCreateProject_StampsControllerLabel(t *testing.T) {
	scheme := newServerTestScheme(t)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	handler := NewResourceHandler(k8sClient, "default", nil, "ctrl-a")

	body := []byte(`{"name":"proj-10","team":"alpha-team","repos":[{"url":"https://x/a.git","access":"rw"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.CreateProject(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}
	var proj v1beta1.Project
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-10", Namespace: "default"}, &proj); err != nil {
		t.Fatalf("get project: %v", err)
	}
	if got := proj.Labels[v1beta1.LabelController]; got != "ctrl-a" {
		t.Fatalf("expected controller label ctrl-a, got %q", got)
	}
}

