import { describe, it, expect } from "vitest";
import type { Task } from "../../../api/types.ts";
import { applyFilters, defaultFilterState } from "./filters.ts";

function makeTask(overrides: Partial<Task> = {}): Task {
  return {
    id: "001",
    title: "Test task",
    status: "pending",
    priority: "medium",
    effort: "small",
    type: "feature",
    dependencies: null,
    tags: null,
    phase: "",
    group: "",
    owner: "",
    parent: "",
    created: "2026-01-01",
    body: "",
    file_path: "tasks/001-test.md",
    ...overrides,
  };
}

const sampleTasks: Task[] = [
  makeTask({ id: "001", status: "pending", priority: "high", type: "feature", tags: ["api"], effort: "small" }),
  makeTask({ id: "002", status: "in-progress", priority: "medium", type: "bug", tags: ["web", "api"], effort: "medium" }),
  makeTask({ id: "003", status: "completed", priority: "low", type: "chore", tags: ["docs"], effort: "large" }),
  makeTask({ id: "004", status: "blocked", priority: "critical", type: "feature", tags: null, effort: "" }),
  makeTask({ id: "005", status: "pending", priority: "", type: "", tags: null, effort: "" }),
];

describe("applyFilters", () => {
  it("returns all tasks with default filter state", () => {
    const result = applyFilters(sampleTasks, defaultFilterState());
    expect(result).toHaveLength(sampleTasks.length);
  });

  it("filters by single status", () => {
    const filters = { ...defaultFilterState(), selectedStatuses: new Set(["pending"]) };
    const result = applyFilters(sampleTasks, filters);
    expect(result.map((t) => t.id)).toEqual(["001", "005"]);
  });

  it("filters by in-review status", () => {
    const tasksWithReview = [
      ...sampleTasks,
      makeTask({ id: "006", status: "in-review", priority: "medium", type: "feature" }),
    ];
    const filters = { ...defaultFilterState(), selectedStatuses: new Set(["in-review"]) };
    const result = applyFilters(tasksWithReview, filters);
    expect(result.map((t) => t.id)).toEqual(["006"]);
  });

  it("filters by multiple statuses", () => {
    const filters = { ...defaultFilterState(), selectedStatuses: new Set(["pending", "blocked"]) };
    const result = applyFilters(sampleTasks, filters);
    expect(result.map((t) => t.id)).toEqual(["001", "004", "005"]);
  });

  it("filters by priority", () => {
    const filters = { ...defaultFilterState(), selectedPriorities: new Set(["high"]) };
    const result = applyFilters(sampleTasks, filters);
    expect(result.map((t) => t.id)).toEqual(["001", "005"]);
  });

  it("filters by type", () => {
    const filters = { ...defaultFilterState(), selectedTypes: new Set(["bug"]) };
    const result = applyFilters(sampleTasks, filters);
    expect(result.map((t) => t.id)).toEqual(["002", "005"]);
  });

  it("filters by tags (OR among selected tags)", () => {
    const filters = { ...defaultFilterState(), selectedTags: new Set(["api"]) };
    const result = applyFilters(sampleTasks, filters);
    expect(result.map((t) => t.id)).toEqual(["001", "002"]);
  });

  it("filters by multiple tags (OR logic)", () => {
    const filters = { ...defaultFilterState(), selectedTags: new Set(["docs", "web"]) };
    const result = applyFilters(sampleTasks, filters);
    expect(result.map((t) => t.id)).toEqual(["002", "003"]);
  });

  it("filters by effort", () => {
    const filters = { ...defaultFilterState(), selectedEffort: new Set(["small"]) };
    const result = applyFilters(sampleTasks, filters);
    expect(result.map((t) => t.id)).toEqual(["001"]);
  });

  it("treats a full custom vocabulary as no effort filter", () => {
    const efforts = ["xs", "s", "m", "l", "xl"];
    const custom = [
      makeTask({ id: "101", effort: "xs" }),
      makeTask({ id: "102", effort: "xl" }),
    ];
    const filters = { ...defaultFilterState(efforts), selectedEffort: new Set(efforts) };
    const result = applyFilters(custom, filters, efforts);
    expect(result.map((t) => t.id)).toEqual(["101", "102"]);
  });

  it("filters by a value from a custom vocabulary", () => {
    const efforts = ["xs", "s", "m", "l", "xl"];
    const custom = [
      makeTask({ id: "101", effort: "xs" }),
      makeTask({ id: "102", effort: "xl" }),
    ];
    const filters = { ...defaultFilterState(efforts), selectedEffort: new Set(["xl"]) };
    const result = applyFilters(custom, filters, efforts);
    expect(result.map((t) => t.id)).toEqual(["102"]);
  });

  it("applies intersection of multiple filters (status AND priority AND type)", () => {
    const filters = {
      ...defaultFilterState(),
      selectedStatuses: new Set(["pending", "in-progress"]),
      selectedPriorities: new Set(["medium", "high"]),
      selectedTypes: new Set(["feature", "bug"]),
    };
    const result = applyFilters(sampleTasks, filters);
    expect(result.map((t) => t.id)).toEqual(["001", "002", "005"]);
  });

  it("applies intersection of all filter criteria", () => {
    const filters = {
      selectedStatuses: new Set(["pending", "in-progress"]),
      selectedPriorities: new Set(["high", "medium"]),
      selectedTypes: new Set(["feature", "bug"]),
      selectedTags: new Set(["api"]),
      selectedEffort: new Set(["small"]),
      selectedPhases: new Set<string>(),
      globalFilter: "",
    };
    const result = applyFilters(sampleTasks, filters);
    expect(result.map((t) => t.id)).toEqual(["001"]);
  });

  it("filters by selected phases", () => {
    const tasksWithPhases = [
      makeTask({ id: "001", phase: "mvp" }),
      makeTask({ id: "002", phase: "v2" }),
      makeTask({ id: "003", phase: "" }),
    ];
    const filters = { ...defaultFilterState(), selectedPhases: new Set(["mvp"]) };
    const result = applyFilters(tasksWithPhases, filters);
    expect(result.map((t) => t.id)).toEqual(["001"]);
  });

  it("excludes tasks without phase when phase filter is active", () => {
    const tasksWithPhases = [
      makeTask({ id: "001", phase: "mvp" }),
      makeTask({ id: "002", phase: "" }),
    ];
    const filters = { ...defaultFilterState(), selectedPhases: new Set(["mvp"]) };
    const result = applyFilters(tasksWithPhases, filters);
    expect(result.map((t) => t.id)).toEqual(["001"]);
  });

  it("returns empty array when no tasks match", () => {
    const filters = { ...defaultFilterState(), selectedStatuses: new Set(["cancelled"]) };
    const result = applyFilters(sampleTasks, filters);
    expect(result).toEqual([]);
  });

  it("tasks without priority pass through priority filter", () => {
    const filters = { ...defaultFilterState(), selectedPriorities: new Set(["critical"]) };
    const result = applyFilters(sampleTasks, filters);
    expect(result.map((t) => t.id)).toEqual(["004", "005"]);
  });

  it("tasks without type pass through type filter", () => {
    const filters = { ...defaultFilterState(), selectedTypes: new Set(["chore"]) };
    const result = applyFilters(sampleTasks, filters);
    expect(result.map((t) => t.id)).toEqual(["003", "005"]);
  });

  it("tasks without tags are excluded when tag filter is active", () => {
    const filters = { ...defaultFilterState(), selectedTags: new Set(["api"]) };
    const result = applyFilters(sampleTasks, filters);
    expect(result.every((t) => t.tags !== null)).toBe(true);
  });

  it("tasks without effort are excluded when effort filter is active", () => {
    const filters = { ...defaultFilterState(), selectedEffort: new Set(["medium"]) };
    const result = applyFilters(sampleTasks, filters);
    expect(result.map((t) => t.id)).toEqual(["002"]);
  });
});

describe("applyFilters worktree overlay", () => {
  it("matches the status filter against the effective status", () => {
    const tasks = [
      makeTask({ id: "010", status: "pending", effective_status: "in-progress" }),
      makeTask({ id: "011", status: "pending" }),
    ];
    const filters = { ...defaultFilterState(), selectedStatuses: new Set(["in-progress"]) };
    expect(applyFilters(tasks, filters).map((t) => t.id)).toEqual(["010"]);
  });

  it("excludes a task whose local status matches but effective status does not", () => {
    const tasks = [makeTask({ id: "010", status: "pending", effective_status: "completed" })];
    const filters = { ...defaultFilterState(), selectedStatuses: new Set(["pending"]) };
    expect(applyFilters(tasks, filters)).toEqual([]);
  });
});
