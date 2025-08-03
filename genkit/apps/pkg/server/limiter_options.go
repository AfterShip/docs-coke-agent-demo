package server

import errors2 "github.com/mingyuans/errors"

type LimiterOption struct {
	RateForm        string `json:"rate_form" mapstructure:"rate_form"`
	RedisPrefixName string `json:"redis_prefix_name" mapstructure:"redis_prefix_name"`
}

func NewLimiterOption() *LimiterOption {
	return &LimiterOption{
		RateForm:        "10-S",
		RedisPrefixName: "redis:lock:",
	}
}

func (o *LimiterOption) Validate() []error {
	var errors = make([]error, 0)

	if len(o.RedisPrefixName) == 0 {
		errors = append(errors, errors2.New("LimiterOption redis prefix name must not be empty"))
	}

	return errors
}
