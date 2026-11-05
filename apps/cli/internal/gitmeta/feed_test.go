package gitmeta

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/driangle/taskmd/sdk/go/feed"
)

func TestFeedLog_NonRepository(t *testing.T) {
	t.Chdir(t.TempDir())
	_, err := FeedLog([]string{"log"})
	if !errors.Is(err, feed.ErrNotGitRepository) {
		t.Fatalf("got %v", err)
	}
}

func TestFeedLog_EmptyRepositoryError(t *testing.T) {
	t.Chdir(t.TempDir())
	if out, err := exec.Command("git", "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", out, err)
	}
	_, err := FeedLog([]string{"log"})
	if err == nil || errors.Is(err, feed.ErrNotGitRepository) {
		t.Fatalf("unexpected error: %v", err)
	}
}
