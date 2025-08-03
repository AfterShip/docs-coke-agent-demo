package server

type Server interface {
	Prepare() (PreparedServer, error)
}

type PreparedServer interface {
	Run(stopChan chan error) error
	OnShutdown(shutdownManager string) error
}
