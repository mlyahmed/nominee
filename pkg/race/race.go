package race

import (
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/proxy"
	"github/mlyahmed.io/nominee/pkg/service"
)

// Racer ...
type Racer interface {
	Run(service.Service) error
	StopChan() nominee.StopChan
	Cleanup()
}

// Observer ...
type Observer interface {
	Observe(proxy proxy.Proxy) error
	StopChan() nominee.StopChan
	Cleanup()
}
