package taskcontext

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

// ScopeMap maps scope names to their file paths.
type ScopeMap map[string][]string

// FileEntry represents a single file in the resolved context.
type FileEntry struct {
	Path    string `json:"path" yaml:"path"`
	Source  string `json:"source" yaml:"source"`
	Exists  bool   `json:"exists" yaml:"exists"`
	Content string `json:"content,omitempty" yaml:"content,omitempty"`
	Lines   int    `json:"lines,omitempty" yaml:"lines,omitempty"`
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

// checkExistence stats each file path relative to projectRoot and sets Exists.
func checkExistence(files []FileEntry, projectRoot string) {
	for i := range files {
		fullPath := filepath.Join(projectRoot, files[i].Path)
		_, err := os.Stat(fullPath)
		files[i].Exists = err == nil
	}
}

// expandDirectories replaces directory entries with their individual files.
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
		// Walk the directory and add individual files
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			result = append(result, f)
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			childPath := filepath.Join(f.Path, entry.Name())
			if seen[childPath] {
				continue
			}
			seen[childPath] = true
			result = append(result, FileEntry{
				Path:   childPath,
				Source: f.Source,
				Exists: true,
			})
		}
	}
	return result
}

// inlineContents reads file contents and counts lines for each existing file.
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
		data, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}
		content := string(data)
		files[i].Content = content
		files[i].Lines = countLines(content)
	}
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
