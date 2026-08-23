package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/driangle/taskmd/apps/cli/internal/worktree"
)

// ProjectEntry represents a registered project for the API.
type ProjectEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// ProjectResolverFunc resolves a project ID to its task scan directory and phases.
type ProjectResolverFunc func(id string) (scanDir string, phases []PhaseInfo, err error)

// ErrProjectNotFound indicates the requested project ID is not in the registry.
var ErrProjectNotFound = errors.New("project not found")

type contextKey string

const projectDPKey contextKey = "projectDP"
const projectPhasesKey contextKey = "projectPhases"

type projectContext struct {
	dp     *DataProvider
	phases []PhaseInfo
}

// ProjectResolver caches DataProviders per project ID.
type ProjectResolver struct {
	resolve ProjectResolverFunc
	verbose bool
	wt      worktree.Builder

	mu    sync.RWMutex
	cache map[string]*projectContext
}

// NewProjectResolver creates a resolver with lazy caching. Each project's
// DataProvider builds the worktree overlay for its own scan dir.
func NewProjectResolver(resolve ProjectResolverFunc, verbose bool, wt worktree.Builder) *ProjectResolver {
	return &ProjectResolver{
		resolve: resolve,
		verbose: verbose,
		wt:      wt,
		cache:   make(map[string]*projectContext),
	}
}

// get returns a cached projectContext for the given ID, creating one on first access.
func (pr *ProjectResolver) get(id string) (*projectContext, error) {
	pr.mu.RLock()
	if ctx, ok := pr.cache[id]; ok {
		pr.mu.RUnlock()
		return ctx, nil
	}
	pr.mu.RUnlock()

	pr.mu.Lock()
	defer pr.mu.Unlock()

	// Double-check after acquiring write lock.
	if ctx, ok := pr.cache[id]; ok {
		return ctx, nil
	}

	scanDir, phases, err := pr.resolve(id)
	if err != nil {
		return nil, err
	}

	ctx := &projectContext{
		dp:     NewDataProviderWithWorktrees(scanDir, pr.verbose, pr.wt),
		phases: phases,
	}
	pr.cache[id] = ctx
	return ctx, nil
}

// projectMiddleware resolves ?project=<id> and injects the DataProvider into request context.
func projectMiddleware(resolver *ProjectResolver, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := r.URL.Query().Get("project")
		if projectID == "" || resolver == nil {
			next.ServeHTTP(w, r)
			return
		}

		pctx, err := resolver.get(projectID)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, ErrProjectNotFound) {
				status = http.StatusNotFound
			}
			writeError(w, status, fmt.Sprintf("project %q: %s", projectID, err.Error()), nil)
			return
		}

		ctx := context.WithValue(r.Context(), projectDPKey, pctx.dp)
		ctx = context.WithValue(ctx, projectPhasesKey, pctx.phases)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// effectiveDP returns the project-scoped DataProvider from request context,
// falling back to the default.
func effectiveDP(r *http.Request, defaultDP *DataProvider) *DataProvider {
	if dp, ok := r.Context().Value(projectDPKey).(*DataProvider); ok {
		return dp
	}
	return defaultDP
}

// effectivePhases returns the project-scoped phases from request context,
// falling back to the default.
func effectivePhases(r *http.Request, defaultPhases []PhaseInfo) []PhaseInfo {
	if phases, ok := r.Context().Value(projectPhasesKey).([]PhaseInfo); ok {
		return phases
	}
	return defaultPhases
}
