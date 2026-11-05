package gitmeta

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/driangle/taskmd/sdk/go/feed"
)

// FeedLog reads Git history, distinguishing a non-repository from other failures.
func FeedLog(args []string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && strings.Contains(string(exitErr.Stderr), "fatal: not a git repository") {
			return "", feed.ErrNotGitRepository
		}
		return "", err
	}
	return string(out), nil
}
