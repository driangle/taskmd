---
id: "062"
title: "Implement task editing interface in web app"
status: pending
priority: high
effort: large
dependencies: []
tags:
  - web
  - frontend
  - feature
  - editing
  - mvp
created: 2026-02-12
---

# Implement Task Editing Interface in Web App

## Objective

Add the ability for users to edit tasks directly from the web interface. Users should be able to open any task, modify both frontmatter metadata and the markdown body, and save changes back to the filesystem. The interface should validate frontmatter fields against the taskmd schema and provide a user-friendly editing experience.

## Context

Currently, the web interface is read-only. Users can view tasks, filter them, and visualize them in various ways (board, table, graph), but cannot make any modifications. Adding editing capabilities will make the web interface a full-featured task management tool.

## Tasks

### UI Components

- [ ] Create TaskEditor component for editing task details
  - Modal or drawer interface for editing
  - Tabbed or sectioned layout (metadata vs markdown body)
  - Responsive design for mobile and desktop
- [ ] Build FrontmatterForm component
  - Input fields for all frontmatter fields (id, title, status, priority, effort)
  - Dropdown/select for enum fields (status, priority, effort)
  - Tags input with autocomplete from existing tags
  - Dependencies selector with task ID search/autocomplete
  - Date picker for created field
  - Group input/selector
- [ ] Build MarkdownEditor component
  - Textarea or rich markdown editor
  - Syntax highlighting for markdown
  - Preview pane (optional but nice to have)
  - Character/word count
- [ ] Add "Edit" button/action to task views
  - In TaskTable rows
  - In task detail view
  - In board cards
  - In graph nodes (optional)

### API Endpoints

- [ ] Create PUT/PATCH `/api/tasks/:id` endpoint
  - Accept task updates (frontmatter + body)
  - Validate against schema
  - Write changes to filesystem
  - Return updated task or validation errors
- [ ] Add validation middleware
  - Validate required fields (id, title, status)
  - Validate enum values (status, priority, effort)
  - Validate dependencies reference existing tasks
  - Check for circular dependencies
  - Return clear error messages
- [ ] Handle file operations safely
  - Read current file contents
  - Parse frontmatter and body
  - Update only changed fields
  - Preserve file formatting where possible
  - Handle file write errors gracefully

### Validation & Error Handling

- [ ] Implement client-side validation
  - Required field validation
  - Enum value validation
  - Format validation (dates, IDs)
  - Dependency validation
  - Real-time validation feedback
- [ ] Implement server-side validation
  - Reuse existing validator from CLI
  - Validate against taskmd specification
  - Check for duplicate IDs
  - Validate circular dependencies
  - Return structured error responses
- [ ] Add error UI components
  - Field-level error messages
  - Form-level error summary
  - Toast/notification for save success/failure
  - Conflict resolution for concurrent edits

### State Management

- [ ] Add edit state to task context/store
  - Track which task is being edited
  - Store draft changes before save
  - Handle optimistic updates
  - Revert on error or cancel
- [ ] Implement save workflow
  - Show loading state during save
  - Handle save success (update cache, close editor)
  - Handle save failure (show errors, keep editor open)
  - Confirm before discarding unsaved changes

### Testing

- [ ] Unit tests for validation logic
- [ ] Integration tests for API endpoints
- [ ] E2E tests for edit workflow
  - Open task editor
  - Modify fields
  - Save changes
  - Verify changes persisted
- [ ] Test error scenarios
  - Invalid field values
  - Duplicate IDs
  - Missing dependencies
  - File write failures

### Documentation

- [ ] Update user guide with editing instructions
- [ ] Add API documentation for edit endpoints
- [ ] Document validation rules
- [ ] Add inline help text for complex fields

## Acceptance Criteria

- ✅ Users can click "Edit" on any task to open the editor
- ✅ Editor displays current task frontmatter in editable form fields
- ✅ Editor displays current task markdown body in editable textarea
- ✅ Status field is a dropdown with valid values: pending, in-progress, completed, blocked, cancelled
- ✅ Priority field is a dropdown with valid values: low, medium, high, critical
- ✅ Effort field is a dropdown with valid values: small, medium, large
- ✅ Tags can be added/removed with autocomplete from existing tags
- ✅ Dependencies can be added/removed with task ID search
- ✅ Client-side validation prevents invalid values
- ✅ Server-side validation catches any issues missed by client
- ✅ Clear error messages for validation failures
- ✅ Changes are saved to the filesystem when user clicks "Save"
- ✅ File format and structure are preserved (YAML frontmatter + markdown body)
- ✅ Success notification shown after save
- ✅ Task list/board/view updates with new values after save
- ✅ Cancel button discards changes without saving
- ✅ Unsaved changes warning if user tries to close editor
- ✅ All existing tests pass
- ✅ New tests cover editing functionality

## Implementation Notes

### Frontmatter Schema Constraints

From the taskmd specification, enforce these rules:

**Required fields:**
- `id` (string) - Unique identifier
- `title` (string) - Task title
- `status` (enum) - One of: pending, in-progress, completed, blocked, cancelled

