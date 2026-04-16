package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/FelipeVel/drumkit-int/config"
	"github.com/FelipeVel/drumkit-int/internal/model"
)

// ---------------------------------------------------------------------------
// Turvo external API shapes (anti-corruption layer — private to this file)
// ---------------------------------------------------------------------------

type turvoStatusCode struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type turvoStatus struct {
	Code turvoStatusCode `json:"code"`
}

type turvoCustomerEntry struct {
	ID       int `json:"id"`
	Customer struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"customer"`
	Deleted bool `json:"deleted"`
}

type turvoCarrierEntry struct {
	ID      int `json:"id"`
	Carrier struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"carrier"`
	Deleted bool `json:"deleted"`
}

type turvoLane struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type turvoLoad struct {
	ID            int                  `json:"id"`
	CustomID      string               `json:"customId"`
	Status        turvoStatus          `json:"status"`
	LtlShipment   bool                 `json:"ltlShipment"`
	StartDate     time.Time            `json:"startDate"`
	EndDate       time.Time            `json:"endDate"`
	Lane          turvoLane            `json:"lane"`
	CustomerOrder []turvoCustomerEntry `json:"customerOrder"`
	CarrierOrder  []turvoCarrierEntry  `json:"carrierOrder"`
	Created       time.Time            `json:"created"`
	Updated       time.Time            `json:"updated"`
}

type turvoLoadsResponse struct {
	Shipments  []turvoLoad `json:"shipments"`
	Pagination struct {
		Total int `json:"total"`
		Pages int `json:"pages"`
		Page  int `json:"page"`
		Limit int `json:"limit"`
	} `json:"pagination"`
}

type turvoAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds until the token expires
}

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

// ---------------------------------------------------------------------------
// TurvoLoadRepository
// ---------------------------------------------------------------------------

// TurvoLoadRepository implements LoadRepository by delegating to the Turvo
// external TMS API. It manages bearer token caching and automatically
// re-authenticates on token expiry or 401 responses.
type TurvoLoadRepository struct {
	cfg        *config.Config
	httpClient *http.Client
	cache      *tokenCache
}

// NewTurvoLoadRepository constructs a Turvo-backed LoadRepository.
// The caller supplies the *http.Client so timeouts and transports are
// configured once in main.go and shared across all outbound calls.
func NewTurvoLoadRepository(cfg *config.Config, client *http.Client) *TurvoLoadRepository {
	return &TurvoLoadRepository{
		cfg:        cfg,
		httpClient: client,
		cache:      &tokenCache{},
	}
}

// GetAll fetches all loads from Turvo and maps them to the internal model.
func (r *TurvoLoadRepository) GetAll() ([]model.Load, error) {
	url := r.cfg.TurvoBaseURL + "/shipments/list"

	body, statusCode, err := r.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("turvo GetAll: unexpected status %d: %s", statusCode, body)
	}

	type turvoLoadsResponseFull struct {
		Status  string             `json:"status"`
		Details turvoLoadsResponse `json:"details"`
	}

	var response turvoLoadsResponseFull
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("turvo GetAll: decode response: %w", err)
	}

	fmt.Println(response.Details.Shipments)

	loads := make([]model.Load, 0, len(response.Details.Shipments))
	for _, tl := range response.Details.Shipments {
		loads = append(loads, turvoToModel(tl))
	}
	return loads, nil
}

// Create sends a new load to Turvo and returns the created entity.
func (r *TurvoLoadRepository) Create(load model.Load) (model.Load, error) {
	url := r.cfg.TurvoBaseURL + "/shipments"

	payload := modelToTurvoPayload(load)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return model.Load{}, fmt.Errorf("turvo Create: marshal payload: %w", err)
	}

	body, statusCode, err := r.doRequest(http.MethodPost, url, payloadBytes)
	if err != nil {
		return model.Load{}, err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return model.Load{}, fmt.Errorf("turvo Create: unexpected status %d: %s", statusCode, body)
	}

	var tl turvoLoad
	if err := json.Unmarshal(body, &tl); err != nil {
		return model.Load{}, fmt.Errorf("turvo Create: decode response: %w", err)
	}

	return turvoToModel(tl), nil
}

