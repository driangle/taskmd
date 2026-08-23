export interface Task {
  id: string;
  title: string;
  status: string;
  priority: string;
  effort: string;
  type: string;
  dependencies: string[] | null;
  tags: string[] | null;
  phase: string;
  group: string;
  owner: string;
  parent: string;
  created: string;
  body: string;
  file_path: string;
  worklog_entries?: number;
  worklog_updated?: string;
  // Worktree overlay provenance. Present only when the server merges
  // multiple git worktrees; `status`/`owner` still describe the local copy.
  effective_status?: string;
  effective_owner?: string;
  /** Winning copy's worktree name; empty/absent when the local copy wins. */
  worktree?: string;
  branch?: string;
  /** True when the task exists only in a sibling worktree. */
  remote_only?: boolean;
  /** Per-worktree copies; sent on the detail endpoint when copies diverge. */
  worktrees?: WorktreeCopy[];
}

export interface WorktreeCopy {
  worktree?: string;
  branch?: string;
  status: string;
  owner?: string;
  local?: boolean;
}

export interface WorklogEntry {
  timestamp: string;
  content: string;
}

export interface BoardGroup {
  group: string;
  count: number;
  tasks: BoardTask[];
}

export interface BoardTask {
  id: string;
  title: string;
  status: string;
  priority?: string;
  effort?: string;
  type?: string;
  phase?: string;
  tags?: string[];
  /** Winning copy's worktree name; absent when the local copy wins. */
  worktree?: string;
  /** True when the task exists only in a sibling worktree. */
  remote_only?: boolean;
}

export interface GraphData {
  nodes: GraphNode[];
  edges: GraphEdge[];
  cycles?: string[][];
}

export interface GraphNode {
  id: string;
  title: string;
  status: string;
  priority?: string;
  group?: string;
}

export interface GraphEdge {
  from: string;
  to: string;
}

export interface TagInfo {
  tag: string;
  count: number;
}

export interface Stats {
  total_tasks: number;
  tasks_by_status: Record<string, number>;
  tasks_by_priority: Record<string, number>;
  tasks_by_effort: Record<string, number>;
  tasks_by_phase: Record<string, number>;
  blocked_tasks_count: number;
  critical_path_length: number;
  max_dependency_depth: number;
  avg_dependencies_per_task: number;
  tags_by_count: TagInfo[];
}

export interface ValidationResult {
  issues: ValidationIssue[];
  errors: number;
  warnings: number;
}

export interface ValidationIssue {
  level: "error" | "warning";
  task_id?: string;
  file_path?: string;
  message: string;
}

export interface Recommendation {
  rank: number;
  id: string;
  title: string;
  file_path: string;
  status: string;
  priority: string;
  effort: string;
  score: number;
  reasons: string[];
  downstream_count: number;
  on_critical_path: boolean;
}

export interface TaskUpdateRequest {
  title?: string;
  status?: string;
  priority?: string;
  effort?: string;
  type?: string;
  phase?: string;
  owner?: string;
  parent?: string;
  tags?: string[];
  body?: string;
}

export interface TrackTask {
  id: string;
  title: string;
  priority?: string;
  effort?: string;
  score: number;
  file_path: string;
  touches?: string[];
}

export interface Track {
  id: number;
  tasks: TrackTask[];
  scopes: string[];
}

export interface TracksResult {
  tracks: Track[];
  flexible: TrackTask[];
  warnings?: string[];
}

export interface SearchResult {
  id: string;
  title: string;
  status: string;
  file_path: string;
  match_location: string;
  snippet: string;
}

export interface FeedEntry {
  source: string;
  hash?: string;
  author?: string;
  timestamp: string;
  message: string;
  taskID?: string;
  files?: FeedFileChange[];
}

export interface FeedFileChange {
  path: string;
  status: string;
  taskID?: string;
  taskStatus?: string;
  fieldChanges?: FeedFieldChange[];
  subtaskChanges?: FeedSubtaskChange[];
}

export interface FeedFieldChange {
  field: string;
  oldValue: string;
  newValue: string;
}

export interface FeedSubtaskChange {
  text: string;
  done: boolean;
}

export interface ApiError {
  error: string;
  details?: string[];
}
