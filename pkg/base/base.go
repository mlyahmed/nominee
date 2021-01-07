package base

// DoneChan ...
type DoneChan <-chan struct{}

// Doer ...
type Doer interface {
	Done() DoneChan
}

// Cleaner ...
type Cleaner interface {
	Cleanup()
}

type Reseter interface {
	Reset()
}
