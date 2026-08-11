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

describe("BoardFilterBar", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders Filters button initially collapsed", () => {
    render(<BoardFilterBar {...defaultProps()} />);

    expect(screen.getByText("Filters")).toBeInTheDocument();
    expect(screen.queryByText("Priority:")).not.toBeInTheDocument();
    expect(screen.queryByText("Effort:")).not.toBeInTheDocument();
    expect(screen.queryByText("Type:")).not.toBeInTheDocument();
  });

  it("clicking Filters expands the filter section", () => {
    render(<BoardFilterBar {...defaultProps()} />);

    fireEvent.click(screen.getByText("Filters"));

    expect(screen.getByText("Priority:")).toBeInTheDocument();
    expect(screen.getByText("Effort:")).toBeInTheDocument();
    expect(screen.getByText("Type:")).toBeInTheDocument();
  });

  it("hides Status pills when groupBy is status", () => {
    render(<BoardFilterBar {...defaultProps({ groupBy: "status" })} />);

    fireEvent.click(screen.getByText("Filters"));

    expect(screen.queryByText("Status:")).not.toBeInTheDocument();
    expect(screen.getByText("Priority:")).toBeInTheDocument();
    expect(screen.getByText("Effort:")).toBeInTheDocument();
    expect(screen.getByText("Type:")).toBeInTheDocument();
  });

  it("hides Priority pills when groupBy is priority", () => {
    render(<BoardFilterBar {...defaultProps({ groupBy: "priority" })} />);

    fireEvent.click(screen.getByText("Filters"));

    expect(screen.getByText("Status:")).toBeInTheDocument();
    expect(screen.queryByText("Priority:")).not.toBeInTheDocument();
    expect(screen.getByText("Effort:")).toBeInTheDocument();
    expect(screen.getByText("Type:")).toBeInTheDocument();
  });

  it("calls onStatusesChange when clicking a status pill", () => {
    const onStatusesChange = vi.fn();
    render(
      <BoardFilterBar
        {...defaultProps({ groupBy: "priority", onStatusesChange })}
      />
    );

    fireEvent.click(screen.getByText("Filters"));
    fireEvent.click(screen.getByText("pending"));

    expect(onStatusesChange).toHaveBeenCalledTimes(1);
    const result = onStatusesChange.mock.calls[0][0] as Set<string>;
    expect(result).toBeInstanceOf(Set);
    expect(result.has("pending")).toBe(false);
  });

  it("calls onStatusesChange with all statuses when clicking the all button", () => {
    const onStatusesChange = vi.fn();
    const partial = new Set(["pending"]);
    render(
      <BoardFilterBar
        {...defaultProps({
          groupBy: "priority",
          selectedStatuses: partial,
          onStatusesChange,
        })}
      />
    );

    fireEvent.click(screen.getByText("Filters"));

    const allButtons = screen.getAllByText("all");
    // Status row is the first row when groupBy !== "status"
    fireEvent.click(allButtons[0]);

    expect(onStatusesChange).toHaveBeenCalledTimes(1);
    const result = onStatusesChange.mock.calls[0][0] as Set<string>;
    expect(result.size).toBe(STATUSES.length);
    for (const s of STATUSES) {
      expect(result.has(s)).toBe(true);
    }
  });

  it("shows Clear filters button when filters are active", () => {
    const partial = new Set(["pending"]);
    render(
      <BoardFilterBar
        {...defaultProps({
          groupBy: "priority",
          selectedStatuses: partial,
        })}
      />
    );

    expect(screen.getByText("Clear filters")).toBeInTheDocument();
  });

  it("resets all filters when clicking Clear filters", () => {
    const onStatusesChange = vi.fn();
    const onPrioritiesChange = vi.fn();
    const onEffortsChange = vi.fn();
    const onTypesChange = vi.fn();
    const onTagsChange = vi.fn();

    render(
      <BoardFilterBar
        {...defaultProps({
          groupBy: "priority",
          selectedStatuses: new Set(["pending"]),
          onStatusesChange,
          onPrioritiesChange,
          onEffortsChange,
          onTypesChange,
          selectedTags: new Set(["bug"]),
          onTagsChange,
        })}
      />
    );

    fireEvent.click(screen.getByText("Clear filters"));

    expect(onStatusesChange).toHaveBeenCalledWith(new Set(STATUSES));
    expect(onPrioritiesChange).toHaveBeenCalledWith(new Set(PRIORITIES));
    expect(onEffortsChange).toHaveBeenCalledWith(new Set(DEFAULT_EFFORTS));
    expect(onTypesChange).toHaveBeenCalledWith(new Set(TYPES));
    expect(onTagsChange).toHaveBeenCalledWith(new Set());
  });

  it("does not show Clear filters when all filters are at defaults", () => {
    render(<BoardFilterBar {...defaultProps()} />);

    expect(screen.queryByText("Clear filters")).not.toBeInTheDocument();
  });

  it("shows TagAutocomplete when availableTags has items and groupBy is not tag", () => {
    render(
      <BoardFilterBar
        {...defaultProps({
          groupBy: "status",
          availableTags: ["frontend", "backend"],
        })}
      />
    );

    fireEvent.click(screen.getByText("Filters"));

    expect(screen.getByTestId("tag-autocomplete")).toBeInTheDocument();
  });

  it("hides TagAutocomplete when groupBy is tag", () => {
    render(
      <BoardFilterBar
        {...defaultProps({
          groupBy: "tag",
          availableTags: ["frontend", "backend"],
        })}
      />
    );

    fireEvent.click(screen.getByText("Filters"));

    expect(screen.queryByTestId("tag-autocomplete")).not.toBeInTheDocument();
  });

  it("hides TagAutocomplete when availableTags is empty", () => {
    render(
      <BoardFilterBar
        {...defaultProps({
          groupBy: "status",
          availableTags: [],
        })}
      />
    );

    fireEvent.click(screen.getByText("Filters"));

    expect(screen.queryByTestId("tag-autocomplete")).not.toBeInTheDocument();
  });

});
