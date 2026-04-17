package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FelipeVel/drumkit-int/internal/dto"
	"github.com/FelipeVel/drumkit-int/internal/model"
)

// ---------------------------------------------------------------------------
// Mock
// ---------------------------------------------------------------------------

type mockLoadRepository struct {
	getAllFn func() ([]model.Load, error)
	createFn func(model.Load) (int, error)
}

func (m *mockLoadRepository) GetAll() ([]model.Load, error) { return m.getAllFn() }
func (m *mockLoadRepository) Create(l model.Load) (int, error) { return m.createFn(l) }

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func svc(getAllFn func() ([]model.Load, error), createFn func(model.Load) (int, error)) *LoadService {
	repo := &mockLoadRepository{getAllFn: getAllFn, createFn: createFn}
	return NewLoadService(repo)
}

var errRepo = errors.New("repository failure")

// ---------------------------------------------------------------------------
// GetAll
// ---------------------------------------------------------------------------

func TestGetAll_Success(t *testing.T) {
	appt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	loads := []model.Load{
		{
			ExternalTMSLoadID: "42",
			FreightLoadID:     "FL-001",
			Status:            "Covered",
			LtlShipment:       false,
			Customer:          model.Party{ExternalTMSId: "77", Name: "Acme Corp", City: "Chicago", State: "IL"},
			BillTo:            model.Party{Name: "Acme Billing", AddressLine1: "100 Main St"},
			Pickup:            model.StopParty{Party: model.Party{City: "Dallas", State: "TX"}, ApptTime: appt},
			Consignee:         model.StopParty{Party: model.Party{City: "Houston", State: "TX"}, ApptTime: appt},
			Carrier:           model.CarrierInfo{Name: "Fast Freight", FirstDriverName: "John Doe", FirstDriverPhone: "5550001111"},
			TotalWeight:       1500.0,
			RouteMiles:        240.5,
		},
	}

	s := svc(func() ([]model.Load, error) { return loads, nil }, nil)

	resp, err := s.GetAll()

	require.NoError(t, err)
	require.Len(t, resp, 1)

	r := resp[0]
	assert.Equal(t, "42", r.ExternalTMSLoadID)
	assert.Equal(t, "FL-001", r.FreightLoadID)
	assert.Equal(t, "Covered", r.Status)
	assert.Equal(t, "77", r.Customer.ExternalTMSId)
	assert.Equal(t, "Acme Corp", r.Customer.Name)
	assert.Equal(t, "Chicago", r.Customer.City)
	assert.Equal(t, "Acme Billing", r.BillTo.Name)
	assert.Equal(t, "100 Main St", r.BillTo.AddressLine1)
	assert.Equal(t, "Dallas", r.Pickup.City)
	assert.Equal(t, "Houston", r.Consignee.City)
	assert.Equal(t, "Fast Freight", r.Carrier.Name)
	assert.Equal(t, "John Doe", r.Carrier.FirstDriverName)
	assert.Equal(t, 1500.0, r.TotalWeight)
	assert.Equal(t, 240.5, r.RouteMiles)
}

func TestGetAll_EmptyList(t *testing.T) {
	s := svc(func() ([]model.Load, error) { return []model.Load{}, nil }, nil)

	resp, err := s.GetAll()

	require.NoError(t, err)
	assert.Empty(t, resp)
}

func TestGetAll_NilList(t *testing.T) {
	s := svc(func() ([]model.Load, error) { return nil, nil }, nil)

	resp, err := s.GetAll()

	require.NoError(t, err)
	assert.Empty(t, resp)
}

func TestGetAll_RepoError(t *testing.T) {
	s := svc(func() ([]model.Load, error) { return nil, errRepo }, nil)

	resp, err := s.GetAll()

	assert.Nil(t, resp)
	require.ErrorIs(t, err, errRepo)
}

