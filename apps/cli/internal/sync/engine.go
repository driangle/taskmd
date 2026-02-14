package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/driangle/taskmd/apps/cli/internal/nextid"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

// ConflictStrategy controls how conflicts are resolved during sync.
const (
	ConflictSkip   = "skip"   // default: skip conflicting tasks
	ConflictRemote = "remote" // overwrite local with remote
	ConflictLocal  = "local"  // keep local, update state hashes
)

// Engine orchestrates syncing tasks from external sources.
type Engine struct {
	ConfigDir        string
	Verbose          bool
	DryRun           bool
	ConflictStrategy string
}

// SyncAction describes a single sync operation.
type SyncAction struct {
	ExternalID string
	LocalID    string
	FilePath   string
	Title      string
	Reason     string
}

// SyncError describes an error during sync.
type SyncError struct {
	ExternalID string
	Title      string
	Err        error
}

// SyncResult holds the outcome of a sync run.
type SyncResult struct {
	Created   []SyncAction
	Updated   []SyncAction
	Skipped   []SyncAction
	Conflicts []SyncAction
	Errors    []SyncError
}

// RunSync syncs tasks for a single source configuration.
func (e *Engine) RunSync(srcCfg SourceConfig) (*SyncResult, error) {
	src, err := GetSource(srcCfg.Name)
	if err != nil {
		return nil, err
	}

	if err := src.ValidateConfig(srcCfg); err != nil {
		return nil, fmt.Errorf("invalid config for source %q: %w", srcCfg.Name, err)
	}

	externalTasks, err := src.FetchTasks(srcCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks from %q: %w", srcCfg.Name, err)
	}

	state, err := LoadState(e.ConfigDir, srcCfg.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	existingIDs, err := e.scanExistingIDs(srcCfg.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to scan existing tasks: %w", err)
	}

	result := &SyncResult{}
	now := time.Now()

	for _, ext := range externalTasks {
		action, syncErr := e.syncTask(ext, srcCfg, state, existingIDs, now)
		if syncErr != nil {
			result.Errors = append(result.Errors, SyncError{
				ExternalID: ext.ExternalID,
				Title:      ext.Title,
				Err:        syncErr,
			})
			continue
		}
		switch action.Reason {
		case "created":
			result.Created = append(result.Created, action)
			existingIDs = append(existingIDs, action.LocalID)
		case "updated":
			result.Updated = append(result.Updated, action)
		case "conflict":
			result.Conflicts = append(result.Conflicts, action)
		default:
			result.Skipped = append(result.Skipped, action)
		}
	}

	if !e.DryRun {
		state.LastSync = now
		if err := SaveState(e.ConfigDir, srcCfg.Name, state); err != nil {
			return result, fmt.Errorf("failed to save state: %w", err)
		}
	}

	return result, nil
}

func (e *Engine) syncTask(
	ext ExternalTask,
	srcCfg SourceConfig,
	state *SyncState,
	existingIDs []string,
	now time.Time,
) (SyncAction, error) {
	mapped := MapExternalTask(ext, srcCfg.FieldMap)
	extHash := HashExternalTask(ext)
	outputDir := srcCfg.OutputDir

	ts, exists := state.Tasks[ext.ExternalID]

	if !exists {
		return e.createTask(ext, mapped, extHash, outputDir, srcCfg.Name, state, existingIDs, now)
	}

	return e.updateTask(ext, mapped, extHash, ts, state, now)
}

func (e *Engine) createTask(
	ext ExternalTask,
	mapped MappedTask,
	extHash, outputDir, sourceName string,
	state *SyncState,
	existingIDs []string,
	now time.Time,
) (SyncAction, error) {
	nextResult := nextid.Calculate(existingIDs)
	newID := nextResult.NextID

	action := SyncAction{
		ExternalID: ext.ExternalID,
		LocalID:    newID,
		Title:      ext.Title,
		Reason:     "created",
	}

	if e.DryRun {
		return action, nil
	}

	filePath, err := WriteTaskFile(outputDir, newID, mapped, ext.ExternalID, sourceName)
	if err != nil {
		return SyncAction{}, err
	}

	localHash, err := HashLocalFile(filePath)
	if err != nil {
		return SyncAction{}, fmt.Errorf("failed to hash new file: %w", err)
	}

	action.FilePath = filePath

	state.Tasks[ext.ExternalID] = TaskState{
		ExternalID:   ext.ExternalID,
		LocalID:      newID,
		FilePath:     filePath,
		ExternalHash: extHash,
		LocalHash:    localHash,
		LastSynced:   now,
	}

	return action, nil
}

func (e *Engine) updateTask(
	ext ExternalTask,
	mapped MappedTask,
	extHash string,
	ts TaskState,
	state *SyncState,
	now time.Time,
) (SyncAction, error) {
	action := SyncAction{
		ExternalID: ext.ExternalID,
		LocalID:    ts.LocalID,
		FilePath:   ts.FilePath,
		Title:      ext.Title,
	}

	extChanged := extHash != ts.ExternalHash

	localChanged, err := e.localFileChanged(ts)
	if err != nil {
		if os.IsNotExist(err) {
			return e.recreateMissingFile(action, ext, mapped, extHash, ts, state, now)
		}
		return SyncAction{}, fmt.Errorf("failed to check local file: %w", err)
	}

	if !extChanged && !localChanged {
		action.Reason = "skipped"
		return action, nil
	}

	if localChanged {
		return e.resolveConflict(action, ext, mapped, extHash, ts, state, now)
	}

	return e.applyExternalUpdate(action, ext, mapped, extHash, ts, state, now)
}

func (e *Engine) resolveConflict(
	action SyncAction,
	ext ExternalTask,
	mapped MappedTask,
	extHash string,
	ts TaskState,
	state *SyncState,
	now time.Time,
) (SyncAction, error) {
	switch e.ConflictStrategy {
	case ConflictRemote:
		return e.applyExternalUpdate(action, ext, mapped, extHash, ts, state, now)
	case ConflictLocal:
		return e.acceptLocal(action, extHash, ts, state, now)
	default:
		action.Reason = "conflict"
		return action, nil
	}
}

func (e *Engine) acceptLocal(
	action SyncAction,
	extHash string,
	ts TaskState,
	state *SyncState,
	now time.Time,
) (SyncAction, error) {
	action.Reason = "updated"
	if e.DryRun {
		return action, nil
	}

	localHash, err := HashLocalFile(ts.FilePath)
	if err != nil {
		return SyncAction{}, fmt.Errorf("failed to hash local file: %w", err)
	}

	ts.ExternalHash = extHash
	ts.LocalHash = localHash
	ts.LastSynced = now
	state.Tasks[action.ExternalID] = ts

	return action, nil
}

func (e *Engine) recreateMissingFile(
	action SyncAction,
	ext ExternalTask,
	mapped MappedTask,
	extHash string,
	ts TaskState,
	state *SyncState,
	now time.Time,
) (SyncAction, error) {
	action.Reason = "created"
	if e.DryRun {
		return action, nil
	}

	filePath, err := WriteTaskFile(filepath.Dir(ts.FilePath), ts.LocalID, mapped, ext.ExternalID, state.Source)
	if err != nil {
		return SyncAction{}, err
	}

	localHash, err := HashLocalFile(filePath)
	if err != nil {
		return SyncAction{}, err
	}

	ts.FilePath = filePath
	ts.ExternalHash = extHash
	ts.LocalHash = localHash
	ts.LastSynced = now
	state.Tasks[ext.ExternalID] = ts

	return action, nil
}

func (e *Engine) applyExternalUpdate(
	action SyncAction,
	ext ExternalTask,
	mapped MappedTask,
	extHash string,
	ts TaskState,
	state *SyncState,
	now time.Time,
) (SyncAction, error) {
	action.Reason = "updated"
	if e.DryRun {
		return action, nil
	}

	if err := UpdateSyncedTaskFile(ts.FilePath, mapped); err != nil {
		return SyncAction{}, fmt.Errorf("failed to update task file: %w", err)
	}

	localHash, err := HashLocalFile(ts.FilePath)
	if err != nil {
		return SyncAction{}, fmt.Errorf("failed to hash updated file: %w", err)
	}

	ts.ExternalHash = extHash
	ts.LocalHash = localHash
	ts.LastSynced = now
	state.Tasks[ext.ExternalID] = ts

	return action, nil
}

func (e *Engine) localFileChanged(ts TaskState) (bool, error) {
	currentHash, err := HashLocalFile(ts.FilePath)
	if err != nil {
		return false, err
	}
	return currentHash != ts.LocalHash, nil
}

func (e *Engine) scanExistingIDs(outputDir string) ([]string, error) {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return nil, nil
	}

	taskScanner := scanner.NewScanner(outputDir, e.Verbose)
	result, err := taskScanner.Scan()
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(result.Tasks))
	for i, t := range result.Tasks {
		ids[i] = t.ID
	}
	return ids, nil
}
