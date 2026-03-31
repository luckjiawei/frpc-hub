package hooks

import (
	"context"
	"fmt"
	"time"

	"podux/internal/application/integration"
	"podux/internal/application/tunnel"

	"github.com/pocketbase/pocketbase/core"
)

// registerProxyHooks handles fh_proxies events.
func registerProxyHooks(app core.App, cloudflareService *integration.CloudflareService, frpcProvider tunnel.Provider) {
	// ensureFrpTunnel finds or creates the th_tunnels record (type=frp) for the
	// given serverId. One server maps to exactly one FRP tunnel.
	ensureFrpTunnel := func(serverId string) (string, error) {
		existing, err := app.FindFirstRecordByFilter(
			"th_tunnels",
			fmt.Sprintf(`serverId = "%s" && type = "frp"`, serverId),
		)
		if err == nil {
			return existing.Id, nil
		}

		collection, err := app.FindCollectionByNameOrId("th_tunnels")
		if err != nil {
			return "", fmt.Errorf("th_tunnels collection not found: %w", err)
		}
		record := core.NewRecord(collection)
		record.Set("serverId", serverId)
		record.Set("type", "frp")
		if err := app.Save(record); err != nil {
			return "", fmt.Errorf("save frp tunnel: %w", err)
		}
		app.Logger().Info("Auto-created FRP tunnel", "tunnelId", record.Id, "serverId", serverId)
		return record.Id, nil
	}

	// Ensure FRP tunnel exists (and set tunnelId) on proxy create/update.
	// Primary path: tunnelId is set directly by the caller.
	// Fallback path: derive tunnel from the deprecated serverId field for backwards compat.
	// Note: proxyType is always a protocol (tcp/udp/http/https); tunnel type is in th_tunnels.type.
	bindFrpTunnel := func(e *core.RecordEvent) error {
		if e.Record.GetString("tunnelId") != "" {
			return e.Next()
		}
		// tunnelId not set — fall back to serverId (deprecated).
		serverId := e.Record.GetString("serverId")
		if serverId == "" {
			return e.Next()
		}
		tunnelId, err := ensureFrpTunnel(serverId)
		if err != nil {
			app.Logger().Error("ensure FRP tunnel failed", "error", err, "serverId", serverId)
			return e.Next()
		}
		e.Record.Set("tunnelId", tunnelId)
		return e.Next()
	}
	app.OnRecordCreate("fh_proxies").BindFunc(bindFrpTunnel)
	app.OnRecordUpdate("fh_proxies").BindFunc(bindFrpTunnel)

	// Sync Cloudflare tunnel ingress on proxy create/update/delete.
	// Tunnel type is resolved via tunnelId → th_tunnels.type; integrationId lives on th_tunnels.
	syncCloudflare := func(e *core.RecordEvent) error {
		tunnelId := e.Record.GetString("tunnelId")
		if tunnelId == "" {
			return e.Next()
		}
		tunnelRecord, err := app.FindRecordById("th_tunnels", tunnelId)
		if err != nil || tunnelRecord.GetString("type") != "cloudflare" {
			return e.Next()
		}
		integrationID := tunnelRecord.GetString("integrationId")
		if integrationID == "" {
			return e.Next()
		}
		go func(id string) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := cloudflareService.SyncTunnelForProxy(ctx, id); err != nil {
				app.Logger().Error("cloudflare tunnel sync failed", "error", err, "integrationId", id)
			}
		}(integrationID)
		return e.Next()
	}
	app.OnRecordAfterCreateSuccess("fh_proxies").BindFunc(syncCloudflare)
	app.OnRecordAfterUpdateSuccess("fh_proxies").BindFunc(syncCloudflare)
	app.OnRecordAfterDeleteSuccess("fh_proxies").BindFunc(syncCloudflare)

	// Reload frpc when an FRP proxy is updated: resolve server via tunnelId → th_tunnels.serverId.
	app.OnRecordAfterUpdateSuccess("fh_proxies").BindFunc(func(e *core.RecordEvent) error {
		tunnelId := e.Record.GetString("tunnelId")
		if tunnelId == "" {
			return e.Next()
		}
		tunnelRecord, err := app.FindRecordById("th_tunnels", tunnelId)
		if err != nil {
			return e.Next()
		}
		serverId := tunnelRecord.GetString("serverId")
		if serverId != "" && frpcProvider.IsRunning(serverId) {
			app.Logger().Info("Proxy updated, reloading frpc configuration", "proxyId", e.Record.Id, "serverId", serverId)
			if err := frpcProvider.Reload(serverId); err != nil {
				app.Logger().Error("Failed to reload frpc after proxy update", "error", err, "serverId", serverId)
			}
		}
		return e.Next()
	})

	// EnsureServerTunnels runs at startup:
	// 1. Scans all fh_servers and ensures each has a th_tunnels record (type=frp).
	// 2. Links any proxies still using the deprecated serverId field to their server's tunnel.
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		go func() {
			servers, err := app.FindAllRecords("fh_servers")
			if err != nil {
				app.Logger().Error("ensureServerTunnels: failed to list servers", "error", err)
				return
			}

			for _, server := range servers {
				serverId := server.Id

				tunnelId, err := ensureFrpTunnel(serverId)
				if err != nil {
					app.Logger().Error("ensureServerTunnels: ensure tunnel failed", "error", err, "serverId", serverId)
					continue
				}

				// Link proxies that still reference this server via the deprecated serverId field.
				proxies, err := app.FindRecordsByFilter(
					"fh_proxies",
					fmt.Sprintf(`serverId = "%s" && tunnelId = ""`, serverId),
					"", 0, 0, nil,
				)
				if err != nil {
					app.Logger().Error("ensureServerTunnels: query proxies failed", "error", err, "serverId", serverId)
					continue
				}
				for _, proxy := range proxies {
					proxy.Set("tunnelId", tunnelId)
					if err := app.Save(proxy); err != nil {
						app.Logger().Error("ensureServerTunnels: save proxy failed", "error", err, "proxyId", proxy.Id)
					}
				}
				if len(proxies) > 0 {
					app.Logger().Info("ensureServerTunnels: linked proxies to tunnel",
						"serverId", serverId, "tunnelId", tunnelId, "count", len(proxies))
				}
			}

			app.Logger().Info("ensureServerTunnels: done", "serverCount", len(servers))
		}()
		return e.Next()
	})
}
