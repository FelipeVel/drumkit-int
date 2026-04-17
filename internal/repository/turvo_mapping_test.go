package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/FelipeVel/drumkit-int/internal/model"
)

// ---------------------------------------------------------------------------
// turvoToModel
// ---------------------------------------------------------------------------

func TestTurvoToModel_FullLoad(t *testing.T) {
	start := time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 12, 17, 0, 0, 0, time.UTC)

	tl := turvoLoad{
		ID:          906,
		CustomID:    "31436-60978",
		LtlShipment: true,
		StartDate:   start,
		EndDate:     end,
		Status:      turvoStatus{Code: turvoKeyValue{Key: "2102", Value: "Covered"}},
		CustomerOrder: []turvoCustomerEntry{
			{ID: 939, Customer: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}{ID: 77, Name: "Whole Foods"}, Deleted: false},
		},
		CarrierOrder: []turvoCarrierEntry{
			{ID: 628, Carrier: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}{ID: 65, Name: "Fast Freight"}, Deleted: false},
		},
	}

	load := turvoToModel(tl)

	assert.Equal(t, "906", load.ExternalTMSLoadID)
	assert.Equal(t, "31436-60978", load.FreightLoadID)
	assert.Equal(t, "Covered", load.Status)
	assert.True(t, load.LtlShipment)
	assert.Equal(t, start, load.StartDate)
	assert.Equal(t, end, load.EndDate)
	assert.Equal(t, "77", load.Customer.ExternalTMSId)
	assert.Equal(t, "Whole Foods", load.Customer.Name)
	assert.Equal(t, "Fast Freight", load.Carrier.Name)
}

func TestTurvoToModel_DeletedCustomerOrderSkipped(t *testing.T) {
	tl := turvoLoad{
		ID: 1,
		CustomerOrder: []turvoCustomerEntry{
			{ID: 1, Customer: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}{ID: 99, Name: "Ghost"}, Deleted: true},
		},
	}

	load := turvoToModel(tl)

	assert.Empty(t, load.Customer.ExternalTMSId)
	assert.Empty(t, load.Customer.Name)
}

func TestTurvoToModel_DeletedCarrierOrderSkipped(t *testing.T) {
	tl := turvoLoad{
		ID: 1,
		CarrierOrder: []turvoCarrierEntry{
			{ID: 1, Carrier: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}{ID: 99, Name: "Ghost Carrier"}, Deleted: true},
		},
	}

	load := turvoToModel(tl)

	assert.Empty(t, load.Carrier.Name)
}

func TestTurvoToModel_EmptyOrders(t *testing.T) {
	tl := turvoLoad{ID: 5, CustomID: "FL-005"}

	load := turvoToModel(tl)

	assert.Equal(t, "5", load.ExternalTMSLoadID)
	assert.Empty(t, load.Customer.Name)
	assert.Empty(t, load.Carrier.Name)
}

// ---------------------------------------------------------------------------
// turvoShipmentToModel
// ---------------------------------------------------------------------------

