package hooks

import (
	"podux/internal/application/integration"
	"podux/internal/application/tunnel"

	"github.com/pocketbase/pocketbase/core"
)

// Register wires all PocketBase record hooks for the application.
func Register(app core.App, cloudflareService *integration.CloudflareService, frpcProvider tunnel.Provider) {
	registerIntegrationHooks(app)
	registerProxyHooks(app, cloudflareService, frpcProvider)
}
