package local

import "github.com/dgraph-io/badger/v3"

type Option struct {
	Badger *badger.Options `json:"badger" mapstructure:"badger" validate:"omitempty,required"`
}

func NewOption() *Option {
	badgerOption := badger.DefaultOptions("/tmp/badger")
	return &Option{
		Badger: &badgerOption,
	}
}
