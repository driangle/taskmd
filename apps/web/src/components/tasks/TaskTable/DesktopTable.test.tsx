import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import type { Table, Row } from "@tanstack/react-table";
import { DesktopTable } from "./DesktopTable.tsx";
import type { Task } from "../../../api/types.ts";
import { createTask, resetFixtureCounter } from "../../../test-utils/fixtures.ts";

const mockNavigate = vi.fn();
vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

function createMockTable(
  headers: { id: string; header: string; isSorted: false | "asc" | "desc"; className?: string }[],
) {
  return {
    getHeaderGroups: () => [
      {
        id: "hg-0",
        headers: headers.map((h) => ({
          id: h.id,
          column: {
            getToggleSortingHandler: () => vi.fn(),
            getIsSorted: () => h.isSorted,
            columnDef: {
              header: h.header,
              meta: h.className ? { className: h.className } : undefined,
            },
          },
          getContext: () => ({}),
        })),
      },
    ],
  } as unknown as Table<Task>;
}

function createMockRow(task: Task, cells: { id: string; value: string }[] = []) {
  return {
    id: `row-${task.id}`,
    original: task,
    getVisibleCells: () =>
      cells.map((c) => ({
        id: c.id,
        column: { columnDef: { cell: () => c.value, meta: undefined } },
        getContext: () => ({}),
      })),
  } as unknown as Row<Task>;
}

const defaultHeaders = [
  { id: "id", header: "ID", isSorted: false as const },
  { id: "title", header: "Title", isSorted: false as const },
];

function renderDesktopTable({
  rows = [] as Row<Task>[],
  table = createMockTable(defaultHeaders),
  columns = { length: defaultHeaders.length },
  clearFilters = vi.fn(),
} = {}) {
  return render(
    <MemoryRouter>
      <DesktopTable
        table={table}
        rows={rows}
        columns={columns}
        clearFilters={clearFilters}
      />
    </MemoryRouter>,
  );
}

describe("DesktopTable", () => {
  beforeEach(() => {
    resetFixtureCounter();
    mockNavigate.mockClear();
  });

  it("shows empty state message when rows is empty", () => {
    renderDesktopTable({ rows: [] });
    expect(screen.getByText("No tasks match your filters.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Clear filters" })).toBeInTheDocument();
  });

  it("calls clearFilters when Clear filters button is clicked", () => {
    const clearFilters = vi.fn();
    renderDesktopTable({ rows: [], clearFilters });
    fireEvent.click(screen.getByRole("button", { name: "Clear filters" }));
    expect(clearFilters).toHaveBeenCalledOnce();
  });

  it("renders task cell data when rows are provided", () => {
    const task = createTask({ id: "042", title: "My Task" });
    const row = createMockRow(task, [
      { id: "col-id", value: "042" },
      { id: "col-title", value: "My Task" },
    ]);
    renderDesktopTable({ rows: [row] });
    expect(screen.getByText("042")).toBeInTheDocument();
    expect(screen.getByText("My Task")).toBeInTheDocument();
  });

  it("navigates to task detail when onActivate is triggered via Enter key", () => {
    const task = createTask({ id: "007" });
    const row = createMockRow(task, [{ id: "col-id", value: "007" }]);
    renderDesktopTable({ rows: [row] });

    const grid = screen.getByRole("grid");
    // Focus first item via ArrowDown, then activate with Enter
    fireEvent.keyDown(grid, { key: "ArrowDown" });
    fireEvent.keyDown(grid, { key: "Enter" });
    expect(mockNavigate).toHaveBeenCalledWith("/tasks/007");
  });

  it("marks an ascending-sorted column header with aria-sort", () => {
    const headers = [{ id: "id", header: "ID", isSorted: "asc" as const }];
    renderDesktopTable({ table: createMockTable(headers), columns: { length: 1 } });
    expect(screen.getByRole("columnheader")).toHaveAttribute("aria-sort", "ascending");
  });

  it("marks a descending-sorted column header with aria-sort", () => {
    const headers = [{ id: "id", header: "ID", isSorted: "desc" as const }];
    renderDesktopTable({ table: createMockTable(headers), columns: { length: 1 } });
    expect(screen.getByRole("columnheader")).toHaveAttribute("aria-sort", "descending");
  });

  it("marks an unsorted column header with aria-sort none", () => {
    const headers = [{ id: "id", header: "ID", isSorted: false as const }];
    renderDesktopTable({ table: createMockTable(headers), columns: { length: 1 } });
    expect(screen.getByRole("columnheader")).toHaveAttribute("aria-sort", "none");
  });

  it("marks the focused row with aria-selected", () => {
    const task = createTask();
    const row = createMockRow(task, [{ id: "col-id", value: task.id }]);
    renderDesktopTable({ rows: [row] });

    const grid = screen.getByRole("grid");
    fireEvent.keyDown(grid, { key: "ArrowDown" });

    const taskRow = document.getElementById("task-row-0");
    expect(taskRow).toHaveAttribute("aria-selected", "true");
  });
});
