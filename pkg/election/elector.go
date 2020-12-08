package election

type Elector interface {
	Run() error
	Cleanup()
	StopCh() <-chan struct{}
}
