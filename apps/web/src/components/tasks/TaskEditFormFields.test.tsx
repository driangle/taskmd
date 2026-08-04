import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { FieldGrid, MetadataFields } from "./TaskEditFormFields.tsx";
import { STATUSES, PRIORITIES, EFFORTS, TYPES } from "./TaskTable/constants.ts";

function optionValues(select: HTMLElement): string[] {
  return Array.from((select as HTMLSelectElement).options).map((o) => o.value);
}

describe("FieldGrid", () => {
  function renderFieldGrid(overrides = {}) {
    const defaults = {
      status: "pending",
      onStatusChange: vi.fn(),
      priority: "medium",
      onPriorityChange: vi.fn(),
      effort: "small",
      onEffortChange: vi.fn(),
      taskType: "feature",
      onTaskTypeChange: vi.fn(),
      inputClasses: "test-input",
    };
    const props = { ...defaults, ...overrides };
    return { ...render(<FieldGrid {...props} />), props };
  }

  it("renders all four labelled select fields", () => {
    renderFieldGrid();
    expect(screen.getByLabelText("Status")).toBeInTheDocument();
    expect(screen.getByLabelText("Priority")).toBeInTheDocument();
    expect(screen.getByLabelText("Effort")).toBeInTheDocument();
    expect(screen.getByLabelText("Type")).toBeInTheDocument();
    expect(screen.getAllByRole("combobox")).toHaveLength(4);
  });

  it("renders all status options", () => {
    renderFieldGrid();
    expect(optionValues(screen.getByLabelText("Status"))).toEqual(STATUSES);
  });

  it("exposes all 6 spec statuses including in-review", () => {
    renderFieldGrid();
    const options = optionValues(screen.getByLabelText("Status"));
    expect(options).toEqual(
      expect.arrayContaining([
        "pending",
        "in-progress",
        "in-review",
        "completed",
        "blocked",
        "cancelled",
      ]),
    );
    expect(options).toHaveLength(6);
  });

  it("renders priority options with empty default", () => {
    renderFieldGrid();
    const select = screen.getByLabelText("Priority") as HTMLSelectElement;
    expect(optionValues(select)).toEqual(["", ...PRIORITIES]);
    expect(select.options[0].textContent).toBe("-");
  });

  it("renders effort options with empty default", () => {
    renderFieldGrid();
    const select = screen.getByLabelText("Effort") as HTMLSelectElement;
    expect(optionValues(select)).toEqual(["", ...EFFORTS]);
    expect(select.options[0].textContent).toBe("-");
  });

  it("renders type options with empty default", () => {
    renderFieldGrid();
    const select = screen.getByLabelText("Type") as HTMLSelectElement;
    expect(optionValues(select)).toEqual(["", ...TYPES]);
    expect(select.options[0].textContent).toBe("-");
  });

  it("calls onStatusChange when status is changed", async () => {
    const { props } = renderFieldGrid();
    await userEvent.selectOptions(screen.getByLabelText("Status"), "completed");
    expect(props.onStatusChange).toHaveBeenCalledWith("completed");
  });

  it("calls onPriorityChange when priority is changed", async () => {
    const { props } = renderFieldGrid();
    await userEvent.selectOptions(screen.getByLabelText("Priority"), "high");
    expect(props.onPriorityChange).toHaveBeenCalledWith("high");
  });

  it("calls onEffortChange when effort is changed", async () => {
    const { props } = renderFieldGrid();
    await userEvent.selectOptions(screen.getByLabelText("Effort"), "large");
    expect(props.onEffortChange).toHaveBeenCalledWith("large");
  });

  it("calls onTaskTypeChange when type is changed", async () => {
    const { props } = renderFieldGrid();
    await userEvent.selectOptions(screen.getByLabelText("Type"), "bug");
    expect(props.onTaskTypeChange).toHaveBeenCalledWith("bug");
  });
});

describe("MetadataFields", () => {
  function renderMetadataFields(overrides = {}) {
    const defaults = {
      phase: "",
      onPhaseChange: vi.fn(),
      owner: "",
      onOwnerChange: vi.fn(),
      parent: "",
      onParentChange: vi.fn(),
      tags: "",
      onTagsChange: vi.fn(),
      inputClasses: "test-input",
    };
    const props = { ...defaults, ...overrides };
    return { ...render(<MetadataFields {...props} />), props };
  }

  it("renders all four labelled input fields", () => {
    renderMetadataFields();
    expect(screen.getByLabelText("Phase")).toBeInTheDocument();
    expect(screen.getByLabelText("Owner")).toBeInTheDocument();
    expect(screen.getByLabelText("Parent")).toBeInTheDocument();
    expect(screen.getByLabelText("Tags (comma-separated)")).toBeInTheDocument();
    expect(screen.getAllByRole("textbox")).toHaveLength(4);
  });

  it("renders correct placeholders", () => {
    renderMetadataFields();
    expect(screen.getByPlaceholderText("e.g. v1.0")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("e.g. alice")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("e.g. 045")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("e.g. backend, api, feature")).toBeInTheDocument();
  });

  it("calls onOwnerChange when owner input changes", async () => {
    const { props } = renderMetadataFields();
    await userEvent.type(screen.getByLabelText("Owner"), "bob");
    expect(props.onOwnerChange).toHaveBeenCalled();
  });

  it("calls onParentChange when parent input changes", async () => {
    const { props } = renderMetadataFields();
    await userEvent.type(screen.getByLabelText("Parent"), "042");
    expect(props.onParentChange).toHaveBeenCalled();
  });

  it("calls onTagsChange when tags input changes", async () => {
    const { props } = renderMetadataFields();
    await userEvent.type(screen.getByLabelText("Tags (comma-separated)"), "test");
    expect(props.onTagsChange).toHaveBeenCalled();
  });
});
