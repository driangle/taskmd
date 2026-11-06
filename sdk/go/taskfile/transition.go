package taskfile

import (
	"time"

	"github.com/driangle/taskmd/sdk/go/model"
)

// dateLayout is the frontmatter date format for completed_at and cancelled_at.
const dateLayout = "2006-01-02"

// ApplyStatusTransitionDates sets or clears completed_at/cancelled_at on req
// to match the status change it makes to task. Every writer that changes a
// task's status should call it before UpdateTaskFile, so all surfaces agree:
//
//   - to completed: completed_at is set to now; cancelled_at is removed
//   - to cancelled: cancelled_at is set to now; completed_at is removed
//   - from completed/cancelled to any other status: both dates are removed
//
// It does nothing when req does not change the status.
func ApplyStatusTransitionDates(task *model.Task, req *UpdateRequest, now time.Time) {
	if req.Status == nil {
		return
	}
	today := now.Format(dateLayout)
	newStatus := model.Status(*req.Status)

	switch {
	case newStatus == model.StatusCompleted:
		req.Completed = &today
		req.RemoveFields = append(req.RemoveFields, "cancelled_at")
	case newStatus == model.StatusCancelled:
		req.CancelledAt = &today
		req.RemoveFields = append(req.RemoveFields, "completed_at")
	case task.Status.IsResolved():
		// Reopening: the task was completed/cancelled and is moving to a non-terminal status.
		req.RemoveFields = append(req.RemoveFields, "completed_at", "cancelled_at")
	}
}
