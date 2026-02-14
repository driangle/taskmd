# Web User Guide

Complete guide to using the taskmd web interface.

## What You'll Learn

- Starting the web server
- Navigating the interface
- Using all views (Tasks, Board, Graph, Stats)
- Filtering and searching
- Keyboard shortcuts
- Live reload functionality

## Getting Started

### Starting the Web Server

**Basic command:**
```bash
taskmd web start
```

The server will start on `http://localhost:8080`.

**With options:**
```bash
# Auto-open browser
taskmd web start --open

# Custom port
taskmd web start --port 3000 --open

# Specific tasks directory
taskmd web start --dir ./my-tasks --open

# Development mode (CORS for Vite)
taskmd web start --dev
```

**What you'll see:**
```
Starting taskmd web server...
Server running at http://localhost:8080
Watching for changes in: ./tasks
Press Ctrl+C to stop
```

### Accessing the Interface

Open your browser to:
- Default: `http://localhost:8080`
- Custom port: `http://localhost:[port]`

Or use `--open` flag to launch automatically.

### Live Reload

The interface automatically updates when task files change:
- Edit task files in your text editor
- Save the file
- Web interface updates immediately

No page refresh needed!

## Interface Overview

### Navigation

The top navigation bar provides quick access to all views:

- **taskmd** (logo) - Click to return to task list
- **Tasks** - Main task list view
- **Board** - Kanban board view
- **Graph** - Dependency visualization
- **Stats** - Project statistics

### Views

#### 1. Tasks View (Default)

**URL:** `http://localhost:8080/tasks`

The main task list in a sortable, filterable table.

**Features:**
- Sortable columns (click headers)
- Search/filter functionality
- Status badges with colors
- Priority indicators
- Clickable task IDs and titles
- Dependency counts

**Columns:**
- **ID** - Task identifier (clickable)
- **Title** - Task name (clickable)
- **Status** - Current status with color badge
- **Priority** - Priority level
- **Effort** - Estimated effort
- **Tags** - Task tags
- **Deps** - Dependency count (clickable)

**Sorting:**
- Click any column header to sort
- Click again to reverse sort order
- Default sort: by ID

**Searching:**
- Use the search box at the top
- Searches across: ID, title, tags
- Real-time filtering as you type

**Filtering:**
- Filter by status using dropdown
- Filter by priority
- Filter by tags
- Combine multiple filters

**Task Details:**
- Click task ID or title to view full details
- See complete markdown content
- View dependencies
- Check subtasks

#### 2. Board View (Kanban)

**URL:** `http://localhost:8080/board`

Visual board with tasks organized in columns.

**Features:**
- Cards grouped by selected field
- Column headers with task counts
- Color-coded by status
- Scroll horizontally for many columns

**Grouping Options:**

Use the "Group by" dropdown to organize by:

**Status (default):**
- Columns: pending, in-progress, completed, blocked, cancelled
- Color-coded for quick status overview
- Standard kanban workflow

**Priority:**
- Columns: critical, high, medium, low
- Helps prioritize work visually
- Identify high-priority tasks at a glance

**Effort:**
- Columns: small, medium, large
- Plan capacity and quick wins
- Balance workload

**Group:**
- Columns: task groups (e.g., cli, web, docs)
- Organize by project area
- Team-based views

**Tag:**
- Columns: one per unique tag
- Tasks may appear in multiple columns
- Feature-based organization

**Card Contents:**
- Task title (clickable)
- Task ID
- Priority (if set)
- Click to view full details

**Best for:**
- Visual project overview
- Sprint planning
- Status at a glance
- Team standup discussions

#### 3. Graph View

**URL:** `http://localhost:8080/graph`

Interactive dependency visualization using Mermaid diagrams.

**Features:**
- Visual dependency graph
- Nodes represent tasks
- Arrows show dependencies
- Color-coded by status
- Interactive exploration

**Node Colors:**
- 🟡 Yellow - pending
- 🔵 Blue - in-progress
- 🟢 Green - completed
- 🔴 Red - blocked

**Interactions:**
- **Hover** - Highlight task
- **Click** - View task details (future enhancement)
- **Pan** - Drag to move view
- **Zoom** - Mouse wheel or pinch

**Understanding the Graph:**
- **Arrows** - Point from dependency to dependent
  - Task A → Task B means B depends on A
