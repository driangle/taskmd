package taskcontext

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

// skipDirs are directories always skipped during directory expansion.
var skipDirs = map[string]bool{
	"node_modules": true,
	".git":         true,
	"vendor":       true,
	"dist":         true,
	"build":        true,
	".next":        true,
	"__pycache__":  true,
	".venv":        true,
}

// generatedFiles are files whose content is never useful to inline.
var generatedFiles = map[string]bool{
	"go.sum":            true,
	"package-lock.json": true,
	"pnpm-lock.yaml":    true,
	"yarn.lock":         true,
	"Gemfile.lock":      true,
	"Cargo.lock":        true,
	"poetry.lock":       true,
	"composer.lock":     true,
	"Pipfile.lock":      true,
}

// ScopeMap maps scope names to their file paths.
type ScopeMap map[string][]string

// FileEntry represents a single file in the resolved context.
type FileEntry struct {
	Path      string `json:"path" yaml:"path"`
	Source    string `json:"source" yaml:"source"`
	Exists    bool   `json:"exists" yaml:"exists"`
	IsDir     bool   `json:"is_dir,omitempty" yaml:"is_dir,omitempty"`
	Binary    bool   `json:"binary,omitempty" yaml:"binary,omitempty"`
	Generated bool   `json:"generated,omitempty" yaml:"generated,omitempty"`
	Content   string `json:"content,omitempty" yaml:"content,omitempty"`
	Lines     int    `json:"lines,omitempty" yaml:"lines,omitempty"`
}

// DepEntry represents a dependency task in the context output.
type DepEntry struct {
	ID     string `json:"id" yaml:"id"`
	Title  string `json:"title" yaml:"title"`
	Status string `json:"status" yaml:"status"`
}

