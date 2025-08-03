package testing2

type ShippingAddress struct {
	City   string `json:"city" db:"city"`
	Street string `json:"street" db:"street"`
}
