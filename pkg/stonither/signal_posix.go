package stonither

import (
	"os"
	"syscall"
)

// ShutdownSignals ...
var ShutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
