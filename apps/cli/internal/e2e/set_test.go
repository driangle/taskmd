//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSet_DependsOn_Basic(t *testing.T) {
	dir := setupTaskDir(t)
	writeTask(t, dir, "001-setup.md", "001", "Setup", "pending", nil)
	writeTask(t, dir, "002-auth.md", "002", "Auth", "pending", []string{"001"})
	writeTask(t, dir, "003-ui.md", "003", "UI", "pending", []string{"002"})

	result := mustRun(t, dir, "set", "003", "--depends-on", "001,002")

	if !strings.Contains(result.Stdout, "dependencies:") {
		t.Errorf("Expected dependencies change in output, got: %s", result.Stdout)
	}

	content, err := os.ReadFile(filepath.Join(dir, "003-ui.md"))
	if err != nil {
		t.Fatalf("failed to read task file: %v", err)
	}
	fileStr := string(content)
	if !strings.Contains(fileStr, `"001"`) || !strings.Contains(fileStr, `"002"`) {
		t.Errorf("Expected file to contain both dep IDs, got:\n%s", fileStr)
	}
}

func TestSet_DependsOn_InvalidID(t *testing.T) {
	dir := setupTaskDir(t)
	writeTask(t, dir, "001-setup.md", "001", "Setup", "pending", nil)

	result := run(t, dir, "set", "001", "--depends-on", "999")

	if result.ExitCode == 0 {
		t.Fatal("Expected non-zero exit code for non-existent dependency")
	}
	combined := result.Stdout + result.Stderr
	if !strings.Contains(combined, "not found") {
		t.Errorf("Expected 'not found' in output, got: %s", combined)
	}
}

func TestSet_DependsOn_CircularDep(t *testing.T) {
	dir := setupTaskDir(t)
	writeTask(t, dir, "001-setup.md", "001", "Setup", "pending", nil)
	writeTask(t, dir, "002-auth.md", "002", "Auth", "pending", []string{"001"})
	writeTask(t, dir, "003-ui.md", "003", "UI", "pending", []string{"002"})

	// Setting 001 to depend on 003 creates a cycle: 001->003->002->001
	result := run(t, dir, "set", "001", "--depends-on", "003")

	if result.ExitCode == 0 {
		t.Fatal("Expected non-zero exit code for circular dependency")
	}
	combined := result.Stdout + result.Stderr
	if !strings.Contains(combined, "circular dependency") {
		t.Errorf("Expected 'circular dependency' in output, got: %s", combined)
	}
}

func TestSet_DependsOn_WithStatus(t *testing.T) {
	dir := setupTaskDir(t)
	writeTask(t, dir, "001-setup.md", "001", "Setup", "pending", nil)
	writeTask(t, dir, "002-auth.md", "002", "Auth", "pending", nil)

	result := mustRun(t, dir, "set", "002", "--depends-on", "001", "--status", "blocked")

	if !strings.Contains(result.Stdout, "status:") {
		t.Error("Expected status change in output")
	}
	if !strings.Contains(result.Stdout, "dependencies:") {
		t.Error("Expected dependencies change in output")
	}

	content, err := os.ReadFile(filepath.Join(dir, "002-auth.md"))
	if err != nil {
		t.Fatalf("failed to read task file: %v", err)
	}
	fileStr := string(content)
	if !strings.Contains(fileStr, "status: blocked") {
		t.Error("Expected file to contain updated status")
	}
	if !strings.Contains(fileStr, `"001"`) {
		t.Error("Expected file to contain dependency")
	}
}

func TestSet_DependsOn_Clear(t *testing.T) {
	dir := setupTaskDir(t)
	writeTask(t, dir, "001-setup.md", "001", "Setup", "pending", nil)
	writeTask(t, dir, "002-auth.md", "002", "Auth", "pending", []string{"001"})

	result := mustRun(t, dir, "set", "002", "--depends-on", "")

	if !strings.Contains(result.Stdout, "dependencies:") {
		t.Errorf("Expected dependencies change in output, got: %s", result.Stdout)
	}

	content, err := os.ReadFile(filepath.Join(dir, "002-auth.md"))
	if err != nil {
		t.Fatalf("failed to read task file: %v", err)
	}
	fileStr := string(content)
	if strings.Contains(fileStr, "dependencies:") {
		t.Errorf("Expected dependencies line to be removed, got:\n%s", fileStr)
	}
}

