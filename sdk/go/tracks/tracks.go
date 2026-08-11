package tracks

import (
	"sort"

	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/filter"
	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/next"
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

// Track represents a group of scope-overlapping tasks that must run sequentially.
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
	Filters       []string
	KnownScopes   map[string]bool
	ArchivedTasks []*model.Task
	Scope         string
	// Efforts is the project's effort vocabulary. The zero value means the
	// default small, medium, large.
	Efforts effort.Scale
}

type scored struct {
	task  *model.Task
	score int
}

// Assign groups actionable tasks into parallel tracks based on scope overlap.
func Assign(tasks []*model.Task, opts Options) (*Result, error) {
	items, err := scoreActionable(tasks, opts)
	if err != nil {
		return nil, err
	}

	warnings := validateScopes(items, opts.KnownScopes)

	if opts.Scope != "" {
		return assignScope(items, tasks, opts.Scope, warnings), nil
	}

	depComponents := buildDependencyComponents(tasks)
	assignedTracks, flexible := assignTracks(items, depComponents)
	if assignedTracks == nil {
		assignedTracks = []Track{}
	}

	return &Result{
		Tracks:   assignedTracks,
		Flexible: toTrackTasks(flexible),
		Warnings: warnings,
	}, nil
}

// assignScope returns a single track containing tasks related to the given scope.
// It finds seed tasks whose touches list contains the scope, then expands via
// dependency components to include any actionable task sharing a component with a seed.
func assignScope(items []scored, allTasks []*model.Task, scope string, warnings []string) *Result {
	depComponents := buildDependencyComponents(allTasks)

	// Find seed tasks that directly touch the scope.
	seedComponents := make(map[string]bool)
	seedIDs := make(map[string]bool)
	for _, it := range items {
		for _, s := range it.task.Touches {
			if filter.MatchScope(scope, s) {
				seedIDs[it.task.ID] = true
				if comp, ok := depComponents[it.task.ID]; ok {
					seedComponents[comp] = true
				}
				break
			}
		}
	}

	// Expand: include any actionable task sharing a dependency component with a seed.
	var matched []scored
	for _, it := range items {
		if seedIDs[it.task.ID] {
			matched = append(matched, it)
			continue
		}
		if comp, ok := depComponents[it.task.ID]; ok && seedComponents[comp] {
			matched = append(matched, it)
		}
	}

	result := &Result{
		Tracks:   []Track{},
		Flexible: []TrackTask{},
		Warnings: warnings,
	}

	if len(matched) == 0 {
		return result
	}

	track := newTrack(1)
	for _, it := range matched {
		addToTrack(&track, it)
	}
	result.Tracks = []Track{track}
	return result
}