func makeShipment() turvoShipment {
	appt := time.Date(2024, 3, 1, 10, 0, 0, 0, time.UTC)
	return turvoShipment{
		ID:          12102,
		CustomID:    "31483-39621",
		LtlShipment: false,
		StartDate:   turvoShipmentDate{Date: appt},
		EndDate:     turvoShipmentDate{Date: appt.Add(48 * time.Hour)},
		Status:      turvoStatus{Code: turvoKeyValue{Key: "2112", Value: "Completed"}},
		CustomerOrder: []turvoShipmentCustomerOrder{
			{
				ID: 12121,
				Customer: struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				}{ID: 64080, Name: "Dhananjay M"},
				FreightTerms: turvoShipmentFreightTerms{
					BillTo: turvoShipmentBillTo{
						ID:   "44bad6fe-401e-42d6-948d-0a14e2158d77",
						Name: "Client 123",
						Address: turvoShipmentBillToAddress{
							Line1:   "135 York Street",
							City:    struct{ Name string `json:"name"` }{Name: "Brooklyn"},
							State:   struct{ Name string `json:"name"` }{Name: "NY"},
							Country: struct{ Name string `json:"name"` }{Name: "US"},
							Zip:     "11201",
						},
						Phone:   "5550001234",
						Contact: "Jane Smith",
					},
				},
				Route: []turvoShipmentRoute{
					{
						StopType: turvoKeyValue{Value: "Pickup"},
						Location: turvoShipmentRouteLocation{ID: 54987},
						Address:  turvoShipmentRouteAddress{Line1: "1035 Hilltop Dr", City: "Itasca", State: "IL", Zip: "60143", Country: "US"},
						Timezone: "America/Chicago",
						Appointment: turvoShipmentRouteAppointment{Start: appt},
						Deleted: false,
					},
					{
						StopType: turvoKeyValue{Value: "Delivery"},
						Location: turvoShipmentRouteLocation{ID: 54988},
						Address:  turvoShipmentRouteAddress{Line1: "90 Finegan Rd", City: "Del Rio", State: "TX", Zip: "78840", Country: "US"},
						Timezone: "America/Chicago",
						Appointment: turvoShipmentRouteAppointment{Start: appt.Add(48 * time.Hour)},
						Deleted: false,
					},
				},
				Deleted: false,
			},
		},
		CarrierOrder: []turvoShipmentCarrierOrder{
			{
				ID: 4849,
				Carrier: struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				}{ID: 65, Name: "Fast Freight"},
				Drivers: []turvoShipmentDriver{
					{Context: turvoShipmentDriverContext{Name: "John Doe"}, Phone: turvoShipmentDriverPhone{Number: "5550001111"}, Deleted: false},
					{Context: turvoShipmentDriverContext{Name: "Jane Roe"}, Phone: turvoShipmentDriverPhone{Number: "5550002222"}, Deleted: false},
				},
				ExternalIds: []turvoShipmentExternalID{
					{Type: turvoKeyValue{Key: "7605"}, Value: "TRUCK-99", Deleted: false},
					{Type: turvoKeyValue{Key: "7606"}, Value: "TRAIL-88", Deleted: false},
				},
				Deleted: false,
			},
		},
	}
}

func TestTurvoShipmentToModel_FullShipment(t *testing.T) {
	load := turvoShipmentToModel(makeShipment())

	assert.Equal(t, "12102", load.ExternalTMSLoadID)
	assert.Equal(t, "31483-39621", load.FreightLoadID)
	assert.Equal(t, "Completed", load.Status)
	assert.False(t, load.LtlShipment)

	// customer
	assert.Equal(t, "64080", load.Customer.ExternalTMSId)
	assert.Equal(t, "Dhananjay M", load.Customer.Name)

	// bill-to
	assert.Equal(t, "44bad6fe-401e-42d6-948d-0a14e2158d77", load.BillTo.ExternalTMSId)
	assert.Equal(t, "Client 123", load.BillTo.Name)
	assert.Equal(t, "135 York Street", load.BillTo.AddressLine1)
	assert.Equal(t, "Brooklyn", load.BillTo.City)
	assert.Equal(t, "NY", load.BillTo.State)
	assert.Equal(t, "11201", load.BillTo.Zipcode)
	assert.Equal(t, "US", load.BillTo.Country)
	assert.Equal(t, "5550001234", load.BillTo.Phone)
	assert.Equal(t, "Jane Smith", load.BillTo.Contact)

	// pickup
	assert.Equal(t, "Itasca", load.Pickup.City)
	assert.Equal(t, "IL", load.Pickup.State)
	assert.Equal(t, "1035 Hilltop Dr", load.Pickup.AddressLine1)
	assert.Equal(t, "54987", load.Pickup.WarehouseID)
	assert.Equal(t, "America/Chicago", load.Pickup.Timezone)

	// delivery
	assert.Equal(t, "Del Rio", load.Consignee.City)
	assert.Equal(t, "TX", load.Consignee.State)
	assert.Equal(t, "54988", load.Consignee.WarehouseID)

	// carrier
	assert.Equal(t, "65", load.Carrier.ExternalTMSId)
	assert.Equal(t, "Fast Freight", load.Carrier.Name)
	assert.Equal(t, "John Doe", load.Carrier.FirstDriverName)
	assert.Equal(t, "5550001111", load.Carrier.FirstDriverPhone)
	assert.Equal(t, "Jane Roe", load.Carrier.SecondDriverName)
	assert.Equal(t, "5550002222", load.Carrier.SecondDriverPhone)
	assert.Equal(t, "TRUCK-99", load.Carrier.ExternalTMSTruckID)
	assert.Equal(t, "TRAIL-88", load.Carrier.ExternalTMSTrailerID)
}

