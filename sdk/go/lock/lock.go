// Package lock provides lightweight, per-task advisory locking so that
// multiple processes (parallel agents, the web server, and CLI invocations)
// can safely share a single task directory.
//
// A lock is scoped to a single task ID via a sidecar lock file
// (<scanDir>/.taskmd/locks/<id>.lock). Holding the lock around a
// read-modify-write of a task file (or an append to its worklog) prevents
// lost updates between concurrent writers. Locks on different task IDs never
// contend, so the common case — agents working on different tasks — is fully
// parallel.
//
// Two layers guard a critical section:
//
//   - an in-process mutex keyed by the absolute lock path, which serializes
//     goroutines within a single process (e.g. concurrent web requests); and
//   - an OS advisory lock on the sidecar file (flock on unix), which
//     serializes across processes (e.g. the web server racing a CLI `set`).
//
// The OS advisory lock is released automatically if the holding process dies,
// so there are no stale locks to clean up. Cross-process locking relies on
// flock and is therefore only guaranteed on unix (darwin, linux) local
// filesystems; on other platforms locking degrades to in-process only (see
// flock_other.go). Advisory locks on networked filesystems (NFS) are
// unreliable and unsupported.
package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DefaultTimeout is the default maximum time to wait for a lock before failing.
const DefaultTimeout = 5 * time.Second

// pollInterval is how often we retry a non-blocking lock attempt.
const pollInterval = 5 * time.Millisecond

// dirPerm and filePerm are the permissions for the locks directory and files.
const (
	dirPerm  os.FileMode = 0o755
	filePerm os.FileMode = 0o644
)

// procMutexes serializes goroutines within this process, keyed by absolute
// lock path. flock alone does not reliably arbitrate between two file
// descriptors opened by the same process, so we layer an in-process mutex on
// top of it.
var (
	procMu      sync.Mutex
	procMutexes = map[string]*sync.Mutex{}
)

// Lock is a held advisory lock. Release must be called to free it.
type Lock struct {
	f      *os.File
	inProc *sync.Mutex
}

// TaskLockPath returns the sidecar lock file path for a task ID within a scan
// directory, e.g. TaskLockPath("/repo/tasks", "042") ->
// "/repo/tasks/.taskmd/locks/042.lock".
func TaskLockPath(scanDir, taskID string) string {
	return filepath.Join(scanDir, ".taskmd", "locks", taskID+".lock")
}

// Acquire obtains an exclusive lock on lockPath, blocking up to timeout.
// It returns an error if the lock cannot be obtained within the timeout.
func Acquire(lockPath string, timeout time.Duration) (*Lock, error) {
	abs, err := filepath.Abs(lockPath)
	if err != nil {
		abs = lockPath
	}

	if err := os.MkdirAll(filepath.Dir(abs), dirPerm); err != nil {
		return nil, fmt.Errorf("create lock directory: %w", err)
	}

	deadline := time.Now().Add(timeout)

	m := procMutex(abs)
	if err := acquireProcMutex(m, deadline); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(abs, os.O_CREATE|os.O_RDWR, filePerm)
	if err != nil {
		m.Unlock()
		return nil, fmt.Errorf("open lock file: %w", err)
	}

	if err := flockExclusive(f, deadline); err != nil {
		_ = f.Close()
		m.Unlock()
		return nil, err
	}

	return &Lock{f: f, inProc: m}, nil
}

// Release frees the lock. It is safe to call once; subsequent calls are no-ops.
func (l *Lock) Release() error {
	if l == nil || l.f == nil {
		return nil
	}
	err := flockUnlock(l.f)
	if cerr := l.f.Close(); err == nil {
		err = cerr
	}
	l.f = nil
	if l.inProc != nil {
		l.inProc.Unlock()
		l.inProc = nil
	}
	return err
}

// procMutex returns the shared in-process mutex for a lock path, creating it on
// first use.
func procMutex(key string) *sync.Mutex {
	procMu.Lock()
	defer procMu.Unlock()
	m, ok := procMutexes[key]
	if !ok {
		m = &sync.Mutex{}
		procMutexes[key] = m
	}
	return m
}

// acquireProcMutex tries to lock m, retrying until the deadline.
func acquireProcMutex(m *sync.Mutex, deadline time.Time) error {
	for {
		if m.TryLock() {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for in-process lock")
		}
		time.Sleep(pollInterval)
	}
}
