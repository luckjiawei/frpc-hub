package httphandler

import (
	"net/http"

	"podux/internal/application/tunnel"

	"github.com/pocketbase/pocketbase/core"
)

// TunnelHandler provides launch/terminate/reload/log-stream endpoints for any
// tunnel.Provider implementation. The URL prefix (e.g. "/api/frpc" or
// "/api/cloudflare") is supplied at construction time so that the same handler
// code serves both backends without URL collisions.
type TunnelHandler struct {
	app      core.App
	provider tunnel.Provider
	prefix   string
}

func NewTunnelHandler(app core.App, provider tunnel.Provider, prefix string) *TunnelHandler {
	return &TunnelHandler{app: app, provider: provider, prefix: prefix}
}

func (h *TunnelHandler) RegisterHandlers(e *core.ServeEvent) {
	e.Router.POST(h.prefix+"/launch", requireAuth(func(e *core.RequestEvent) error {
		var body struct {
			ID string `json:"id"`
		}
		if err := e.BindBody(&body); err != nil {
			return e.JSON(400, map[string]string{"error": "invalid request body"})
		}
		if body.ID == "" {
			return e.JSON(400, map[string]string{"error": "id is required"})
		}
		if !validID.MatchString(body.ID) {
			return e.JSON(400, map[string]string{"error": "invalid id format"})
		}
		if err := h.provider.Launch(body.ID); err != nil {
			return e.JSON(500, map[string]string{"error": err.Error()})
		}
		return e.JSON(200, map[string]string{"message": "tunnel started"})
	}))

	e.Router.GET(h.prefix+"/logs/stream", requireAuth(func(e *core.RequestEvent) error {
		id := e.Request.URL.Query().Get("id")
		if id == "" {
			return e.JSON(400, map[string]string{"error": "id is required"})
		}
		if !validID.MatchString(id) {
			return e.JSON(400, map[string]string{"error": "invalid id format"})
		}

		w := e.Response
		tunnel.WriteSSEHeaders(w)

		flusher, ok := w.(http.Flusher)
		if !ok {
			return e.JSON(500, map[string]string{"error": "streaming not supported"})
		}

		h.provider.StreamLog(id, e.Request.Context(), w, flusher)
		return nil
	}, h.app))

	e.Router.POST(h.prefix+"/terminate", requireAuth(func(e *core.RequestEvent) error {
		var body struct {
			ID string `json:"id"`
		}
		if err := e.BindBody(&body); err != nil {
			return e.JSON(400, map[string]string{"error": "invalid request body"})
		}
		if body.ID == "" {
			return e.JSON(400, map[string]string{"error": "id is required"})
		}
		if !validID.MatchString(body.ID) {
			return e.JSON(400, map[string]string{"error": "invalid id format"})
		}
		if err := h.provider.Terminate(body.ID); err != nil {
			return e.JSON(500, map[string]string{"error": err.Error()})
		}
		return e.JSON(200, map[string]string{"message": "tunnel terminated"})
	}))

	e.Router.POST(h.prefix+"/reload", requireAuth(func(e *core.RequestEvent) error {
		var body struct {
			ID string `json:"id"`
		}
		if err := e.BindBody(&body); err != nil {
			return e.JSON(400, map[string]string{"error": "invalid request body"})
		}
		if body.ID == "" {
			return e.JSON(400, map[string]string{"error": "id is required"})
		}
		if !validID.MatchString(body.ID) {
			return e.JSON(400, map[string]string{"error": "invalid id format"})
		}
		if !h.provider.IsRunning(body.ID) {
			return e.JSON(200, map[string]string{"message": "not running, skipped"})
		}
		if err := h.provider.Reload(body.ID); err != nil {
			return e.JSON(500, map[string]string{"error": err.Error()})
		}
		return e.JSON(200, map[string]string{"message": "tunnel reloaded"})
	}))
}