func TestTurvoShipmentToModel_DeletedCustomerOrderSkipped(t *testing.T) {
	s := makeShipment()
	s.CustomerOrder[0].Deleted = true

	load := turvoShipmentToModel(s)

	assert.Empty(t, load.Customer.Name)
	assert.Empty(t, load.BillTo.Name)
	assert.Empty(t, load.Pickup.City)
}

func TestTurvoShipmentToModel_DeletedCarrierOrderSkipsToNext(t *testing.T) {
	s := makeShipment()
	s.CarrierOrder[0].Deleted = true
	s.CarrierOrder = append(s.CarrierOrder, turvoShipmentCarrierOrder{
		ID: 9999,
		Carrier: struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}{ID: 200, Name: "Backup Carrier"},
		Deleted: false,
	})

	load := turvoShipmentToModel(s)

	assert.Equal(t, "Backup Carrier", load.Carrier.Name)
}

func TestTurvoShipmentToModel_DeletedRoutesSkipped(t *testing.T) {
	s := makeShipment()
	s.CustomerOrder[0].Route[0].Deleted = true // pickup deleted

	load := turvoShipmentToModel(s)

	assert.Empty(t, load.Pickup.City)
	assert.Equal(t, "Del Rio", load.Consignee.City) // delivery still mapped
}

func TestTurvoShipmentToModel_DeletedDriversSkipped(t *testing.T) {
	s := makeShipment()
	s.CarrierOrder[0].Drivers[0].Deleted = true // first driver deleted

	load := turvoShipmentToModel(s)

	// second driver becomes first
	assert.Equal(t, "Jane Roe", load.Carrier.FirstDriverName)
	assert.Empty(t, load.Carrier.SecondDriverName)
}

func TestTurvoShipmentToModel_DeletedExternalIdsSkipped(t *testing.T) {
	s := makeShipment()
	s.CarrierOrder[0].ExternalIds[0].Deleted = true

	load := turvoShipmentToModel(s)

	assert.Empty(t, load.Carrier.ExternalTMSTruckID)
	assert.Equal(t, "TRAIL-88", load.Carrier.ExternalTMSTrailerID)
}

func TestTurvoShipmentToModel_NoCarrierOrders(t *testing.T) {
	s := makeShipment()
	s.CarrierOrder = nil

	load := turvoShipmentToModel(s)

	assert.Empty(t, load.Carrier.Name)
}

// ---------------------------------------------------------------------------
// turvoCustomerToModel
// ---------------------------------------------------------------------------

func TestTurvoCustomerToModel_FullDetails(t *testing.T) {
	tc := turvoCustomerDetails{
		ID:   77,
		Name: "Whole Foods Market",
		Contact: turvoContactDetails{Name: "John Buyer"},
		Address: []turvoContactAddressResponse{
			{Line1: "100 Main St", City: "Chicago", State: "IL", Zip: "60601", Country: "US", IsPrimary: true, Deleted: false},
		},
		Email: []turvoContactEmailResponse{
			{Email: "billing@wholefood.com", IsPrimary: true, Deleted: false},
		},
		Phone: []turvoContactPhoneResponse{
			{Number: "3121234567", IsPrimary: true, Deleted: false},
		},
	}

	c := turvoCustomerToModel(tc)

	assert.Equal(t, "77", c.ExternalTMSId)
	assert.Equal(t, "Whole Foods Market", c.Name)
	assert.Equal(t, "100 Main St", c.AddressLine1)
	assert.Equal(t, "Chicago", c.City)
	assert.Equal(t, "IL", c.State)
	assert.Equal(t, "60601", c.Zipcode)
	assert.Equal(t, "US", c.Country)
	assert.Equal(t, "billing@wholefood.com", c.Email)
	assert.Equal(t, "3121234567", c.Phone)
	assert.Equal(t, "John Buyer", c.Contact)
}

func TestTurvoCustomerToModel_NonPrimaryContactsSkipped(t *testing.T) {
	tc := turvoCustomerDetails{
		ID:   10,
		Name: "Acme",
		Address: []turvoContactAddressResponse{
			{Line1: "ignored", IsPrimary: false, Deleted: false},
			{Line1: "100 Primary St", IsPrimary: true, Deleted: false},
		},
		Email: []turvoContactEmailResponse{
			{Email: "ignore@x.com", IsPrimary: false, Deleted: false},
			{Email: "primary@x.com", IsPrimary: true, Deleted: false},
		},
		Phone: []turvoContactPhoneResponse{
			{Number: "0000000000", IsPrimary: false, Deleted: false},
			{Number: "1111111111", IsPrimary: true, Deleted: false},
		},
	}

	c := turvoCustomerToModel(tc)

	assert.Equal(t, "100 Primary St", c.AddressLine1)
	assert.Equal(t, "primary@x.com", c.Email)
	assert.Equal(t, "1111111111", c.Phone)
}

