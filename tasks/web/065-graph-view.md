---
id: "065"
title: "Implement Graph view for task dependencies"
status: completed
priority: medium
effort: large
dependencies: ["007"]
tags:
  - web
  - typescript
  - visualization
  - graph
  - mvp
created: 2026-02-12
---

# Implement Graph View for Task Dependencies

## Objective

Implement a graph visualization view in the web app that displays task dependencies as a directed acyclic graph (DAG), similar to the CLI `graph` command output. This provides users with a visual understanding of task relationships and project structure.

## Tasks

- [ ] Research and choose a graph visualization library:
  - Options: D3.js, vis.js, cytoscape.js, react-flow, mermaid
  - Consider: performance, interactivity, styling flexibility, bundle size
- [ ] Create graph data structure from tasks
  - Parse task dependencies
  - Build nodes (tasks) and edges (dependencies)
  - Handle missing or invalid dependencies
- [ ] Create `GraphView` component
  - Render nodes as task cards/boxes
  - Render edges as arrows/lines
  - Color-code by status (pending, in-progress, completed, blocked)
- [ ] Implement interactive features:
  - Pan and zoom
  - Click task to view details
  - Hover to highlight dependencies
  - Drag nodes to reposition (optional)
- [ ] Add graph layout options:
  - Hierarchical (top-to-bottom or left-to-right)
  - Force-directed
  - Circular (optional)
- [ ] Add filtering controls:
  - Filter by status
  - Filter by tags
  - Show/hide completed tasks
  - Search/highlight specific tasks
- [ ] Add graph controls:
  - Zoom in/out buttons
  - Fit to screen
  - Reset layout
  - Export as image (optional)
- [ ] Style nodes based on task properties:
  - Status colors
  - Priority indicators
  - Show task ID and title
  - Blocked tasks indicator
- [ ] Handle edge cases:
  - Circular dependencies (show error/warning)
  - Large graphs (100+ tasks)
  - Disconnected components
  - Tasks with no dependencies
- [ ] Add graph statistics panel:
  - Total tasks
  - Connected components
  - Longest dependency chain
  - Blocked tasks count
- [ ] Implement responsive layout for mobile
- [ ] Add loading states and error handling
- [ ] Write tests for graph data transformation
- [ ] Update navigation to include Graph view
- [ ] Update documentation

## Acceptance Criteria

- Graph view accessible from main navigation
- Tasks displayed as nodes with ID and title
- Dependencies shown as directed edges
- Different colors for different task statuses
- Interactive: click task to view details, pan/zoom
- Filter controls work correctly
- Performance is acceptable for 100+ tasks
- Responsive design works on mobile
- Handles edge cases gracefully (circular deps, missing tasks)
- Graph can be exported or shared (optional)
- All tests pass

## Implementation Notes

### Graph Library Recommendations

**react-flow** (Recommended):
- ✅ React-specific, good performance
- ✅ Built-in interactivity (pan, zoom, drag)
- ✅ Customizable node components
- ✅ Good TypeScript support
- ❌ Larger bundle size

**cytoscape.js**:
- ✅ Powerful, mature library
- ✅ Many layout algorithms
- ✅ Good for large graphs
- ❌ Not React-specific (needs wrapper)

**mermaid**:
- ✅ Simple, declarative syntax
- ✅ Good for static graphs
- ❌ Less interactive
- ❌ Limited customization

### Data Structure

```typescript
interface GraphNode {
  id: string;
  label: string;
  status: TaskStatus;
  priority?: string;
  tags?: string[];
  blocked?: boolean;
}

interface GraphEdge {
  from: string; // source task ID
  to: string;   // target task ID
  type: 'dependency' | 'blocks';
}

interface TaskGraph {
  nodes: GraphNode[];
  edges: GraphEdge[];
  stats: {
    totalTasks: number;
    components: number;
    longestChain: number;
    blockedCount: number;
  };
}
```

### Layout Considerations

- **Hierarchical**: Best for showing task flow and dependencies clearly
- **Force-directed**: Good for exploring relationships, but can be messy
- **Automatic**: Use hierarchical by default, let users switch

### Performance

For large graphs (100+ tasks):
- Use virtualization or level-of-detail rendering
- Lazy-load task details
- Debounce filtering/search
- Consider WebGL rendering for very large graphs

### Styling

```typescript
// Node colors by status
const statusColors = {
  pending: '#gray',
  'in-progress': '#yellow',
  completed: '#green',
  blocked: '#red',
  cancelled: '#gray-dark'
};

// Priority indicators
const priorityBorder = {
  high: 'solid 3px',
  medium: 'solid 2px',
  low: 'solid 1px'
};
```

## Examples

### API Endpoint

```typescript
// GET /api/graph?status=pending,in-progress
{
  nodes: [
    { id: "001", label: "Task 1", status: "completed" },
    { id: "002", label: "Task 2", status: "in-progress", blocked: false },
    { id: "003", label: "Task 3", status: "pending", blocked: true }
  ],
  edges: [
    { from: "001", to: "002", type: "dependency" },
    { from: "002", to: "003", type: "dependency" }
  ],
  stats: {
    totalTasks: 3,
    components: 1,
    longestChain: 3,
    blockedCount: 1
  }
}
```

### Component Usage

```typescript
import { GraphView } from '@/components/graph/GraphView';

function ProjectGraph() {
  return (
    <GraphView
      projectId={projectId}
      layout="hierarchical"
      onTaskClick={(taskId) => router.push(`/tasks/${taskId}`)}
      filters={{ status: ['pending', 'in-progress'] }}
    />
  );
}
```

## References

- CLI graph command: `internal/cli/graph.go`
- Graph package: `internal/graph/`
- Task dependencies: stored in frontmatter `dependencies` field
- Similar views: Gantt charts, Kanban boards (for inspiration)
