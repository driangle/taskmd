//go:build unix

package lock

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// flockExclusive acquires an exclusive advisory lock on f, retrying the
// non-blocking flock until the deadline passes.
func flockExclusive(f *os.File, deadline time.Time) error {
	fd := int(f.Fd())
	for {
		err := syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return nil
		}
		if err != syscall.EWOULDBLOCK {
			return fmt.Errorf("flock: %w", err)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for file lock on %s", f.Name())
		}
		time.Sleep(pollInterval)
	}
}

// flockUnlock releases the advisory lock held on f.
func flockUnlock(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
