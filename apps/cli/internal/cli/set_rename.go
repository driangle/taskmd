package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/slug"
	"github.com/driangle/taskmd/sdk/go/taskfile"
)

// renamePlan describes the file-level side effects of a `set --title`, so the
// same values drive both the --dry-run preview and the real write.
type renamePlan struct {
	newPath string           // destination path, empty when the file stays put
	heading taskfile.Heading // the body heading as it was before the update
}

// planRename inspects the task file and works out what a title change implies
// for the filename and the body heading. It performs no writes.
func planRename(task *model.Task, req taskfile.UpdateRequest) (renamePlan, error) {
	if req.Title == nil {
		return renamePlan{heading: taskfile.Heading{Line: -1}}, nil
	}

	heading, err := taskfile.FindTitleHeading(task.FilePath)
	if err != nil {
		return renamePlan{}, err
	}

	newPath, err := resolveRenamePath(task, *req.Title)
	if err != nil {
		return renamePlan{}, err
	}

	return renamePlan{newPath: newPath, heading: heading}, nil
}

// resolveRenamePath computes the destination filename for --rename, mirroring
// the `<id>-<slug>.md` convention that `taskmd add` uses. It returns an empty
// path when the file does not need to move.
func resolveRenamePath(task *model.Task, newTitle string) (string, error) {
	if !setRename {
		return "", nil
	}

	suffix := slug.Slugify(newTitle)
	if suffix == "" {
		return "", fmt.Errorf("cannot rename: title %q produces an empty filename slug", newTitle)
	}

	newPath := filepath.Join(filepath.Dir(task.FilePath), fmt.Sprintf("%s-%s.md", task.ID, suffix))
	if newPath == task.FilePath {
		return "", nil
	}
	if _, err := os.Stat(newPath); err == nil {
		return "", fmt.Errorf("cannot rename: %s already exists", newPath)
	}

	return newPath, nil
}

// applyRename rewrites the body heading and moves the file, in that order, so a
// failed heading rewrite leaves the file where the scanner expects it.
func applyRename(task *model.Task, req taskfile.UpdateRequest, plan renamePlan) error {
	if req.Title == nil {
		return nil
	}

	if _, err := taskfile.RewriteTitleHeading(task.FilePath, task.Title, *req.Title); err != nil {
		return err
	}

	if plan.newPath == "" {
		return nil
	}
	if err := os.Rename(task.FilePath, plan.newPath); err != nil {
		return fmt.Errorf("failed to rename task file: %w", err)
	}

	return nil
}

// headingSkipped reports whether the body has a heading that will be left alone
// because it no longer matches the task's title.
func (p renamePlan) headingSkipped(oldTitle string) bool {
	return p.heading.Found && p.heading.Text != oldTitle
}
