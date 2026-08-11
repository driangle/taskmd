import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { FilterBar, type FilterBarProps } from "./FilterBar.tsx";
import { STATUSES, PRIORITIES, DEFAULT_EFFORTS, TYPES } from "./constants.ts";

function buildDefaultProps(overrides: Partial<FilterBarProps> = {}): FilterBarProps {
  return {
    globalFilter: "",
    onGlobalFilterChange: vi.fn(),
    selectedStatuses: new Set(STATUSES),
    onToggleStatus: vi.fn(),
    onSelectAllStatuses: vi.fn(),
    selectedPriorities: new Set(PRIORITIES),
    onTogglePriority: vi.fn(),
    onSelectAllPriorities: vi.fn(),
    efforts: DEFAULT_EFFORTS,
    selectedEffort: new Set(DEFAULT_EFFORTS),
    onToggleEffort: vi.fn(),
    onSelectAllEffort: vi.fn(),
    selectedTypes: new Set(TYPES),
    onToggleType: vi.fn(),
    onSelectAllTypes: vi.fn(),
    selectedTags: new Set<string>(),
    onRemoveTag: vi.fn(),
    selectedPhases: new Set<string>(),
    availablePhases: [],
    onTogglePhase: vi.fn(),
    onClearFilters: vi.fn(),
    hasActiveFilters: false,
    ...overrides,
  };
}