- **Chains** - Long paths show critical paths
- **Bottlenecks** - Tasks with many outgoing arrows block others
- **Clusters** - Groups of related tasks

**Best for:**
- Understanding dependencies
- Identifying critical paths
- Finding blockers
- Planning parallel work
- Detecting circular dependencies

#### 4. Stats View

**URL:** `http://localhost:8080/stats`

Project metrics and analytics.

**Metrics Displayed:**

**Overview:**
- Total tasks
- Completion rate (percentage)
- Tasks by status breakdown
  - Pending count
  - In-progress count
  - Completed count
  - Blocked count

**Priority Breakdown:**
- Tasks by priority level
- Critical, high, medium, low counts
- Distribution visualization

**Effort Breakdown:**
- Tasks by effort estimate
- Small, medium, large counts
- Capacity planning data

**Dependency Analysis:**
- Critical path length (longest chain)
- Maximum dependency depth
- Average dependencies per task
- Blocked tasks count

**Best for:**
- Project health checks
- Progress tracking
- Sprint retrospectives
- Stakeholder reports
- Capacity planning

## Common Workflows

### Daily Task Management

**Morning Check-in:**
1. Open web interface: `taskmd web start --open`
2. Check Stats view for project health
3. Switch to Board view (Group by: priority)
4. Identify today's priorities
5. Click tasks to review details

**During Work:**
1. Edit task files in your editor
2. Mark status as `in-progress`
3. Check off subtasks as completed
4. Watch web interface update automatically

**End of Day:**
1. Review Board view
2. Update any task statuses
3. Check Stats for completion progress
4. Plan tomorrow's priorities

### Weekly Planning

**Monday Planning:**
1. Launch web interface
2. **Stats view** - Review last week's progress
3. **Board view** - Group by priority
4. **Graph view** - Identify dependencies and blockers
5. **Tasks view** - Filter by `status=pending` and `priority=high`
6. Plan week's priorities

**Friday Review:**
1. **Stats view** - Check completion metrics
2. **Board view** - Group by status
3. **Tasks view** - Filter by `status=completed`
4. Celebrate wins!

### Project Planning

**Initial Planning:**
1. Create task files with dependencies
2. Open web interface
3. **Graph view** - Visualize project structure
4. Identify critical path
5. **Board view** - Group by effort for capacity planning
6. Adjust priorities and dependencies as needed

**Ongoing Management:**
1. **Tasks view** - Regular task list reviews
2. **Board view** - Visual status tracking
3. **Graph view** - Dependency management
4. **Stats view** - Progress monitoring

### Team Collaboration

**Stand-ups:**
1. Share screen with web interface
2. **Board view** - Group by status
3. Discuss in-progress and blocked tasks
4. **Graph view** - Discuss dependencies

**Sprint Reviews:**
1. **Stats view** - Show metrics
2. **Tasks view** - Filter by `status=completed`
3. Demonstrate completed work
4. Review velocity

**Sprint Planning:**
1. **Graph view** - Understand dependencies
2. **Board view** - Group by effort
3. **Tasks view** - Filter and prioritize
4. Assign tasks

## Features in Detail

### Task List Features

**Search/Filter:**
```
Search box: Type to filter by ID, title, or tags
Status dropdown: Filter by specific status
Priority filter: Show only high-priority tasks
```

**Sorting:**
- Click column headers to sort
- Supported columns: ID, Title, Status, Priority, Effort
- Click again to reverse sort

**Task Cards:**
- Compact view for many tasks
- Color-coded status badges
- Priority indicators
- Dependency counts (click to see deps)

**Clicking Tasks:**
- Click ID or title → Full task details
- See complete markdown
- View all metadata
- Read acceptance criteria

### Board Features

**Group By Options:**

Each grouping provides different insights:

1. **Status** - Progress tracking
   - See work distribution
   - Identify bottlenecks
   - Standard kanban flow

2. **Priority** - Urgency planning
   - Focus on critical tasks
   - Balance priorities
   - Emergency vs. planned work

3. **Effort** - Capacity planning
   - Quick wins (small tasks)
   - Major projects (large tasks)
   - Sprint sizing

4. **Group** - Area organization
   - Feature teams
   - Project components
   - Responsibility areas

5. **Tag** - Cross-cutting views
   - Frontend vs. backend
   - Bug vs. feature
   - Client-specific work

**Visual Design:**
- Color-coded columns (status grouping)
- Card hover effects
- Responsive layout
- Horizontal scroll for many columns

