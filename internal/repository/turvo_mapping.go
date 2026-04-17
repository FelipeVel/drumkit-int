package repository

import (
	"fmt"
	"math/rand"

	"github.com/FelipeVel/drumkit-int/internal/model"
)

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

// turvoLoadDetailsToModel maps the shipment creation response details to the
// internal domain model. Dates are extracted from the turvoShipmentDate wrapper.
func turvoLoadDetailsToModel(tl turvoLoadDetails) model.Load {
	load := model.Load{
		ExternalTMSLoadID: fmt.Sprintf("%d", tl.ID),
		FreightLoadID:     tl.CustomID,
		Status:            tl.Status.Code.Value,
		LtlShipment:       tl.LtlShipment,
		StartDate:         tl.StartDate.Date,
		EndDate:           tl.EndDate.Date,
	}

	if len(tl.CustomerOrder) > 0 && !tl.CustomerOrder[0].Deleted {
		load.Customer = model.Party{
			ExternalTMSId: fmt.Sprintf("%d", tl.CustomerOrder[0].Customer.ID),
			Name:          tl.CustomerOrder[0].Customer.Name,
		}
	}

	if len(tl.CarrierOrder) > 0 && !tl.CarrierOrder[0].Deleted {
		load.Carrier = model.CarrierInfo{
			Name:          tl.CarrierOrder[0].Carrier.Name,
			ExternalTMSId: fmt.Sprintf("%d", tl.CarrierOrder[0].Carrier.ID),
		}
	}

	return load
}

// turvoShipmentToModel maps a full Turvo shipment detail to the internal model.
// This is the inverse of modelToTurvoPayload and provides richer data than turvoToModel:
// it includes pickup/delivery addresses and carrier driver details.
func turvoShipmentToModel(ts turvoShipment) model.Load {
	load := model.Load{
		ExternalTMSLoadID: fmt.Sprintf("%d", ts.ID),
		FreightLoadID:     ts.CustomID,
		Status:            ts.Status.Code.Value,
		LtlShipment:       ts.LtlShipment,
		StartDate:         ts.StartDate.Date,
		EndDate:           ts.EndDate.Date,
	}

	if len(ts.CustomerOrder) > 0 && !ts.CustomerOrder[0].Deleted {
		co := ts.CustomerOrder[0]
		load.Customer = model.Party{
			ExternalTMSId: fmt.Sprintf("%d", co.Customer.ID),
			Name:          co.Customer.Name,
		}
		bt := co.FreightTerms.BillTo
		load.BillTo = model.Party{
			ExternalTMSId: bt.ID,
			Name:          bt.Name,
			AddressLine1:  bt.Address.Line1,
			City:          bt.Address.City.Name,
			State:         bt.Address.State.Name,
			Zipcode:       bt.Address.Zip,
			Country:       bt.Address.Country.Name,
			Phone:         bt.Phone,
			Contact:       bt.Contact,
		}
		for _, route := range co.Route {
			if route.Deleted {
				continue
			}
			stop := model.StopParty{
				Party: model.Party{
					AddressLine1: route.Address.Line1,
					AddressLine2: route.Address.Line2,
					City:         route.Address.City,
					State:        route.Address.State,
					Zipcode:      route.Address.Zip,
					Country:      route.Address.Country,
				},
				Timezone:    route.Timezone,
				WarehouseID: fmt.Sprintf("%d", route.Location.ID),
				ApptTime:    route.Appointment.Start,
			}
			switch route.StopType.Value {
			case "Pickup":
				load.Pickup = stop
			case "Delivery":
				load.Consignee = stop
			}
		}
	}

	for _, co := range ts.CarrierOrder {
		if co.Deleted {
			continue
		}
		load.Carrier = model.CarrierInfo{
			ExternalTMSId: fmt.Sprintf("%d", co.Carrier.ID),
			Name:          co.Carrier.Name,
		}
		driverIdx := 0
		for _, d := range co.Drivers {
			if d.Deleted {
				continue
			}
			if driverIdx == 0 {
				load.Carrier.FirstDriverName = d.Context.Name
				load.Carrier.FirstDriverPhone = d.Phone.Number
			} else if driverIdx == 1 {
				load.Carrier.SecondDriverName = d.Context.Name
				load.Carrier.SecondDriverPhone = d.Phone.Number
			}
			driverIdx++
		}
		for _, ext := range co.ExternalIds {
			if ext.Deleted {
				continue
			}
			switch ext.Type.Key {
			case "7605":
				load.Carrier.ExternalTMSTruckID = ext.Value
			case "7606":
				load.Carrier.ExternalTMSTrailerID = ext.Value
			}
		}
		break
	}

	return load
}

