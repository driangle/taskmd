import { describe, it, expect, vi, type Mock } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { KeyboardList } from "./KeyboardList";

function renderList({
  itemCount = 3,
  onActivate = vi.fn(),
  role,
  ariaLabel,
}: {
  itemCount?: number;
  onActivate?: Mock<(index: number) => void>;
  role?: string;
  ariaLabel?: string;
} = {}) {
  render(
    <KeyboardList
      itemCount={itemCount}
      onActivate={onActivate}
      role={role}
      aria-label={ariaLabel}
    >
      {(idx) => <span data-testid="focused">{idx}</span>}
    </KeyboardList>,
  );
  const list = screen.getByRole(role ?? "list");
  return { list, onActivate };
}

describe("KeyboardList", () => {
  it("renders with correct role and aria-label", () => {
    renderList({ ariaLabel: "Task list" });
    const list = screen.getByRole("list");
    expect(list).toBeInTheDocument();
    expect(list).toHaveAttribute("aria-label", "Task list");
  });

  it("uses default role of list", () => {
    renderList();
    expect(screen.getByRole("list")).toBeInTheDocument();
  });

  it("uses custom role when provided", () => {
    renderList({ role: "listbox" });
    expect(screen.getByRole("listbox")).toBeInTheDocument();
  });

  it("ArrowDown advances focused index from -1 to 0", () => {
    const { list } = renderList();
    expect(screen.getByTestId("focused")).toHaveTextContent("-1");

    fireEvent.keyDown(list, { key: "ArrowDown" });
    expect(screen.getByTestId("focused")).toHaveTextContent("0");
  });

  it("ArrowDown wraps from last item to first", () => {
    const { list } = renderList({ itemCount: 3 });

    // Navigate to last item (index 2)
    fireEvent.keyDown(list, { key: "ArrowDown" }); // -1 -> 0
    fireEvent.keyDown(list, { key: "ArrowDown" }); // 0 -> 1
    fireEvent.keyDown(list, { key: "ArrowDown" }); // 1 -> 2
    expect(screen.getByTestId("focused")).toHaveTextContent("2");

    // Wrap to first
    fireEvent.keyDown(list, { key: "ArrowDown" }); // 2 -> 0
    expect(screen.getByTestId("focused")).toHaveTextContent("0");
  });

  it("ArrowUp wraps from first item to last", () => {
    const { list } = renderList({ itemCount: 3 });

    // Go to first item
    fireEvent.keyDown(list, { key: "ArrowDown" }); // -1 -> 0
    expect(screen.getByTestId("focused")).toHaveTextContent("0");

    // Wrap to last
    fireEvent.keyDown(list, { key: "ArrowUp" }); // 0 -> 2
    expect(screen.getByTestId("focused")).toHaveTextContent("2");
  });

  it("Home jumps to index 0", () => {
    const { list } = renderList({ itemCount: 3 });

    // Navigate to index 2
    fireEvent.keyDown(list, { key: "End" });
    expect(screen.getByTestId("focused")).toHaveTextContent("2");

    // Jump to start
    fireEvent.keyDown(list, { key: "Home" });
    expect(screen.getByTestId("focused")).toHaveTextContent("0");
  });

  it("End jumps to last index", () => {
    const { list } = renderList({ itemCount: 5 });

    fireEvent.keyDown(list, { key: "End" });
    expect(screen.getByTestId("focused")).toHaveTextContent("4");
  });

  it("Enter calls onActivate with current focused index", () => {
    const onActivate = vi.fn();
    const { list } = renderList({ onActivate });

    fireEvent.keyDown(list, { key: "ArrowDown" }); // -1 -> 0
    fireEvent.keyDown(list, { key: "ArrowDown" }); // 0 -> 1
    fireEvent.keyDown(list, { key: "Enter" });

    expect(onActivate).toHaveBeenCalledTimes(1);
    expect(onActivate).toHaveBeenCalledWith(1);
  });

  it("Enter does NOT call onActivate when no item is focused", () => {
    const onActivate = vi.fn();
    const { list } = renderList({ onActivate });

    // focusedIndex is -1, press Enter
    fireEvent.keyDown(list, { key: "Enter" });
    expect(onActivate).not.toHaveBeenCalled();
  });

  it("blur resets focused index to -1", () => {
    const { list } = renderList();

    fireEvent.keyDown(list, { key: "ArrowDown" });
    expect(screen.getByTestId("focused")).toHaveTextContent("0");

    fireEvent.blur(list);
    expect(screen.getByTestId("focused")).toHaveTextContent("-1");
  });

  it("does nothing when itemCount is 0", () => {
    const onActivate = vi.fn();
    renderList({ itemCount: 0, onActivate });
    const list = screen.getByRole("list");

    fireEvent.keyDown(list, { key: "ArrowDown" });
    expect(screen.getByTestId("focused")).toHaveTextContent("-1");

    fireEvent.keyDown(list, { key: "ArrowUp" });
    expect(screen.getByTestId("focused")).toHaveTextContent("-1");

    fireEvent.keyDown(list, { key: "Enter" });
    expect(onActivate).not.toHaveBeenCalled();
  });

  it("children render function receives focusedIndex", () => {
    const childrenFn = vi.fn((idx: number) => (
      <span data-testid="focused">{idx}</span>
    ));

    render(
      <KeyboardList itemCount={3} onActivate={vi.fn()}>
        {childrenFn}
      </KeyboardList>,
    );

    expect(childrenFn).toHaveBeenCalledWith(-1);

    fireEvent.keyDown(screen.getByRole("list"), { key: "ArrowDown" });
    expect(childrenFn).toHaveBeenCalledWith(0);
  });
});
