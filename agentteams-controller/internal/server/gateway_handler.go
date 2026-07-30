package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/agentscope-ai/AgentTeams/agentteams-controller/internal/gateway"
	"github.com/agentscope-ai/AgentTeams/agentteams-controller/internal/httputil"
)

// GatewayHandler handles /api/v1/gateway/* requests using the unified gateway.Client.
type GatewayHandler struct {
	gw gateway.Client
	// aiGatewayDomains holds the domains to set on provider-specific AI routes.
	aiGatewayDomains []string
}

func NewGatewayHandler(gw gateway.Client, aiGatewayDomains ...string) *GatewayHandler {
	return &GatewayHandler{gw: gw, aiGatewayDomains: aiGatewayDomains}
}

func (h *GatewayHandler) CreateConsumer(w http.ResponseWriter, r *http.Request) {
	if h.gw == nil {
		httputil.WriteError(w, http.StatusNotImplemented, "no gateway backend available")
		return
	}

	var req CreateConsumerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	result, err := h.gw.EnsureConsumer(r.Context(), gateway.ConsumerRequest{
		Name:          req.Name,
		CredentialKey: req.CredentialKey,
	})
	if err != nil {
		log.Printf("[ERROR] create consumer %s: %v", req.Name, err)
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, ConsumerResponse{
		Name:       req.Name,
		ConsumerID: result.ConsumerID,
		APIKey:     result.APIKey,
		Status:     result.Status,
	})
}

