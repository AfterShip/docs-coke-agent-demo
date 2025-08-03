package testing1

type Address struct {
	City   string `json:"city" db:"city" spanner:"city"  readonly:"true"`
	Street string `json:"street" db:"street" spanner:"street"  readonly:"true"`
}

type User struct {
	Name string `json:"name" db:"name"`
	Address
}
