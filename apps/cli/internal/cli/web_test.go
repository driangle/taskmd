package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestWebStart_NonExistentDirectory(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("web", "start", "--task-dir", "/nonexistent/path/that/does/not/exist")
	if res.Err == nil {
		t.Fatal("expected error for non-existent directory")
	}

	if !strings.Contains(res.Err.Error(), "not a valid directory") {
		t.Errorf("expected 'not a valid directory' error, got: %v", res.Err)
	}
}

func TestWebStart_FileInsteadOfDirectory(t *testing.T) {
	repo := newTaskRepo(t, nil)
	filePath := repo.Write("not-a-dir.txt", "hello")

	res := repo.Run("web", "start", "--task-dir", filePath)
	if res.Err == nil {
		t.Fatal("expected error when task-dir points to a file")
	}

	if !strings.Contains(res.Err.Error(), "not a valid directory") {
		t.Errorf("expected 'not a valid directory' error, got: %v", res.Err)
	}
}

func TestWebStart_HostFlagDefaultsToLoopback(t *testing.T) {
	flag := webStartCmd.Flags().Lookup("host")
	if flag == nil {
		t.Fatal("expected a --host flag on `web start`")
	}
	// The default is the security posture of the command: an unauthenticated
	// API that can rewrite task files must not be on every interface unless
	// the operator asked for that.
	if flag.DefValue != "127.0.0.1" {
		t.Errorf("expected --host to default to 127.0.0.1, got %q", flag.DefValue)
	}
}

func TestWebStart_HostFlagIsBoundToViper(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	// viper.Reset() in resetCLIState drops init()'s bindings, so restore them
	// through the production function rather than restating them here: this
	// fails if bindWebFlags ever stops binding web.host.
	bindWebFlags()

	if err := webStartCmd.Flags().Set("host", "192.168.1.5"); err != nil {
		t.Fatalf("failed to set --host: %v", err)
	}
	t.Cleanup(func() { _ = webStartCmd.Flags().Set("host", "127.0.0.1") })

	if got := viper.GetString("web.host"); got != "192.168.1.5" {
		t.Errorf("expected --host to reach viper as web.host, got %q", got)
	}
}

func TestWebStart_HostFromConfigFile(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	projectDir := t.TempDir()
	createConfigFile(t, projectDir, `
web:
  host: 10.0.0.1
`)

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	initConfig()

	if got := viper.GetString("web.host"); got != "10.0.0.1" {
		t.Errorf("expected web.host from .taskmd.yaml, got %q", got)
	}
}

func TestWebStart_HelpDocumentsTheDefaultBind(t *testing.T) {
	long := webStartCmd.Long
	if !strings.Contains(long, "127.0.0.1") {
		t.Errorf("expected `web start` help to state the default bind, got %q", long)
	}
	if !strings.Contains(long, "unauthenticated") {
		t.Errorf("expected `web start` help to note the API is unauthenticated, got %q", long)
	}
}
