package web

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/worktree"
)

func TestProjectResolver_Get_CachesResult(t *testing.T) {
	var callCount atomic.Int32
	resolver := NewProjectResolver(func(_ string) (string, []PhaseInfo, error) {
		callCount.Add(1)
		return t.TempDir(), []PhaseInfo{{ID: "p1"}}, nil
	}, false, worktree.Builder{})

	// First call should invoke resolve.
	ctx1, err := resolver.get("proj-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx1 == nil || ctx1.dp == nil {
		t.Fatal("expected non-nil context and data provider")
	}
	if callCount.Load() != 1 {
		t.Fatalf("expected resolve called once, got %d", callCount.Load())
	}

	// Second call should return cached.
	ctx2, err := resolver.get("proj-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx1.dp != ctx2.dp {
		t.Fatal("expected same DataProvider from cache")
	}
	if callCount.Load() != 1 {
		t.Fatalf("expected resolve called once (cached), got %d", callCount.Load())
	}
}

func TestProjectResolver_Get_DifferentProjects(t *testing.T) {
	resolver := NewProjectResolver(func(_ string) (string, []PhaseInfo, error) {
		return t.TempDir(), nil, nil
	}, false, worktree.Builder{})

	ctxA, err := resolver.get("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctxB, err := resolver.get("b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctxA.dp == ctxB.dp {
		t.Fatal("expected different DataProviders for different projects")
	}
}

func TestProjectResolver_Get_NotFoundError(t *testing.T) {
	resolver := NewProjectResolver(func(_ string) (string, []PhaseInfo, error) {
		return "", nil, ErrProjectNotFound
	}, false, worktree.Builder{})

	_, err := resolver.get("unknown")
	if err == nil {
		t.Fatal("expected error for unknown project")
	}
	if err != ErrProjectNotFound {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectResolver_Get_ConcurrentAccess(t *testing.T) {
	var callCount atomic.Int32
	resolver := NewProjectResolver(func(_ string) (string, []PhaseInfo, error) {
		callCount.Add(1)
		return t.TempDir(), nil, nil
	}, false, worktree.Builder{})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := resolver.get("proj")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	// Resolve should be called at most a few times (race between goroutines
	// before the first write lock), but always at least once.
	if callCount.Load() < 1 {
		t.Fatal("expected resolve called at least once")
	}
}
