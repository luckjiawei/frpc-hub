package hooks

import (
	"context"
	"time"

	"podux/internal/application/integration"

	"github.com/pocketbase/pocketbase/core"
)

// registerIntegrationHooks handles fh_integrations events.
func registerIntegrationHooks(app core.App) {
	// Auto-validate integration credentials after creation.
	// Runs in a goroutine so the create response is not blocked.
	app.OnRecordAfterCreateSuccess("fh_integrations").BindFunc(func(e *core.RecordEvent) error {
		id := e.Record.Id
		intType := e.Record.GetString("integrationsType")
		credentials := e.Record.GetString("credentials")

		go func(id, intType, credentials string) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			logger := app.Logger().With(
				"integrationId", id,
				"integrationType", intType,
				"source", "auto-validate",
			)

			record, err := app.FindRecordById("fh_integrations", id)
			if err != nil {
				logger.Error("auto-validate: record not found", "error", err)
				return
			}

			switch intType {
			case "cloudflare":
				accounts, validateErr := integration.ValidateCloudflare(ctx, credentials, logger)
				if validateErr != nil {
					record.Set("status", "invalid")
					logger.Warn("auto-validate: cloudflare validation failed", "error", validateErr)
				} else {
					record.Set("status", "active")
					record.Set("metadata", map[string]any{"accounts": accounts})
					logger.Info("auto-validate: cloudflare validation passed", "accountCount", len(accounts))
				}
			default:
				logger.Debug("auto-validate: skipping unsupported type", "type", intType)
				return
			}

			if saveErr := app.Save(record); saveErr != nil {
				logger.Error("auto-validate: failed to save record", "error", saveErr)
			}
		}(id, intType, credentials)

		return e.Next()
	})
}