// Result holds the resolved context for a task.
type Result struct {
	TaskID       string      `json:"task_id" yaml:"task_id"`
	Title        string      `json:"title" yaml:"title"`
	TaskBody     string      `json:"task_body,omitempty" yaml:"task_body,omitempty"`
	Files        []FileEntry `json:"files" yaml:"files"`
	Dependencies []DepEntry  `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
}

// Options configures context resolution behavior.
type Options struct {
	Scopes         ScopeMap
	ProjectRoot    string
	Resolve        bool // expand directory paths to individual files
	IncludeContent bool
	MaxFiles       int
}

// Resolve builds the context result for a task.
func Resolve(task *model.Task, opts Options) (*Result, error) {
	files := resolveScopeFiles(task.Touches, opts.Scopes)
	files = append(files, resolveExplicitFiles(task.Context)...)
	files = deduplicateFiles(files)
	checkExistence(files, opts.ProjectRoot)

	if opts.Resolve {
		files = expandDirectories(files, opts.ProjectRoot)
	}

	if opts.IncludeContent {
		inlineContents(files, opts.ProjectRoot)
	}

	if opts.MaxFiles > 0 {
		files = capFiles(files, opts.MaxFiles)
	}

	result := &Result{
		TaskID: task.ID,
		Title:  task.Title,
		Files:  files,
	}

	body := strings.TrimSpace(task.Body)
	if body != "" {
		result.TaskBody = body
	}

	return result, nil
}

// resolveScopeFiles maps touches to file paths via scope definitions.
func resolveScopeFiles(touches []string, scopes ScopeMap) []FileEntry {
	var files []FileEntry
	for _, scope := range touches {
		paths, ok := scopes[scope]
		if !ok {
			continue
		}
		for _, p := range paths {
			files = append(files, FileEntry{
				Path:   p,
				Source: "scope:" + scope,
			})
		}
	}
	return files
}

// resolveExplicitFiles creates entries from the task's context field.
func resolveExplicitFiles(contextPaths []string) []FileEntry {
	files := make([]FileEntry, len(contextPaths))
	for i, p := range contextPaths {
		files[i] = FileEntry{
			Path:   p,
			Source: "explicit",
		}
	}
	return files
}

// deduplicateFiles removes duplicate paths, keeping the first occurrence.
func deduplicateFiles(files []FileEntry) []FileEntry {
	seen := make(map[string]bool, len(files))
	var result []FileEntry
	for _, f := range files {
		if seen[f.Path] {
			continue
		}
		seen[f.Path] = true
		result = append(result, f)
	}
	return result
}

// checkExistence stats each file path relative to projectRoot and sets Exists and IsDir.
func checkExistence(files []FileEntry, projectRoot string) {
	for i := range files {
		fullPath := filepath.Join(projectRoot, files[i].Path)
		info, err := os.Stat(fullPath)
		files[i].Exists = err == nil
		if err == nil {
			files[i].IsDir = info.IsDir()
		}
	}
}

// expandDirectories replaces directory entries with their individual files recursively.
// It skips known junk directories and gitignored files.
func expandDirectories(files []FileEntry, projectRoot string) []FileEntry {
	seen := make(map[string]bool)
	for _, f := range files {
		seen[f.Path] = true
	}

	var result []FileEntry
	for _, f := range files {
		fullPath := filepath.Join(projectRoot, f.Path)
		info, err := os.Stat(fullPath)
		if err != nil || !info.IsDir() {
			result = append(result, f)
			continue
		}
		expanded := walkDirectory(fullPath, projectRoot, f.Source, seen)
		result = append(result, expanded...)
	}

	return filterGitIgnored(result, projectRoot)
}

// walkDirectory recursively collects files from a directory, skipping known junk directories.
func walkDirectory(dir, projectRoot, source string, seen map[string]bool) []FileEntry {
	var result []FileEntry
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, relErr := filepath.Rel(projectRoot, path)
		if relErr != nil || seen[rel] {
			return nil
		}
		seen[rel] = true
		result = append(result, FileEntry{
			Path:   rel,
			Source: source,
			Exists: true,
		})
		return nil
	})
	return result
}

// filterGitIgnored removes gitignored files from the list using git check-ignore.
// Falls back gracefully (returns input unchanged) if git is unavailable or the project is not a repo.
func filterGitIgnored(files []FileEntry, projectRoot string) []FileEntry {
	if len(files) == 0 {
		return files
	}

	var paths []string
	for _, f := range files {
		paths = append(paths, f.Path)
	}

	cmd := exec.Command("git", "check-ignore", "--stdin")
	cmd.Dir = projectRoot
	cmd.Stdin = strings.NewReader(strings.Join(paths, "\n"))

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		// exit code 1 = no paths are ignored; other errors = git unavailable
		return files
	}

	ignored := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if line != "" {
			ignored[line] = true
		}
	}

	var filtered []FileEntry
	for _, f := range files {
		if !ignored[f.Path] {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// inlineContents reads file contents and counts lines for each existing text file.
// Binary files and known generated files (lock files, etc.) are skipped.
func inlineContents(files []FileEntry, projectRoot string) {
	for i := range files {
		if !files[i].Exists {
			continue
		}
		fullPath := filepath.Join(projectRoot, files[i].Path)
		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() {
			continue
		}
		if generatedFiles[filepath.Base(files[i].Path)] {
			files[i].Generated = true
			continue
		}
		data, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}
		if isBinary(data) {
			files[i].Binary = true
			continue
		}
		content := string(data)
		files[i].Content = content
		files[i].Lines = countLines(content)
	}
}

// isBinary reports whether data looks like binary content.
// It checks for null bytes in the first 8KB, which is the same heuristic git uses.
func isBinary(data []byte) bool {
	check := data
	if len(check) > 8192 {
		check = check[:8192]
	}
	for _, b := range check {
		if b == 0 {
			return true
		}
	}
	return false
}

// countLines returns the number of lines in content, handling trailing newlines correctly.
func countLines(content string) int {
	if content == "" {
		return 0
	}
	n := strings.Count(content, "\n")
	// If file ends with newline, the last "line" is empty — don't count it
	if strings.HasSuffix(content, "\n") {
		return n
	}
	return n + 1
}

// capFiles truncates the file list to maxFiles entries.
func capFiles(files []FileEntry, maxFiles int) []FileEntry {
	if len(files) <= maxFiles {
		return files
	}
	return files[:maxFiles]
}
