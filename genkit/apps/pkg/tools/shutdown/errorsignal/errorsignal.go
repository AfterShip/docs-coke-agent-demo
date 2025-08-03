package errorsignal

import (
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/tools/shutdown"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
	"go.uber.org/zap"
	"os"
)

const Name string = ""

type ErrorSignalManager struct {
	stopChan chan error
}

func (e *ErrorSignalManager) GetName() string {
	return "ErrorSignalManager"
}

func (e *ErrorSignalManager) Start(gs shutdown.GSInterface) error {
	go func() {
		// Block until a error is received.
		err := <-e.stopChan
		log.Warn("start to shutdown as an error received", zap.Error(err))
		gs.StartShutdown(e)
	}()
	return nil
}

func (e *ErrorSignalManager) ShutdownStart() error {
	return nil
}

func (e *ErrorSignalManager) ShutdownFinish() error {
	os.Exit(0)
	return nil
}

func NewErrorSignalManager(stopChan chan error) *ErrorSignalManager {
	return &ErrorSignalManager{
		stopChan: stopChan,
	}
}
