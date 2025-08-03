package testing1

type Man struct {
	Name           string `json:"name" db:"name"`
	BillingAddress struct {
		City string `json:"city" db:"city"`
	} `json:"billing_address" db:"billing_address"`
}