func TestSet_CompletedDate_AutoSet(t *testing.T) {
	dir := setupTaskDir(t)
	writeTask(t, dir, "001-setup.md", "001", "Setup", "pending", nil)

	result := mustRun(t, dir, "set", "001", "--status", "completed")

	if !strings.Contains(result.Stdout, "completed_at:") {
		t.Errorf("Expected completed_at date in output, got: %s", result.Stdout)
	}

	content, err := os.ReadFile(filepath.Join(dir, "001-setup.md"))
	if err != nil {
		t.Fatalf("failed to read task file: %v", err)
	}
	fileStr := string(content)
	if !strings.Contains(fileStr, "status: completed") {
		t.Error("Expected status to be completed")
	}
	if !strings.Contains(fileStr, "completed_at: ") {
		t.Errorf("Expected completed_at date in file, got:\n%s", fileStr)
	}
}

func TestSet_CompletedDate_ClearedOnReopen(t *testing.T) {
	dir := setupTaskDir(t)

	// Write a task that is already completed with a completed_at date
	taskContent := fmt.Sprintf(`---
id: "001"
title: "Setup"
status: completed
priority: medium
effort: small
dependencies: []
tags: ["e2e"]
completed_at: 2026-03-01
created: 2026-01-01
---

# Setup
`)
	path := filepath.Join(dir, "001-setup.md")
	if err := os.WriteFile(path, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	result := mustRun(t, dir, "set", "001", "--status", "pending")

	if !strings.Contains(result.Stdout, "completed_at:") {
		t.Errorf("Expected completed_at change in output, got: %s", result.Stdout)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read task file: %v", err)
	}
	fileStr := string(content)
	if !strings.Contains(fileStr, "status: pending") {
		t.Error("Expected status to be pending")
	}
	if strings.Contains(fileStr, "completed_at:") {
		t.Errorf("Expected completed_at field to be removed, got:\n%s", fileStr)
	}
}

func TestSet_CancelledDate_AutoSet(t *testing.T) {
	dir := setupTaskDir(t)
	writeTask(t, dir, "001-setup.md", "001", "Setup", "pending", nil)

	result := mustRun(t, dir, "set", "001", "--status", "cancelled")

	content, err := os.ReadFile(filepath.Join(dir, "001-setup.md"))
	if err != nil {
		t.Fatalf("failed to read task file: %v", err)
	}
	fileStr := string(content)
	if !strings.Contains(fileStr, "cancelled_at: ") {
		t.Errorf("Expected cancelled_at date for cancelled status, got:\n%s", fileStr)
	}
	if strings.Contains(fileStr, "completed_at:") {
		t.Error("Expected no completed_at field when cancelling")
	}
	_ = result
}

func TestSet_Title_Rename(t *testing.T) {
	dir := setupTaskDir(t)
	writeTask(t, dir, "001-setup.md", "001", "Setup", "pending", nil)

	result := mustRun(t, dir, "set", "001", "--title", "Bootstrap the project", "--rename")

	if !strings.Contains(result.Stdout, "file: 001-setup.md -> 001-bootstrap-the-project.md") {
		t.Errorf("Expected rename in output, got: %s", result.Stdout)
	}

	if _, err := os.Stat(filepath.Join(dir, "001-setup.md")); err == nil {
		t.Error("Expected the original file to be gone")
	}

	content, err := os.ReadFile(filepath.Join(dir, "001-bootstrap-the-project.md"))
	if err != nil {
		t.Fatalf("failed to read renamed task file: %v", err)
	}
	fileStr := string(content)
	if !strings.Contains(fileStr, `title: "Bootstrap the project"`) {
		t.Errorf("Expected updated frontmatter title, got:\n%s", fileStr)
	}
	if !strings.Contains(fileStr, "# Bootstrap the project") {
		t.Errorf("Expected updated body heading, got:\n%s", fileStr)
	}

	// The task must still be reachable by ID from its new path.
	getResult := mustRun(t, dir, "get", "001")
	if !strings.Contains(getResult.Stdout, "Bootstrap the project") {
		t.Errorf("Expected the renamed task to be resolvable by ID, got: %s", getResult.Stdout)
	}
}

func TestSet_Title_RenameRequiresTitle(t *testing.T) {
	dir := setupTaskDir(t)
	writeTask(t, dir, "001-setup.md", "001", "Setup", "pending", nil)

	result := run(t, dir, "set", "001", "--rename")

	if result.ExitCode == 0 {
		t.Fatal("Expected non-zero exit code for --rename without --title")
	}
	combined := result.Stdout + result.Stderr
	if !strings.Contains(combined, "--rename requires --title") {
		t.Errorf("Expected an actionable error, got: %s", combined)
	}
}
