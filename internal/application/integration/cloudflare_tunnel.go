package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// TunnelInfo holds a Cloudflare tunnel's identity and secret token.
// It is persisted inside fh_integrations.metadata under the key "tunnel".
type TunnelInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

// IngressRoute maps a public hostname to an internal service URL.
type IngressRoute struct {
	Hostname string
	Service  string // e.g. "http://127.0.0.1:3000"
}

// EnsureTunnel returns the existing tunnel for the integration, or creates a new
// one named "podux" under the given Cloudflare account.
// It returns the up-to-date TunnelInfo; callers must persist it back to metadata.
func EnsureTunnel(ctx context.Context, apiToken, accountID string, existing *TunnelInfo, logger *slog.Logger) (*TunnelInfo, error) {
	if existing != nil && existing.ID != "" && existing.Token != "" {
		logger.Debug("cloudflare tunnel already exists", "tunnelId", existing.ID)
		return existing, nil
	}

	logger.Info("creating cloudflare tunnel", "accountId", accountID)

	body, _ := json.Marshal(map[string]any{
		"name":       "podux",
		"config_src": "cloudflare",
	})

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/cfd_tunnel", accountID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		Result struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Token string `json:"token"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if !result.Success {
		if len(result.Errors) > 0 {
			return nil, fmt.Errorf("cloudflare: %s (code %d)", result.Errors[0].Message, result.Errors[0].Code)
		}
		return nil, fmt.Errorf("cloudflare: create tunnel failed")
	}

	info := &TunnelInfo{
		ID:    result.Result.ID,
		Name:  result.Result.Name,
		Token: result.Result.Token,
	}
	logger.Info("cloudflare tunnel created", "tunnelId", info.ID, "tunnelName", info.Name)
	return info, nil
}

// SyncIngress pushes the full ingress configuration (all routes + catch-all) to
// the Cloudflare tunnel. A catch-all "http_status:404" rule is always appended.
func SyncIngress(ctx context.Context, apiToken, accountID, tunnelID string, routes []IngressRoute, logger *slog.Logger) error {
	logger.Info("syncing cloudflare tunnel ingress", "tunnelId", tunnelID, "routeCount", len(routes))

	type ingressEntry struct {
		Hostname string `json:"hostname,omitempty"`
		Service  string `json:"service"`
	}

	ingress := make([]ingressEntry, 0, len(routes)+1)
	for _, r := range routes {
		ingress = append(ingress, ingressEntry{Hostname: r.Hostname, Service: r.Service})
	}
	// Cloudflare requires a catch-all entry with no hostname
	ingress = append(ingress, ingressEntry{Service: "http_status:404"})

	payload := map[string]any{
		"config": map[string]any{
			"ingress": ingress,
		},
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/accounts/%s/cfd_tunnel/%s/configurations",
		accountID, tunnelID,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if !result.Success {
		if len(result.Errors) > 0 {
			return fmt.Errorf("cloudflare: %s (code %d)", result.Errors[0].Message, result.Errors[0].Code)
		}
		return fmt.Errorf("cloudflare: sync ingress failed")
	}

	logger.Info("cloudflare tunnel ingress synced", "tunnelId", tunnelID, "routeCount", len(routes))
	return nil
}
