package repository

import "fmt"

// ---------------------------------------------------------------------------
// Turvo status codes
// ---------------------------------------------------------------------------

// turvoStatusCodes maps Turvo numeric status codes to their human-readable labels.
var turvoStatusCodes = map[int]string{
	2100: "Quote active",
	2101: "Tendered",
	2102: "Covered",
	2103: "Dispatched",
	2104: "At pickup",
	2105: "En route",
	2106: "At delivery",
	2107: "Delivered",
	2108: "Ready for billing",
	2109: "Processing",
	2110: "Carrier paid",
	2111: "Customer paid",
	2112: "Completed",
	2113: "Canceled",
	2114: "Quote inactive",
	2115: "Picked up",
	2116: "Route Complete",
	2117: "Tender - offered",
	2118: "Tender - accepted",
	2119: "Tender - rejected",
	2120: "Draft",
	2121: "Shipment Ready",
	2123: "Acquiring Location",
	2124: "Customs Hold",
	2125: "Arrived",
	2126: "Available",
	2127: "Out Gated",
	2129: "In Gated",
	2131: "Arriving to Port",
	2132: "Berthing",
	2133: "Unloading",
	2134: "Ramped",
	2135: "Deramped",
	2136: "Departed",
	2137: "Held",
	2138: "Out for Delivery",
	2139: "In TransShipment",
	2140: "On Hold",
	2141: "Interline",
}

// turvoStatusByValue is the reverse index: label → string key, built once at init time.
var turvoStatusByValue = func() map[string]string {
	m := make(map[string]string, len(turvoStatusCodes))
	for code, label := range turvoStatusCodes {
		m[label] = fmt.Sprintf("%d", code)
	}
	return m
}()
