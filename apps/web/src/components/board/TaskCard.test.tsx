import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { TaskCard } from "./TaskCard.tsx";
import type { BoardTask } from "../../api/types.ts";

function makeTask(overrides: Partial<BoardTask> = {}): BoardTask {
  return {
    id: "042",
    title: "Implement feature X",
    status: "pending",
    priority: "high",
    phase: "mvp",
    ...overrides,
  };
}

function renderCard(props: Partial<React.ComponentProps<typeof TaskCard>> = {}) {
  const defaults = {
    task: makeTask(),
    sourceGroup: "pending",
    canDrag: true,
  };
  return render(
    <MemoryRouter>
      <TaskCard {...defaults} {...props} />
    </MemoryRouter>,
  );
}

describe("TaskCard", () => {
  it("renders task title as a link", () => {
    renderCard();
    expect(screen.getByRole("link", { name: "Implement feature X" })).toHaveAttribute(
      "href",
      "/tasks/042",
    );
  });

  it("renders task id", () => {
    renderCard();
    expect(screen.getByText("042")).toBeInTheDocument();
  });

  it("renders priority when present", () => {
    renderCard();
    expect(screen.getByText("high")).toBeInTheDocument();
  });

  it("does not render priority when absent", () => {
    renderCard({ task: makeTask({ priority: undefined }) });
    expect(screen.queryByText("high")).not.toBeInTheDocument();
  });

  it("sets draggable=true when canDrag is true", () => {
    const { container } = renderCard({ canDrag: true });
    const card = container.firstElementChild!;
    expect(card).toHaveAttribute("draggable", "true");
  });

  it("sets draggable=false when canDrag is false", () => {
    const { container } = renderCard({ canDrag: false });
    const card = container.firstElementChild!;
    expect(card).toHaveAttribute("draggable", "false");
  });

  describe("drag events", () => {
    it("dragStart sets data and effectAllowed", () => {
      const { container } = renderCard();
      const card = container.firstElementChild!;

      const store: Record<string, string> = {};
      const dragStartEvent = new Event("dragstart", { bubbles: true });
      Object.defineProperty(dragStartEvent, "dataTransfer", {
        value: {
          setData: (key: string, val: string) => { store[key] = val; },
          effectAllowed: "",
        },
      });

      fireEvent(card, dragStartEvent);

      expect(store["text/plain"]).toBe("042");
      expect(store["application/x-source-group"]).toBe("pending");
    });

    it("dragStart adds document-level listeners and dragEnd removes them", () => {
      const addSpy = vi.spyOn(document, "addEventListener");
      const removeSpy = vi.spyOn(document, "removeEventListener");

      const { container } = renderCard();
      const card = container.firstElementChild!;

      // dragStart
      const startEvent = new Event("dragstart", { bubbles: true });
      Object.defineProperty(startEvent, "dataTransfer", {
        value: {
          setData: vi.fn(),
          effectAllowed: "",
        },
      });
      fireEvent(card, startEvent);

      expect(addSpy).toHaveBeenCalledWith("dragover", expect.any(Function));
      expect(addSpy).toHaveBeenCalledWith("drop", expect.any(Function));

      // dragEnd
      fireEvent.dragEnd(card);

      expect(removeSpy).toHaveBeenCalledWith("dragover", expect.any(Function));
      expect(removeSpy).toHaveBeenCalledWith("drop", expect.any(Function));

      addSpy.mockRestore();
      removeSpy.mockRestore();
    });

    it("marks the card as dragging while a drag is in progress", () => {
      const { container } = renderCard();
      const card = container.firstElementChild!;

      expect(card).toHaveAttribute("data-dragging", "false");

      const startEvent = new Event("dragstart", { bubbles: true });
      Object.defineProperty(startEvent, "dataTransfer", {
        value: { setData: vi.fn(), effectAllowed: "" },
      });
      fireEvent(card, startEvent);

      expect(card).toHaveAttribute("data-dragging", "true");

      fireEvent.dragEnd(card);
      expect(card).toHaveAttribute("data-dragging", "false");
    });

    it("cleans up document listeners on unmount while dragging", () => {
      const removeSpy = vi.spyOn(document, "removeEventListener");

      const { container, unmount } = renderCard();
      const card = container.firstElementChild!;

      // Start dragging
      const startEvent = new Event("dragstart", { bubbles: true });
      Object.defineProperty(startEvent, "dataTransfer", {
        value: { setData: vi.fn(), effectAllowed: "" },
      });
      fireEvent(card, startEvent);

      removeSpy.mockClear();
      unmount();

      expect(removeSpy).toHaveBeenCalledWith("dragover", expect.any(Function));
      expect(removeSpy).toHaveBeenCalledWith("drop", expect.any(Function));

      removeSpy.mockRestore();
    });
  });

  it("marks the card as focused when focused", () => {
    const { container } = renderCard({ focused: true });
    const card = container.firstElementChild!;
    expect(card).toHaveAttribute("data-focused", "true");
  });

  it("is not marked focused when not focused", () => {
    const { container } = renderCard({ focused: false });
    const card = container.firstElementChild!;
    expect(card).toHaveAttribute("data-focused", "false");
  });
});
