package repository

import "time"

// ---------------------------------------------------------------------------
// Turvo external API shapes (anti-corruption layer — private to this package)
// ---------------------------------------------------------------------------

type turvoKeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type turvoStatus struct {
	Code turvoKeyValue `json:"code"`
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

type turvoContactAddress struct {
	Line1     string        `json:"line1"`
	Line2     string        `json:"line2,omitempty"`
	City      string        `json:"city"`
	State     string        `json:"state"`
	Zip       string        `json:"zip"`
	Country   string        `json:"country"`
	Type      turvoKeyValue `json:"type"`
	IsPrimary bool          `json:"isPrimary"`
}

type turvoContactEmail struct {
	Email     string        `json:"email"`
	IsPrimary bool          `json:"isPrimary"`
	Type      turvoKeyValue `json:"type"`
}

type turvoContactPhone struct {
	Number    string        `json:"number"`
	Extension int           `json:"extension,omitempty"`
	Country   turvoKeyValue `json:"country"`
	IsPrimary bool          `json:"isPrimary"`
	Type      turvoKeyValue `json:"type"`
}

type turvoContactContext struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type turvoContact struct {
	Name    string                `json:"name"`
	Title   string                `json:"title,omitempty"`
	Address []turvoContactAddress `json:"address,omitempty"`
	Email   []turvoContactEmail   `json:"email,omitempty"`
	Phone   []turvoContactPhone   `json:"phone,omitempty"`
	Context []turvoContactContext `json:"context,omitempty"`
	Role    []turvoKeyValue       `json:"role,omitempty"`
}

// ---------------------------------------------------------------------------
// Turvo contact response shapes
// ---------------------------------------------------------------------------

type turvoContactAddressResponse struct {
	ID        string        `json:"id"`
	Line1     string        `json:"line1"`
	Line2     string        `json:"line2,omitempty"`
	City      string        `json:"city"`
	State     string        `json:"state"`
	Zip       string        `json:"zip"`
	Country   string        `json:"country"`
	Type      turvoKeyValue `json:"type"`
	IsPrimary bool          `json:"isPrimary"`
	Deleted   bool          `json:"deleted"`
}

type turvoContactEmailResponse struct {
	ID        string        `json:"id"`
	Email     string        `json:"email"`
	IsPrimary bool          `json:"isPrimary"`
	Type      turvoKeyValue `json:"type"`
	Deleted   bool          `json:"deleted"`
}

type turvoContactPhoneResponse struct {
	ID        string        `json:"id"`
	Number    string        `json:"number"`
	Extension int           `json:"extension,omitempty"`
	Country   turvoKeyValue `json:"country"`
	IsPrimary bool          `json:"isPrimary"`
	Type      turvoKeyValue `json:"type"`
	Deleted   bool          `json:"deleted"`
}

type turvoContactContextResponse struct {
	AssociationID int    `json:"associationId"`
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Deleted       bool   `json:"deleted"`
}

type turvoContactRoleResponse struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Deleted bool   `json:"deleted"`
}

type turvoContactDetails struct {
	ID       int                           `json:"id"`
	Name     string                        `json:"name"`
	Title    string                        `json:"title,omitempty"`
	NickName string                        `json:"nickName,omitempty"`
	DOB      string                        `json:"dob,omitempty"`
	Address  []turvoContactAddressResponse `json:"address,omitempty"`
	Email    []turvoContactEmailResponse   `json:"email,omitempty"`
	Phone    []turvoContactPhoneResponse   `json:"phone,omitempty"`
	Context  []turvoContactContextResponse `json:"context,omitempty"`
	Role     []turvoContactRoleResponse    `json:"role,omitempty"`
}

// turvoAPIResponse is the generic envelope returned by all Turvo write/read
// endpoints: { "Status": "...", "details": <T> }.
type turvoAPIResponse[T any] struct {
	Status  string `json:"Status"`
	Details T      `json:"details"`
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

// ---------------------------------------------------------------------------
// Turvo shipment creation response shapes
// ---------------------------------------------------------------------------

type turvoErrorDetails struct {
	ErrorMessage string `json:"errorMessage"`
	ErrorCode    string `json:"errorCode"`
}

// turvoShipmentDate wraps a timestamp and timezone as returned by Turvo on
// shipment creation (the API returns dates as objects, not bare timestamps).
type turvoShipmentDate struct {
	Date     time.Time `json:"date"`
	TimeZone string    `json:"timeZone"`
}

// turvoLoadDetails is the shape of the `details` field in the shipment
// creation response. Dates differ from turvoLoad (they are objects here).
type turvoLoadDetails struct {
	ID            int                  `json:"id"`
	CustomID      string               `json:"customId"`
	Status        turvoStatus          `json:"status"`
	LtlShipment   bool                 `json:"ltlShipment"`
	StartDate     turvoShipmentDate    `json:"startDate"`
	EndDate       turvoShipmentDate    `json:"endDate"`
	Lane          turvoLane            `json:"lane"`
	CustomerOrder []turvoCustomerEntry `json:"customerOrder"`
	CarrierOrder  []turvoCarrierEntry  `json:"carrierOrder"`
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

// turvoCustomerDetails is the shape of the `details` field in the customer response.
type turvoCustomerDetails struct {
	ID            int                           `json:"id"`
	Name          string                        `json:"name"`
	TaxID         string                        `json:"taxId"`
	Status        turvoStatus                   `json:"status"`
	ParentAccount struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"parentAccount"`
	Contact turvoContactDetails           `json:"contact"`
	Address []turvoContactAddressResponse `json:"address"`
	Email   []turvoContactEmailResponse   `json:"email"`
	Phone   []turvoContactPhoneResponse   `json:"phone"`
}

type turvoAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds until the token expires
}
