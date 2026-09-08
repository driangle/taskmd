package worktree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
	return b.Overlay(scanDir, siblings, localTasks), nil
}

// Overlay merges localTasks with scans of the given sibling worktrees, or
// returns nil (overlay inactive) when there are none. Callers that already
// discovered siblings (e.g. to derive watch dirs) use this to avoid a second
// discovery.
func (b Builder) Overlay(scanDir string, siblings []gitmeta.Worktree, localTasks []*model.Task) *Overlay {
	if len(siblings) == 0 {
		return nil
	}
	scanned := b.scanSiblings(siblings)
	localTasks = AttributeNestedSiblingCopies(localTasks, scanDir, siblings)
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

// scanSiblings scans each sibling's tasks dir concurrently, skipping siblings
// that fail to scan. Sibling scans are independent and I/O-bound, so a repo
// with many worktrees scans in roughly the time of its slowest sibling rather
// than the sum of all of them. Results are written by index and compacted
// afterwards, so the merged order is worktree order regardless of completion
// order. Sibling duplicate-ID warnings are that worktree's own concern and are
// not reported here.
//
// Verbose scans run serially: the scanner logs progress straight to stderr, so
// concurrent scans would interleave those lines nondeterministically. Verbose
// is a debugging aid where readable output beats speed.
func (b Builder) scanSiblings(siblings []gitmeta.Worktree) []SiblingTasks {
	results := make([]*SiblingTasks, len(siblings))
	if b.Verbose {
		for i, wt := range siblings {
			results[i] = b.scanSibling(wt)
		}
	} else {
		var wg sync.WaitGroup
		for i, wt := range siblings {
			wg.Add(1)
			go func(i int, wt gitmeta.Worktree) {
				defer wg.Done()
				results[i] = b.scanSibling(wt)
			}(i, wt)
		}
		wg.Wait()
	}

	scanned := make([]SiblingTasks, 0, len(siblings))
	for _, r := range results {
		if r != nil {
			scanned = append(scanned, *r)
		}
	}
	return scanned
}

// scanSibling scans one sibling worktree, returning nil when it cannot be
// scanned so the overlay simply omits it.
func (b Builder) scanSibling(wt gitmeta.Worktree) *SiblingTasks {
	result, err := scanner.NewScanner(wt.TasksDir, b.Verbose, b.IgnoreDirs).Scan()
	if err != nil {
		if b.Verbose {
			fmt.Fprintf(os.Stderr, "Warning: skipping worktree %s: %v\n", wt.Root, err)
		}
		return nil
	}
	return &SiblingTasks{WT: wt, Tasks: result.Tasks}
}

// AttributeNestedSiblingCopies drops local-scan copies whose file path lies
// inside a sibling worktree's root: a non-hidden checkout nested in the scan
// root gets double-scanned, and those files belong to that worktree, not this
// one (spec §8). The sibling's own scan already carries them, so dropping the
// local-scan copies attributes them instead of flagging duplicates.
//
// Only siblings nested *inside* scanDir can double-scan, so only those are
// considered. A sibling whose root is an ancestor of scanDir — the primary
// checkout, when this worktree lives under it (`repo/.claude/worktrees/x`) —
// contains every local task path without having scanned any of them; treating
// that as nesting would drop the entire local task list.
func AttributeNestedSiblingCopies(local []*model.Task, scanDir string, siblings []gitmeta.Worktree) []*model.Task {
	nested := nestedSiblings(scanDir, siblings)
	if len(nested) == 0 {
		return local
	}
	attributed := make([]*model.Task, 0, len(local))
	for _, task := range local {
		if !insideAnyWorktreeRoot(task.FilePath, nested) {
			attributed = append(attributed, task)
		}
	}
	return attributed
}

// nestedSiblings returns the siblings whose root lies inside scanDir, i.e. the
// ones the local scan could have picked up.
func nestedSiblings(scanDir string, siblings []gitmeta.Worktree) []gitmeta.Worktree {
	root, err := filepath.Abs(scanDir)
	if err != nil {
		return nil
	}
	var nested []gitmeta.Worktree
	for _, wt := range siblings {
		wtRoot, err := filepath.Abs(wt.Root)
		if err != nil {
			continue
		}
		if isUnder(wtRoot, root) {
			nested = append(nested, wt)
		}
	}
	return nested
}

// isUnder reports whether path lies strictly inside dir.
func isUnder(path, dir string) bool {
	return strings.HasPrefix(filepath.Clean(path), filepath.Clean(dir)+string(filepath.Separator))
}

// insideAnyWorktreeRoot reports whether path lies under any sibling's root.
func insideAnyWorktreeRoot(path string, siblings []gitmeta.Worktree) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	for _, wt := range siblings {
		if isUnder(abs, wt.Root) {
			return true
		}
	}
	return false
}
