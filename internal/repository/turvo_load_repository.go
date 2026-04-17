package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FelipeVel/drumkit-int/config"
	"github.com/FelipeVel/drumkit-int/internal/model"
)

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

	var response turvoAPIResponse[turvoLoadsResponse]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("turvo GetAll: decode response: %w", err)
	}

	loads := make([]model.Load, len(response.Details.Shipments))
	for i, tl := range response.Details.Shipments {
		loads[i] = turvoToModel(tl)
	}

	if err := r.enrichShipments(loads); err != nil {
		return nil, err
	}

	if err := r.enrichCustomers(loads); err != nil {
		return nil, err
	}

	return loads, nil
}

// GetShipment fetches a single shipment by ID from Turvo and maps it to the internal model.
func (r *TurvoLoadRepository) GetShipment(id int) (model.Load, error) {
	url := fmt.Sprintf("%s/shipments/%d", r.cfg.TurvoBaseURL, id)

	body, statusCode, err := r.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return model.Load{}, err
	}
	if statusCode != http.StatusOK {
		return model.Load{}, fmt.Errorf("turvo GetShipment: unexpected status %d: %s", statusCode, body)
	}

	var response turvoAPIResponse[turvoShipment]
	if err := json.Unmarshal(body, &response); err != nil {
		return model.Load{}, fmt.Errorf("turvo GetShipment: decode response: %w", err)
	}

	return turvoShipmentToModel(response.Details), nil
}

// enrichShipments replaces each load's basic data with the full shipment detail
// by calling GetShipment concurrently for every load that has an ExternalTMSLoadID.
func (r *TurvoLoadRepository) enrichShipments(loads []model.Load) error {
	type result struct {
		idx  int
		load model.Load
		err  error
	}

	resultc := make(chan result, len(loads))
	var wg sync.WaitGroup

	for i, load := range loads {
		if load.ExternalTMSLoadID == "" {
			continue
		}
		id, err := strconv.Atoi(load.ExternalTMSLoadID)
		if err != nil {
			continue
		}
		wg.Add(1)
		go func(i, id int) {
			defer wg.Done()
			l, err := r.GetShipment(id)
			resultc <- result{idx: i, load: l, err: err}
		}(i, id)
	}

	go func() {
		wg.Wait()
		close(resultc)
	}()

	for res := range resultc {
		if res.err != nil {
			return fmt.Errorf("turvo GetAll: enrich shipment: %w", res.err)
		}
		loads[res.idx] = res.load
	}

	return nil
}

// enrichCustomers fetches full customer details concurrently for each load that
// has a non-empty ExternalTMSId and fills the load's Customer Party in place.
func (r *TurvoLoadRepository) enrichCustomers(loads []model.Load) error {
	type result struct {
		idx int
		c   model.Customer
		err error
	}

	resultc := make(chan result, len(loads))
	var wg sync.WaitGroup

	for i, load := range loads {
		if load.Customer.ExternalTMSId == "" {
			continue
		}
		id, err := strconv.Atoi(load.Customer.ExternalTMSId)
		if err != nil {
			continue
		}
		wg.Add(1)
		go func(i, id int) {
			defer wg.Done()
			c, err := r.GetCustomer(id)
			resultc <- result{idx: i, c: c, err: err}
		}(i, id)
	}

	go func() {
		wg.Wait()
		close(resultc)
	}()

	for res := range resultc {
		fmt.Println(res.c)
		if res.err != nil {
			return fmt.Errorf("turvo GetAll: enrich customer: %w", res.err)
		}
		loads[res.idx].Customer = model.Party{
			ExternalTMSId: res.c.ExternalTMSId,
			Name:          res.c.Name,
			AddressLine1:  res.c.AddressLine1,
			AddressLine2:  res.c.AddressLine2,
			City:          res.c.City,
			State:         res.c.State,
			Zipcode:       res.c.Zipcode,
			Country:       res.c.Country,
			Contact:       res.c.Contact,
			Phone:         res.c.Phone,
			Email:         res.c.Email,
		}
	}

	return nil
}

// GetCustomer fetches a customer by ID from Turvo and maps it to the internal model.
func (r *TurvoLoadRepository) GetCustomer(id int) (model.Customer, error) {
	url := fmt.Sprintf("%s/customers/%d", r.cfg.TurvoBaseURL, id)

	body, statusCode, err := r.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return model.Customer{}, err
	}
	if statusCode != http.StatusOK {
		return model.Customer{}, fmt.Errorf("turvo GetCustomer: unexpected status %d: %s", statusCode, body)
	}

	var response turvoAPIResponse[turvoCustomerDetails]
	if err := json.Unmarshal(body, &response); err != nil {
		return model.Customer{}, fmt.Errorf("turvo GetCustomer: decode response: %w", err)
	}

	return turvoCustomerToModel(response.Details), nil
}