// doRequest executes an authenticated HTTP request against the Turvo API.
// On 401 it invalidates the token cache and retries once.
func (r *TurvoLoadRepository) doRequest(method, url string, payload []byte) ([]byte, int, error) {
	body, status, err := r.executeRequest(method, url, payload)
	if err != nil {
		return nil, 0, err
	}

	if status == http.StatusUnauthorized {
		slog.Warn("turvo: received 401, invalidating token and retrying")
		r.cache.invalidate()
		body, status, err = r.executeRequest(method, url, payload)
		if err != nil {
			return nil, 0, err
		}
	}

	return body, status, nil
}

// executeRequest performs a single HTTP call with auth headers and structured logging.
func (r *TurvoLoadRepository) executeRequest(method, url string, payload []byte) ([]byte, int, error) {
	token, err := r.cache.getToken(r.cfg, r.httpClient)
	if err != nil {
		return nil, 0, fmt.Errorf("turvo: get auth token: %w", err)
	}

	var bodyReader io.Reader
	if payload != nil {
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("turvo: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Api-Key", r.cfg.TurvoAPIKey)
	req.Header.Set("Content-Type", "application/json")

	slog.Info("outbound request",
		slog.String("direction", "outbound"),
		slog.String("method", method),
		slog.String("url", url),
		slog.String("body", truncate(string(payload), 500)),
	)

	start := time.Now()
	resp, err := r.httpClient.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return nil, 0, fmt.Errorf("turvo: execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("turvo: read response: %w", err)
	}

	slog.Info("outbound response",
		slog.String("direction", "outbound"),
		slog.String("method", method),
		slog.String("url", url),
		slog.Int("status", resp.StatusCode),
		slog.Int64("latency_ms", latency),
		slog.String("body", truncate(string(respBody), 500)),
	)

	return respBody, resp.StatusCode, nil
}

// ---------------------------------------------------------------------------
// Mapping helpers (Turvo ↔ internal model)
// ---------------------------------------------------------------------------

// turvoToModel maps a Turvo API load to the internal domain model.
// Only fields that Turvo provides are populated; the rest are left as zero values.
func turvoToModel(tl turvoLoad) model.Load {
	load := model.Load{
		ExternalTMSLoadID: fmt.Sprintf("%d", tl.ID),
		FreightLoadID:     tl.CustomID,
		Status:            tl.Status.Code.Value,
		LtlShipment:       tl.LtlShipment,
		StartDate:         tl.StartDate,
		EndDate:           tl.EndDate,
		Lane: model.Lane{
			Origin:      tl.Lane.Start,
			Destination: tl.Lane.End,
		},
	}

	if len(tl.CustomerOrder) > 0 && !tl.CustomerOrder[0].Deleted {
		load.Customer = model.Party{
			ExternalTMSId: fmt.Sprintf("%d", tl.CustomerOrder[0].Customer.ID),
			Name:          tl.CustomerOrder[0].Customer.Name,
		}
	}

	if len(tl.CarrierOrder) > 0 && !tl.CarrierOrder[0].Deleted {
		load.Carrier = model.CarrierInfo{
			Name: tl.CarrierOrder[0].Carrier.Name,
		}
	}

	return load
}

// modelToTurvoPayload converts an internal Load to the JSON payload shape
// expected by Turvo's shipment creation endpoint.
func modelToTurvoPayload(load model.Load) map[string]any {
	payload := map[string]any{
		"ltlShipment": load.LtlShipment,
		"startDate": map[string]any{
			"date":     load.StartDate,
			"timeZone": "America/New_York",
		},
		"endDate": map[string]any{
			"date":     load.EndDate,
			"timeZone": "America/New_York",
		},
		"lane": map[string]any{
			"start": load.Lane.Origin,
			"end":   load.Lane.Destination,
		},
		"status": map[string]any{
			"code": map[string]any{
				"value": load.Status,
				"key":   "2102",
			},
		},
		"customerOrder": []map[string]any{
			{"customer": map[string]any{"name": load.Customer.Name}},
		},
	}

	if load.Carrier.Name != "" {
		payload["carrierOrder"] = []map[string]any{
			{"carrier": map[string]any{"name": load.Carrier.Name}},
		}
	}

	return payload
}

// truncate shortens a string to maxLen characters for safe log output.
func truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "…"
}
