package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/parser"
)

// defaultSkipDirs are directories always skipped during scanning.
var defaultSkipDirs = []string{
	"node_modules",
	"vendor",
	"dist",
	"build",
	".next",
	".nuxt",
	"out",
	"target",
	"__pycache__",
	"archive",
}

// Scanner scans directories for markdown task files
type Scanner struct {
	rootDir    string
	verbose    bool
	ignoreDirs map[string]bool
}

// NewScanner creates a new directory scanner.
// ignoreDirs specifies additional directory names to skip during scanning.
func NewScanner(rootDir string, verbose bool, ignoreDirs []string) *Scanner {
	ignoreMap := make(map[string]bool, len(defaultSkipDirs)+len(ignoreDirs))
	for _, d := range defaultSkipDirs {
		ignoreMap[d] = true
	}
	for _, d := range ignoreDirs {
		ignoreMap[d] = true
	}
	return &Scanner{
		rootDir:    rootDir,
		verbose:    verbose,
		ignoreDirs: ignoreMap,
	}
}

// ScanResult contains the results of a directory scan
type ScanResult struct {
	Tasks  []*model.Task
	Errors []ScanError
}

// ScanError represents an error encountered during scanning
type ScanError struct {
	FilePath string
	Error    error
}

// Scan walks the directory tree and finds all markdown files with task frontmatter
//
//nolint:gocognit,funlen // TODO: refactor to reduce complexity
func (s *Scanner) Scan() (*ScanResult, error) {
	result := &ScanResult{
		Tasks:  make([]*model.Task, 0),
		Errors: make([]ScanError, 0),
	}

	// Resolve absolute path
	absRoot, err := filepath.Abs(s.rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path %s: %w", s.rootDir, err)
	}

	if s.verbose {
		fmt.Fprintf(os.Stderr, "Scanning directory: %s\n", absRoot)
	}

	// Walk the directory tree
	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, ScanError{
				FilePath: path,
				Error:    fmt.Errorf("access error: %w", err),
			})
			return nil // Continue walking despite errors
		}

		// Skip directories
		if d.IsDir() {
			// Skip hidden directories and configured/default ignore patterns
			name := d.Name()
			if s.shouldSkipDirectory(name) {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .md files
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		// Try to parse as a task file
		task, err := parser.ParseTaskFile(path)
		if err != nil {
			// Not all .md files are tasks, so we silently skip parse errors
			// unless verbose mode is enabled
			if s.verbose {
				fmt.Fprintf(os.Stderr, "Skipping %s: %v\n", path, err)
			}
			return nil
		}

		// If task has a group field in frontmatter, use it
		// Otherwise, derive group from directory structure
		if task.Group == "" {
			task.Group = deriveGroupFromPath(absRoot, path)
		}

		result.Tasks = append(result.Tasks, task)

		if s.verbose {
			fmt.Fprintf(os.Stderr, "Found task: %s - %s\n", task.ID, task.Title)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("directory walk failed: %w", err)
	}

	if s.verbose {
		fmt.Fprintf(os.Stderr, "Scan complete. Found %d tasks\n", len(result.Tasks))
	}

	return result, nil
}

// shouldSkipDirectory determines if a directory should be skipped during scanning.
func (s *Scanner) shouldSkipDirectory(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	return s.ignoreDirs[name]
}

// deriveGroupFromPath derives a group name from the file's directory path
// relative to the root scan directory
func deriveGroupFromPath(rootDir, filePath string) string {
	// Get the directory containing the file
	dir := filepath.Dir(filePath)

	// Get relative path from root
	relPath, err := filepath.Rel(rootDir, dir)
	if err != nil || relPath == "." {
		return ""
	}

	// Use the immediate parent directory as the group name
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return ""
}
