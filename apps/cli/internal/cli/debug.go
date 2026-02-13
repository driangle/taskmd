package cli

import (
	"fmt"
	"os"
)

// debugLog prints a debug message to stderr when debug mode is enabled.
func debugLog(format string, args ...any) {
	if !debug {
		return
	}
	fmt.Fprintf(os.Stderr, "[debug] "+format+"\n", args...)
}