// turvoCustomerToModel maps a Turvo customer details response to the internal domain model.
func turvoCustomerToModel(tc turvoCustomerDetails) model.Customer {
	c := model.Customer{
		ExternalTMSId: fmt.Sprintf("%d", tc.ID),
		Name:          tc.Name,
	}

	for _, a := range tc.Address {
		if a.IsPrimary && !a.Deleted {
			c.AddressLine1 = a.Line1
			c.AddressLine2 = a.Line2
			c.City = a.City
			c.State = a.State
			c.Zipcode = a.Zip
			c.Country = a.Country
			break
		}
	}

	for _, e := range tc.Email {
		if e.IsPrimary && !e.Deleted {
			c.Email = e.Email
			break
		}
	}

	for _, p := range tc.Phone {
		if p.IsPrimary && !p.Deleted {
			c.Phone = p.Number
			break
		}
	}

	c.Contact = tc.Contact.Name

	return c
}

// modelToTurvoPayload converts an internal Load to the JSON payload shape
// expected by Turvo's shipment creation endpoint.
func modelToTurvoPayload(load model.Load) map[string]any {
	statusValue := load.Status
	statusKey, ok := turvoStatusByValue[statusValue]
	if !ok {
		// Default to Processing when the status has no Turvo mapping.
		statusKey = "2109"
		statusValue = turvoStatusCodes[2109]
	}

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
			"start": fmt.Sprintf("%s,%s", load.Pickup.City, load.Pickup.State),
			"end":   fmt.Sprintf("%s,%s", load.Consignee.City, load.Consignee.State),
		},
		"status": map[string]any{
			"code": map[string]any{
				"value": statusValue,
				"key":   statusKey,
			},
		},
	}

	payload["customerOrder"] = []map[string]any{
		{
			"customerOrderSourceId": rand.Int63n(1_000_000_000),
			"customer": map[string]any{
				"name": load.Customer.Name,
				"id":   load.Customer.ExternalTMSId,
			},
			"contacts": []map[string]any{
				{
					"id":   1,
					"name": load.Customer.Name,
					"email": map[string]any{
						"email": load.Customer.Email,
						"type": map[string]any{
							"key":   "1051",
							"value": "Other",
						},
					},
					"phone": []map[string]any{
						{
							"number": load.Customer.Phone,
							"type": map[string]any{
								"key":   "1005",
								"value": "Other",
							},
						},
					},
				},
			},
			"freightTerms": map[string]any{
				"billTo": map[string]any{
					"id":      load.BillTo.ExternalTMSId,
					"billTo":  load.BillTo.Name,
					"address": load.BillTo.AddressLine1,
					"emails":  []string{load.BillTo.Email},
					"phone":   load.BillTo.Phone,
					"contact": load.BillTo.Contact,
				},
			},
		},
	}

	payload["globalRoute"] = []map[string]any{
		{
			"stopType": map[string]any{
				"key":   "1500",
				"value": "Pickup",
			},
			"sequence": 1,
			"location": map[string]any{
				"id": load.Pickup.WarehouseID,
			},
			"appointment": map[string]any{
				"date": load.Pickup.ApptTime,
			},
			"notes":    load.Pickup.ApptNote,
			"timezone": load.Pickup.Timezone,
		},
		{
			"stopType": map[string]any{
				"key":   "1501",
				"value": "Delivery",
			},
			"sequence": 2,
			"location": map[string]any{
				"id": load.Consignee.WarehouseID,
			},
			"appointment": map[string]any{
				"date": load.Consignee.ApptTime,
			},
			"notes":    load.Consignee.ApptNote,
			"timezone": load.Consignee.Timezone,
		},
	}

	payload["carrierOrder"] = []map[string]any{
		{
			"carrier": map[string]any{
				"id":   load.Carrier.ExternalTMSId,
				"name": load.Carrier.Name,
			},
		},
	}

	return payload
}