### Graph Features

**Dependency Visualization:**
- See the full task network
- Understand relationships
- Identify critical paths
- Spot potential issues

**Color Coding:**
- Completed tasks (green)
- In-progress tasks (blue)
- Pending tasks (yellow)
- Blocked tasks (red)

**Use Cases:**
- Planning parallel work streams
- Finding tasks that unblock others
- Understanding project structure
- Detecting circular dependencies

### Stats Features

**Key Metrics:**

1. **Completion Rate**
   - Percentage of completed tasks
   - Overall project progress
   - Velocity indicator

2. **Status Distribution**
   - Work in progress
   - Backlog size
   - Blocked count

3. **Priority Analysis**
   - Focus areas
   - Risk assessment
   - Planning data

4. **Dependency Metrics**
   - Critical path length
   - Project duration estimate
   - Complexity indicators

**Practical Use:**
- Daily health check
- Weekly reviews
- Stakeholder updates
- Process improvements

## Live Reload

### How It Works

The server watches task files and sends updates via Server-Sent Events (SSE):

1. You edit a task file in your editor
2. Server detects the file change
3. Server sends event to browser
4. Browser refetches data automatically
5. UI updates without page refresh

### What Triggers Reload

- Creating new `.md` files in tasks directory
- Modifying existing task files
- Deleting task files
- Moving tasks between directories

### Troubleshooting Live Reload

**If updates don't appear:**

1. **Check console** (F12 → Console)
   - Look for SSE connection messages
   - Check for errors

2. **Verify file save**
   - Ensure file is actually saved
   - Some editors use temporary files

3. **Check file location**
   - File must be in watched directory
   - Subdirectories are watched

4. **Refresh manually**
   - Press F5 if needed
   - Should reconnect automatically

5. **Restart server**
   ```bash
   # Stop with Ctrl+C, then restart
   taskmd web start --open
   ```

## Keyboard Shortcuts

### Navigation

- **Click links** - Navigate between views
- **Browser back/forward** - Navigate history
- **Ctrl+Click** - Open in new tab

### General

- **F5** - Refresh page (rarely needed with live reload)
- **F12** - Open developer tools
- **Ctrl+F** - Find in page

### Future Enhancements

Planned keyboard shortcuts:
- `j/k` - Navigate tasks up/down
- `Enter` - Open task details
- `/` - Focus search
- `Esc` - Clear filters/search
- `g + t` - Go to Tasks view
- `g + b` - Go to Board view
- `g + g` - Go to Graph view
- `g + s` - Go to Stats view

## Tips and Best Practices

### Performance

**Large Projects:**
- Web interface handles 100s of tasks well
- Graph view may be slow with >200 tasks
- Use filtering to reduce displayed tasks
- Consider organizing tasks in subdirectories

**Browser Recommendations:**
- Chrome/Edge - Best performance
- Firefox - Good performance
- Safari - Good performance
- Use modern browser versions

### Workflow Tips

**1. Keep web interface open**
- Background tab stays updated
- Quick reference while coding
- Check progress easily

**2. Use multiple views**
- Tasks for detailed work
- Board for planning
- Graph for architecture
- Stats for reporting

**3. Share with team**
- Run on shared server
- Port forward for remote access
- Use for demos and reviews

**4. Combine with CLI**
- Use CLI for automation
- Use web for exploration
- Best of both worlds

### Organization Tips

**Tags for Grouping:**
```yaml
tags:
  - frontend
  - backend
  - database
  - docs
```

Then use Board view → Group by: tag

**Consistent Naming:**
- Use standard status values
- Use standard priority levels
- Use consistent tag names

**Dependencies:**
- Keep chains short
- Use Graph view to verify
- Break up long chains

## Troubleshooting

### Server Won't Start

**Error: "Port already in use"**
```bash
# Check what's using the port
lsof -i :8080

# Use different port
taskmd web start --port 3000
```

**Error: "Permission denied"**
```bash
# Don't use privileged ports (< 1024)
# Use ports 3000-9999
taskmd web start --port 8080
```

### No Tasks Showing

**Check:**
1. Correct directory: `--dir ./tasks`
2. Files have `.md` extension
3. Files have valid YAML frontmatter
4. Browser console for errors (F12)

**Debug:**
```bash
# Validate tasks from CLI
taskmd validate ./tasks

# Check verbose output
taskmd web start --verbose
```

