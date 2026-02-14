# Web Interface

Complete guide to using the taskmd web dashboard.

## Getting Started

### Starting the Web Server

```bash
# Basic start
taskmd web start

# Auto-open browser
taskmd web start --open

# Custom port
taskmd web start --port 3000 --open

# Specific tasks directory
taskmd web start --dir ./my-tasks --open
```

The server starts on `http://localhost:8080` by default.

### Live Reload

The interface automatically updates when task files change:

1. Edit a task file in your text editor
2. Save the file
3. The web interface updates immediately via Server-Sent Events (SSE)

No page refresh needed.

## Views

### Tasks View

**URL:** `/tasks`

The main task list in a sortable, filterable table.

**Features:**
- **Sortable columns** - click headers to sort (ID, Title, Status, Priority, Effort)
- **Search** - real-time filtering across ID, title, and tags
- **Status filtering** - dropdown to filter by status
- **Clickable tasks** - click ID or title to view full details
- **Dependency counts** - see how many dependencies each task has

### Board View (Kanban)

**URL:** `/board`

Visual board with tasks organized in columns.

**Group by options:**

| Grouping | Columns | Best for |
|----------|---------|----------|
| Status | pending, in-progress, completed, blocked, cancelled | Standard kanban workflow |
| Priority | critical, high, medium, low | Prioritization planning |
| Effort | small, medium, large | Capacity planning |
| Group | Task groups (cli, web, docs...) | Team-based views |
| Tag | One per unique tag | Feature-based organization |

### Graph View

**URL:** `/graph`

Interactive dependency visualization using Mermaid diagrams.

- Nodes represent tasks, arrows show dependencies
- Color-coded by status (yellow=pending, blue=in-progress, green=completed, red=blocked)
- Useful for understanding dependencies, finding critical paths, and spotting blockers

### Stats View

**URL:** `/stats`

Project metrics and analytics:

- **Overview** - total tasks, completion rate, status breakdown
- **Priority breakdown** - tasks by priority level
- **Effort breakdown** - tasks by effort estimate
- **Dependency analysis** - critical path length, max depth, average dependencies

## Common Workflows

### Daily Task Management

1. Open web interface: `taskmd web start --open`
2. Check **Stats** view for project health
3. Switch to **Board** view (Group by: priority) to identify today's priorities
4. Edit task files in your editor - watch the web UI update automatically
5. Review **Board** view at end of day

### Weekly Planning

1. **Stats** view - review progress
2. **Board** view - group by priority
3. **Graph** view - identify dependencies and blockers
4. **Tasks** view - filter by `status=pending` and `priority=high`

### Team Collaboration

- Share screen with the web interface during standups
- Use **Board** view grouped by status for status discussions
- Use **Graph** view to discuss dependencies
- Use **Stats** view for sprint reviews

## API Access

The web interface uses a JSON API you can access directly:

```bash
# Get all tasks
curl http://localhost:8080/api/tasks

# Get board data
curl http://localhost:8080/api/board?groupBy=status

# Get graph data
curl http://localhost:8080/api/graph

# Get statistics
curl http://localhost:8080/api/stats
```

## Advanced Usage

### Remote Access

```bash
# Start server
taskmd web start --port 8080

# Port forward via SSH
ssh -L 8080:localhost:8080 user@remote-host

# Access from local browser
open http://localhost:8080
```

### Multiple Projects

Run separate instances on different ports:

```bash
taskmd web start --dir ~/project1/tasks --port 8081
taskmd web start --dir ~/project2/tasks --port 8082
```

## Troubleshooting

### Server Won't Start

```bash
# Check if port is in use
lsof -i :8080

# Use a different port
taskmd web start --port 3000
```

### No Tasks Showing

1. Verify the correct directory: `--dir ./tasks`
2. Ensure files have `.md` extension and valid YAML frontmatter
3. Check browser console (F12) for errors
4. Run `taskmd validate ./tasks` from the CLI

### Live Reload Not Working

1. Check browser console (F12) for SSE connection messages
2. Verify file is saved (some editors use temporary files)
3. Try refreshing the page manually
4. Restart the server
