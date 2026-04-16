package model

import "time"

// Lane represents the origin-destination pair for a freight lane.
type Lane struct {
	Origin      string
	Destination string
}

// Load is the internal domain entity representing a freight load.
// It is the canonical shape used across service and repository layers.
// No JSON tags — this is not a wire type.
type Load struct {
	ExternalTMSLoadID string
	FreightLoadID     string
	Status            string
	LtlShipment       bool
	StartDate         time.Time
	EndDate           time.Time
	Lane              Lane
	Customer          Party
	BillTo            Party
	Pickup            StopParty
	Consignee         StopParty
	Carrier           CarrierInfo
	RateData          RateData
	Specifications    Specifications
	InPalletCount     int
	OutPalletCount    int
	NumCommodities    int
	TotalWeight       float64
	BillableWeight    float64
	PoNums            string
	Operator          string
	RouteMiles        float64
}

// Party holds contact and address information for a load participant
// (customer, bill-to, etc.).
type Party struct {
	ExternalTMSId string
	Name          string
	AddressLine1  string
	AddressLine2  string
	City          string
	State         string
	Zipcode       string
	Country       string
	Contact       string
	Phone         string
	Email         string
	RefNumber     string
}

// StopParty extends Party with scheduling information for pickup/delivery stops.
type StopParty struct {
	Party
	BusinessHours string
	ReadyTime     time.Time
	ApptTime      time.Time
	ApptNote      string
	Timezone      string
	WarehouseID   string
	MustDeliver   string
}

// CarrierInfo holds all carrier and driver details for a load.
type CarrierInfo struct {
	McNumber                 string
	DotNumber                string
	Name                     string
	Phone                    string
	Dispatcher               string
	SealNumber               string
	Scac                     string
	FirstDriverName          string
	FirstDriverPhone         string
	SecondDriverName         string
	SecondDriverPhone        string
	Email                    string
	DispatchCity             string
	DispatchState            string
	ExternalTMSTruckID       string
	ExternalTMSTrailerID     string
	ConfirmationSentTime     time.Time
	ConfirmationReceivedTime time.Time
	DispatchedTime           time.Time
	ExpectedPickupTime       time.Time
	PickupStart              time.Time
	PickupEnd                time.Time
	ExpectedDeliveryTime     time.Time
	DeliveryStart            time.Time
	DeliveryEnd              time.Time
	SignedBy                 string
	ExternalTMSId            string
}

// RateData holds financial rate information for a load.
type RateData struct {
	CustomerRateType  string
	CustomerNumHours  float64
	CustomerLhRateUsd float64
	FscPercent        float64
	FscPerMile        float64
	CarrierRateType   string
	CarrierNumHours   float64
	CarrierLhRateUsd  float64
	CarrierMaxRate    float64
	NetProfitUsd      float64
	ProfitPercent     float64
}

// Specifications holds special handling and equipment requirements for a load.
type Specifications struct {
	MinTempFahrenheit float64
	MaxTempFahrenheit float64
	LiftgatePickup    bool
	LiftgateDelivery  bool
	InsidePickup      bool
	InsideDelivery    bool
	Tarps             bool
	Oversized         bool
	Hazmat            bool
	Straps            bool
	Permits           bool
	Escorts           bool
	Seal              bool
	CustomBonded      bool
	Labor             bool
}
