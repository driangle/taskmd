package cli

import (
	"fmt"
	"os"

	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/scanner"
)

// newTaskScanner constructs a task scanner from the resolved scan dir and
// global flags. It is the single place the scanner is configured, so scan
// behavior can evolve here rather than at every call site.
func newTaskScanner(scanDir string, flags GlobalFlags) *scanner.Scanner {
	return scanner.NewScanner(scanDir, flags.Verbose, flags.IgnoreDirs)
}

// scanTasks scans scanDir for tasks, reporting scan errors to stderr when
// verbose is enabled and warning about duplicate IDs. It returns the
// discovered tasks.
func scanTasks(scanDir string, flags GlobalFlags) ([]*model.Task, error) {
	result, err := newTaskScanner(scanDir, flags).Scan()
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	reportScanErrors(result.Errors, flags.Verbose)
	warnDuplicateIDs(result.Tasks)

	return result.Tasks, nil
}

// scanActiveAndArchived scans scanDir for active tasks and also scans the
// archive, applying the same error reporting and duplicate-ID warnings as
// scanTasks. It is used by commands that need both sets (e.g. next, tracks).
func scanActiveAndArchived(scanDir string, flags GlobalFlags) (active, archived []*model.Task, err error) {
	taskScanner := newTaskScanner(scanDir, flags)
	result, err := taskScanner.Scan()
	if err != nil {
		return nil, nil, fmt.Errorf("scan failed: %w", err)
	}

	reportScanErrors(result.Errors, flags.Verbose)
	warnDuplicateIDs(result.Tasks)

	archived, err = taskScanner.ScanArchive()
	if err != nil {
		return nil, nil, fmt.Errorf("archive scan failed: %w", err)
	}

	return result.Tasks, archived, nil
}

// reportScanErrors prints scan errors to stderr when verbose is enabled.
func reportScanErrors(errs []scanner.ScanError, verbose bool) {
	if !verbose || len(errs) == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, "\nWarning: encountered %d errors during scan:\n", len(errs))
	for _, scanErr := range errs {
		fmt.Fprintf(os.Stderr, "  %s: %v\n", scanErr.FilePath, scanErr.Error)
	}
	fmt.Fprintln(os.Stderr)
}
