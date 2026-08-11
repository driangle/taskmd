import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { BoardFilterBar, BoardFilterBarProps } from "./BoardFilterBar.tsx";
import { STATUSES, PRIORITIES, DEFAULT_EFFORTS, TYPES } from "../tasks/TaskTable/constants.ts";

vi.mock("./TagAutocomplete.tsx", () => ({
  TagAutocomplete: () => <div data-testid="tag-autocomplete" />,
}));

function defaultProps(overrides: Partial<BoardFilterBarProps> = {}): BoardFilterBarProps {
  return {
    groupBy: "status",
    selectedStatuses: new Set(STATUSES),
    onStatusesChange: vi.fn(),
    selectedPriorities: new Set(PRIORITIES),
    onPrioritiesChange: vi.fn(),
    efforts: DEFAULT_EFFORTS,
    selectedEfforts: new Set(DEFAULT_EFFORTS),
    onEffortsChange: vi.fn(),
    selectedTypes: new Set(TYPES),
    onTypesChange: vi.fn(),
    selectedTags: new Set(),
    onTagsChange: vi.fn(),
    availableTags: [],
    ...overrides,
  };
}

describe("BoardFilterBar pill callbacks", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls onEffortsChange when clicking an effort pill", () => {
    const onEffortsChange = vi.fn();
    render(
      <BoardFilterBar
        {...defaultProps({ groupBy: "status", onEffortsChange })}
      />
    );
    fireEvent.click(screen.getByText("Filters"));
    fireEvent.click(screen.getByText("small"));
    expect(onEffortsChange).toHaveBeenCalledTimes(1);
    const result = onEffortsChange.mock.calls[0][0] as Set<string>;
    expect(result.has("small")).toBe(false);
  });

  it("calls onEffortsChange with all efforts when clicking all button in Effort row", () => {
    const onEffortsChange = vi.fn();
    render(
      <BoardFilterBar
        {...defaultProps({
          groupBy: "status",
          selectedEfforts: new Set(["small"]),
          onEffortsChange,
        })}
      />
    );
    fireEvent.click(screen.getByText("Filters"));
    const allButtons = screen.getAllByText("all");
    fireEvent.click(allButtons[1]);
    expect(onEffortsChange).toHaveBeenCalledTimes(1);
    const result = onEffortsChange.mock.calls[0][0] as Set<string>;
    expect(result.size).toBe(DEFAULT_EFFORTS.length);
  });

  it("calls onTypesChange when clicking a type pill", () => {
    const onTypesChange = vi.fn();
    render(
      <BoardFilterBar
        {...defaultProps({ groupBy: "status", onTypesChange })}
      />
    );
    fireEvent.click(screen.getByText("Filters"));
    fireEvent.click(screen.getByText("feature"));
    expect(onTypesChange).toHaveBeenCalledTimes(1);
    const result = onTypesChange.mock.calls[0][0] as Set<string>;
    expect(result.has("feature")).toBe(false);
  });

  it("calls onTypesChange with all types when clicking all button in Type row", () => {
    const onTypesChange = vi.fn();
    render(
      <BoardFilterBar
        {...defaultProps({
          groupBy: "status",
          selectedTypes: new Set(["feature"]),
          onTypesChange,
        })}
      />
    );
    fireEvent.click(screen.getByText("Filters"));
    const allButtons = screen.getAllByText("all");
    fireEvent.click(allButtons[2]);
    expect(onTypesChange).toHaveBeenCalledTimes(1);
    const result = onTypesChange.mock.calls[0][0] as Set<string>;
    expect(result.size).toBe(TYPES.length);
  });

  it("calls onPrioritiesChange when clicking a priority pill", () => {
    const onPrioritiesChange = vi.fn();
    render(
      <BoardFilterBar
        {...defaultProps({ groupBy: "status", onPrioritiesChange })}
      />
    );
    fireEvent.click(screen.getByText("Filters"));
    fireEvent.click(screen.getByText("high"));
    expect(onPrioritiesChange).toHaveBeenCalledTimes(1);
    const result = onPrioritiesChange.mock.calls[0][0] as Set<string>;
    expect(result.has("high")).toBe(false);
  });

  it("calls onPrioritiesChange with all priorities when clicking all button in Priority row", () => {
    const onPrioritiesChange = vi.fn();
    render(
      <BoardFilterBar
        {...defaultProps({
          groupBy: "status",
          selectedPriorities: new Set(["high"]),
          onPrioritiesChange,
        })}
      />
    );
    fireEvent.click(screen.getByText("Filters"));
    const allButtons = screen.getAllByText("all");
    fireEvent.click(allButtons[0]);
    expect(onPrioritiesChange).toHaveBeenCalledTimes(1);
    const result = onPrioritiesChange.mock.calls[0][0] as Set<string>;
    expect(result.size).toBe(PRIORITIES.length);
  });

  it("shows hasActiveFilters dot when tags are selected", () => {
    render(
      <BoardFilterBar
        {...defaultProps({
          selectedTags: new Set(["bug"]),
        })}
      />
    );
    expect(screen.getByText("Clear filters")).toBeInTheDocument();
  });
});
