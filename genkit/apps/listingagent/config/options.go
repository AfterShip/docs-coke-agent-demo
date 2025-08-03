package config

import (
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/internal/infra/cache"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/internal/infra/externalapi"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/internal/infra/mcp"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/llm"
	genericserver "github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/server"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
	"github.com/go-playground/validator/v10"
	cliflag "github.com/marmotedu/component-base/pkg/cli/flag"
)

type Options struct {
	GenericServerRunOptions *genericserver.RunOptions             `json:"server"   mapstructure:"server" validate:"omitempty"`
	InsecureServing         *genericserver.InsecureServingOptions `json:"insecure" mapstructure:"insecure" validate:"required"`
	SecureServing           *genericserver.SecureServingOptions   `json:"secure"   mapstructure:"secure" validate:"omitempty"`
	Log                     *log.Options                          `json:"log" mapstructure:"log" validate:"omitempty"`
	ExternalAPIs            *externalapi.Option                   `json:"external_apis" mapstructure:"external_apis" validate:"omitempty,required"`
	Cache                   *cache.Option                         `json:"cache" mapstructure:"cache" validate:"required"`
	Limiter                 *genericserver.LimiterOption          `json:"limiter"  mapstructure:"limiter" validate:"required"`
	MCP                     *mcp.Option                           `json:"mcp" mapstructure:"mcp" validate:"required"`
	LLM                     *llm.Option                           `json:"llm" mapstructure:"llm" validate:"omitempty"`
}

func NewOptions() *Options {
	o := Options{
		GenericServerRunOptions: genericserver.NewRunOptions(),
		InsecureServing:         genericserver.NewInsecureServingOptions(),
		SecureServing:           genericserver.NewSecureServingOptions(),
		Log:                     log.NewOptions(),
		ExternalAPIs:            externalapi.NewOption(),
		Limiter:                 genericserver.NewLimiterOption(),
		Cache:                   cache.NewOption(),
		MCP:                     mcp.NewOption(),
		LLM:                     llm.NewOption(),
	}
	return &o
}

func (o *Options) Validate() []error {
	var errs []error

	validatorInstance := validator.New()
	err := validatorInstance.Struct(o)
	errs = append(errs, err)

	return errs
}

func (o *Options) Flags() cliflag.NamedFlagSets {
	//这里将 options 配置为 cmd flags，允许 CMD 启动时候手动指定 value.ß
	fss := cliflag.NamedFlagSets{}
	//走配置文件即可，不需要走命令行参数；需要走 CMD 参数的话，可以参考下面的代码
	//o.GenericServerRunOptions.AddFlags(fss.FlagSet("server"))
	//o.InsecureServing.AddFlags(fss.FlagSet("insecure"))
	//o.SecureServing.AddFlags(fss.FlagSet("secure"))
	//o.FeatureOptions.AddFlags(fss.FlagSet("feature"))
	return fss
}
