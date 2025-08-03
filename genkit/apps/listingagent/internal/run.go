package internal

import (
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/config"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/internal/infra"
	server2 "github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/server"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/tools/shutdown"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/tools/shutdown/errorsignal"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/tools/shutdown/posixsignal"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
)

func Run(cfg *config.Config) error {
	var shutdownManager = shutdown.New()
	//注册信号管理器
	shutdownManager.AddShutdownManager(posixsignal.NewPosixSignalManager())
	//注册错误信号管理器
	shutdownChan := make(chan error, 1)
	shutdownManager.AddShutdownManager(errorsignal.NewErrorSignalManager(shutdownChan))
	//启动信号管理器
	if err := shutdownManager.Start(); err != nil {
		log.Fatalf("start shutdown manager failed: %s", err.Error())
		return err
	}

	//构建需要的服务，对于 api 类型的服务来说就只有 infra 和 api 服务两个
	var servers = buildServers(cfg)
	//准备服务
	var preparedServers []server2.PreparedServer
	for _, server := range servers {
		preparedServer, err := server.Prepare()
		if err != nil {
			return err
		}
		preparedServers = append(preparedServers, preparedServer)
	}
	//启动服务,如果启动失败，通过 shutdownChan 通知 shutdownManager 进行优雅退出
	var runError error = nil
	for _, preparedServer := range preparedServers {
		if err := preparedServer.Run(shutdownChan); err != nil {
			runError = err
			shutdownChan <- err
			break
		}
		shutdownManager.AddShutdownCallback(preparedServer)
	}

	if runError == nil {
		log.Info("All servers started.")
		//阻塞整个进程，避免退出了
		shutdownManager.Wait()
	}

	return runError
}

func buildServers(cfg *config.Config) []server2.Server {
	return []server2.Server{
		//依赖的组件服务
		infra.NewInfraServer(cfg),
		//对外暴露的 API 服务
		newAPIServer(cfg),
	}
}
