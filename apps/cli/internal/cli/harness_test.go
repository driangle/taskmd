package cli

// Shared command-test harness for the cli package.
//
// This file centralizes the three responsibilities that used to be duplicated
// across ~79 per-command test helpers (resetXFlags / captureXOutput /
// createXTestFiles):
//
//   1. resetCLIState  — restore all mutable global flag/viper state to defaults,
//   2. taskRepo.Run   — execute a command and capture stdout/stderr,
//   3. newTaskRepo    — build an isolated temp task directory from fixtures.
//
// Typical usage in a migrated test:
//
//	repo := newTaskRepo(t, map[string]string{"001-setup.md": "..."})
//	res := repo.Run("get", "001", "--format", "json")
//	if res.Err != nil { t.Fatalf("get failed: %v", res.Err) }
//	// assert on res.Stdout / res.Stderr
//
// Hermeticity: Run invokes the target command's RunE directly (after parsing
// flags through the command tree) rather than rootCmd.Execute(). This is
// deliberate — Execute triggers cobra.OnInitialize(initConfig), which walks the
// filesystem (and falls back to $HOME/.taskmd.yaml) for config discovery and
// would pollute tests with the developer's real config. Running RunE directly
// with a fresh viper keeps every test isolated to its temp repo.
//
// Parallelism: Run swaps the process-global os.Stdout/os.Stderr, so command
// tests driven through Run must NOT call t.Parallel() yet. Pure-function unit
// tests (no Run, no global flag mutation) may. Making read-only command tests
// parallel-safe requires routing command output through cmd.OutOrStdout(), which
// is tracked as a separate follow-up (see task 01kz3nmka).

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// fixturesFS embeds the shared canonical task fixtures under testdata/ so tests
// can seed a repo without depending on the working directory. Each immediate
// subdirectory is one named fixture set; see testdata/README.md.
//
//go:embed testdata
var fixturesFS embed.FS

// cliResult captures the observable outcome of running a CLI command in a test.
type cliResult struct {
	Stdout string
	Stderr string
	Err    error
}

// resetCLIState restores all mutable global CLI state to its defaults so each
// test (or each Run) starts from a clean slate. It replaces the ~35 per-command
// resetXFlags helpers plus resetViper with a single canonical reset.
func resetCLIState() {
	viper.Reset()
	resetFlagTree(rootCmd)
	// Non-flag globals with test-relevant defaults.
	taskDir = "."
	cfgFile = ""
}

// resetFlagTree restores every flag in the command tree to its registered
// default and clears its Changed marker. Because flags are bound to package
// globals (e.g. StringVar(&getFormat, ...)), resetting the flag value also
// resets the bound global — that is how a single reset covers all commands.
//
// Note: the sole StringSlice flag (graph --exclude-status) keeps an internal
// "changed" bit that pflag does not expose, so repeatedly passing that specific
// flag across consecutive graph runs can append rather than replace. Every other
// flag (scalars and the 16 StringArray flags) resets cleanly. Address graph's
// case when graph_test.go is migrated (e.g. one Run per exclude-status value).
func resetFlagTree(cmd *cobra.Command) {
	reset := func(fs *pflag.FlagSet) {
		fs.VisitAll(func(f *pflag.Flag) {
			if sv, ok := f.Value.(pflag.SliceValue); ok {
				_ = sv.Replace(parseDefaultSlice(f.DefValue))
			} else {
				_ = f.Value.Set(f.DefValue)
			}
			f.Changed = false
		})
	}
	reset(cmd.PersistentFlags())
	reset(cmd.Flags())
	for _, sub := range cmd.Commands() {
		resetFlagTree(sub)
	}
}

// parseDefaultSlice converts a pflag slice DefValue ("[a,b]" / "[]") back into a
// []string so slice flags can be reset to their registered default.
func parseDefaultSlice(def string) []string {
	def = strings.TrimPrefix(def, "[")
	def = strings.TrimSuffix(def, "]")
	if def == "" {
		return nil
	}
	return strings.Split(def, ",")
}

// captureOutput runs fn with os.Stdout and os.Stderr redirected to pipes and
// returns whatever fn wrote to each. Output is drained concurrently so writes
// larger than the OS pipe buffer cannot deadlock.
func captureOutput(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()

	origOut, origErr := os.Stdout, os.Stderr
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stderr pipe: %v", err)
	}
	os.Stdout, os.Stderr = outW, errW
	defer func() { os.Stdout, os.Stderr = origOut, origErr }()

	outCh := drainPipe(outR)
	errCh := drainPipe(errR)

	fn()

	_ = outW.Close()
	_ = errW.Close()
	stdout = <-outCh
	stderr = <-errCh
	_ = outR.Close()
	_ = errR.Close()
	return stdout, stderr
}

