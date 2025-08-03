package basehttp

import (
	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
	"github.com/go-resty/resty/v2"
	"net/http"
	"time"
)

type RestyOption struct {
	Timeout     time.Duration `json:"timeout" mapstructure:"timeout" validate:"gt=0"`
	RetryCount  int           `json:"retry_count" mapstructure:"retry_count" validate:"gte=0"`
	EnableTrace bool          `json:"enable_trace" mapstructure:"enable_trace"`
	Proxy       string        `json:"proxy" mapstructure:"proxy"`
}

func NewRestyClient(option RestyOption) *resty.Client {
	restyCli := resty.NewWithClient(&http.Client{
		Timeout: option.Timeout,
	})

	if len(option.Proxy) != 0 {
		restyCli.SetProxy(option.Proxy)
	}

	restyCli.SetRetryCount(option.RetryCount)

	if option.EnableTrace {
		restyCli.EnableTrace()
	}

	restyCli.SetLogger(log.WithName("[Resty]"))
	return restyCli
}

func NewRestyOption() RestyOption {
	return RestyOption{
		Timeout:     25 * time.Second,
		RetryCount:  3,
		EnableTrace: false,
		Proxy:       "",
	}
}
