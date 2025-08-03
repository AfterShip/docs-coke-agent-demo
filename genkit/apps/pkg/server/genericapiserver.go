// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"fmt"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/middleware"
	"github.com/mingyuans/errors"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"

	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

// GenericAPIServer contains state for an api server.
// type GenericAPIServer gin.Engine.
type GenericAPIServer struct {
	middlewares []string

	defaultAPIs []string

	// See gin.mode
	mode string
	// SecureServingInfo holds configuration of the TLS server.
	SecureServingInfo *SecureServingOptions

	// InsecureServingInfo holds configuration of the insecure HTTP server.
	InsecureServingInfo *InsecureServingOptions

	// ShutdownTimeout is the timeout used for server shutdown. This specifies the timeout before server
	// gracefully shutdown returns.
	ShutdownTimeout time.Duration

	*gin.Engine
	enableMetrics   bool
	enableProfiling bool
	// wrapper for gin.Engine

	insecureServer, secureServer *http.Server
}

func initGenericAPIServer(s *GenericAPIServer) {
	s.Setup()
	//we placed all middlewares in pkg/middleware directory.
	s.InstallNecessaryMiddlewares()
	s.InstallGenericAPIs()
}

// InstallGenericAPIs install generic apis.
func (s *GenericAPIServer) InstallGenericAPIs() {
	// install metric handler
	if s.enableMetrics {
		prometheus := ginprometheus.NewPrometheus("gin")
		prometheus.Use(s.Engine)
	}

	// install pprof handler
	if s.enableProfiling {
		pprof.Register(s.Engine)
	}

	for _, name := range s.defaultAPIs {
		installAPI, ok := DefaultAPIs[name]
		if !ok {
			log.Warnf("can not find preset api: %s", name)
			continue
		}

		log.Debugf("install preset api: %s", name)
		installAPI(s)
	}
}

// Setup do some setup work for gin engine.
func (s *GenericAPIServer) Setup() {
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Infof("%-6s %-s --> %s (%d handlers)", httpMethod, absolutePath, handlerName, nuHandlers)
	}
}

// InstallNecessaryMiddlewares install generic middlewares.
func (s *GenericAPIServer) InstallNecessaryMiddlewares() {
	// necessary middlewares
	s.Use(gin.Recovery())
	s.Use(middleware.RequestID())
	s.Use(middleware.Logger())
	s.Use(middleware.AccessLogger())
	s.Use(middleware.Secure)
	s.Use(middleware.Cors())

	// install custom middlewares
	for _, m := range s.middlewares {
		mw, ok := middleware.Middlewares[m]
		if !ok {
			log.Warnf("can not find middleware: %s", m)
			continue
		}

		log.Infof("install middleware: %s", m)
		s.Use(mw)
	}
}

func (s *GenericAPIServer) startInsecureServer(errChan chan error) {
	// For scalability, use custom HTTP configuration mode here
	s.insecureServer = &http.Server{
		Addr:    s.InsecureServingInfo.Address(),
		Handler: s,
	}
	log.Infof("Start to listening the incoming requests on http address: %s", s.InsecureServingInfo.Address())
	if err := s.insecureServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("Server stopped",
			zap.String("address", s.InsecureServingInfo.Address()),
			zap.Error(err))
		errChan <- err
	}
	log.Infof("Server on %s stopped", s.InsecureServingInfo.Address())
}

func (s *GenericAPIServer) startSecureServer(errChan chan error) {
	key, cert := s.SecureServingInfo.ServerCert.CertKey.KeyFile, s.SecureServingInfo.ServerCert.CertKey.CertFile
	if cert == "" || key == "" || s.SecureServingInfo.BindPort == 0 {
		log.Info("will not start secure server as cert is not set")
		return
	}

	log.Infof("Start to listening the incoming requests on https address: %s", s.SecureServingInfo.Address())

	if err := s.secureServer.ListenAndServeTLS(cert, key); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("Server stopped",
			zap.String("address", s.InsecureServingInfo.Address()),
			zap.Error(err))
		errChan <- err
	}

	log.Infof("Server on %s stopped", s.SecureServingInfo.Address())
}

func (s *GenericAPIServer) pingHTTPServerIfHealthzEnabled() error {
	// Ping the server to make sure the router is working.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var isHealthzEnabled = false
	for _, name := range s.defaultAPIs {
		if name == "healthz" {
			isHealthzEnabled = true
		}
	}

	if isHealthzEnabled {
		if err := s.ping(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *GenericAPIServer) Run(stopChan chan error) error {
	var insecureServerErrorChan = make(chan error)
	go s.startInsecureServer(insecureServerErrorChan)

	var secureServerErrorChan = make(chan error)
	go s.startSecureServer(secureServerErrorChan)

	select {
	case err := <-secureServerErrorChan:
		return err
	case err := <-insecureServerErrorChan:
		return err
	case <-time.After(1 * time.Second):
	}

	if err := s.pingHTTPServerIfHealthzEnabled(); err != nil {
		return err
	}

	go func() {
		select {
		case err := <-secureServerErrorChan:
			stopChan <- err
		case err := <-insecureServerErrorChan:
			stopChan <- err
		}
	}()
	return nil
}

//// Run spawns the http server. It only returns when the port cannot be listened on initially.
//func (s *GenericAPIServer) Run() error {
//	return s.RunWithWaiting()
//}

// Close graceful shutdown the api server.
func (s *GenericAPIServer) Close() {
	log.Info("Try to shutdown the server gracefully...")
	// The context is used to inform the server it has 10 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if s.secureServer != nil {
		if err := s.secureServer.Shutdown(ctx); err != nil {
			log.Warnf("Shutdown secure server failed: %s", err.Error())
		}
	}

	if s.insecureServer != nil {
		if err := s.insecureServer.Shutdown(ctx); err != nil {
			log.Warnf("Shutdown insecure server failed: %s", err.Error())
		}
	}
}

// ping pings the http server to make sure the router is working.
func (s *GenericAPIServer) ping(ctx context.Context) error {
	url := fmt.Sprintf("http://%s/healthz", s.InsecureServingInfo.Address())
	if strings.Contains(s.InsecureServingInfo.Address(), "0.0.0.0") {
		url = fmt.Sprintf("http://127.0.0.1:%s/healthz", strings.Split(s.InsecureServingInfo.Address(), ":")[1])
	}

	for {
		// Change NewRequest to NewRequestWithContext and pass context it
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		// Ping the server by sending a GET request to `/healthz`.
		// nolint: gosec
		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			log.Info("The router has been deployed successfully.")
			//goland:noinspection GoUnhandledErrorResult
			resp.Body.Close()
			return nil
		}

		// Sleep for a second to continue the next ping.
		log.Info("Waiting for the router, retry in 1 second.")
		time.Sleep(1 * time.Second)

		select {
		case <-ctx.Done():
			log.Fatal("can not ping http server within the specified time interval.")
		default:
		}
	}
	// return fmt.Errorf("the router has no response, or it might took too long to start up")
}
