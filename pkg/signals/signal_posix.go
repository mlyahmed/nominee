package signals

import (
	"os"
	"syscall"
)

// ShutdownSignals ...
var ShutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
