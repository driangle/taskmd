// Command verify runs a single named correctness check for the get-task eval
// suite.
//
// Usage: cd .verify && GOWORK=off go run . <check-name>
//
// It is its own module so it stays out of the repo's go.work, and it chdirs up
// to the project root so checks can address `tasks/` directly.
//
// Unlike the add-task suite, most checks here grade the agent's *reported
// output* rather than the filesystem: `get-task` is read-only, so there is no
// created file to inspect. skival's `check_output` verifier pipes the agent's
// final text to stdin, and the output checks read it from there. The one
// filesystem check is `no-mutation`, which asserts the agent changed nothing.
//
// Exit code 0 means the check passed; any non-zero exit means it failed and
// skival records the sample as incorrect. Failure reasons go to stderr so they
// show up in the run log.
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		exitf("usage: verify <check-name> (available: %s)", strings.Join(checkNames(), ", "))
	}

	name := os.Args[1]
	check, ok := checks[name]
	if !ok {
		exitf("unknown check %q (available: %s)", name, strings.Join(checkNames(), ", "))
	}

	if err := chdirProjectRoot(); err != nil {
		exitf("%v", err)
	}

	if err := check(); err != nil {
		exitf("%s: %v", name, err)
	}

	fmt.Printf("PASS %s\n", name)
}

// chdirProjectRoot makes the taskmd project directory the working directory,
// whether the verifier was launched from the project root or from .verify.
func chdirProjectRoot() error {
	for _, dir := range []string{".", ".."} {
		if info, err := os.Stat(dir + "/tasks"); err == nil && info.IsDir() {
			return os.Chdir(dir)
		}
	}
	return fmt.Errorf("no taskmd project found: neither ./tasks nor ../tasks exists")
}

func checkNames() []string {
	names := make([]string, 0, len(checks))
	for name := range checks {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "FAIL "+format+"\n", args...)
	os.Exit(1)
}
