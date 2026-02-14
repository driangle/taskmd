# Why taskmd?

Common questions about why taskmd exists, how it fits alongside other tools, and why it's built the way it is.

## The Core Premise

### Q: Why plain markdown files instead of a database or SaaS tool?

**A:** Because files are the lowest common denominator. Every editor can open them. Every AI assistant can read and write them. Every version control system can track them. There's no server to run, no account to create, no sync to configure, no subscription to manage. Your tasks are just files in a directory — portable, inspectable, and yours. When you `mkdir tasks/` and create a `.md` file, you're done. That's the entire setup.

### Q: Why build another task management tool when Jira, Linear, GitHub Issues, etc. already exist?

**A:** Those tools are designed for teams coordinating through a web interface. taskmd is designed for developers coordinating with AI. When you're pair-programming with an AI assistant, it can't click through Jira — but it can read a file that's right there in the repo. taskmd optimizes for the workflow where your task context and your code context are the same thing: files in a repository. It's not a replacement for your team's project management tool. It's a development-time companion that keeps your working backlog where your work actually happens.

### Q: Why should tasks live inside my code repository?

**A:** Because tasks and code evolve together. When you branch to build a feature, the tasks for that feature travel with the branch. When you open a PR, reviewers can see what tasks were completed alongside the code changes. When you roll back a release, the task state rolls back too. There's no drift between "what the tracker says" and "what the code does." Your Git history becomes a complete record of both *what changed* and *why*.

## AI Coding Assistants

### Q: Why does the task format matter for AI coding assistants?

**A:** AI assistants like Claude Code, Codex, Cursor, and Windsurf already know how to read and write markdown files — it's one of their most natural capabilities. A task file sitting in your repo is instantly accessible to any AI tool with zero integration work. No API tokens, no MCP servers, no OAuth flows, no plugins. The AI reads the task, understands the context (because the code is right there too), does the work, and updates the task status. The file *is* the integration layer.

### Q: Won't AI assistants just get better at integrating with tools like Jira or Linear via APIs and MCPs?

**A:** They will, and that's great. But there's always a cost to indirection. An API call to fetch a Jira ticket requires authentication, network access, error handling, and a translation step between the external format and the local context. A file in the repo requires none of that. It's already loaded, already in context, already diffable. As AI tools improve, they'll integrate with everything — but the simplest, fastest, most reliable integration will always be "read the file that's right here."

### Q: Why not let the AI assistant manage tasks in its own memory or context?

**A:** Because context windows reset between sessions. When you close your terminal and come back tomorrow, the AI's internal state is gone. When a different team member picks up the work, they start with a blank slate. When you switch between AI tools, nothing carries over. Task files persist across sessions, across tools, and across people. They're the durable layer that gives continuity to an inherently ephemeral interaction model.

### Q: How is this different from built-in AI task features like Claude Code's TodoWrite or Cursor's task tracking?

**A:** Those features are session-scoped and tool-specific. They help the AI organize its thinking during a single work session, but the data lives inside the tool and disappears when the session ends. taskmd is the opposite: tool-agnostic, persistent, and human-owned. Any AI assistant can read and write taskmd files. The tasks survive across sessions, across tools, and across time. You're building a durable backlog, not a temporary scratchpad.

## The Format

### Q: Why YAML frontmatter + markdown body instead of pure JSON, TOML, or a custom format?

**A:** It's the best of both worlds. The YAML frontmatter gives you structured, machine-parseable metadata that tools can sort, filter, graph, and validate. The markdown body gives you freeform space for objectives, acceptance criteria, subtasks, and notes — the kind of rich context that doesn't fit into fields. Developers already know both formats. AI assistants parse both natively. Git diffs both cleanly. And every text editor on earth highlights both correctly.

### Q: Why have a spec at all? Why not just freeform markdown notes?

**A:** Freeform notes are great for thinking. They're terrible for tooling. Without a consistent structure, you can't sort tasks by priority, filter by status, visualize dependencies, validate references, or suggest what to work on next. The spec is deliberately minimal — three required fields — so you get the power of structured data without the burden of filling out forms. The structure enables the CLI, the web UI, the graph visualization, and the AI workflow. The freeform body is still right there for everything else.

## The Philosophy

### Q: Why local-first with no server dependency?

**A:** Because the fastest, most reliable service is no service. taskmd works on an airplane, in a coffee shop with bad wifi, and on a CI runner with no internet access. There are no accounts, no permissions, no outages, no breaking API changes, no pricing tiers. The cost of running taskmd is zero, forever. And if you stop using it, your tasks are still perfectly readable markdown files — not data locked in a proprietary format behind a defunct API.

### Q: Why should developers own their task data instead of a platform?

**A:** Because platforms come and go, but plain text endures. Your task files will be readable in 20 years with nothing more than a text editor. They're not tied to a vendor's business model, pricing changes, or acquisition. You can grep them, script them, pipe them, back them up, and move them anywhere. When your data is just files, you're never locked in — and you're never locked out.

### Q: If AI tools keep getting better, won't they eventually make task management unnecessary?

**A:** Better AI makes structured task data *more* valuable, not less. A capable AI assistant with access to a well-organized backlog can autonomously pick up the next task, understand its dependencies, do the work, and mark it complete. Without that structured backlog, even the best AI has to start every session with "so, what are we working on?" The clearer and more structured your task data, the more an AI can do with it — and the less you have to babysit the process. taskmd isn't overhead that AI will eliminate. It's the interface through which AI becomes genuinely autonomous.
