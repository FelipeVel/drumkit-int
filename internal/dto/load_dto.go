package dto

import "time"

// PartyDTO is the wire representation of a load participant (customer, bill-to, etc.).
type PartyDTO struct {
	ExternalTMSId string `json:"externalTMSId,omitempty"`
	Name          string `json:"name,omitempty"`
	AddressLine1  string `json:"addressLine1,omitempty"`
	AddressLine2  string `json:"addressLine2,omitempty"`
	City          string `json:"city,omitempty"`
	State         string `json:"state,omitempty"`
	Zipcode       string `json:"zipcode,omitempty"`
	Country       string `json:"country,omitempty"`
	Contact       string `json:"contact,omitempty"`
	Phone         string `json:"phone,omitempty"`
	Email         string `json:"email,omitempty"`
	RefNumber     string `json:"refNumber,omitempty"`
}

// StopPartyDTO extends PartyDTO with scheduling fields for pickup/delivery stops.
type StopPartyDTO struct {
	PartyDTO
	BusinessHours string    `json:"businessHours,omitempty"`
	ReadyTime     time.Time `json:"readyTime,omitempty"`
	ApptTime      time.Time `json:"apptTime,omitempty"`
	ApptNote      string    `json:"apptNote,omitempty"`
	Timezone      string    `json:"timezone,omitempty"`
	WarehouseId   string    `json:"warehouseId,omitempty"`
	MustDeliver   string    `json:"mustDeliver,omitempty"`
}

// CarrierDTO is the wire representation of carrier and driver information.
type CarrierDTO struct {
	McNumber                 string    `json:"mcNumber,omitempty"`
	DotNumber                string    `json:"dotNumber,omitempty"`
	Name                     string    `json:"name,omitempty"`
	Phone                    string    `json:"phone,omitempty"`
	Dispatcher               string    `json:"dispatcher,omitempty"`
	SealNumber               string    `json:"sealNumber,omitempty"`
	Scac                     string    `json:"scac,omitempty"`
	FirstDriverName          string    `json:"firstDriverName,omitempty"`
	FirstDriverPhone         string    `json:"firstDriverPhone,omitempty"`
	SecondDriverName         string    `json:"secondDriverName,omitempty"`
	SecondDriverPhone        string    `json:"secondDriverPhone,omitempty"`
	Email                    string    `json:"email,omitempty"`
	DispatchCity             string    `json:"dispatchCity,omitempty"`
	DispatchState            string    `json:"dispatchState,omitempty"`
	ExternalTMSTruckId       string    `json:"externalTMSTruckId,omitempty"`
	ExternalTMSTrailerId     string    `json:"externalTMSTrailerId,omitempty"`
	ConfirmationSentTime     time.Time `json:"confirmationSentTime,omitempty"`
	ConfirmationReceivedTime time.Time `json:"confirmationReceivedTime,omitempty"`
	DispatchedTime           time.Time `json:"dispatchedTime,omitempty"`
	ExpectedPickupTime       time.Time `json:"expectedPickupTime,omitempty"`
	PickupStart              time.Time `json:"pickupStart,omitempty"`
	PickupEnd                time.Time `json:"pickupEnd,omitempty"`
	ExpectedDeliveryTime     time.Time `json:"expectedDeliveryTime,omitempty"`
	DeliveryStart            time.Time `json:"deliveryStart,omitempty"`
	DeliveryEnd              time.Time `json:"deliveryEnd,omitempty"`
	SignedBy                 string    `json:"signedBy,omitempty"`
	ExternalTMSId            string    `json:"externalTMSId,omitempty"`
}

