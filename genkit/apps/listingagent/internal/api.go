package internal

import (
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/config"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/internal/infra/cache"
	basemodel "github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/model/code"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/middleware"
	coreServer "github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/server"
	"github.com/gin-gonic/gin"
	"github.com/mingyuans/errors"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
)

type apiServer struct {
	conf *config.Config
}

type preparedAPIServer struct {
	restAPIServer *coreServer.GenericAPIServer
}

func initAppPrivateMiddlewares(apiServer *coreServer.GenericAPIServer, cfg *config.Config) {
	//apiServer.Use(getLimitMiddleware(cfg))
}

func getLimitMiddleware(cfg *config.Config) gin.HandlerFunc {
	errorHandler := mgin.ErrorHandler(func(c *gin.Context, err error) {
		coreServer.NewRestfulResponseBuilder(c).
			Error(errors.WithCode(basemodel.ErrUnknown, "")).
			SendJSON()
	})

	limitReachedHandler := mgin.LimitReachedHandler(func(c *gin.Context) {
		coreServer.NewRestfulResponseBuilder(c).
			Error(errors.WithCode(basemodel.ErrTooManyRequests, "")).
			SendJSON()
	})

	return middleware.LimitWithRedis(
		cfg.Limiter.RateForm,
		cache.GetRedisInfra(),
		cfg.Limiter.RedisPrefixName,
		limitReachedHandler,
		errorHandler)
}

func newAPIServer(cfg *config.Config) apiServer {
	return apiServer{
		conf: cfg,
	}
}

func buildGenericConfig(cfg *config.Config) *coreServer.Config {
	genericConfig := coreServer.NewConfig()
	if cfg.SecureServing != nil {
		genericConfig.SecureServing = cfg.SecureServing
	}
	if cfg.InsecureServing != nil {
		genericConfig.InsecureServing = cfg.InsecureServing
	}
	if cfg.GenericServerRunOptions != nil {
		genericConfig.RunInfo = cfg.GenericServerRunOptions
	}
	return genericConfig
}

func initGenericAPIServer(cfg *config.Config) (*coreServer.GenericAPIServer, error) {
	genericConfig := buildGenericConfig(cfg)

	genericServer, err := genericConfig.Complete().New()
	if err != nil {
		return nil, err
	}

	initAppPrivateMiddlewares(genericServer, cfg)
	return genericServer, nil
}

func (s apiServer) Prepare() (coreServer.PreparedServer, error) {
	genericAPIServer, err := initGenericAPIServer(s.conf)
	if err != nil {
		return preparedAPIServer{}, err
	}

	InstallControllers(genericAPIServer.Engine, *s.conf)

	return preparedAPIServer{
		restAPIServer: genericAPIServer,
	}, nil
}

func (s preparedAPIServer) OnShutdown(string) error {
	if s.restAPIServer != nil {
		s.restAPIServer.Close()
	}
	return nil
}

func (s preparedAPIServer) Run(shutdownChan chan error) error {
	return s.restAPIServer.Run(shutdownChan)
}
