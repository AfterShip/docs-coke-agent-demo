package testing2

type Order struct {
	ID   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}
