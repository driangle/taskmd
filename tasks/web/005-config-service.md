---
id: "005"
title: "Config service (project path persistence)"
status: completed
priority: high
effort: small
dependencies:
  - "002"
tags:
  - core
  - config
created: 2026-02-08
---

# Config Service (Project Path Persistence)

## Objective

Build the config service that manages the application's persistent configuration — primarily the list of project folder paths the user has added. Config is stored in `~/.md-task-tracker/config.json`.

## Tasks

- [ ] Create `src/lib/config.ts`
- [ ] Define the config file path: `~/.md-task-tracker/config.json`
- [ ] Implement `loadConfig(): Promise<AppConfig>`
  - Read and parse the config file
  - If the file doesn't exist, return a default config (`{ projects: [], activeProjectId: null }`)
  - Validate the structure and handle corruption gracefully
- [ ] Implement `saveConfig(config: AppConfig): Promise<void>`
  - Write the config as formatted JSON
  - Create the `~/.md-task-tracker/` directory if it doesn't exist
- [ ] Implement `addProject(name: string, path: string): Promise<Project>`
  - Generate a unique project ID (e.g., nanoid or slugified name)
  - Validate that the path exists and is a directory
  - Add to the projects array and save
- [ ] Implement `removeProject(projectId: string): Promise<void>`
  - Remove from the projects array (does NOT delete the folder)
  - If the removed project was active, set `activeProjectId` to null
- [ ] Implement `setActiveProject(projectId: string): Promise<void>`
  - Set the `activeProjectId` in config

## Acceptance Criteria

- Config persists across server restarts
- Adding a project with a non-existent path returns an error
- Removing a project does not delete any files on disk
- First-run experience works (no config file → creates one)
- Config file is human-readable JSON
