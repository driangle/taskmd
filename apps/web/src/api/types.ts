export interface Task {
  id: string;
  title: string;
  status: string;
  priority: string;
  effort: string;
  dependencies: string[] | null;
  tags: string[] | null;
  group: string;
  owner: string;
  parent: string;
  created: string;
  body: string;
  file_path: string;
  worklog_entries?: number;
  worklog_updated?: string;
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
  tags?: string[];
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

export interface ApiError {
  error: string;
  details?: string[];
}
