import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import type { CellContext, Row } from "@tanstack/react-table";
import { createTaskColumns } from "./columns.tsx";
import type { Task } from "../../../api/types.ts";
import { createTask, resetFixtureCounter } from "../../../test-utils/fixtures.ts";

vi.mock("./Badges.tsx", () => ({
  StatusBadge: ({ status }: { status: string }) => <span data-testid="status-badge">{status}</span>,
  PriorityBadge: ({ priority }: { priority: string }) => <span data-testid="priority-badge">{priority}</span>,
  TypeBadge: ({ type }: { type: string }) => <span data-testid="type-badge">{type}</span>,
  PhaseBadge: ({ phase }: { phase: string }) => <span data-testid="phase-badge">{phase}</span>,
  BlockedStatusBadge: ({ dependencies }: { dependencies: string[] | null }) => (
    <span data-testid="blocked-badge">{dependencies?.join(",") ?? "none"}</span>
  ),
}));

function makeCellContext<K extends keyof Task>(task: Task, key: K): CellContext<Task, Task[K]> {
  return {
    getValue: () => task[key],
    row: { original: task } as Row<Task>,
    column: {} as never,
    cell: {} as never,
    table: {} as never,
    renderValue: () => task[key],
  } as unknown as CellContext<Task, Task[K]>;
}

function renderCell(node: React.ReactNode) {
  return render(<MemoryRouter>{node}</MemoryRouter>);
}

function getColumns(selectedTags = new Set<string>(), toggleTag = vi.fn()) {
  return createTaskColumns(selectedTags, toggleTag);
}

describe("columns tags", () => {
  beforeEach(() => {
    resetFixtureCounter();
  });

  it("renders tag buttons when tags are present", () => {
    const task = createTask({ tags: ["backend", "api"] });
    const cols = getColumns();
    const tagsCol = cols.find((c) => (c as { accessorKey?: string }).accessorKey === "tags")!;
    const cellFn = tagsCol.cell as (ctx: CellContext<Task, string[]>) => React.ReactNode;
    renderCell(cellFn(makeCellContext(task, "tags") as CellContext<Task, string[]>));
    expect(screen.getByRole("button", { name: "backend" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "api" })).toBeInTheDocument();
  });

  it("renders '-' when tags is null", () => {
    const task = createTask({ tags: null });
    const cols = getColumns();
    const tagsCol = cols.find((c) => (c as { accessorKey?: string }).accessorKey === "tags")!;
    const cellFn = tagsCol.cell as (ctx: CellContext<Task, string[] | null>) => React.ReactNode;
    renderCell(cellFn(makeCellContext(task, "tags") as CellContext<Task, string[] | null>));
    expect(screen.getByText("-")).toBeInTheDocument();
  });

  it("calls toggleTag when a tag button is clicked", () => {
    const toggleTag = vi.fn();
    const task = createTask({ tags: ["frontend"] });
    const cols = getColumns(new Set(), toggleTag);
    const tagsCol = cols.find((c) => (c as { accessorKey?: string }).accessorKey === "tags")!;
    const cellFn = tagsCol.cell as (ctx: CellContext<Task, string[]>) => React.ReactNode;
    renderCell(cellFn(makeCellContext(task, "tags") as CellContext<Task, string[]>));
    fireEvent.click(screen.getByRole("button", { name: "frontend" }));
    expect(toggleTag).toHaveBeenCalledWith("frontend");
  });

  it("marks the tag button as pressed when tag is in selectedTags", () => {
    const task = createTask({ tags: ["backend"] });
    const cols = getColumns(new Set(["backend"]));
    const tagsCol = cols.find((c) => (c as { accessorKey?: string }).accessorKey === "tags")!;
    const cellFn = tagsCol.cell as (ctx: CellContext<Task, string[]>) => React.ReactNode;
    renderCell(cellFn(makeCellContext(task, "tags") as CellContext<Task, string[]>));
    expect(screen.getByRole("button", { name: "backend" })).toHaveAttribute("aria-pressed", "true");
  });

  it("marks the tag button as not pressed when tag is not in selectedTags", () => {
    const task = createTask({ tags: ["backend"] });
    const cols = getColumns(new Set());
    const tagsCol = cols.find((c) => (c as { accessorKey?: string }).accessorKey === "tags")!;
    const cellFn = tagsCol.cell as (ctx: CellContext<Task, string[]>) => React.ReactNode;
    renderCell(cellFn(makeCellContext(task, "tags") as CellContext<Task, string[]>));
    expect(screen.getByRole("button", { name: "backend" })).toHaveAttribute("aria-pressed", "false");
  });
});
