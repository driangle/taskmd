package search

import (
	"strings"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

// Result represents a single search match.
type Result struct {
	ID            string `json:"id" yaml:"id"`
	Title         string `json:"title" yaml:"title"`
	Status        string `json:"status" yaml:"status"`
	FilePath      string `json:"file_path" yaml:"file_path"`
	MatchLocation string `json:"match_location" yaml:"match_location"`
	Snippet       string `json:"snippet" yaml:"snippet"`
}

// Search performs case-insensitive full-text search across task titles and bodies.
func Search(tasks []*model.Task, query string) []Result {
	lowerQuery := strings.ToLower(query)
	var results []Result

	for _, task := range tasks {
		titleMatch := strings.Contains(strings.ToLower(task.Title), lowerQuery)
		bodyMatch := strings.Contains(strings.ToLower(task.Body), lowerQuery)

		if !titleMatch && !bodyMatch {
			continue
		}

		location := matchLocation(titleMatch, bodyMatch)
		snippet := extractSnippet(task, lowerQuery, bodyMatch)

		results = append(results, Result{
			ID:            task.ID,
			Title:         task.Title,
			Status:        string(task.Status),
			FilePath:      task.FilePath,
			MatchLocation: location,
			Snippet:       snippet,
		})
	}

	return results
}

func matchLocation(titleMatch, bodyMatch bool) string {
	if titleMatch && bodyMatch {
		return "title,body"
	}
	if titleMatch {
		return "title"
	}
	return "body"
}

func extractSnippet(task *model.Task, lowerQuery string, bodyMatch bool) string {
	if bodyMatch {
		return ExtractBodySnippet(task.Body, lowerQuery)
	}
	return task.Title
}

// ExtractBodySnippet returns a short snippet of body text around the first match.
func ExtractBodySnippet(body, lowerQuery string) string {
	lowerBody := strings.ToLower(body)
	idx := strings.Index(lowerBody, lowerQuery)
	if idx < 0 {
		return ""
	}

	const radius = 40

	start := max(idx-radius, 0)
	end := min(idx+len(lowerQuery)+radius, len(body))

	// Trim to word boundaries
	if start > 0 {
		if spaceIdx := strings.IndexByte(body[start:], ' '); spaceIdx >= 0 {
			start += spaceIdx + 1
		}
	}
	if end < len(body) {
		if spaceIdx := strings.LastIndexByte(body[:end], ' '); spaceIdx >= 0 {
			end = spaceIdx
		}
	}

	snippet := body[start:end]
	// Collapse whitespace
	snippet = strings.Join(strings.Fields(snippet), " ")

	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(body) {
		snippet = snippet + "..."
	}

	return snippet
}
