package tracks

import (
	"sort"

	"github.com/driangle/taskmd/apps/cli/internal/filter"
	"github.com/driangle/taskmd/apps/cli/internal/graph"
	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/next"
)

// TrackTask holds task metadata for output.
type TrackTask struct {
	ID       string   `json:"id" yaml:"id"`
	Title    string   `json:"title" yaml:"title"`
	Priority string   `json:"priority,omitempty" yaml:"priority,omitempty"`
	Effort   string   `json:"effort,omitempty" yaml:"effort,omitempty"`
	Score    int      `json:"score" yaml:"score"`
	FilePath string   `json:"file_path" yaml:"file_path"`
	Touches  []string `json:"touches,omitempty" yaml:"touches,omitempty"`
}

// Track represents a group of non-overlapping tasks that can run in sequence.
type Track struct {
	ID     int         `json:"id" yaml:"id"`
	Tasks  []TrackTask `json:"tasks" yaml:"tasks"`
	Scopes []string    `json:"scopes" yaml:"scopes"`

	// scopeSet is internal for fast overlap checks during assignment.
	scopeSet map[string]bool
}

// Result holds the output of the tracks algorithm.
type Result struct {
	Tracks   []Track     `json:"tracks" yaml:"tracks"`
	Flexible []TrackTask `json:"flexible" yaml:"flexible"`
	Warnings []string    `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}

// Options controls track assignment behaviour.
type Options struct {
	Filters     []string
	KnownScopes map[string]bool
}

type scored struct {
	task  *model.Task
	score int
}

// Assign groups actionable tasks into parallel tracks based on scope overlap.
func Assign(tasks []*model.Task, opts Options) (*Result, error) {
	items, err := scoreActionable(tasks, opts.Filters)
	if err != nil {
		return nil, err
	}

	warnings := validateScopes(items, opts.KnownScopes)
	withTouches, flexible := splitByTouches(items)
	assignedTracks := assignTracks(withTouches)

	return &Result{
		Tracks:   assignedTracks,
		Flexible: toTrackTasks(flexible),
		Warnings: warnings,
	}, nil
}

func scoreActionable(tasks []*model.Task, filters []string) ([]scored, error) {
	taskMap := next.BuildTaskMap(tasks)
	criticalPath := next.CalculateCriticalPathTasks(tasks, taskMap)
	downstreamCounts := computeDownstreamCounts(tasks)

	candidates := tasks
	if len(filters) > 0 {
		var err error
		candidates, err = filter.Apply(candidates, filters)
		if err != nil {
			return nil, err
		}
	}

	var items []scored
	for _, t := range candidates {
		if next.IsActionable(t, taskMap) {
			s, _ := next.ScoreTask(t, criticalPath, downstreamCounts)
			items = append(items, scored{task: t, score: s})
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].score != items[j].score {
			return items[i].score > items[j].score
		}
		return items[i].task.ID < items[j].task.ID
	})

	return items, nil
}

func validateScopes(items []scored, knownScopes map[string]bool) []string {
	if knownScopes == nil {
		return nil
	}
	var warnings []string
	seen := make(map[string]bool)
	for _, it := range items {
		for _, scope := range it.task.Touches {
			if !knownScopes[scope] && !seen[scope] {
				seen[scope] = true
				warnings = append(warnings, "unknown scope: "+scope)
			}
		}
	}
	return warnings
}

func splitByTouches(items []scored) (withTouches, flexible []scored) {
	for _, it := range items {
		if len(it.task.Touches) > 0 {
			withTouches = append(withTouches, it)
		} else {
			flexible = append(flexible, it)
		}
	}
	return
}

func assignTracks(items []scored) []Track {
	var tracks []Track
	for _, it := range items {
		placed := false
		for ti := range tracks {
			if !overlaps(&tracks[ti], it.task.Touches) {
				addToTrack(&tracks[ti], it)
				placed = true
				break
			}
		}
		if !placed {
			t := newTrack(len(tracks) + 1)
			addToTrack(&t, it)
			tracks = append(tracks, t)
		}
	}
	return tracks
}

func computeDownstreamCounts(tasks []*model.Task) map[string]int {
	g := graph.NewGraph(tasks)
	counts := make(map[string]int, len(tasks))
	for _, t := range tasks {
		counts[t.ID] = len(g.GetDownstream(t.ID))
	}
	return counts
}

func newTrack(id int) Track {
	return Track{
		ID:       id,
		scopeSet: make(map[string]bool),
	}
}

func overlaps(track *Track, touches []string) bool {
	for _, scope := range touches {
		if track.scopeSet[scope] {
			return true
		}
	}
	return false
}

func addToTrack(track *Track, it scored) {
	track.Tasks = append(track.Tasks, TrackTask{
		ID:       it.task.ID,
		Title:    it.task.Title,
		Priority: string(it.task.Priority),
		Effort:   string(it.task.Effort),
		Score:    it.score,
		FilePath: it.task.FilePath,
		Touches:  it.task.Touches,
	})
	for _, scope := range it.task.Touches {
		if !track.scopeSet[scope] {
			track.scopeSet[scope] = true
			track.Scopes = append(track.Scopes, scope)
		}
	}
}

func toTrackTasks(items []scored) []TrackTask {
	out := make([]TrackTask, len(items))
	for i, it := range items {
		out[i] = TrackTask{
			ID:       it.task.ID,
			Title:    it.task.Title,
			Priority: string(it.task.Priority),
			Effort:   string(it.task.Effort),
			Score:    it.score,
			FilePath: it.task.FilePath,
			Touches:  it.task.Touches,
		}
	}
	return out
}