// drainPipe reads r to EOF on a goroutine and delivers the result on the channel.
func drainPipe(r io.Reader) <-chan string {
	ch := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		ch <- buf.String()
	}()
	return ch
}

// taskRepo is an isolated temporary task directory used by command tests.
type taskRepo struct {
	t   *testing.T
	Dir string
}

// newTaskRepo creates a temp directory seeded with the given task files
// (relative filename -> file content) and returns a handle for running commands
// against it. It replaces the ~50 per-command createXTestFiles helpers. Pass nil
// for an empty repo and add files later with Write.
func newTaskRepo(t *testing.T, files map[string]string) *taskRepo {
	t.Helper()
	r := &taskRepo{t: t, Dir: t.TempDir()}
	for name, content := range files {
		r.Write(name, content)
	}
	return r
}

// newTaskRepoFromFixture creates a temp repo seeded from a named fixture set
// under testdata/ (see testdata/README.md). It replaces hand-written inline
// task-YAML maps for the canonical, recurring task shapes. Layer additional
// one-off files on afterward with Write, or overlay another set with SeedFixture.
func newTaskRepoFromFixture(t *testing.T, set string) *taskRepo {
	t.Helper()
	r := &taskRepo{t: t, Dir: t.TempDir()}
	r.SeedFixture(set)
	return r
}

// SeedFixture copies every task file from the named testdata/ fixture set into
// the repo, preserving the set's relative directory layout (so nested sets like
// subdir-projects land in cli/ and backend/ subdirs). It fails the test if the
// set does not exist, so a typo surfaces immediately rather than as a silently
// empty repo.
func (r *taskRepo) SeedFixture(set string) {
	r.t.Helper()
	root := path.Join("testdata", set)
	entries, err := fs.ReadDir(fixturesFS, root)
	if err != nil || len(entries) == 0 {
		r.t.Fatalf("fixture set %q not found under testdata/", set)
	}
	walkErr := fs.WalkDir(fixturesFS, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		content, readErr := fixturesFS.ReadFile(p)
		if readErr != nil {
			return readErr
		}
		// embed paths always use forward slashes; strip the set root prefix to
		// get the repo-relative name and let Write handle OS-native joining.
		r.Write(strings.TrimPrefix(p, root+"/"), string(content))
		return nil
	})
	if walkErr != nil {
		r.t.Fatalf("seed fixture %q: %v", set, walkErr)
	}
}

// Write creates or overwrites a repo-relative file, creating parent dirs, and
// returns its absolute path.
func (r *taskRepo) Write(name, content string) string {
	r.t.Helper()
	path := filepath.Join(r.Dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		r.t.Fatalf("mkdir for %s: %v", name, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		r.t.Fatalf("write %s: %v", name, err)
	}
	return path
}

// Path returns the absolute path of a repo-relative file.
func (r *taskRepo) Path(name string) string {
	return filepath.Join(r.Dir, name)
}

// Run executes a CLI command against this repo, capturing stdout and stderr.
// Example: repo.Run("get", "001", "--format", "json").
//
// It resets all global CLI state, locates the command and parses its flags
// through the command tree, points the CLI at this repo (unless the test passed
// its own --task-dir/--dir), then invokes the command's RunE directly. Flag
// parse errors and cobra Args-validation errors are returned in Result.Err, so
// error-path tests need no special casing.
func (r *taskRepo) Run(args ...string) cliResult {
	r.t.Helper()
	resetCLIState()

	cmd, flagArgs, findErr := rootCmd.Find(args)
	if findErr != nil {
		return cliResult{Err: findErr}
	}

	var runErr error
	stdout, stderr := captureOutput(r.t, func() {
		if perr := cmd.ParseFlags(flagArgs); perr != nil {
			runErr = perr
			return
		}
		positional := cmd.Flags().Args()
		if verr := cmd.ValidateArgs(positional); verr != nil {
			runErr = verr
			return
		}
		// Point the CLI at this repo unless the test overrode the directory.
		if !flagChanged(cmd, "task-dir") && !flagChanged(cmd, "dir") {
			taskDir = r.Dir
		}
		runErr = invoke(cmd, positional)
	})
	return cliResult{Stdout: stdout, Stderr: stderr, Err: runErr}
}

// invoke calls the command's run function, preferring RunE.
func invoke(cmd *cobra.Command, args []string) error {
	switch {
	case cmd.RunE != nil:
		return cmd.RunE(cmd, args)
	case cmd.Run != nil:
		cmd.Run(cmd, args)
		return nil
	default:
		return fmt.Errorf("command %q has no run function", cmd.Name())
	}
}

// flagChanged reports whether the named flag was explicitly set on cmd.
func flagChanged(cmd *cobra.Command, name string) bool {
	if f := cmd.Flags().Lookup(name); f != nil {
		return f.Changed
	}
	return false
}
