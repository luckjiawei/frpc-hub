package httphandler

import (
	"net/http"
	"regexp"

	"github.com/pocketbase/pocketbase/core"
)

var validID = regexp.MustCompile(`^[a-z0-9]{15}$`)

// requireAuth wraps a handler with authentication check.
// Pass an optional core.App to also accept a ?token= query param (needed for EventSource/SSE).
func requireAuth(next func(*core.RequestEvent) error, apps ...core.App) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if e.Auth == nil && len(apps) > 0 {
			if token := e.Request.URL.Query().Get("token"); token != "" {
				if record, err := apps[0].FindAuthRecordByToken(token, core.TokenTypeAuth); err == nil {
					e.Auth = record
				}
			}
		}
		if e.Auth == nil {
			return e.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		return next(e)
	}
}
