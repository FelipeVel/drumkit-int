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