func (h *GatewayHandler) BindConsumer(w http.ResponseWriter, r *http.Request) {
	if h.gw == nil {
		httputil.WriteError(w, http.StatusNotImplemented, "no gateway backend available")
		return
	}

	consumerName := r.PathValue("id")
	if consumerName == "" {
		httputil.WriteError(w, http.StatusBadRequest, "consumer name is required")
		return
	}

	if err := h.gw.AuthorizeAIRoutes(r.Context(), consumerName, ""); err != nil {
		log.Printf("[ERROR] bind consumer %s: %v", consumerName, err)
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GatewayHandler) DeleteConsumer(w http.ResponseWriter, r *http.Request) {
	if h.gw == nil {
		httputil.WriteError(w, http.StatusNotImplemented, "no gateway backend available")
		return
	}

	consumerName := r.PathValue("id")
	if consumerName == "" {
		httputil.WriteError(w, http.StatusBadRequest, "consumer name is required")
		return
	}

	if err := h.gw.DeleteConsumer(r.Context(), consumerName); err != nil {
		log.Printf("[ERROR] delete consumer %s: %v", consumerName, err)
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Provider management endpoints ---

func (h *GatewayHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	if h.gw == nil {
		httputil.WriteError(w, http.StatusNotImplemented, "no gateway backend available")
		return
	}

	providers, err := h.gw.ListAIProviders(r.Context())
	if err != nil {
		log.Printf("[ERROR] list providers: %v", err)
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := ProviderListResponse{
		Providers: make([]ProviderResponse, 0, len(providers)),
		Total:     len(providers),
	}
	for _, p := range providers {
		resp.Providers = append(resp.Providers, ProviderResponse{
			Name:  p.Name,
			Type:  p.Type,
			Route: p.Route,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *GatewayHandler) RegisterProvider(w http.ResponseWriter, r *http.Request) {
	if h.gw == nil {
		httputil.WriteError(w, http.StatusNotImplemented, "no gateway backend available")
		return
	}

	var req RegisterProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	// Validate inputs.
	if req.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if strings.Contains(req.Name, "/") {
		httputil.WriteError(w, http.StatusBadRequest, "name must not contain '/'")
		return
	}
	if req.Name == "default" {
		httputil.WriteError(w, http.StatusBadRequest, "name 'default' is reserved")
		return
	}
	if req.URL == "" {
		httputil.WriteError(w, http.StatusBadRequest, "url is required")
		return
	}
	if req.Key == "" {
		httputil.WriteError(w, http.StatusBadRequest, "key is required")
		return
	}

	// Parse domain, port, protocol from the provider's base URL.
	parsedURL, err := url.Parse(req.URL)
	if err != nil || parsedURL.Host == "" {
		httputil.WriteError(w, http.StatusBadRequest, "url must be a valid HTTP(S) URL")
		return
	}
	protocol := parsedURL.Scheme // "http" or "https"
	domain := parsedURL.Hostname()
	port := 443
	if protocol == "http" {
		port = 80
	}
	if parsedURL.Port() != "" {
		port, _ = strconv.Atoi(parsedURL.Port())
	}

	ctx := r.Context()

	// Step 1: DNS service source.
	if err := h.gw.EnsureServiceSource(ctx, req.Name, domain, port, protocol); err != nil {
		log.Printf("[ERROR] register provider %s: service source: %v", req.Name, err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create service source: "+err.Error())
		return
	}

	// Step 2: AI provider.
	providerReq := gateway.AIProviderRequest{
		Name:     req.Name,
		Type:     "openai",
		Tokens:   []string{req.Key},
		Protocol: "openai/v1",
		Raw: map[string]interface{}{
			"openaiCustomUrl":         req.URL,
			"openaiCustomServiceName": req.Name + ".dns",
			"openaiCustomServicePort": port,
		},
	}
	if err := h.gw.EnsureAIProvider(ctx, providerReq); err != nil {
		log.Printf("[ERROR] register provider %s: AI provider: %v", req.Name, err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create AI provider: "+err.Error())
		return
	}

	// Step 3: Provider-specific AI route.
	routeName := "agentteams-" + req.Name + "-route"
	routeReq := gateway.ProviderRouteRequest{
		Name:             routeName,
		Provider:         req.Name,
		Domains:          h.aiGatewayDomains,
		ModelPrefix:      req.Name + "/",
		AllowedConsumers: []string{"manager"},
	}
	if err := h.gw.CreateProviderRoute(ctx, routeReq); err != nil {
		log.Printf("[ERROR] register provider %s: AI route: %v", req.Name, err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create AI route: "+err.Error())
		return
	}

	// Never echo the API key in the response.
	httputil.WriteJSON(w, http.StatusCreated, ProviderResponse{
		Name:  req.Name,
		Type:  "openai",
		Route: routeName,
	})
}

func (h *GatewayHandler) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	if h.gw == nil {
		httputil.WriteError(w, http.StatusNotImplemented, "no gateway backend available")
		return
	}

	name := r.PathValue("name")
	if name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "provider name is required")
		return
	}
	if name == "default" {
		httputil.WriteError(w, http.StatusBadRequest, "cannot delete the default provider")
		return
	}

	ctx := r.Context()
	var errs []string

	// Delete in reverse order: route -> provider -> service-source.
	if err := h.gw.DeleteProviderRoute(ctx, name); err != nil {
		errs = append(errs, fmt.Sprintf("route: %v", err))
	}
	if err := h.gw.DeleteAIProvider(ctx, name); err != nil {
		errs = append(errs, fmt.Sprintf("provider: %v", err))
	}
	if err := h.gw.DeleteServiceSource(ctx, name); err != nil {
		errs = append(errs, fmt.Sprintf("service-source: %v", err))
	}

	if len(errs) > 0 {
		log.Printf("[ERROR] delete provider %s: partial failure: %s", name, strings.Join(errs, "; "))
		httputil.WriteError(w, http.StatusInternalServerError, "partial delete failure: "+strings.Join(errs, "; "))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListProviderModels proxies the upstream provider's /v1/models endpoint and
// returns a sorted list of model ids. Tokens never leave the controller —
// the upstream call is made server-side with the stored credential.
func (h *GatewayHandler) ListProviderModels(w http.ResponseWriter, r *http.Request) {
	if h.gw == nil {
		httputil.WriteError(w, http.StatusNotImplemented, "no gateway backend available")
		return
	}
	name := r.PathValue("name")
	if name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "provider name is required")
		return
	}

	detail, err := h.gw.GetAIProvider(r.Context(), name)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "provider not found: "+err.Error())
		return
	}
	if len(detail.Tokens) == 0 {
		httputil.WriteError(w, http.StatusConflict, "provider has no stored tokens")
		return
	}

	// Resolve the upstream base URL.
	baseURL := ""
	if detail.RawConfigs != nil {
		if u, ok := detail.RawConfigs["openaiCustomUrl"].(string); ok {
			baseURL = u
		}
	}
	if baseURL == "" {
		httputil.WriteError(w, http.StatusConflict, "provider has no upstream base URL")
		return
	}

	modelsURL := strings.TrimRight(baseURL, "/") + "/models"
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, modelsURL, nil)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "build models request: "+err.Error())
		return
	}
	req.Header.Set("Authorization", "Bearer "+detail.Tokens[0])

	resp, err := client.Do(req)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "upstream models request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		httputil.WriteError(w, http.StatusBadGateway, fmt.Sprintf("upstream returned HTTP %d", resp.StatusCode))
		return
	}

	// Parse OpenAI-style {"data":[{"id":...}]} — fall back to a bare array of
	// strings/objects for non-conforming providers.
	var ids []string
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	var openaiResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &openaiResp); err == nil && len(openaiResp.Data) > 0 {
		for _, m := range openaiResp.Data {
			if m.ID != "" {
				ids = append(ids, m.ID)
			}
		}
	} else {
		var bare []json.RawMessage
		if err := json.Unmarshal(bodyBytes, &bare); err == nil {
			for _, raw := range bare {
				var s string
				if json.Unmarshal(raw, &s) == nil {
					ids = append(ids, s)
					continue
				}
				var obj struct {
					ID string `json:"id"`
				}
				if json.Unmarshal(raw, &obj) == nil && obj.ID != "" {
					ids = append(ids, obj.ID)
				}
			}
		}
	}

	sort.Strings(ids)
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"models": ids})
}
