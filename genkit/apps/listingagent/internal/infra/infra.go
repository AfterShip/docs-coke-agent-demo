package infra

import (
	"context"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/config"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/internal/infra/cache"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/internal/infra/externalapi"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/infra/lock"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/server"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
	"go.uber.org/zap"
	"os"
)

// 这里配置 RedisBusiness, Spanner 等等额外的服务组件；
type infraServer struct {
	cfg *config.Config
}

type PreparedInfraServer struct {
	rootContext context.Context
	cancelFunc  context.CancelFunc
}

func NewInfraServer(cfg *config.Config) server.Server {
	return infraServer{
		cfg: cfg,
	}
}

func (s infraServer) Prepare() (server.PreparedServer, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	err := initComponents(ctx, s.cfg)
	return PreparedInfraServer{
		rootContext: ctx,
		cancelFunc:  cancelFunc,
	}, err
}

func initComponents(ctx context.Context, cfg *config.Config) error {
	initFunctions := []func(context.Context, *config.Config) error{
		//initCache,
		//initRedisLock,
		initExternalAPI,
	}

	for _, initFunc := range initFunctions {
		if err := initFunc(ctx, cfg); err != nil {
			return err
		}
	}

	return nil
}

func initExternalAPI(ctx context.Context, cfg *config.Config) error {
	apiKey := os.Getenv("AM_API_KEY")
	if apiKey == "" {
		panic("Please set the AM_API_KEY environment variable")
	}
	cfg.ExternalAPIs.AfterShipAPI.APIKey = apiKey
	return externalapi.InitExternalAPIs(ctx, *cfg.ExternalAPIs)
}

func initCache(ctx context.Context, cfg *config.Config) error {
	return cache.InitCache(ctx, cfg.Cache)
}

func initRedisLock(ctx context.Context, cfg *config.Config) error {
	return lock.InitRedisLock(cache.GetRedisInfra())
}

func (s PreparedInfraServer) OnShutdown(shutdownManager string) error {
	log.Info("infra server shutdown", zap.String("shutdownManager", shutdownManager))
	s.cancelFunc()
	cache.Close()
	return nil
}

func (s PreparedInfraServer) Run(shutdownChan chan error) error {
	return nil
}
