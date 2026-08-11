import type {
  Task,
  BoardGroup,
  BoardTask,
  Stats,
  GraphData,
  GraphNode,
  GraphEdge,
  ValidationResult,
  ValidationIssue,
  Recommendation,
  WorklogEntry,
  SearchResult,
  Track,
  TrackTask,
  TracksResult,
} from "../api/types.ts";
import { DEFAULT_EFFORTS } from "../components/tasks/TaskTable/constants.ts";

// --- Task ---

let taskCounter = 0;

export function createTask(overrides: Partial<Task> = {}): Task {
  const n = ++taskCounter;
  return {
    id: `${String(n).padStart(3, "0")}`,
    title: `Task ${n}`,
    status: "pending",
    priority: "medium",
    effort: "medium",
    type: "feature",
    dependencies: null,
    tags: null,
    phase: "",
    group: "cli",
    owner: "",
    parent: "",
    created: "2026-01-15",
    body: "",
    file_path: `tasks/cli/${String(n).padStart(3, "0")}-task-${n}.md`,
    ...overrides,
  };
}

// --- Board ---

export function createBoardTask(overrides: Partial<BoardTask> = {}): BoardTask {
  const n = ++taskCounter;
  return {
    id: `${String(n).padStart(3, "0")}`,
    title: `Task ${n}`,
    status: "pending",
    ...overrides,
  };
}

export function createBoardGroup(overrides: Partial<BoardGroup> & { tasks?: BoardTask[] } = {}): BoardGroup {
  const tasks = overrides.tasks ?? [createBoardTask(), createBoardTask()];
  return {
    group: "pending",
    count: tasks.length,
    tasks,
    ...overrides,
  };
}

// --- Stats ---

export function createStats(overrides: Partial<Stats> = {}): Stats {
  return {
    total_tasks: 10,
    tasks_by_status: { pending: 5, "in-progress": 3, completed: 2 },
    tasks_by_priority: { high: 3, medium: 4, low: 3 },
    tasks_by_effort: { small: 4, medium: 3, large: 3 },
    tasks_by_phase: {},
    blocked_tasks_count: 1,
    critical_path_length: 3,
    max_dependency_depth: 2,
    avg_dependencies_per_task: 0.5,
    tags_by_count: [
      { tag: "backend", count: 4 },
      { tag: "frontend", count: 3 },
    ],
    ...overrides,
  };
}

// --- Graph ---

export function createGraphNode(overrides: Partial<GraphNode> = {}): GraphNode {
  const n = ++taskCounter;
  return {
    id: `${String(n).padStart(3, "0")}`,
    title: `Task ${n}`,
    status: "pending",
    ...overrides,
  };
}

export function createGraphEdge(overrides: Partial<GraphEdge> = {}): GraphEdge {
  return {
    from: "001",
    to: "002",
    ...overrides,
  };
}

export function createGraphData(overrides: Partial<GraphData> = {}): GraphData {
  return {
    nodes: [
      createGraphNode({ id: "001", title: "First" }),
      createGraphNode({ id: "002", title: "Second" }),
    ],
    edges: [createGraphEdge({ from: "001", to: "002" })],
    ...overrides,
  };
}

// --- Validation ---

export function createValidationIssue(overrides: Partial<ValidationIssue> = {}): ValidationIssue {
  return {
    level: "error",
    task_id: "001",
    file_path: "tasks/cli/001-task.md",
    message: "Missing required field: title",
    ...overrides,
  };
}

export function createValidationResult(overrides: Partial<ValidationResult> = {}): ValidationResult {
  return {
    issues: [],
    errors: 0,
    warnings: 0,
    ...overrides,
  };
}

// --- Recommendation ---

export function createRecommendation(overrides: Partial<Recommendation> = {}): Recommendation {
  const n = ++taskCounter;
  return {
    rank: 1,
    id: `${String(n).padStart(3, "0")}`,
    title: `Recommended Task ${n}`,
    file_path: `tasks/cli/${String(n).padStart(3, "0")}-task.md`,
    status: "pending",
    priority: "high",
    effort: "small",
    score: 85,
    reasons: ["High priority", "No blockers"],
    downstream_count: 2,
    on_critical_path: true,
    ...overrides,
  };
}

// --- Worklog ---

export function createWorklogEntry(overrides: Partial<WorklogEntry> = {}): WorklogEntry {
  return {
    timestamp: "2026-01-15T10:30:00Z",
    content: "Started implementation.",
    ...overrides,
  };
}

// --- Search ---

export function createSearchResult(overrides: Partial<SearchResult> = {}): SearchResult {
  const n = ++taskCounter;
  return {
    id: `${String(n).padStart(3, "0")}`,
    title: `Search Result ${n}`,
    status: "pending",
    file_path: `tasks/cli/${String(n).padStart(3, "0")}-task.md`,
    match_location: "title",
    snippet: `...result ${n}...`,
    ...overrides,
  };
}

// --- Tracks ---

export function createTrackTask(overrides: Partial<TrackTask> = {}): TrackTask {
  const n = ++taskCounter;
  return {
    id: `${String(n).padStart(3, "0")}`,
    title: `Track Task ${n}`,
    score: 50,
    file_path: `tasks/cli/${String(n).padStart(3, "0")}-task.md`,
    ...overrides,
  };
}

export function createTrack(overrides: Partial<Track> = {}): Track {
  return {
    id: 1,
    tasks: [createTrackTask(), createTrackTask()],
    scopes: ["cli"],
    ...overrides,
  };
}

export function createTracksResult(overrides: Partial<TracksResult> = {}): TracksResult {
  return {
    tracks: [createTrack()],
    flexible: [createTrackTask()],
    ...overrides,
  };
}

// --- Config ---

export function createConfig(
  overrides: Partial<{
    readonly: boolean;
    version: string;
    phases: { id: string; name: string; description: string }[];
    efforts: string[];
  }> = {},
) {
  return {
    readonly: false,
    version: "1.0.0",
    phases: [],
    efforts: DEFAULT_EFFORTS,
    ...overrides,
  };
}

/** Reset the internal counter (call in beforeEach if you need deterministic IDs). */
export function resetFixtureCounter() {
  taskCounter = 0;
}