func TestGetAll_MultipleLoads(t *testing.T) {
	loads := []model.Load{
		{ExternalTMSLoadID: "1", FreightLoadID: "FL-001", Status: "Covered"},
		{ExternalTMSLoadID: "2", FreightLoadID: "FL-002", Status: "Completed"},
		{ExternalTMSLoadID: "3", FreightLoadID: "FL-003", Status: "In Transit"},
	}

	s := svc(func() ([]model.Load, error) { return loads, nil }, nil)

	resp, err := s.GetAll()

	require.NoError(t, err)
	require.Len(t, resp, 3)
	assert.Equal(t, "FL-001", resp[0].FreightLoadID)
	assert.Equal(t, "FL-002", resp[1].FreightLoadID)
	assert.Equal(t, "FL-003", resp[2].FreightLoadID)
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func newCreateRequest() dto.CreateLoadRequest {
	appt := time.Date(2024, 3, 10, 8, 0, 0, 0, time.UTC)
	return dto.CreateLoadRequest{
		FreightLoadID: "FL-999",
		Status:        "Covered",
		Customer:      dto.CreateCustomerDTO{ExternalTMSId: "77", Name: "Acme Corp", Email: "acme@example.com", Phone: "5550000001"},
		BillTo:        dto.PartyDTO{Name: "Acme Billing"},
		Pickup: dto.CreatePickupDTO{
			City:        "Dallas",
			State:       "TX",
			ReadyTime:   &appt,
			Timezone:    "America/Chicago",
			WarehouseId: "WH-01",
		},
		Consignee: dto.CreateConsigneeDTO{
			City:        "Houston",
			State:       "TX",
			ApptTime:    &appt,
			Timezone:    "America/Chicago",
			WarehouseId: "WH-02",
		},
		Carrier: dto.CarrierDTO{
			Name:             "Fast Freight",
			FirstDriverName:  "John Doe",
			FirstDriverPhone: "5550001111",
		},
		TotalWeight: 2000.0,
		RouteMiles:  300.0,
	}
}

func TestCreate_Success(t *testing.T) {
	var capturedLoad model.Load
	s := svc(nil, func(l model.Load) (int, error) {
		capturedLoad = l
		return 1234, nil
	})

	req := newCreateRequest()
	resp, err := s.Create(req)

	require.NoError(t, err)
	assert.Equal(t, 1234, resp.Id)
	assert.NotEmpty(t, resp.CreatedAt)

	// verify model was mapped correctly from the DTO
	assert.Equal(t, "FL-999", capturedLoad.FreightLoadID)
	assert.Equal(t, "Covered", capturedLoad.Status)
	assert.Equal(t, "Acme Corp", capturedLoad.Customer.Name)
	assert.Equal(t, "acme@example.com", capturedLoad.Customer.Email)
	assert.Equal(t, "Dallas", capturedLoad.Pickup.City)
	assert.Equal(t, "Houston", capturedLoad.Consignee.City)
	assert.Equal(t, "WH-01", capturedLoad.Pickup.WarehouseID)
	assert.Equal(t, "America/Chicago", capturedLoad.Pickup.Timezone)
	assert.Equal(t, "Fast Freight", capturedLoad.Carrier.Name)
	assert.Equal(t, "John Doe", capturedLoad.Carrier.FirstDriverName)
	assert.Equal(t, 2000.0, capturedLoad.TotalWeight)
}

func TestCreate_RepoError(t *testing.T) {
	s := svc(nil, func(l model.Load) (int, error) { return 0, errRepo })

	resp, err := s.Create(newCreateRequest())

	assert.Equal(t, dto.CreateLoadResponse{}, resp)
	require.ErrorIs(t, err, errRepo)
}

func TestCreate_ResponseTimestamp(t *testing.T) {
	before := time.Now().UTC()
	s := svc(nil, func(l model.Load) (int, error) { return 99, nil })

	resp, err := s.Create(newCreateRequest())
	after := time.Now().UTC()

	require.NoError(t, err)

	ts, parseErr := time.Parse(time.RFC3339Nano, resp.CreatedAt)
	require.NoError(t, parseErr, "CreatedAt must be RFC3339Nano")
	assert.True(t, !ts.Before(before) && !ts.After(after), "CreatedAt must be within test window")
}

func TestCreate_StartDateFromPickupReadyTime(t *testing.T) {
	ready := time.Date(2024, 5, 1, 6, 0, 0, 0, time.UTC)
	appt := time.Date(2024, 5, 3, 14, 0, 0, 0, time.UTC)

	var capturedLoad model.Load
	s := svc(nil, func(l model.Load) (int, error) {
		capturedLoad = l
		return 1, nil
	})

	req := newCreateRequest()
	req.Pickup.ReadyTime = &ready
	req.Consignee.ApptTime = &appt

	_, err := s.Create(req)
	require.NoError(t, err)

	assert.Equal(t, ready, capturedLoad.StartDate)
	assert.Equal(t, appt, capturedLoad.EndDate)
}

func TestGetAll_StopPartyReadyTimeMapped(t *testing.T) {
	ready := time.Date(2024, 6, 1, 7, 0, 0, 0, time.UTC)
	appt := time.Date(2024, 6, 3, 15, 0, 0, 0, time.UTC)
	loads := []model.Load{
		{
			Pickup:    model.StopParty{ReadyTime: ready, ApptTime: appt},
			Consignee: model.StopParty{ReadyTime: ready, ApptTime: appt},
		},
	}
	s := svc(func() ([]model.Load, error) { return loads, nil }, nil)

	resp, err := s.GetAll()

	require.NoError(t, err)
	require.Len(t, resp, 1)
	assert.Equal(t, &ready, resp[0].Pickup.ReadyTime)
	assert.Equal(t, &appt, resp[0].Pickup.ApptTime)
	assert.Equal(t, &ready, resp[0].Consignee.ReadyTime)
}

func TestGetAll_CarrierTimestampsMapped(t *testing.T) {
	ts := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	loads := []model.Load{
		{
			Carrier: model.CarrierInfo{
				ConfirmationSentTime:     ts,
				ConfirmationReceivedTime: ts,
				DispatchedTime:           ts,
				ExpectedPickupTime:       ts,
				PickupStart:              ts,
				PickupEnd:                ts,
				ExpectedDeliveryTime:     ts,
				DeliveryStart:            ts,
				DeliveryEnd:              ts,
			},
		},
	}
	s := svc(func() ([]model.Load, error) { return loads, nil }, nil)

	resp, err := s.GetAll()

	require.NoError(t, err)
	c := resp[0].Carrier
	assert.Equal(t, &ts, c.ConfirmationSentTime)
	assert.Equal(t, &ts, c.ConfirmationReceivedTime)
	assert.Equal(t, &ts, c.DispatchedTime)
	assert.Equal(t, &ts, c.ExpectedPickupTime)
	assert.Equal(t, &ts, c.PickupStart)
	assert.Equal(t, &ts, c.PickupEnd)
	assert.Equal(t, &ts, c.ExpectedDeliveryTime)
	assert.Equal(t, &ts, c.DeliveryStart)
	assert.Equal(t, &ts, c.DeliveryEnd)
}

func TestCreate_CarrierTimestampsRoundTrip(t *testing.T) {
	ts := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	var capturedLoad model.Load
	s := svc(nil, func(l model.Load) (int, error) {
		capturedLoad = l
		return 1, nil
	})

	req := newCreateRequest()
	req.Carrier.ConfirmationSentTime = &ts
	req.Carrier.ConfirmationReceivedTime = &ts
	req.Carrier.DispatchedTime = &ts
	req.Carrier.ExpectedPickupTime = &ts
	req.Carrier.PickupStart = &ts
	req.Carrier.PickupEnd = &ts
	req.Carrier.ExpectedDeliveryTime = &ts
	req.Carrier.DeliveryStart = &ts
	req.Carrier.DeliveryEnd = &ts

	_, err := s.Create(req)
	require.NoError(t, err)

	c := capturedLoad.Carrier
	assert.Equal(t, ts, c.ConfirmationSentTime)
	assert.Equal(t, ts, c.ConfirmationReceivedTime)
	assert.Equal(t, ts, c.DispatchedTime)
	assert.Equal(t, ts, c.ExpectedPickupTime)
	assert.Equal(t, ts, c.PickupStart)
	assert.Equal(t, ts, c.PickupEnd)
	assert.Equal(t, ts, c.ExpectedDeliveryTime)
	assert.Equal(t, ts, c.DeliveryStart)
	assert.Equal(t, ts, c.DeliveryEnd)
}

// ---------------------------------------------------------------------------
// formatLaneStop
// ---------------------------------------------------------------------------

func TestFormatLaneStop(t *testing.T) {
	tests := []struct {
		city, state, want string
	}{
		{"Chicago", "IL", "Chicago, IL"},
		{"Dallas", "", "Dallas"},
		{"", "TX", "TX"},
		{"", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, formatLaneStop(tt.city, tt.state))
		})
	}
}

func TestCreate_ZeroTimestampsWhenNil(t *testing.T) {
	var capturedLoad model.Load
	s := svc(nil, func(l model.Load) (int, error) {
		capturedLoad = l
		return 1, nil
	})

	req := newCreateRequest()
	req.Pickup.ReadyTime = nil
	req.Consignee.ApptTime = nil

	_, err := s.Create(req)
	require.NoError(t, err)

	assert.True(t, capturedLoad.StartDate.IsZero())
	assert.True(t, capturedLoad.EndDate.IsZero())
}
