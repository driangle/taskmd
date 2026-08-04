import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { BoardColumn } from "./BoardColumn.tsx";
import type { BoardGroup } from "../../api/types.ts";

function makeGroup(overrides: Partial<BoardGroup> = {}): BoardGroup {
  return {
    group: "pending",
    count: 2,
    tasks: [
      { id: "001", title: "First task", status: "pending", priority: "high" },
      { id: "002", title: "Second task", status: "pending" },
    ],
    ...overrides,
  };
}

function renderColumn(props: Partial<React.ComponentProps<typeof BoardColumn>> = {}) {
  const defaults = {
    group: makeGroup(),
    canDrag: true,
    onTaskDrop: vi.fn(),
  };
  return render(
    <MemoryRouter>
      <BoardColumn {...defaults} {...props} />
    </MemoryRouter>,
  );
}

function makeDragEvent(_type: string, data: Record<string, string> = {}) {
  const store: Record<string, string> = { ...data };
  return {
    preventDefault: vi.fn(),
    dataTransfer: {
      dropEffect: "",
      getData: (key: string) => store[key] ?? "",
      setData: (key: string, val: string) => { store[key] = val; },
    },
    currentTarget: null as unknown as EventTarget,
    relatedTarget: null as unknown as EventTarget,
  };
}

describe("BoardColumn", () => {
  it("renders group name and count", () => {
    renderColumn();
    expect(screen.getByText("pending")).toBeInTheDocument();
    expect(screen.getByText("(2)")).toBeInTheDocument();
  });

  it("renders task cards", () => {
    renderColumn();
    expect(screen.getByText("First task")).toBeInTheDocument();
    expect(screen.getByText("Second task")).toBeInTheDocument();
  });

  it("renders 'No tasks' when group is empty", () => {
    renderColumn({ group: makeGroup({ tasks: [], count: 0 }) });
    expect(screen.getByText("No tasks")).toBeInTheDocument();
  });

  it("labels the column with its group name", () => {
    renderColumn({
      group: makeGroup({
        group: "in-review",
        count: 1,
        tasks: [{ id: "001", title: "Review task", status: "in-review" }],
      }),
    });
    expect(screen.getByRole("group", { name: "in-review" })).toBeInTheDocument();
  });

  describe("drag handlers", () => {
    it("marks the column as an active drop target on dragOver", () => {
      renderColumn();
      const col = screen.getByRole("group");
      const event = makeDragEvent("dragover");
      event.currentTarget = col;
      fireEvent.dragOver(col, event);
      expect(col).toHaveAttribute("data-drop-active", "true");
    });

    it("handleDrop calls onTaskDrop when source !== target", () => {
      const onTaskDrop = vi.fn();
      renderColumn({ onTaskDrop });
      const col = screen.getByRole("group");

      // Simulate drop with different source group
      const dropEvent = new Event("drop", { bubbles: true }) as unknown as DragEvent;
      const store: Record<string, string> = {
        "text/plain": "001",
        "application/x-source-group": "in-progress",
      };
      Object.defineProperty(dropEvent, "dataTransfer", {
        value: {
          getData: (key: string) => store[key] ?? "",
        },
      });
      Object.defineProperty(dropEvent, "preventDefault", { value: vi.fn() });

      fireEvent(col, dropEvent);
      expect(onTaskDrop).toHaveBeenCalledWith("001", "in-progress", "pending");
    });

    it("handleDrop does NOT call onTaskDrop when source === target", () => {
      const onTaskDrop = vi.fn();
      renderColumn({ onTaskDrop });
      const col = screen.getByRole("group");

      const dropEvent = new Event("drop", { bubbles: true }) as unknown as DragEvent;
      const store: Record<string, string> = {
        "text/plain": "001",
        "application/x-source-group": "pending",
      };
      Object.defineProperty(dropEvent, "dataTransfer", {
        value: { getData: (key: string) => store[key] ?? "" },
      });
      Object.defineProperty(dropEvent, "preventDefault", { value: vi.fn() });

      fireEvent(col, dropEvent);
      expect(onTaskDrop).not.toHaveBeenCalled();
    });

    it("handleDragLeave removes highlight", () => {
      renderColumn();
      const col = screen.getByRole("group");

      // First trigger dragOver with a proper dataTransfer to set highlight
      const overEvent = new Event("dragover", { bubbles: true });
      Object.defineProperty(overEvent, "dataTransfer", {
        value: { dropEffect: "" },
      });
      Object.defineProperty(overEvent, "preventDefault", { value: vi.fn() });
      fireEvent(col, overEvent);
      expect(col).toHaveAttribute("data-drop-active", "true");

      // Then trigger dragLeave with relatedTarget outside column
      fireEvent.dragLeave(col, { relatedTarget: document.body });
      expect(col).toHaveAttribute("data-drop-active", "false");
    });
  });

  describe("canDrag=false", () => {
    it("does not call onTaskDrop on drop", () => {
      const onTaskDrop = vi.fn();
      renderColumn({ canDrag: false, onTaskDrop });
      const col = screen.getByRole("group");

      const dropEvent = new Event("drop", { bubbles: true }) as unknown as DragEvent;
      const store: Record<string, string> = {
        "text/plain": "001",
        "application/x-source-group": "in-progress",
      };
      Object.defineProperty(dropEvent, "dataTransfer", {
        value: { getData: (key: string) => store[key] ?? "" },
      });
      Object.defineProperty(dropEvent, "preventDefault", { value: vi.fn() });

      fireEvent(col, dropEvent);
      expect(onTaskDrop).not.toHaveBeenCalled();
    });

    it("does not become an active drop target on dragOver", () => {
      renderColumn({ canDrag: false });
      const col = screen.getByRole("group");
      fireEvent.dragOver(col);
      expect(col).toHaveAttribute("data-drop-active", "false");
    });
  });
});