// Create sends a new load to Turvo and returns the created entity.
// Both driver contacts are registered concurrently before the shipment is created.
func (r *TurvoLoadRepository) Create(load model.Load) (int, error) {
	contactIDs, err := r.registerDriverContacts(load)
	if err != nil {
		return 0, err
	}

	url := r.cfg.TurvoBaseURL + "/shipments"

	payload := modelToTurvoPayload(load)

	if carrierOrders, ok := payload["carrierOrder"].([]map[string]any); ok && len(carrierOrders) > 0 {
		contacts := make([]map[string]any, 0, len(contactIDs))
		for _, id := range contactIDs {
			if id != 0 {
				contacts = append(contacts, map[string]any{"id": id})
			}
		}
		if len(contacts) > 0 {
			carrierOrders[0]["contacts"] = contacts
		}
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("turvo Create: marshal payload: %w", err)
	}

	body, statusCode, err := r.doRequest(http.MethodPost, url, payloadBytes)
	if err != nil {
		r.cleanupDriverContacts(contactIDs)
		return 0, err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		r.cleanupDriverContacts(contactIDs)
		return 0, fmt.Errorf("turvo Create: unexpected status %d: %s", statusCode, body)
	}

	var envelope turvoAPIResponse[json.RawMessage]
	if err := json.Unmarshal(body, &envelope); err != nil {
		r.cleanupDriverContacts(contactIDs)
		return 0, fmt.Errorf("turvo Create: decode response: %w", err)
	}

	if envelope.Status != "SUCCESS" {
		var errDetails turvoErrorDetails
		_ = json.Unmarshal(envelope.Details, &errDetails)
		r.cleanupDriverContacts(contactIDs)
		return 0, fmt.Errorf("turvo Create: shipment creation failed: %s (code: %s)", errDetails.ErrorMessage, errDetails.ErrorCode)
	}

	var details turvoLoadDetails
	if err := json.Unmarshal(envelope.Details, &details); err != nil {
		r.cleanupDriverContacts(contactIDs)
		return 0, fmt.Errorf("turvo Create: decode load details: %w", err)
	}

	return details.ID, nil
}

// cleanupDriverContacts deletes previously registered driver contacts on shipment failure.
func (r *TurvoLoadRepository) cleanupDriverContacts(ids []int) {
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if err := r.DeleteContact(id); err != nil {
			slog.Warn("turvo: failed to delete driver contact during rollback", slog.Int("contactID", id), slog.String("error", err.Error()))
		}
	}
}

// DeleteContact removes a contact from Turvo by ID.
func (r *TurvoLoadRepository) DeleteContact(id int) error {
	url := fmt.Sprintf("%s/contacts/%d", r.cfg.TurvoBaseURL, id)

	body, statusCode, err := r.doRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusNoContent {
		return fmt.Errorf("turvo DeleteContact: unexpected status %d: %s", statusCode, body)
	}
	return nil
}

// CreateContact sends a new contact to Turvo and returns the created entity.
func (r *TurvoLoadRepository) CreateContact(contact turvoContact) (turvoContactDetails, error) {
	url := r.cfg.TurvoBaseURL + "/contacts"

	payloadBytes, err := json.Marshal(contact)
	if err != nil {
		return turvoContactDetails{}, fmt.Errorf("turvo CreateContact: marshal payload: %w", err)
	}

	body, statusCode, err := r.doRequest(http.MethodPost, url, payloadBytes)
	if err != nil {
		return turvoContactDetails{}, err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return turvoContactDetails{}, fmt.Errorf("turvo CreateContact: unexpected status %d: %s", statusCode, body)
	}

	var resp turvoAPIResponse[turvoContactDetails]
	if err := json.Unmarshal(body, &resp); err != nil {
		return turvoContactDetails{}, fmt.Errorf("turvo CreateContact: decode response: %w", err)
	}

	return resp.Details, nil
}

type driverContactResult struct {
	idx int
	id  int
	err error
}

// registerDriverContacts registers both carrier drivers in Turvo concurrently
// and returns their assigned contact IDs in driver order (first, second).
func (r *TurvoLoadRepository) registerDriverContacts(load model.Load) ([]int, error) {
	drivers := []struct{ name, phone string }{
		{load.Carrier.FirstDriverName, load.Carrier.FirstDriverPhone},
		{load.Carrier.SecondDriverName, load.Carrier.SecondDriverPhone},
	}

	resultc := make(chan driverContactResult, len(drivers))

	for i, d := range drivers {
		go func(i int, d struct{ name, phone string }) {
			if d.name == "" {
				resultc <- driverContactResult{idx: i}
				return
			}
			details, err := r.CreateContact(driverContact(d.name, d.phone))
			resultc <- driverContactResult{idx: i, id: details.ID, err: err}
		}(i, d)
	}

	ids := make([]int, len(drivers))
	for range drivers {
		res := <-resultc
		if res.err != nil {
			return nil, fmt.Errorf("turvo Create: register driver contact: %w", res.err)
		}
		ids[res.idx] = res.id
	}
	return ids, nil
}

// driverContact builds a minimal turvoContact for a carrier driver.
func driverContact(name, phone string) turvoContact {
	c := turvoContact{Name: name}
	if phone != "" {
		c.Phone = []turvoContactPhone{
			{
				Number:    phone,
				IsPrimary: true,
				Country:   turvoKeyValue{Key: "us", Value: "+1"},
				Type:      turvoKeyValue{Key: "1001", Value: "Work"},
			},
		}
	}
	return c
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

// truncate shortens a string to maxLen characters for safe log output.
func truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "…"
}