func TestTurvoCustomerToModel_DeletedContactsSkipped(t *testing.T) {
	tc := turvoCustomerDetails{
		ID:   10,
		Name: "Acme",
		Address: []turvoContactAddressResponse{
			{Line1: "deleted", IsPrimary: true, Deleted: true},
		},
		Email: []turvoContactEmailResponse{
			{Email: "deleted@x.com", IsPrimary: true, Deleted: true},
		},
		Phone: []turvoContactPhoneResponse{
			{Number: "9999999999", IsPrimary: true, Deleted: true},
		},
	}

	c := turvoCustomerToModel(tc)

	assert.Empty(t, c.AddressLine1)
	assert.Empty(t, c.Email)
	assert.Empty(t, c.Phone)
}

func TestTurvoCustomerToModel_EmptySlices(t *testing.T) {
	tc := turvoCustomerDetails{ID: 5, Name: "Minimal"}

	c := turvoCustomerToModel(tc)

	assert.Equal(t, "5", c.ExternalTMSId)
	assert.Equal(t, "Minimal", c.Name)
	assert.Empty(t, c.AddressLine1)
	assert.Empty(t, c.Email)
	assert.Empty(t, c.Phone)
	assert.Empty(t, c.Contact)
}

// ---------------------------------------------------------------------------
// modelToTurvoPayload
// ---------------------------------------------------------------------------

func TestModelToTurvoPayload_KnownStatus(t *testing.T) {
	load := model.Load{
		Status:  "Covered",
		Pickup:  model.StopParty{Party: model.Party{City: "Dallas", State: "TX"}},
		Consignee: model.StopParty{Party: model.Party{City: "Houston", State: "TX"}},
	}

	p := modelToTurvoPayload(load)

	status := p["status"].(map[string]any)["code"].(map[string]any)
	assert.Equal(t, "Covered", status["value"])
	assert.Equal(t, "2102", status["key"])
}

func TestModelToTurvoPayload_UnknownStatusDefaultsToProcessing(t *testing.T) {
	load := model.Load{
		Status:  "not-a-real-status",
		Pickup:  model.StopParty{Party: model.Party{City: "A", State: "B"}},
		Consignee: model.StopParty{Party: model.Party{City: "C", State: "D"}},
	}

	p := modelToTurvoPayload(load)

	status := p["status"].(map[string]any)["code"].(map[string]any)
	assert.Equal(t, "2109", status["key"])
	assert.Equal(t, turvoStatusCodes[2109], status["value"])
}

func TestModelToTurvoPayload_LaneBuiltFromStops(t *testing.T) {
	load := model.Load{
		Status:    "Covered",
		Pickup:    model.StopParty{Party: model.Party{City: "Chicago", State: "IL"}},
		Consignee: model.StopParty{Party: model.Party{City: "Dallas", State: "TX"}},
	}

	p := modelToTurvoPayload(load)

	lane := p["lane"].(map[string]any)
	assert.Equal(t, "Chicago,IL", lane["start"])
	assert.Equal(t, "Dallas,TX", lane["end"])
}

func TestModelToTurvoPayload_CustomerAndCarrierMapped(t *testing.T) {
	load := model.Load{
		Status: "Covered",
		Customer: model.Party{ExternalTMSId: "77", Name: "Acme", Email: "a@b.com", Phone: "5550001"},
		BillTo:   model.Party{ExternalTMSId: "bt-1", Name: "Billing Co"},
		Carrier:  model.CarrierInfo{ExternalTMSId: "65", Name: "Fast Freight"},
		Pickup:   model.StopParty{Party: model.Party{City: "A", State: "B"}},
		Consignee: model.StopParty{Party: model.Party{City: "C", State: "D"}},
	}

	p := modelToTurvoPayload(load)

	co := p["customerOrder"].([]map[string]any)[0]
	cust := co["customer"].(map[string]any)
	assert.Equal(t, "Acme", cust["name"])
	assert.Equal(t, "77", cust["id"])

	carrier := p["carrierOrder"].([]map[string]any)[0]["carrier"].(map[string]any)
	assert.Equal(t, "Fast Freight", carrier["name"])
	assert.Equal(t, "65", carrier["id"])
}
