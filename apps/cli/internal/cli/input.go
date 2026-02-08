package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	// DefaultTaskFile is the default file to read if no file is specified
	DefaultTaskFile = "tasks.md"
)

// InputResolver handles resolving input sources (stdin, explicit file, default file)
type InputResolver struct {
	useStdin bool
	verbose  bool
}

// NewInputResolver creates a new input resolver
func NewInputResolver(useStdin bool, verbose bool) *InputResolver {
	return &InputResolver{
		useStdin: useStdin,
		verbose:  verbose,
	}
}

// ResolveInput determines the input source and returns a reader and cleanup function
// Priority: stdin flag > explicit file argument > default tasks.md file
func (ir *InputResolver) ResolveInput(args []string) (io.Reader, func() error, error) {
	// Case 1: Read from stdin
	if ir.useStdin {
		if ir.verbose {
			fmt.Fprintln(os.Stderr, "Reading from stdin...")
		}
		return os.Stdin, func() error { return nil }, nil
	}

	// Case 2: Explicit file path provided
	if len(args) > 0 {
		filePath := args[0]
		return ir.openFile(filePath)
	}

	// Case 3: Default to tasks.md in current directory
	return ir.openFile(DefaultTaskFile)
}

// openFile opens a file and returns a reader and cleanup function
func (ir *InputResolver) openFile(filePath string) (io.Reader, func() error, error) {
	// Expand relative path to absolute
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve path %s: %w", filePath, err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("file not found: %s", absPath)
	}

	// Open the file
	file, err := os.Open(absPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file %s: %w", absPath, err)
	}

	if ir.verbose {
		fmt.Fprintf(os.Stderr, "Reading from file: %s\n", absPath)
	}

	cleanup := func() error {
		return file.Close()
	}

	return bufio.NewReader(file), cleanup, nil
}

// ReadAll reads all content from the resolved input source
func (ir *InputResolver) ReadAll(args []string) ([]byte, error) {
	reader, cleanup, err := ir.ResolveInput(args)
	if err != nil {
		return nil, err
	}
	defer cleanup() //nolint:errcheck

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	return content, nil
}
