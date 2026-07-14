//go:build !unix

package lock

import (
	"os"
	"time"
)

// flockExclusive is a no-op on platforms without flock support. Cross-process
// locking is not available here; the in-process mutex in Acquire still
// serializes goroutines within a single process. This is a deliberate
// degradation: the primary target for the shared-tracker workflow is unix
// (darwin, linux).
func flockExclusive(_ *os.File, _ time.Time) error {
	return nil
}

// flockUnlock is a no-op on platforms without flock support.
func flockUnlock(_ *os.File) error {
	return nil
}
