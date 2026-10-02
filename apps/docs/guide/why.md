# Why taskmd?

Common questions about why taskmd exists, how it fits alongside other tools, and why it's built the way it is.

## The Core Premise

### Why plain markdown files instead of a database or SaaS tool?

Files are the lowest common denominator. Every editor can open them, every AI assistant can read and write them, and every version control system can track them. There's no server to run, no account to create, and no sync to configure. Your tasks are files in a directory — portable, inspectable, and under your control.

### Why build another task management tool when Jira, Linear, GitHub Issues, etc. already exist?

Those tools are designed for teams coordinating through a web interface. taskmd is designed for developers working with AI assistants. An AI assistant can't click through Jira, but it can read a file in the repo. taskmd keeps your working backlog where your work happens — in the repository. It's not a replacement for your team's project management tool; it's a development-time companion.

### Why store tasks in files at all? Why not just tell the AI what to do directly?

Telling an AI "build me a login page" works for one-shot tasks. But real projects have dozens of interrelated tasks with dependencies, priorities, and acceptance criteria that evolve over days or weeks. Without a written backlog, you're carrying all of that context in your head — and re-explaining it every session. Task files give the AI a persistent, structured view of what needs doing, what's done, and what's blocked. You stop being the bottleneck for context and start directing work at a higher level.

### Why should tasks live inside my code repository?

Tasks and code evolve together. When you branch to build a feature, the tasks for that feature travel with the branch. When you open a PR, reviewers can see what tasks were completed alongside the code changes. When you roll back a release, the task state rolls back too. Git history becomes a record of both *what changed* and *why*.

## AI Coding Assistants

### Why does the task format matter for AI coding assistants?

AI assistants like Claude Code, Codex, Cursor, and Windsurf already know how to read and write markdown. A task file in your repo is accessible to any AI tool without integration work — no API tokens, no MCP servers, no OAuth flows, no plugins. The AI reads the task, sees the surrounding code for context, does the work, and updates the task status. The file itself is the integration layer.

### Won't AI assistants just get better at integrating with tools like Jira or Linear via APIs and MCPs?

They will. But there's a cost to indirection: an API call to fetch a Jira ticket requires authentication, network access, error handling, and a translation step between the external format and local context. A file in the repo is already loaded, already in context, and already diffable. As AI tools improve, they'll integrate with more services — but reading a local file will remain the simplest path.

### Why not let the AI assistant manage tasks in its own memory or context?

Context windows reset between sessions. When you close your terminal and come back later, the AI's internal state is gone. When a different team member picks up the work, they start from scratch. When you switch between AI tools, nothing carries over. Task files persist across sessions, tools, and people. They provide continuity to an otherwise ephemeral interaction model.

### How is this different from built-in AI task features like Claude Code's TodoWrite or Cursor's task tracking?

Those features are session-scoped and tool-specific. They help the AI organize its work during a single session, but the data disappears when the session ends. taskmd files are tool-agnostic, persistent, and human-owned. Any AI assistant can read and write them, and the tasks survive across sessions and tools.

### How does taskmd compare to Claude Code's task list?

Claude Code has a built-in task list that it creates during a session to track its own progress on multi-step work. It's useful — you can see what's pending, in progress, or complete while Claude works. But it's a session-level tool for the AI to organize itself, not a project-level tool for you to organize your backlog.

The key differences: Claude Code's tasks are created by the AI during a session and live in Claude's internal state. They can persist across context compactions within a session, but they're not files in your repo — they don't travel with branches, show up in PRs, or carry structured metadata like priority and dependencies. A different AI tool can't read them. A teammate can't pick them up. They're an execution aid, not a planning system.

taskmd tasks are files you own. They exist before the AI session starts and after it ends. They define *what* to build and *why*, with acceptance criteria the AI can verify against. Claude Code's task list tracks *how* the AI is executing right now. They complement each other — taskmd defines the work, and Claude Code's task list helps it stay organized while doing it.

## The Format

### Why YAML frontmatter + markdown body instead of pure JSON, TOML, or a custom format?

The YAML frontmatter provides structured, machine-parseable metadata that tools can sort, filter, graph, and validate. The markdown body provides freeform space for objectives, acceptance criteria, subtasks, and notes — context that doesn't fit into fields. Developers are familiar with both formats, AI assistants parse both natively, Git diffs both cleanly, and text editors highlight both correctly.

### Why have a spec at all? Why not just freeform markdown notes?

Freeform notes work well for thinking, but they're hard to build tooling around. Without consistent structure, you can't sort tasks by priority, filter by status, visualize dependencies, or validate references. The spec is deliberately minimal — two required fields, `id` and `title` — so you get the benefits of structured data without the overhead of filling out forms. The freeform body is still available for everything else.

## The Philosophy

### Why local-first with no server dependency?

taskmd works offline, on CI runners, and in environments with limited connectivity. There are no accounts, permissions, or external service dependencies. If you stop using it, your tasks remain readable markdown files — not data in a proprietary format behind an API.

### Why should developers own their task data instead of a platform?

Plain text is durable. Your task files will be readable in 20 years with nothing more than a text editor. They're not tied to a vendor's business model or pricing changes. You can grep them, script them, pipe them, back them up, and move them anywhere.

### If AI tools keep getting better, won't they eventually make task management unnecessary?

More capable AI makes structured task data more valuable. An AI assistant with access to a well-organized backlog can pick up the next task, understand its dependencies, do the work, and mark it complete. Without that structured backlog, it has to start every session asking what to work on. The clearer and more structured your task data, the more an AI can do with it autonomously.
