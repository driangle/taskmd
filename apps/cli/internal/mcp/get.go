package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/graph"
	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

// GetInput defines the input schema for the get tool.
type GetInput struct {
	TaskDir string `json:"task_dir,omitempty" jsonschema:"task directory to scan, defaults to current directory"`
	TaskID  string `json:"task_id" jsonschema:"required,task ID to retrieve"`
}

// getOutput is the JSON representation of a task with body included.
type getOutput struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Status       string   `json:"status"`
	Priority     string   `json:"priority,omitempty"`
	Effort       string   `json:"effort,omitempty"`
	Tags         []string `json:"tags"`
	Dependencies []string `json:"dependencies"`
	Touches      []string `json:"touches,omitempty"`
	Parent       string   `json:"parent,omitempty"`
	Created      string   `json:"created,omitempty"`
	FilePath     string   `json:"file_path"`
	Content      string   `json:"content"`
	DependsOn    []depRef `json:"depends_on"`
	Blocks       []depRef `json:"blocks"`
	Children     []depRef `json:"children,omitempty"`
}

// depRef is a lightweight dependency reference.
type depRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func registerGetTool(server *gomcp.Server) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "get",
		Description: "Get full details of a single task by ID, including body content and dependency information",
	}, handleGet)
}

func handleGet(_ context.Context, _ *gomcp.CallToolRequest, input GetInput) (*gomcp.CallToolResult, any, error) {
	if input.TaskID == "" {
		return nil, nil, fmt.Errorf("task_id is required")
	}

	taskDir := input.TaskDir
	if taskDir == "" {
		taskDir = "."
	}

	taskScanner := scanner.NewScanner(taskDir, false, nil)
	result, err := taskScanner.Scan()
	if err != nil {
		return nil, nil, fmt.Errorf("scan failed: %w", err)
	}

	task := findTaskByID(input.TaskID, result.Tasks)
	if task == nil {
		return nil, nil, fmt.Errorf("task not found: %s", input.TaskID)
	}

	out := buildGetOutput(task, result.Tasks)

	data, err := json.Marshal(out)
	if err != nil {
		return nil, nil, fmt.Errorf("json marshal failed: %w", err)
	}

	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func findTaskByID(id string, tasks []*model.Task) *model.Task {
	for _, t := range tasks {
		if t.ID == id {
			return t
		}
	}
	return nil
}

func buildGetOutput(task *model.Task, allTasks []*model.Task) getOutput {
	taskMap := make(map[string]*model.Task, len(allTasks))
	for _, t := range allTasks {
		taskMap[t.ID] = t
	}

	g := graph.NewGraph(allTasks)

	var dependsOn []depRef
	for _, depID := range task.Dependencies {
		ref := depRef{ID: depID}
		if dep, ok := taskMap[depID]; ok {
			ref.Title = dep.Title
		}
		dependsOn = append(dependsOn, ref)
	}

	var blocks []depRef
	for _, blockedID := range g.Adjacency[task.ID] {
		ref := depRef{ID: blockedID}
		if dep, ok := taskMap[blockedID]; ok {
			ref.Title = dep.Title
		}
		blocks = append(blocks, ref)
	}

	var children []depRef
	for _, t := range allTasks {
		if t.Parent == task.ID {
			children = append(children, depRef{ID: t.ID, Title: t.Title})
		}
	}

	created := ""
	if !task.Created.IsZero() {
		created = task.Created.Format("2006-01-02")
	}

	return getOutput{
		ID:           task.ID,
		Title:        task.Title,
		Status:       string(task.Status),
		Priority:     string(task.Priority),
		Effort:       string(task.Effort),
		Tags:         task.Tags,
		Dependencies: task.Dependencies,
		Touches:      task.Touches,
		Parent:       task.Parent,
		Created:      created,
		FilePath:     task.FilePath,
		Content:      task.Body,
		DependsOn:    dependsOn,
		Blocks:       blocks,
		Children:     children,
	}
}
