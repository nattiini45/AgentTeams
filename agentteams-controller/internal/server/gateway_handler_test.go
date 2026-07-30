package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/agentscope-ai/AgentTeams/agentteams-controller/internal/gateway"
)

// fakeProviderGateway implements gateway.Client for provider handler tests.
type fakeProviderGateway struct {
	providers      []gateway.AIProviderInfo
	listErr        error
	deleteErr      error
	ensureSvcErr   error
	ensureProvErr  error
	createRouteErr error

	// Recorded calls
	ensureSvcCalls   []string
	ensureProvCalls  []gateway.AIProviderRequest
	createRouteCalls []gateway.ProviderRouteRequest
	deleteRouteCalls []string
	deleteProvCalls  []string
	deleteSvcCalls   []string
}

func (f *fakeProviderGateway) EnsureConsumer(context.Context, gateway.ConsumerRequest) (*gateway.ConsumerResult, error) {
	return &gateway.ConsumerResult{}, nil
}
func (f *fakeProviderGateway) DeleteConsumer(context.Context, string) error { return nil }
func (f *fakeProviderGateway) AuthorizeAIRoutes(context.Context, string, string) error {
	return nil
}
func (f *fakeProviderGateway) DeauthorizeAIRoutes(context.Context, string, string) error { return nil }
func (f *fakeProviderGateway) ExposePort(context.Context, gateway.PortExposeRequest) error {
	return nil
}
func (f *fakeProviderGateway) UnexposePort(context.Context, gateway.PortExposeRequest) error {
	return nil
}
func (f *fakeProviderGateway) EnsureServiceSource(_ context.Context, name, _ string, _ int, _ string) error {
	f.ensureSvcCalls = append(f.ensureSvcCalls, name)
	return f.ensureSvcErr
}
func (f *fakeProviderGateway) EnsureStaticServiceSource(context.Context, string, string, int) error {
	return nil
}
func (f *fakeProviderGateway) EnsureRoute(context.Context, string, []string, string, int, string) error {
	return nil
}
func (f *fakeProviderGateway) DeleteRoute(context.Context, string) error { return nil }
func (f *fakeProviderGateway) EnsureAIProvider(_ context.Context, req gateway.AIProviderRequest) error {
	f.ensureProvCalls = append(f.ensureProvCalls, req)
	return f.ensureProvErr
}
func (f *fakeProviderGateway) EnsureStreamIdleTimeout(context.Context, int) error { return nil }
func (f *fakeProviderGateway) EnsureAIRoute(context.Context, gateway.AIRouteRequest) error {
	return nil
}
func (f *fakeProviderGateway) ListAIProviders(context.Context) ([]gateway.AIProviderInfo, error) {
	return f.providers, f.listErr
}
func (f *fakeProviderGateway) DeleteAIProvider(_ context.Context, name string) error {
	f.deleteProvCalls = append(f.deleteProvCalls, name)
	return f.deleteErr
}
func (f *fakeProviderGateway) CreateProviderRoute(_ context.Context, req gateway.ProviderRouteRequest) error {
	f.createRouteCalls = append(f.createRouteCalls, req)
	return f.createRouteErr
}
func (f *fakeProviderGateway) DeleteProviderRoute(_ context.Context, name string) error {
	f.deleteRouteCalls = append(f.deleteRouteCalls, name)
	return nil
}
func (f *fakeProviderGateway) DeleteServiceSource(_ context.Context, name string) error {
	f.deleteSvcCalls = append(f.deleteSvcCalls, name)
	return nil
}
func (f *fakeProviderGateway) ResolveModelProvider(context.Context, string) (*gateway.ModelProviderInfo, error) {
	return nil, gateway.ErrUnsupportedOp
}
func (f *fakeProviderGateway) Healthy(context.Context) error { return nil }
func (f *fakeProviderGateway) ListMCPServers(context.Context) ([]gateway.MCPServerInfo, error) {
	return nil, nil
}

func TestListProviders(t *testing.T) {
	gw := &fakeProviderGateway{
		providers: []gateway.AIProviderInfo{
			{Name: "qwen", Type: "qwen", Route: ""},
			{Name: "ollama", Type: "openai", Route: "hiclaw-ollama-route"},
		},
	}
	h := NewGatewayHandler(gw, "aigw-local.agentteams.io")

	req := httptest.NewRequest("GET", "/api/v1/gateway/providers", nil)
	w := httptest.NewRecorder()
	h.ListProviders(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp ProviderListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("total = %d, want 2", resp.Total)
	}
	if resp.Providers[1].Name != "ollama" || resp.Providers[1].Route != "hiclaw-ollama-route" {
		t.Errorf("unexpected provider[1]: %+v", resp.Providers[1])
	}
}