### Live Reload Not Working

**Check browser console:**
1. Press F12
2. Look for SSE connection messages
3. Check for errors

**Try:**
1. Refresh page (F5)
2. Restart server
3. Clear browser cache
4. Try different browser

### Graph Not Rendering

**Mermaid rendering issues:**
1. Wait a few seconds (large graphs take time)
2. Check browser console for errors
3. Try a different browser
4. Reduce task count with filters

### Page Loads Slowly

**For large projects:**
1. Use filters to reduce data
2. Organize tasks in subdirectories
3. Archive completed tasks
4. Check network tab (F12) for slow requests

## Advanced Usage

### Custom Port and Access

**Local development:**
```bash
taskmd web start --port 3000 --dev
```

**Remote access:**
```bash
# Start server
taskmd web start --port 8080

# Port forward via SSH
ssh -L 8080:localhost:8080 user@remote-host

# Access from local browser
open http://localhost:8080
```

### Multiple Projects

**Switch directories:**
```bash
# Project 1
taskmd web start --dir ~/project1/tasks --port 8081

# Project 2
taskmd web start --dir ~/project2/tasks --port 8082
```

**Use tabs** in your browser to manage multiple projects.

### Integration with Development

**Auto-start with project:**

Create a `start-dev.sh` script:
```bash
#!/bin/bash

# Start taskmd in background
taskmd web start --dir ./tasks &
TASKMD_PID=$!

# Start your app
npm run dev

# Cleanup on exit
trap "kill $TASKMD_PID" EXIT
```

### Export and Sharing

**Generate static snapshot:**
```bash
# Create JSON snapshot
taskmd snapshot --derived > public/tasks.json

# Share via any web server
python -m http.server 8000
```

**Screenshots for Reports:**
1. Open web interface
2. Navigate to desired view
3. Take screenshot (OS screenshot tool)
4. Include in documentation

## Mobile Access

### Mobile Browsers

The web interface works on mobile browsers:
- Responsive design
- Touch-friendly
- Scrollable views

**Best views for mobile:**
1. **Tasks** - Works well, but table is wide
2. **Board** - Horizontal scroll, good for quick checks
3. **Stats** - Excellent on mobile
4. **Graph** - May be hard to read on small screens

**Tips for mobile:**
- Use landscape mode for tasks view
- Board view works well in portrait
- Stats view is mobile-optimized
- Graph view may need zoom

## API Access

The web interface uses a JSON API that you can access directly:

### Endpoints

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

### Use Cases

- Custom automation
- External integrations
- Mobile app development
- Reporting tools

## Configuration

### Config File (Coming Soon)

Configuration file support is **planned but not yet implemented**. See [task 056](../../tasks/056-implement-taskmd-yaml-config.md).

### Current Options

Use command-line flags:

```bash
# Set port and auto-open
taskmd web start --port 3000 --open --dir ./tasks

# Or create a shell alias
alias tmweb='taskmd web start --port 8080 --open --dir ./tasks'
```

### Environment Variables

```bash
# Override directory
export TASKMD_DIR=./my-tasks

# Start server
taskmd web start
```

## Getting Help

### Resources

- **[Quick Start Guide](quickstart.md)** - Get started fast
- **[CLI Guide](cli-guide.md)** - Command-line reference
- **[Task Specification](../taskmd_specification.md)** - Task format

### Debugging

**Browser Developer Tools:**
1. Press F12
2. Check Console tab for errors
3. Check Network tab for failed requests
4. Check Application tab for SSE status

**Server Logs:**
```bash
# Run with verbose output
taskmd web start --verbose

# Check for errors in output
```

### Reporting Issues

When reporting problems, include:
1. taskmd version: `taskmd --version`
2. Browser and version
3. Operating system
4. Error messages from console
5. Steps to reproduce

## Future Features

Planned enhancements:

- [ ] **Inline editing** - Edit tasks in web interface
- [ ] **Drag and drop** - Reorder and change status
- [ ] **Dark mode** - Theme toggle
- [ ] **Keyboard shortcuts** - Power user features
- [ ] **Task creation** - Create tasks via web
- [ ] **Filtering UI** - Advanced filter builder
- [ ] **Export from UI** - Download reports
- [ ] **User preferences** - Save view settings
- [ ] **Real-time collaboration** - Multi-user support

---

**Next:** Explore [CLI commands](cli-guide.md) for automation.