// RateDataDTO is the wire representation of financial rate information.
type RateDataDTO struct {
	CustomerRateType  string  `json:"customerRateType,omitempty"`
	CustomerNumHours  float64 `json:"customerNumHours,omitempty"`
	CustomerLhRateUsd float64 `json:"customerLhRateUsd,omitempty"`
	FscPercent        float64 `json:"fscPercent,omitempty"`
	FscPerMile        float64 `json:"fscPerMile,omitempty"`
	CarrierRateType   string  `json:"carrierRateType,omitempty"`
	CarrierNumHours   float64 `json:"carrierNumHours,omitempty"`
	CarrierLhRateUsd  float64 `json:"carrierLhRateUsd,omitempty"`
	CarrierMaxRate    float64 `json:"carrierMaxRate,omitempty"`
	NetProfitUsd      float64 `json:"netProfitUsd,omitempty"`
	ProfitPercent     float64 `json:"profitPercent,omitempty"`
}

// SpecificationsDTO is the wire representation of special handling requirements.
type SpecificationsDTO struct {
	MinTempFahrenheit float64 `json:"minTempFahrenheit,omitempty"`
	MaxTempFahrenheit float64 `json:"maxTempFahrenheit,omitempty"`
	LiftgatePickup    bool    `json:"liftgatePickup,omitempty"`
	LiftgateDelivery  bool    `json:"liftgateDelivery,omitempty"`
	InsidePickup      bool    `json:"insidePickup,omitempty"`
	InsideDelivery    bool    `json:"insideDelivery,omitempty"`
	Tarps             bool    `json:"tarps,omitempty"`
	Oversized         bool    `json:"oversized,omitempty"`
	Hazmat            bool    `json:"hazmat,omitempty"`
	Straps            bool    `json:"straps,omitempty"`
	Permits           bool    `json:"permits,omitempty"`
	Escorts           bool    `json:"escorts,omitempty"`
	Seal              bool    `json:"seal,omitempty"`
	CustomBonded      bool    `json:"customBonded,omitempty"`
	Labor             bool    `json:"labor,omitempty"`
}

// CreateLoadRequest is the JSON body accepted by POST /loads.
type CreateLoadRequest struct {
	FreightLoadID  string            `json:"freightLoadID"  binding:"required"`
	Status         string            `json:"status"         binding:"required"`
	Customer       PartyDTO          `json:"customer"       binding:"required"`
	Pickup         StopPartyDTO      `json:"pickup"         binding:"required"`
	Consignee      StopPartyDTO      `json:"consignee"      binding:"required"`
	BillTo         PartyDTO          `json:"billTo"`
	Carrier        CarrierDTO        `json:"carrier"`
	RateData       RateDataDTO       `json:"rateData"`
	Specifications SpecificationsDTO `json:"specifications"`
	InPalletCount  int               `json:"inPalletCount"`
	OutPalletCount int               `json:"outPalletCount"`
	NumCommodities int               `json:"numCommodities"`
	TotalWeight    float64           `json:"totalWeight"`
	BillableWeight float64           `json:"billableWeight"`
	PoNums         string            `json:"poNums"`
	Operator       string            `json:"operator"`
	RouteMiles     float64           `json:"routeMiles"`
}

// LoadResponse is the JSON shape returned for both GET /loads items and POST /loads result.
type LoadResponse struct {
	ExternalTMSLoadID string            `json:"externalTMSLoadID"`
	FreightLoadID     string            `json:"freightLoadID"`
	Status            string            `json:"status"`
	Customer          PartyDTO          `json:"customer"`
	BillTo            PartyDTO          `json:"billTo"`
	Pickup            StopPartyDTO      `json:"pickup"`
	Consignee         StopPartyDTO      `json:"consignee"`
	Carrier           CarrierDTO        `json:"carrier"`
	RateData          RateDataDTO       `json:"rateData"`
	Specifications    SpecificationsDTO `json:"specifications"`
	InPalletCount     int               `json:"inPalletCount"`
	OutPalletCount    int               `json:"outPalletCount"`
	NumCommodities    int               `json:"numCommodities"`
	TotalWeight       float64           `json:"totalWeight"`
	BillableWeight    float64           `json:"billableWeight"`
	PoNums            string            `json:"poNums"`
	Operator          string            `json:"operator"`
	RouteMiles        float64           `json:"routeMiles"`
}
