package cli

import (
	"fmt"
	"os"

	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/scanner"
)

// scanTasks scans scanDir for tasks, reporting scan errors to stderr when
// verbose is enabled and warning about duplicate IDs. It returns the
// discovered tasks.
func scanTasks(scanDir string, flags GlobalFlags) ([]*model.Task, error) {
	taskScanner := scanner.NewScanner(scanDir, flags.Verbose, flags.IgnoreDirs)
	result, err := taskScanner.Scan()
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	reportScanErrors(result.Errors, flags.Verbose)
	warnDuplicateIDs(result.Tasks)

	return result.Tasks, nil
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
