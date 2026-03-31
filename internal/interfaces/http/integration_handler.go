package httphandler

import (
	"fmt"

	"podux/internal/application/integration"

	"github.com/pocketbase/pocketbase/core"
)

// IntegrationHandler exposes integration-related API routes.
type IntegrationHandler struct {
	app core.App
}

func NewIntegrationHandler(app core.App) *IntegrationHandler {
	return &IntegrationHandler{app: app}
}

func (h *IntegrationHandler) RegisterHandlers(e *core.ServeEvent) {
	// POST /api/integrations/validate
	// Body: { "id": "<record id>" }
	// Validates the credentials of the given integration, updates its status,
	// and writes discovered account info into metadata.
	e.Router.POST("/api/integrations/validate", requireAuth(func(e *core.RequestEvent) error {
		var body struct {
			ID string `json:"id"`
		}
		if err := e.BindBody(&body); err != nil {
			return e.JSON(400, map[string]string{"error": "invalid request body"})
		}
		if !validID.MatchString(body.ID) {
			return e.JSON(400, map[string]string{"error": "invalid id format"})
		}

		record, err := h.app.FindRecordById("fh_integrations", body.ID)
		if err != nil {
			return e.JSON(404, map[string]string{"error": "integration not found"})
		}

		intType := record.GetString("integrationsType")
		name := record.GetString("name")
		credentials := record.GetString("credentials")

		logger := h.app.Logger().With(
			"integrationId", body.ID,
			"integrationName", name,
			"integrationType", intType,
		)
		logger.Info("integration validation requested")

		switch intType {
		case "cloudflare":
			accounts, validateErr := integration.ValidateCloudflare(e.Request.Context(), credentials, logger)
			if validateErr != nil {
				record.Set("status", "invalid")
				if saveErr := h.app.Save(record); saveErr != nil {
					logger.Error("failed to update status to invalid", "error", saveErr)
				} else {
					logger.Info("integration status set to invalid")
				}
				return e.JSON(422, map[string]string{"error": validateErr.Error()})
			}

			record.Set("status", "active")
			record.Set("metadata", map[string]any{"accounts": accounts})
			if saveErr := h.app.Save(record); saveErr != nil {
				logger.Error("failed to update status to active", "error", saveErr)
				return e.JSON(500, map[string]string{"error": "failed to update status"})
			}
			logger.Info("integration validation succeeded, metadata updated", "accountCount", len(accounts))
			return e.JSON(200, map[string]any{
				"message":  "validation successful",
				"status":   "active",
				"accounts": accounts,
			})

		default:
			logger.Warn("unsupported integration type for validation", "type", intType)
			return e.JSON(400, map[string]string{"error": "unsupported integration type: " + intType})
		}
	}))

	// POST /api/integrations/ensure-tunnel
	// Body: { "integrationId": "<id>" }
	// Finds or creates a th_tunnels record (type=cloudflare) for the given integration.
	// Returns { "tunnelId": "<id>" } for the caller to set on fh_proxies.
	e.Router.POST("/api/integrations/ensure-tunnel", requireAuth(func(e *core.RequestEvent) error {
		var body struct {
			IntegrationID string `json:"integrationId"`
		}
		if err := e.BindBody(&body); err != nil {
			return e.JSON(400, map[string]string{"error": "invalid request body"})
		}
		if !validID.MatchString(body.IntegrationID) {
			return e.JSON(400, map[string]string{"error": "invalid integrationId"})
		}

		existing, err := h.app.FindFirstRecordByFilter(
			"th_tunnels",
			fmt.Sprintf(`integrationId = "%s" && type = "cloudflare"`, body.IntegrationID),
		)
		if err == nil {
			return e.JSON(200, map[string]string{"tunnelId": existing.Id})
		}

		collection, err := h.app.FindCollectionByNameOrId("th_tunnels")
		if err != nil {
			return e.JSON(500, map[string]string{"error": "th_tunnels collection not found"})
		}
		record := core.NewRecord(collection)
		record.Set("integrationId", body.IntegrationID)
		record.Set("type", "cloudflare")
		if err := h.app.Save(record); err != nil {
			return e.JSON(500, map[string]string{"error": "failed to create tunnel: " + err.Error()})
		}
		h.app.Logger().Info("Auto-created Cloudflare tunnel", "tunnelId", record.Id, "integrationId", body.IntegrationID)
		return e.JSON(200, map[string]string{"tunnelId": record.Id})
	}))
}
