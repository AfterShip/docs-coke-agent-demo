package apiserver

import (
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/config"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/internal"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/app"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
)

const commandDesc = `The apiserver is a service for agent.`

func NewApp(basename string) *app.App {
	opts := config.NewOptions()
	application := app.NewApp(
		//name 也用于 CMD 的 short description
		basename,
		basename,
		app.WithOptions(opts),
		app.WithDescription(commandDesc),
		app.WithRunFunc(run(opts)),
	)
	return application
}

func run(opts *config.Options) app.RunFunc {
	return func(basename string) error {
		//Init log settings
		log.Init(opts.Log)
		defer log.Flush()

		cfg, err := config.CreateConfigFromOptions(opts)
		if err != nil {
			return err
		}
		return internal.Run(cfg)
	}
}