**Optional fields:**
- `priority` (enum) - One of: low, medium, high, critical
- `effort` (enum) - One of: small, medium, large
- `dependencies` (array of strings) - Task IDs
- `tags` (array of strings) - Lowercase, hyphenated tags
- `group` (string) - Logical grouping
- `created` (date) - ISO 8601 format (YYYY-MM-DD)
- `description` (string) - Brief description

**Validation rules:**
- No duplicate IDs across all tasks
- All dependency IDs must reference existing tasks
- No circular dependencies
- Dates must be valid ISO 8601 format
- Tags should be lowercase and hyphenated

### API Design

**Endpoint:** `PUT /api/tasks/:id`

**Request body:**
```json
{
  "frontmatter": {
    "id": "062",
    "title": "Updated title",
    "status": "in-progress",
    "priority": "high",
    "effort": "large",
    "dependencies": ["001", "048"],
    "tags": ["web", "feature"],
    "created": "2026-02-12"
  },
  "body": "# Updated markdown body\n\n..."
}
```

**Response (success):**
```json
{
  "success": true,
  "task": { ...updated task object... }
}
```

**Response (validation error):**
```json
{
  "success": false,
  "errors": [
    {
      "field": "status",
      "message": "Invalid status value. Must be one of: pending, in-progress, completed, blocked, cancelled"
    }
  ]
}
```

### File Writing Strategy

1. **Read current file** - Get current contents
2. **Parse frontmatter and body** - Extract both parts
3. **Merge changes** - Apply only changed fields
4. **Validate** - Run full validation
5. **Write atomically** - Use temp file + rename for safety
6. **Handle errors** - Return clear error messages

### UI/UX Considerations

- **Keyboard shortcuts** - Ctrl/Cmd+S to save, Esc to cancel
- **Autosave draft** - Save to localStorage to prevent data loss
- **Markdown preview** - Toggle between edit and preview mode
- **Field hints** - Show valid values for enum fields
- **Inline validation** - Real-time feedback as user types
- **Confirm discard** - Warn before losing unsaved changes
- **Loading states** - Clear feedback during save operation
- **Success feedback** - Confirmation that save succeeded

### Technical Architecture

**Frontend:**
- React components with TypeScript
- Form state management (React Hook Form or similar)
- Validation with Zod or Yup
- API calls with fetch or axios
- Optimistic UI updates

**Backend:**
- Express.js endpoint
- Reuse internal/validator from CLI
- File I/O with fs/promises
- YAML parsing with js-yaml
- Proper error handling and status codes

### Security Considerations

- **Path traversal** - Validate file paths to prevent directory traversal
- **Input sanitization** - Sanitize all user input
- **File permissions** - Ensure proper file system permissions
- **Rate limiting** - Prevent abuse of edit endpoint
- **Authentication** - Consider adding auth if deploying publicly

## Examples

### Basic Edit Flow

```typescript
// User clicks "Edit" button on a task
const handleEdit = (taskId: string) => {
  const task = tasks.find(t => t.id === taskId);
  setEditingTask(task);
  setEditorOpen(true);
};

// User modifies status and saves
const handleSave = async (updates: TaskUpdate) => {
  try {
    setLoading(true);
    const response = await fetch(`/api/tasks/${task.id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(updates)
    });

    if (!response.ok) {
      const errors = await response.json();
      setErrors(errors);
      return;
    }

    const updatedTask = await response.json();
    updateTaskInCache(updatedTask);
    setEditorOpen(false);
    showToast('Task updated successfully');
  } catch (error) {
    showToast('Failed to save task', 'error');
  } finally {
    setLoading(false);
  }
};
```

### Validation Example

```typescript
const validateTask = (task: TaskFormData): ValidationError[] => {
  const errors: ValidationError[] = [];

  // Required fields
  if (!task.id) errors.push({ field: 'id', message: 'ID is required' });
  if (!task.title) errors.push({ field: 'title', message: 'Title is required' });
  if (!task.status) errors.push({ field: 'status', message: 'Status is required' });

  // Enum validation
  const validStatuses = ['pending', 'in-progress', 'completed', 'blocked', 'cancelled'];
  if (task.status && !validStatuses.includes(task.status)) {
    errors.push({ field: 'status', message: 'Invalid status value' });
  }

  // Dependencies validation
  if (task.dependencies) {
    const invalidDeps = task.dependencies.filter(id => !taskExists(id));
    if (invalidDeps.length > 0) {
      errors.push({
        field: 'dependencies',
        message: `Invalid dependencies: ${invalidDeps.join(', ')}`
      });
    }
  }

  return errors;
};
```

## References

- Task specification: `docs/taskmd_specification.md`
- Existing validator: `apps/cli/internal/validator/validator.go`
- Web app structure: `apps/web/src/`
- API server: `apps/web/src/api/`

## Related Tasks

- Task 048: Graph view (editing from graph nodes)
- Task 024: Enhanced task filtering (editing filtered results)
- Future: Real-time collaboration
- Future: Version history / undo

## Notes

- Consider adding a "Quick Edit" mode for just changing status/priority without full editor
- May want to add markdown editor toolbar for common formatting
- Consider adding templates for common task structures
- Think about bulk edit operations in the future
- Mobile editing experience should be touch-friendly