describe("FilterBar", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("text filter", () => {
    it("renders the text filter input with the provided value", () => {
      render(<FilterBar {...buildDefaultProps({ globalFilter: "hello" })} />);
      const input = screen.getByPlaceholderText("Filter tasks...");
      expect(input).toHaveValue("hello");
    });

    it("calls onGlobalFilterChange when input value changes", () => {
      const onGlobalFilterChange = vi.fn();
      render(<FilterBar {...buildDefaultProps({ onGlobalFilterChange })} />);
      const input = screen.getByPlaceholderText("Filter tasks...");
      fireEvent.change(input, { target: { value: "search term" } });
      expect(onGlobalFilterChange).toHaveBeenCalledWith("search term");
    });
  });

  describe("Filters toggle", () => {
    it("does not show filter rows when filters are closed", () => {
      render(<FilterBar {...buildDefaultProps()} />);
      expect(screen.queryByText("Status:")).not.toBeInTheDocument();
    });

    it("shows filter rows after clicking Filters button", () => {
      render(<FilterBar {...buildDefaultProps()} />);
      fireEvent.click(screen.getByRole("button", { name: /Filters/i }));
      expect(screen.getByText("Status:")).toBeInTheDocument();
      expect(screen.getByText("Priority:")).toBeInTheDocument();
      expect(screen.getByText("Effort:")).toBeInTheDocument();
      expect(screen.getByText("Type:")).toBeInTheDocument();
    });

    it("hides filter rows again after toggling Filters twice", () => {
      render(<FilterBar {...buildDefaultProps()} />);
      const btn = screen.getByRole("button", { name: /Filters/i });
      fireEvent.click(btn);
      fireEvent.click(btn);
      expect(screen.queryByText("Status:")).not.toBeInTheDocument();
    });
  });

  describe("Phase filter row", () => {
    it("shows Phase filter row when availablePhases has items and filters are open", () => {
      render(
        <FilterBar {...buildDefaultProps({ availablePhases: ["alpha", "beta"] })} />,
      );
      fireEvent.click(screen.getByRole("button", { name: /Filters/i }));
      expect(screen.getByText("Phase:")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "alpha" })).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "beta" })).toBeInTheDocument();
    });

    it("does not show Phase filter row when availablePhases is empty", () => {
      render(<FilterBar {...buildDefaultProps({ availablePhases: [] })} />);
      fireEvent.click(screen.getByRole("button", { name: /Filters/i }));
      expect(screen.queryByText("Phase:")).not.toBeInTheDocument();
    });
  });

  describe("active filters indicator", () => {
    it("shows the active-filters indicator when hasActiveFilters is true", () => {
      render(<FilterBar {...buildDefaultProps({ hasActiveFilters: true })} />);
      expect(screen.getByLabelText("Active filters")).toBeInTheDocument();
    });

    it("does not show the active-filters indicator when hasActiveFilters is false", () => {
      render(<FilterBar {...buildDefaultProps({ hasActiveFilters: false })} />);
      expect(screen.queryByLabelText("Active filters")).not.toBeInTheDocument();
    });

    it("shows Clear filters button when hasActiveFilters is true", () => {
      render(<FilterBar {...buildDefaultProps({ hasActiveFilters: true })} />);
      expect(screen.getByRole("button", { name: "Clear filters" })).toBeInTheDocument();
    });

    it("hides Clear filters button when hasActiveFilters is false", () => {
      render(<FilterBar {...buildDefaultProps({ hasActiveFilters: false })} />);
      expect(screen.queryByRole("button", { name: "Clear filters" })).not.toBeInTheDocument();
    });

    it("calls onClearFilters when Clear filters button is clicked", () => {
      const onClearFilters = vi.fn();
      render(<FilterBar {...buildDefaultProps({ hasActiveFilters: true, onClearFilters })} />);
      fireEvent.click(screen.getByRole("button", { name: "Clear filters" }));
      expect(onClearFilters).toHaveBeenCalledOnce();
    });
  });

  describe("filter toggle callbacks", () => {
    it("calls onToggleStatus when a status button is clicked", () => {
      const onToggleStatus = vi.fn();
      render(
        <FilterBar {...buildDefaultProps({ onToggleStatus })} />,
      );
      fireEvent.click(screen.getByRole("button", { name: /Filters/i }));
      fireEvent.click(screen.getByRole("button", { name: "pending" }));
      expect(onToggleStatus).toHaveBeenCalledWith("pending");
    });

    it("calls onSelectAllStatuses when the all button in Status row is clicked", () => {
      const onSelectAllStatuses = vi.fn();
      render(<FilterBar {...buildDefaultProps({ onSelectAllStatuses })} />);
      fireEvent.click(screen.getByRole("button", { name: /Filters/i }));
      // The "all" buttons — get the first one which is Status
      const allButtons = screen.getAllByRole("button", { name: "all" });
      fireEvent.click(allButtons[0]);
      expect(onSelectAllStatuses).toHaveBeenCalledOnce();
    });

    it("calls onTogglePriority when a priority button is clicked", () => {
      const onTogglePriority = vi.fn();
      render(<FilterBar {...buildDefaultProps({ onTogglePriority })} />);
      fireEvent.click(screen.getByRole("button", { name: /Filters/i }));
      fireEvent.click(screen.getByRole("button", { name: "high" }));
      expect(onTogglePriority).toHaveBeenCalledWith("high");
    });

    it("calls onToggleEffort when an effort button is clicked", () => {
      const onToggleEffort = vi.fn();
      render(<FilterBar {...buildDefaultProps({ onToggleEffort })} />);
      fireEvent.click(screen.getByRole("button", { name: /Filters/i }));
      fireEvent.click(screen.getByRole("button", { name: "small" }));
      expect(onToggleEffort).toHaveBeenCalledWith("small");
    });

    it("calls onToggleType when a type button is clicked", () => {
      const onToggleType = vi.fn();
      render(<FilterBar {...buildDefaultProps({ onToggleType })} />);
      fireEvent.click(screen.getByRole("button", { name: /Filters/i }));
      fireEvent.click(screen.getByRole("button", { name: "bug" }));
      expect(onToggleType).toHaveBeenCalledWith("bug");
    });

    it("calls onTogglePhase when a phase button is clicked", () => {
      const onTogglePhase = vi.fn();
      render(
        <FilterBar
          {...buildDefaultProps({ availablePhases: ["alpha"], onTogglePhase })}
        />,
      );
      fireEvent.click(screen.getByRole("button", { name: /Filters/i }));
      fireEvent.click(screen.getByRole("button", { name: "alpha" }));
      expect(onTogglePhase).toHaveBeenCalledWith("alpha");
    });
  });

  describe("selected tags section", () => {
    it("shows Tags section when selectedTags is non-empty", () => {
      render(
        <FilterBar
          {...buildDefaultProps({ selectedTags: new Set(["backend", "frontend"]) })}
        />,
      );
      expect(screen.getByText("Tags:")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /backend/ })).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /frontend/ })).toBeInTheDocument();
    });

    it("does not show Tags section when selectedTags is empty", () => {
      render(<FilterBar {...buildDefaultProps({ selectedTags: new Set() })} />);
      expect(screen.queryByText("Tags:")).not.toBeInTheDocument();
    });

    it("calls onRemoveTag when a tag remove button is clicked", () => {
      const onRemoveTag = vi.fn();
      render(
        <FilterBar
          {...buildDefaultProps({
            selectedTags: new Set(["backend"]),
            onRemoveTag,
          })}
        />,
      );
      fireEvent.click(screen.getByRole("button", { name: /backend/ }));
      expect(onRemoveTag).toHaveBeenCalledWith("backend");
    });
  });
});