func TestRegisterProvider_Success(t *testing.T) {
	gw := &fakeProviderGateway{}
	h := NewGatewayHandler(gw, "aigw-local.agentteams.io")

	body := `{"name":"ollama","url":"https://ollama.com/v1","key":"sk-test-key"}`
	req := httptest.NewRequest("POST", "/api/v1/gateway/providers", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.RegisterProvider(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", w.Code, w.Body.String())
	}

	// Verify the response does NOT contain the API key.
	if strings.Contains(w.Body.String(), "sk-test-key") {
		t.Error("API key leaked in response body")
	}

	var resp ProviderResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Name != "ollama" || resp.Type != "openai" || resp.Route != "hiclaw-ollama-route" {
		t.Errorf("unexpected response: %+v", resp)
	}

	// Verify gateway calls.
	if len(gw.ensureSvcCalls) != 1 || gw.ensureSvcCalls[0] != "ollama" {
		t.Errorf("EnsureServiceSource calls: %v", gw.ensureSvcCalls)
	}
	if len(gw.ensureProvCalls) != 1 {
		t.Fatalf("EnsureAIProvider calls: %d", len(gw.ensureProvCalls))
	}
	if gw.ensureProvCalls[0].Tokens[0] != "sk-test-key" {
		t.Error("key not passed to EnsureAIProvider")
	}
	if len(gw.createRouteCalls) != 1 {
		t.Fatalf("CreateProviderRoute calls: %d", len(gw.createRouteCalls))
	}
	routeReq := gw.createRouteCalls[0]
	if routeReq.Name != "hiclaw-ollama-route" || routeReq.ModelPrefix != "ollama/" {
		t.Errorf("unexpected route request: %+v", routeReq)
	}
	if len(routeReq.Domains) != 1 || routeReq.Domains[0] != "aigw-local.agentteams.io" {
		t.Errorf("unexpected domains: %v", routeReq.Domains)
	}
}

func TestRegisterProvider_ValidationErrors(t *testing.T) {
	gw := &fakeProviderGateway{}
	h := NewGatewayHandler(gw)

	cases := []struct {
		name string
		body string
		want int
	}{
		{"empty name", `{"name":"","url":"https://x.com/v1","key":"k"}`, http.StatusBadRequest},
		{"slash in name", `{"name":"a/b","url":"https://x.com/v1","key":"k"}`, http.StatusBadRequest},
		{"reserved name", `{"name":"default","url":"https://x.com/v1","key":"k"}`, http.StatusBadRequest},
		{"missing url", `{"name":"ok","url":"","key":"k"}`, http.StatusBadRequest},
		{"missing key", `{"name":"ok","url":"https://x.com/v1","key":""}`, http.StatusBadRequest},
		{"invalid url", `{"name":"ok","url":"not-a-url","key":"k"}`, http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/gateway/providers", strings.NewReader(tc.body))
			w := httptest.NewRecorder()
			h.RegisterProvider(w, req)
			if w.Code != tc.want {
				t.Errorf("status = %d, want %d; body: %s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

func TestDeleteProvider_Success(t *testing.T) {
	gw := &fakeProviderGateway{}
	h := NewGatewayHandler(gw)

	req := httptest.NewRequest("DELETE", "/api/v1/gateway/providers/ollama", nil)
	req.SetPathValue("name", "ollama")
	w := httptest.NewRecorder()
	h.DeleteProvider(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body: %s", w.Code, w.Body.String())
	}
	if len(gw.deleteRouteCalls) != 1 || gw.deleteRouteCalls[0] != "ollama" {
		t.Errorf("DeleteProviderRoute calls: %v", gw.deleteRouteCalls)
	}
	if len(gw.deleteProvCalls) != 1 || gw.deleteProvCalls[0] != "ollama" {
		t.Errorf("DeleteAIProvider calls: %v", gw.deleteProvCalls)
	}
	if len(gw.deleteSvcCalls) != 1 || gw.deleteSvcCalls[0] != "ollama" {
		t.Errorf("DeleteServiceSource calls: %v", gw.deleteSvcCalls)
	}
}

func TestDeleteProvider_ReservedName(t *testing.T) {
	gw := &fakeProviderGateway{}
	h := NewGatewayHandler(gw)

	req := httptest.NewRequest("DELETE", "/api/v1/gateway/providers/default", nil)
	req.SetPathValue("name", "default")
	w := httptest.NewRecorder()
	h.DeleteProvider(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}
