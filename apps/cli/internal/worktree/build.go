package worktree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/driangle/taskmd/apps/cli/internal/gitmeta"
	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/scanner"
)

// Discoverer lists the sibling worktrees of the repo containing scanDir
// (nil when scanDir is not in a repo). It is a seam so overlay consumers can
// inject worktrees in tests without git.
type Discoverer func(scanDir string) ([]gitmeta.Worktree, error)

// DiscoverSiblings is the default Discoverer, backed by git via gitmeta.
func DiscoverSiblings(scanDir string) ([]gitmeta.Worktree, error) {
	id, err := gitmeta.Resolve(scanDir)
	if err != nil || id == nil {
		return nil, err
	}
	worktrees, err := gitmeta.ListWorktrees(id)
	if err != nil {
		return nil, err
	}
	var siblings []gitmeta.Worktree
	for _, wt := range worktrees {
		if !wt.IsLocal {
			siblings = append(siblings, wt)
		}
	}
	return siblings, nil
}

// Builder builds the cross-worktree overlay for a scan directory. The zero
// value is a disabled builder, so surfaces that never configure worktrees keep
// exactly today's behavior.
type Builder struct {
	// Enabled activates the overlay (worktree_scope "unified", the default;
	// "isolated" disables). Even when true, the overlay only forms when the
	// scan dir is inside a git repo with sibling worktrees.
	Enabled bool
	// Discover lists sibling worktrees; nil means DiscoverSiblings.
	Discover Discoverer
	// Verbose enables warnings about skipped sibling scans on stderr.
	Verbose bool
	// IgnoreDirs is passed through to sibling scans, matching the local scan.
	IgnoreDirs []string
}

// discoverer returns the configured Discoverer, defaulting to git discovery.
func (b Builder) discoverer() Discoverer {
	if b.Discover != nil {
		return b.Discover
	}
	return DiscoverSiblings
}

// Build builds the overlay for the local task list, or returns nil when the
// overlay is inactive: disabled, scanDir not in a git repo, or no sibling
// worktrees to merge (in which case behavior is identical to today's).
// Sibling task file paths are left absolute; callers that display them
// relativize afterwards.
func (b Builder) Build(scanDir string, localTasks []*model.Task) (*Overlay, error) {
	siblings, err := b.Siblings(scanDir)
	if err != nil {
		return nil, err
	}
	return b.Overlay(siblings, localTasks), nil
}

// Overlay merges localTasks with scans of the given sibling worktrees, or
// returns nil (overlay inactive) when there are none. Callers that already
// discovered siblings (e.g. to derive watch dirs) use this to avoid a second
// discovery.
func (b Builder) Overlay(siblings []gitmeta.Worktree, localTasks []*model.Task) *Overlay {
	if len(siblings) == 0 {
		return nil
	}
	scanned := b.scanSiblings(siblings)
	localTasks = AttributeNestedSiblingCopies(localTasks, siblings)
	return Merge(localTasks, scanned)
}

// Siblings discovers sibling worktrees when the builder is enabled. Discovery
// failures deactivate the overlay rather than failing the command.
func (b Builder) Siblings(scanDir string) ([]gitmeta.Worktree, error) {
	if !b.Enabled {
		return nil, nil
	}
	siblings, err := b.discoverer()(scanDir)
	if err != nil {
		gitmeta.Debugf("worktree discovery failed, overlay inactive: %v", err)
		return nil, nil
	}
	return siblings, nil
}

// scanSiblings scans each sibling's tasks dir, skipping siblings that fail to
// scan. Sibling duplicate-ID warnings are that worktree's own concern and are
// not reported here.
func (b Builder) scanSiblings(siblings []gitmeta.Worktree) []SiblingTasks {
	var scanned []SiblingTasks
	for _, wt := range siblings {
		result, err := scanner.NewScanner(wt.TasksDir, b.Verbose, b.IgnoreDirs).Scan()
		if err != nil {
			if b.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: skipping worktree %s: %v\n", wt.Root, err)
			}
			continue
		}
		scanned = append(scanned, SiblingTasks{WT: wt, Tasks: result.Tasks})
	}
	return scanned
}

// AttributeNestedSiblingCopies drops local-scan copies whose file path lies
// inside a sibling worktree's root: a non-hidden checkout nested in the scan
// root gets double-scanned, and those files belong to that worktree, not this
// one (spec §8). The sibling's own scan already carries them, so dropping the
// local-scan copies attributes them instead of flagging duplicates.
func AttributeNestedSiblingCopies(local []*model.Task, siblings []gitmeta.Worktree) []*model.Task {
	attributed := make([]*model.Task, 0, len(local))
	for _, task := range local {
		if !insideAnyWorktreeRoot(task.FilePath, siblings) {
			attributed = append(attributed, task)
		}
	}
	return attributed
}

// insideAnyWorktreeRoot reports whether path lies under any sibling's root.
func insideAnyWorktreeRoot(path string, siblings []gitmeta.Worktree) bool {
	cleaned := filepath.Clean(path)
	for _, wt := range siblings {
		if strings.HasPrefix(cleaned, filepath.Clean(wt.Root)+string(filepath.Separator)) {
			return true
		}
	}
	return false
}
