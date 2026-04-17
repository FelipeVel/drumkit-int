package service

import (
	"time"

	"github.com/FelipeVel/drumkit-int/internal/dto"
	"github.com/FelipeVel/drumkit-int/internal/model"
	"github.com/FelipeVel/drumkit-int/internal/repository"
)

// LoadService implements the business logic for load operations.
// It depends on the LoadRepository interface so the underlying data source
// (Turvo or any future provider) is completely transparent to this layer.
type LoadService struct {
	repo repository.LoadRepository
}

// NewLoadService constructs a LoadService with the given repository.
func NewLoadService(repo repository.LoadRepository) *LoadService {
	return &LoadService{repo: repo}
}

// GetAll retrieves all loads from the repository and maps them to response DTOs.
func (s *LoadService) GetAll() ([]dto.LoadResponse, error) {
	loads, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	responses := make([]dto.LoadResponse, 0, len(loads))
	for _, l := range loads {
		responses = append(responses, toResponse(l))
	}
	return responses, nil
}

// Create maps the incoming request DTO to a domain model, persists it via
// the repository, and returns the created entity as a response DTO.
func (s *LoadService) Create(req dto.CreateLoadRequest) (dto.CreateLoadResponse, error) {
	load := toModel(req)

	created, err := s.repo.Create(load)
	if err != nil {
		return dto.CreateLoadResponse{}, err
	}

	return dto.CreateLoadResponse{
		Id:        created,
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}, nil
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

func toResponse(l model.Load) dto.LoadResponse {
	return dto.LoadResponse{
		ExternalTMSLoadID: l.ExternalTMSLoadID,
		FreightLoadID:     l.FreightLoadID,
		Status:            l.Status,
		Customer:          partyToDTO(l.Customer),
		BillTo:            partyToDTO(l.BillTo),
		Pickup:            stopPartyToDTO(l.Pickup),
		Consignee:         stopPartyToDTO(l.Consignee),
		Carrier:           carrierToDTO(l.Carrier),
		RateData:          rateDataToDTO(l.RateData),
		Specifications:    specificationsToDTO(l.Specifications),
		InPalletCount:     l.InPalletCount,
		OutPalletCount:    l.OutPalletCount,
		NumCommodities:    l.NumCommodities,
		TotalWeight:       l.TotalWeight,
		BillableWeight:    l.BillableWeight,
		PoNums:            l.PoNums,
		Operator:          l.Operator,
		RouteMiles:        l.RouteMiles,
	}
}

func toModel(req dto.CreateLoadRequest) model.Load {
	pickup := stopPartyFromDTO(req.Pickup)
	consignee := stopPartyFromDTO(req.Consignee)

	return model.Load{
		FreightLoadID: req.FreightLoadID,
		Status:        req.Status,
		// LtlShipment defaults to false (full truckload) unless Turvo derives otherwise.
		LtlShipment: false,
		// StartDate is when the load is ready for pickup; EndDate is the delivery appointment.
		StartDate:      pickup.ReadyTime,
		EndDate:        consignee.ApptTime,
		Customer:       partyFromDTO(req.Customer),
		BillTo:         partyFromDTO(req.BillTo),
		Pickup:         pickup,
		Consignee:      consignee,
		Carrier:        carrierFromDTO(req.Carrier),
		RateData:       rateDataFromDTO(req.RateData),
		Specifications: specificationsFromDTO(req.Specifications),
		InPalletCount:  req.InPalletCount,
		OutPalletCount: req.OutPalletCount,
		NumCommodities: req.NumCommodities,
		TotalWeight:    req.TotalWeight,
		BillableWeight: req.BillableWeight,
		PoNums:         req.PoNums,
		Operator:       req.Operator,
		RouteMiles:     req.RouteMiles,
	}
}

// formatLaneStop formats a city and state as "city, state" for Turvo's lane field.
// Returns just the city if state is empty, or an empty string if both are empty.
func formatLaneStop(city, state string) string {
	if city == "" {
		return state
	}
	if state == "" {
		return city
	}
	return city + ", " + state
}

func partyToDTO(p model.Party) dto.PartyDTO {
	return dto.PartyDTO{
		ExternalTMSId: p.ExternalTMSId,
		Name:          p.Name,
		AddressLine1:  p.AddressLine1,
		AddressLine2:  p.AddressLine2,
		City:          p.City,
		State:         p.State,
		Zipcode:       p.Zipcode,
		Country:       p.Country,
		Contact:       p.Contact,
		Phone:         p.Phone,
		Email:         p.Email,
		RefNumber:     p.RefNumber,
	}
}

func partyFromDTO(d dto.PartyDTO) model.Party {
	return model.Party{
		ExternalTMSId: d.ExternalTMSId,
		Name:          d.Name,
		AddressLine1:  d.AddressLine1,
		AddressLine2:  d.AddressLine2,
		City:          d.City,
		State:         d.State,
		Zipcode:       d.Zipcode,
		Country:       d.Country,
		Contact:       d.Contact,
		Phone:         d.Phone,
		Email:         d.Email,
		RefNumber:     d.RefNumber,
	}
}

func stopPartyToDTO(sp model.StopParty) dto.StopPartyDTO {
	d := dto.StopPartyDTO{
		PartyDTO:      partyToDTO(sp.Party),
		BusinessHours: sp.BusinessHours,
		ApptNote:      sp.ApptNote,
		Timezone:      sp.Timezone,
		WarehouseId:   sp.WarehouseID,
		MustDeliver:   sp.MustDeliver,
	}
	if !sp.ReadyTime.IsZero() {
		d.ReadyTime = &sp.ReadyTime
	}
	if !sp.ApptTime.IsZero() {
		d.ApptTime = &sp.ApptTime
	}
	return d
}

func stopPartyFromDTO(d dto.StopPartyDTO) model.StopParty {
	sp := model.StopParty{
		Party:         partyFromDTO(d.PartyDTO),
		BusinessHours: d.BusinessHours,
		ApptNote:      d.ApptNote,
		Timezone:      d.Timezone,
		WarehouseID:   d.WarehouseId,
		MustDeliver:   d.MustDeliver,
	}
	if d.ReadyTime != nil {
		sp.ReadyTime = *d.ReadyTime
	}
	if d.ApptTime != nil {
		sp.ApptTime = *d.ApptTime
	}
	return sp
}

func carrierToDTO(c model.CarrierInfo) dto.CarrierDTO {
	d := dto.CarrierDTO{
		McNumber:             c.McNumber,
		DotNumber:            c.DotNumber,
		Name:                 c.Name,
		Phone:                c.Phone,
		Dispatcher:           c.Dispatcher,
		SealNumber:           c.SealNumber,
		Scac:                 c.Scac,
		FirstDriverName:      c.FirstDriverName,
		FirstDriverPhone:     c.FirstDriverPhone,
		SecondDriverName:     c.SecondDriverName,
		SecondDriverPhone:    c.SecondDriverPhone,
		Email:                c.Email,
		DispatchCity:         c.DispatchCity,
		DispatchState:        c.DispatchState,
		ExternalTMSTruckId:   c.ExternalTMSTruckID,
		ExternalTMSTrailerId: c.ExternalTMSTrailerID,
		SignedBy:             c.SignedBy,
		ExternalTMSId:        c.ExternalTMSId,
	}
	if !c.ConfirmationSentTime.IsZero() {
		d.ConfirmationSentTime = &c.ConfirmationSentTime
	}
	if !c.ConfirmationReceivedTime.IsZero() {
		d.ConfirmationReceivedTime = &c.ConfirmationReceivedTime
	}
	if !c.DispatchedTime.IsZero() {
		d.DispatchedTime = &c.DispatchedTime
	}
	if !c.ExpectedPickupTime.IsZero() {
		d.ExpectedPickupTime = &c.ExpectedPickupTime
	}
	if !c.PickupStart.IsZero() {
		d.PickupStart = &c.PickupStart
	}
	if !c.PickupEnd.IsZero() {
		d.PickupEnd = &c.PickupEnd
	}
	if !c.ExpectedDeliveryTime.IsZero() {
		d.ExpectedDeliveryTime = &c.ExpectedDeliveryTime
	}
	if !c.DeliveryStart.IsZero() {
		d.DeliveryStart = &c.DeliveryStart
	}
	if !c.DeliveryEnd.IsZero() {
		d.DeliveryEnd = &c.DeliveryEnd
	}
	return d
}

func carrierFromDTO(d dto.CarrierDTO) model.CarrierInfo {
	c := model.CarrierInfo{
		McNumber:             d.McNumber,
		DotNumber:            d.DotNumber,
		Name:                 d.Name,
		Phone:                d.Phone,
		Dispatcher:           d.Dispatcher,
		SealNumber:           d.SealNumber,
		Scac:                 d.Scac,
		FirstDriverName:      d.FirstDriverName,
		FirstDriverPhone:     d.FirstDriverPhone,
		SecondDriverName:     d.SecondDriverName,
		SecondDriverPhone:    d.SecondDriverPhone,
		Email:                d.Email,
		DispatchCity:         d.DispatchCity,
		DispatchState:        d.DispatchState,
		ExternalTMSTruckID:   d.ExternalTMSTruckId,
		ExternalTMSTrailerID: d.ExternalTMSTrailerId,
		SignedBy:             d.SignedBy,
		ExternalTMSId:        d.ExternalTMSId,
	}
	if d.ConfirmationSentTime != nil {
		c.ConfirmationSentTime = *d.ConfirmationSentTime
	}
	if d.ConfirmationReceivedTime != nil {
		c.ConfirmationReceivedTime = *d.ConfirmationReceivedTime
	}
	if d.DispatchedTime != nil {
		c.DispatchedTime = *d.DispatchedTime
	}
	if d.ExpectedPickupTime != nil {
		c.ExpectedPickupTime = *d.ExpectedPickupTime
	}
	if d.PickupStart != nil {
		c.PickupStart = *d.PickupStart
	}
	if d.PickupEnd != nil {
		c.PickupEnd = *d.PickupEnd
	}
	if d.ExpectedDeliveryTime != nil {
		c.ExpectedDeliveryTime = *d.ExpectedDeliveryTime
	}
	if d.DeliveryStart != nil {
		c.DeliveryStart = *d.DeliveryStart
	}
	if d.DeliveryEnd != nil {
		c.DeliveryEnd = *d.DeliveryEnd
	}
	return c
}

func rateDataToDTO(r model.RateData) dto.RateDataDTO {
	return dto.RateDataDTO{
		CustomerRateType:  r.CustomerRateType,
		CustomerNumHours:  r.CustomerNumHours,
		CustomerLhRateUsd: r.CustomerLhRateUsd,
		FscPercent:        r.FscPercent,
		FscPerMile:        r.FscPerMile,
		CarrierRateType:   r.CarrierRateType,
		CarrierNumHours:   r.CarrierNumHours,
		CarrierLhRateUsd:  r.CarrierLhRateUsd,
		CarrierMaxRate:    r.CarrierMaxRate,
		NetProfitUsd:      r.NetProfitUsd,
		ProfitPercent:     r.ProfitPercent,
	}
}

func rateDataFromDTO(d dto.RateDataDTO) model.RateData {
	return model.RateData{
		CustomerRateType:  d.CustomerRateType,
		CustomerNumHours:  d.CustomerNumHours,
		CustomerLhRateUsd: d.CustomerLhRateUsd,
		FscPercent:        d.FscPercent,
		FscPerMile:        d.FscPerMile,
		CarrierRateType:   d.CarrierRateType,
		CarrierNumHours:   d.CarrierNumHours,
		CarrierLhRateUsd:  d.CarrierLhRateUsd,
		CarrierMaxRate:    d.CarrierMaxRate,
		NetProfitUsd:      d.NetProfitUsd,
		ProfitPercent:     d.ProfitPercent,
	}
}

func specificationsToDTO(s model.Specifications) dto.SpecificationsDTO {
	return dto.SpecificationsDTO{
		MinTempFahrenheit: s.MinTempFahrenheit,
		MaxTempFahrenheit: s.MaxTempFahrenheit,
		LiftgatePickup:    s.LiftgatePickup,
		LiftgateDelivery:  s.LiftgateDelivery,
		InsidePickup:      s.InsidePickup,
		InsideDelivery:    s.InsideDelivery,
		Tarps:             s.Tarps,
		Oversized:         s.Oversized,
		Hazmat:            s.Hazmat,
		Straps:            s.Straps,
		Permits:           s.Permits,
		Escorts:           s.Escorts,
		Seal:              s.Seal,
		CustomBonded:      s.CustomBonded,
		Labor:             s.Labor,
	}
}

func specificationsFromDTO(d dto.SpecificationsDTO) model.Specifications {
	return model.Specifications{
		MinTempFahrenheit: d.MinTempFahrenheit,
		MaxTempFahrenheit: d.MaxTempFahrenheit,
		LiftgatePickup:    d.LiftgatePickup,
		LiftgateDelivery:  d.LiftgateDelivery,
		InsidePickup:      d.InsidePickup,
		InsideDelivery:    d.InsideDelivery,
		Tarps:             d.Tarps,
		Oversized:         d.Oversized,
		Hazmat:            d.Hazmat,
		Straps:            d.Straps,
		Permits:           d.Permits,
		Escorts:           d.Escorts,
		Seal:              d.Seal,
		CustomBonded:      d.CustomBonded,
		Labor:             d.Labor,
	}
}
