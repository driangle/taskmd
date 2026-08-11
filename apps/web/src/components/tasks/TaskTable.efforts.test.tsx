import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import type { Task } from "../../api/types.ts";

vi.mock("../../hooks/use-config.ts", () => ({
  useConfig: () => ({
    readonly: false,
    version: "1.0",
    phases: [],
    efforts: ["xs", "s", "m", "l", "xl"],
  }),
}));

import { TaskTable } from "./TaskTable.tsx";

const CUSTOM = ["xs", "s", "m", "l", "xl"];

function makeTask(overrides: Partial<Task> = {}): Task {
  return {
    id: "001",
    title: "Test task",
    status: "pending",
    priority: "medium",
    effort: "s",
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

const tasks: Task[] = [
  makeTask({ id: "001", effort: "xs" }),
  makeTask({ id: "002", effort: "l" }),
  makeTask({ id: "003", effort: "xl" }),
];

function renderTable(ui: React.ReactElement) {
  return render(<MemoryRouter>{ui}</MemoryRouter>);
}

function openFilters() {
  fireEvent.click(screen.getByRole("button", { name: /Filters/i }));
}

describe("TaskTable with a custom effort vocabulary", () => {
  it("offers exactly the configured effort values as filters", () => {
    renderTable(<TaskTable tasks={tasks} />);
    openFilters();
    for (const value of CUSTOM) {
      expect(screen.getByRole("button", { name: value })).toBeInTheDocument();
    }
  });

  // "medium" is excluded: it is also a priority pill, so its presence says
  // nothing about the effort row.
  it("does not offer the built-in vocabulary", () => {
    renderTable(<TaskTable tasks={tasks} />);
    openFilters();
    for (const value of ["small", "large"]) {
      expect(screen.queryByRole("button", { name: value })).not.toBeInTheDocument();
    }
  });

  it("shows all tasks before any effort filter is touched", () => {
    renderTable(<TaskTable tasks={tasks} />);
    expect(screen.getByText("Showing 3 of 3 tasks")).toBeInTheDocument();
  });

  it("filters by a configured effort value", () => {
    renderTable(<TaskTable tasks={tasks} />);
    openFilters();
    // Deselect everything but "xs".
    for (const value of ["s", "m", "l", "xl"]) {
      fireEvent.click(screen.getByRole("button", { name: value }));
    }
    expect(screen.getByText("Showing 1 of 3 tasks")).toBeInTheDocument();
  });

  it("initializes the effort filter from props", () => {
    renderTable(<TaskTable tasks={tasks} initialEffort={["l", "xl"]} />);
    expect(screen.getByText("Showing 2 of 3 tasks")).toBeInTheDocument();
  });

  it("styles active effort pills inline, since Tailwind cannot class custom values", () => {
    renderTable(<TaskTable tasks={tasks} />);
    openFilters();
    const pill = screen.getByRole("button", { name: "xs" });
    expect(pill.getAttribute("style")).toContain("background-color");
  });
});
