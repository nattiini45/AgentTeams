package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
	routeName := "hiclaw-" + req.Name + "-route"
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
