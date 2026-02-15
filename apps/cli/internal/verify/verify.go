package verify

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

// StepStatus represents the outcome of a single verification step.
type StepStatus string

const (
	StatusPass    StepStatus = "pass"
	StatusFail    StepStatus = "fail"
	StatusPending StepStatus = "pending"
	StatusSkip    StepStatus = "skip"
)

// StepResult holds the outcome of a single verify step.
type StepResult struct {
	Index    int        `json:"index"`
	Type     string     `json:"type"`
	Status   StepStatus `json:"status"`
	Command  string     `json:"command,omitempty"`
	Dir      string     `json:"dir,omitempty"`
	Check    string     `json:"check,omitempty"`
	Stdout   string     `json:"stdout,omitempty"`
	Stderr   string     `json:"stderr,omitempty"`
	ExitCode int        `json:"exit_code,omitempty"`
	Duration string     `json:"duration,omitempty"`
	Warning  string     `json:"warning,omitempty"`
}

// Result holds the aggregate outcome of all verify steps.
type Result struct {
	Steps   []StepResult `json:"steps"`
	Passed  int          `json:"passed"`
	Failed  int          `json:"failed"`
	Pending int          `json:"pending"`
	Skipped int          `json:"skipped"`
}

// HasFailures returns true if any bash step failed.
func (r *Result) HasFailures() bool {
	return r.Failed > 0
}

// Options configures verify execution.
type Options struct {
	ProjectRoot string
	DryRun      bool
	Timeout     time.Duration
	Verbose     bool
	LogFunc     func(format string, args ...any) // called before each step
}

// Run executes the given verify steps and returns the aggregate result.
func Run(steps []model.VerifyStep, opts Options) *Result {
	result := &Result{}
	for i, step := range steps {
		sr := runStep(i, step, opts)
		result.Steps = append(result.Steps, sr)
		switch sr.Status {
		case StatusPass:
			result.Passed++
		case StatusFail:
			result.Failed++
		case StatusPending:
			result.Pending++
		case StatusSkip:
			result.Skipped++
		}
	}
	return result
}

func runStep(index int, step model.VerifyStep, opts Options) StepResult {
	switch step.Type {
	case "bash":
		return runBashStep(index, step, opts)
	case "assert":
		return runAssertStep(index, step)
	default:
		return StepResult{
			Index:   index,
			Type:    step.Type,
			Status:  StatusSkip,
			Warning: fmt.Sprintf("unknown verify type %q — skipped", step.Type),
		}
	}
}

func runBashStep(index int, step model.VerifyStep, opts Options) StepResult {
	sr := StepResult{
		Index:   index,
		Type:    "bash",
		Command: step.Run,
		Dir:     step.Dir,
	}

	if opts.DryRun {
		sr.Status = StatusSkip
		sr.Warning = "dry-run"
		return sr
	}

	if opts.LogFunc != nil {
		dir := step.Dir
		if dir == "" {
			dir = "."
		}
		opts.LogFunc("Running: %s (dir: %s)", step.Run, dir)
	}

	execBashCommand(&sr, step, opts)
	return sr
}

func execBashCommand(sr *StepResult, step model.VerifyStep, opts Options) {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", step.Run)

	dir := opts.ProjectRoot
	if step.Dir != "" {
		dir = opts.ProjectRoot + "/" + step.Dir
	}
	if dir != "" {
		cmd.Dir = dir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	sr.Stdout = stdout.String()
	sr.Stderr = stderr.String()
	sr.Duration = elapsed.Round(time.Millisecond).String()

	if err != nil {
		sr.Status = StatusFail
		if exitErr, ok := err.(*exec.ExitError); ok {
			sr.ExitCode = exitErr.ExitCode()
		} else if ctx.Err() == context.DeadlineExceeded {
			sr.ExitCode = -1
			sr.Stderr = sr.Stderr + "\ncommand timed out"
		} else {
			sr.ExitCode = -1
		}
		return
	}

	sr.Status = StatusPass
	sr.ExitCode = 0
}

func runAssertStep(index int, step model.VerifyStep) StepResult {
	return StepResult{
		Index:  index,
		Type:   "assert",
		Status: StatusPending,
		Check:  step.Check,
	}
}
