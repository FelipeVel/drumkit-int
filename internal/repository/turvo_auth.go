package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/FelipeVel/drumkit-int/config"
)

// ---------------------------------------------------------------------------
// Token cache
// ---------------------------------------------------------------------------

type tokenCache struct {
	mu           sync.Mutex
	accessToken  string
	refreshToken string
	expiresAt    time.Time
}

// getToken returns a valid bearer token, fetching a new one when the cache is
// empty or the token has expired. On 401 responses the caller should call
// invalidate() and then call getToken again to force a fresh authentication.
func (tc *tokenCache) getToken(cfg *config.Config, client *http.Client) (string, error) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.accessToken != "" && time.Now().Before(tc.expiresAt) {
		return tc.accessToken, nil
	}

	return tc.fetchToken(cfg, client)
}

// invalidate clears the cached token so the next getToken call re-authenticates.
func (tc *tokenCache) invalidate() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.accessToken = ""
	tc.expiresAt = time.Time{}
}

// fetchToken calls the Turvo auth endpoint and stores the result.
// Must be called with tc.mu held.
func (tc *tokenCache) fetchToken(cfg *config.Config, client *http.Client) (string, error) {
	payload := map[string]string{
		"grant_type": "password",
		"username":   cfg.TurvoUsername,
		"password":   cfg.TurvoPassword,
		"scope":      "read+trust+write",
		"type":       "business",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("turvo auth: marshal payload: %w", err)
	}

	start := time.Now()
	path := cfg.TurvoBaseURL + "/oauth/token?client_id=" + cfg.TurvoClientID + "&client_secret=" + cfg.TurvoClientSecret
	slog.Info("outbound request",
		slog.String("direction", "outbound"),
		slog.String("method", http.MethodPost),
		slog.String("url", path),
		slog.String("body", string(body)),
	)

	req, err := http.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("turvo auth: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", cfg.TurvoAPIKey)

	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return "", fmt.Errorf("turvo auth: execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	slog.Info("outbound response",
		slog.String("direction", "outbound"),
		slog.String("url", path),
		slog.Int("status", resp.StatusCode),
		slog.Int64("latency_ms", latency),
		slog.String("body", truncate(string(respBody), 500)),
	)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("turvo auth: unexpected status %d: %s", resp.StatusCode, respBody)
	}

	var authResp turvoAuthResponse
	if err := json.Unmarshal(respBody, &authResp); err != nil {
		return "", fmt.Errorf("turvo auth: decode response: %w", err)
	}

	tc.accessToken = authResp.AccessToken
	tc.refreshToken = authResp.RefreshToken
	tc.expiresAt = time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)

	return tc.accessToken, nil
}
