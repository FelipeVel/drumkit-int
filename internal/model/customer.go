package model

// Customer is the internal domain entity representing a TMS customer account.
type Customer struct {
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