func scoreActionable(tasks []*model.Task, opts Options) ([]scored, error) {
	taskMap := next.BuildTaskMap(tasks)

	// Merge archived tasks for dependency resolution only.
	for _, at := range opts.ArchivedTasks {
		if _, exists := taskMap[at.ID]; !exists {
			taskMap[at.ID] = at
		}
	}

	childrenMap := next.BuildChildrenMap(tasks)
	criticalPath := next.CalculateCriticalPathTasks(tasks, taskMap)
	downstreamInfo := next.ComputeDownstreamInfo(tasks)

	candidates := tasks
	if len(opts.Filters) > 0 {
		var err error
		candidates, err = filter.Apply(candidates, opts.Filters, opts.Efforts)
		if err != nil {
			return nil, err
		}
	}

	var items []scored
	for _, t := range candidates {
		if next.IsActionable(t, taskMap, childrenMap) {
			s, _ := next.ScoreTask(t, criticalPath, downstreamInfo, opts.Efforts)
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

// unionFind provides path-compressed union-find over integer indices.
type unionFind struct {
	parent []int
}

func newUnionFind(n int) *unionFind {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	return &unionFind{parent: p}
}

func (uf *unionFind) find(x int) int {
	if uf.parent[x] != x {
		uf.parent[x] = uf.find(uf.parent[x])
	}
	return uf.parent[x]
}

func (uf *unionFind) union(a, b int) {
	ra, rb := uf.find(a), uf.find(b)
	if ra != rb {
		uf.parent[ra] = rb
	}
}

func assignTracks(items []scored, depComponents map[string]string) ([]Track, []scored) {
	n := len(items)
	if n == 0 {
		return nil, nil
	}

	uf := newUnionFind(n)
	unionByScopes(uf, items)
	unionByDeps(uf, items, depComponents)
	return splitGroups(uf, items)
}

func unionByScopes(uf *unionFind, items []scored) {
	scopeFirst := make(map[string]int)
	for i, it := range items {
		for _, scope := range it.task.Touches {
			if first, exists := scopeFirst[scope]; exists {
				uf.union(first, i)
			} else {
				scopeFirst[scope] = i
			}
		}
	}
}

func unionByDeps(uf *unionFind, items []scored, depComponents map[string]string) {
	compFirst := make(map[string]int)
	for i, it := range items {
		rep, ok := depComponents[it.task.ID]
		if !ok {
			continue
		}
		if first, exists := compFirst[rep]; exists {
			uf.union(first, i)
		} else {
			compFirst[rep] = i
		}
	}
}

type itemGroup struct {
	indices   []int
	hasScopes bool
}

func splitGroups(uf *unionFind, items []scored) ([]Track, []scored) {
	groups, order := collectGroups(uf, items)

	var tracks []Track
	var flexible []scored
	for _, root := range order {
		g := groups[root]
		if len(g.indices) > 1 || g.hasScopes {
			track := newTrack(len(tracks) + 1)
			for _, idx := range g.indices {
				addToTrack(&track, items[idx])
			}
			tracks = append(tracks, track)
		} else {
			flexible = append(flexible, items[g.indices[0]])
		}
	}
	return tracks, flexible
}

func collectGroups(uf *unionFind, items []scored) (map[int]*itemGroup, []int) {
	seen := make(map[int]*itemGroup)
	var order []int
	for i := range items {
		root := uf.find(i)
		if g, exists := seen[root]; exists {
			g.indices = append(g.indices, i)
			if len(items[i].task.Touches) > 0 {
				g.hasScopes = true
			}
		} else {
			seen[root] = &itemGroup{
				indices:   []int{i},
				hasScopes: len(items[i].task.Touches) > 0,
			}
			order = append(order, root)
		}
	}
	return seen, order
}

// buildDependencyComponents computes connected components from dependency
// edges treated as undirected. Returns a map from task ID to a representative
// component ID (the lexicographically smallest ID in the component).
func buildDependencyComponents(tasks []*model.Task) map[string]string {
	adj, ids := buildDepAdjacency(tasks)
	return bfsComponents(adj, ids)
}

func buildDepAdjacency(tasks []*model.Task) (map[string][]string, map[string]bool) {
	adj := make(map[string][]string)
	ids := make(map[string]bool)
	for _, t := range tasks {
		ids[t.ID] = true
		for _, dep := range t.Dependencies {
			adj[t.ID] = append(adj[t.ID], dep)
			adj[dep] = append(adj[dep], t.ID)
		}
	}
	return adj, ids
}

func bfsComponents(adj map[string][]string, ids map[string]bool) map[string]string {
	visited := make(map[string]bool)
	components := make(map[string]string)

	for id := range ids {
		if visited[id] {
			continue
		}
		members := bfsFrom(id, adj, visited)
		rep := minString(members)
		for _, m := range members {
			components[m] = rep
		}
	}
	return components
}

func bfsFrom(start string, adj map[string][]string, visited map[string]bool) []string {
	queue := []string{start}
	visited[start] = true
	var members []string
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		members = append(members, cur)
		for _, neighbor := range adj[cur] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}
	return members
}

func minString(ss []string) string {
	m := ss[0]
	for _, s := range ss[1:] {
		if s < m {
			m = s
		}
	}
	return m
}

func newTrack(id int) Track {
	return Track{
		ID:       id,
		Scopes:   []string{},
		scopeSet: make(map[string]bool),
	}
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
