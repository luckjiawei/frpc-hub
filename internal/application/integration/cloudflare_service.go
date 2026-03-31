package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/pocketbase/pocketbase/core"
)

// CloudflareService manages the lifecycle of Cloudflare tunnels and route
// synchronisation for a given integration record.
type CloudflareService struct {
	app    core.App
	logger *slog.Logger
}

func NewCloudflareService(app core.App) *CloudflareService {
	return &CloudflareService{app: app, logger: app.Logger()}
}

// IntegrationMeta is the shape stored in fh_integrations.metadata.
type IntegrationMeta struct {
	Accounts []CloudflareAccount `json:"accounts"`
	Tunnel   *TunnelInfo         `json:"tunnel,omitempty"`
}

// SyncTunnelForProxy is called after a cloudflare proxy is created, updated, or
// deleted. It ensures a tunnel exists for the integration and pushes the current
// full ingress config to Cloudflare.
func (s *CloudflareService) SyncTunnelForProxy(ctx context.Context, integrationID string) error {
	logger := s.logger.With("integrationId", integrationID)

	// 1. Load integration record
	record, err := s.app.FindRecordById("fh_integrations", integrationID)
	if err != nil {
		return fmt.Errorf("find integration: %w", err)
	}

	apiToken := record.GetString("credentials")
	if apiToken == "" {
		return fmt.Errorf("integration has no credentials")
	}

	// 2. Parse metadata
	var meta IntegrationMeta
	rawMeta := record.Get("metadata")
	if rawMeta != nil {
		b, _ := json.Marshal(rawMeta)
		_ = json.Unmarshal(b, &meta)
	}

	if len(meta.Accounts) == 0 {
		return fmt.Errorf("integration has no validated accounts; validate first")
	}
	accountID := meta.Accounts[0].ID

	// 3. Ensure tunnel exists
	tunnel, err := EnsureTunnel(ctx, apiToken, accountID, meta.Tunnel, logger)
	if err != nil {
		return fmt.Errorf("ensure tunnel: %w", err)
	}

	// Persist tunnel info back if it was just created
	if meta.Tunnel == nil || meta.Tunnel.ID == "" {
		meta.Tunnel = tunnel
		metaBytes, _ := json.Marshal(meta)
		var metaMap map[string]any
		_ = json.Unmarshal(metaBytes, &metaMap)
		record.Set("metadata", metaMap)
		if saveErr := s.app.Save(record); saveErr != nil {
			logger.Error("failed to persist tunnel metadata", "error", saveErr)
		}
	}

	// 4. Collect all enabled cloudflare proxies for this integration
	proxies, err := s.app.FindRecordsByFilter(
		"fh_proxies",
		`type = "cloudflare" && integrationId = {:integrationId} && status = "enabled"`,
		"-created", 0, 0,
		map[string]any{"integrationId": integrationID},
	)
	if err != nil {
		return fmt.Errorf("list proxies: %w", err)
	}

	// 5. Build ingress routes
	routes := make([]IngressRoute, 0, len(proxies))
	for _, p := range proxies {
		domains, ok := p.Get("customDomains").([]any)
		if !ok || len(domains) == 0 {
			continue
		}
		localIP := p.GetString("localIP")
		if localIP == "" {
			localIP = "127.0.0.1"
		}
		localPort := p.GetString("localPort")
		if localPort == "" {
			continue
		}
		service := fmt.Sprintf("http://%s:%s", localIP, localPort)
		// Each hostname in customDomains gets its own ingress entry
		for _, d := range domains {
			hostname, _ := d.(string)
			if hostname == "" {
				continue
			}
			routes = append(routes, IngressRoute{Hostname: hostname, Service: service})
		}
	}

	// 6. Push ingress config to Cloudflare
	return SyncIngress(ctx, apiToken, accountID, tunnel.ID, routes, logger)
}
