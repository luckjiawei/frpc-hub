// Package integration provides validation helpers for external service integrations.
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// CloudflareAccount holds the identifier and display name of a Cloudflare account
// returned by the /v4/accounts endpoint.
type CloudflareAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// maskToken returns a redacted representation of a credential token safe for logging.
// e.g. "abcd1234efgh5678" → "abcd****5678"
// Tokens shorter than 8 characters are replaced entirely with "****".
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

// ValidateCloudflare calls GET /v4/accounts to verify the token and returns
// the list of accessible accounts. The caller should persist the result in
// the record metadata.
func ValidateCloudflare(ctx context.Context, token string, logger *slog.Logger) ([]CloudflareAccount, error) {
	logger.Info("cloudflare validation started", "token", maskToken(token))

	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet,
		"https://api.cloudflare.com/client/v4/accounts",
		nil,
	)
	if err != nil {
		logger.Error("cloudflare: failed to build request", "error", err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("cloudflare: network error", "error", err)
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	logger.Debug("cloudflare API response received", "statusCode", resp.StatusCode)

	var body struct {
		Success bool `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		Result []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		logger.Error("cloudflare: failed to decode response", "error", err)
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !body.Success {
		if len(body.Errors) > 0 {
			err := fmt.Errorf("cloudflare: %s (code %d)", body.Errors[0].Message, body.Errors[0].Code)
			logger.Warn("cloudflare validation rejected",
				"code", body.Errors[0].Code,
				"message", body.Errors[0].Message,
				"token", maskToken(token),
			)
			return nil, err
		}
		logger.Warn("cloudflare validation failed without details", "token", maskToken(token))
		return nil, fmt.Errorf("cloudflare validation failed")
	}

	accounts := make([]CloudflareAccount, 0, len(body.Result))
	for _, a := range body.Result {
		accounts = append(accounts, CloudflareAccount{ID: a.ID, Name: a.Name})
	}

	logger.Info("cloudflare validation passed",
		"accountCount", len(accounts),
		"token", maskToken(token),
	)
	return accounts, nil
}
