import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { PhaseSelector } from "./PhaseSelector.tsx";
import { DEFAULT_EFFORTS } from "../tasks/TaskTable/constants.ts";

const mockSetPhase = vi.fn();

vi.mock("../../hooks/use-config.ts", () => ({
  useConfig: vi.fn(),
}));

vi.mock("../../hooks/use-phase.tsx", () => ({
  usePhase: () => ({ phase: null, setPhase: mockSetPhase }),
}));

vi.mock("../../hooks/use-tasks.ts", () => ({
  useTasks: () => ({
    data: [
      { id: "1", phase: "phase-a" },
      { id: "2", phase: "phase-a" },
      { id: "3", phase: "phase-b" },
      { id: "4", phase: "" },
    ],
  }),
}));

import { useConfig } from "../../hooks/use-config.ts";
const mockUseConfig = vi.mocked(useConfig);

describe("PhaseSelector", () => {
  beforeEach(() => {
    mockSetPhase.mockClear();
  });

  it("renders nothing when no phases configured", () => {
    mockUseConfig.mockReturnValue({
      phases: [],
      readonly: false,
      version: "1.0",
      efforts: DEFAULT_EFFORTS,
    });
    const { container } = render(<PhaseSelector />);
    expect(container.innerHTML).toBe("");
  });

  it("renders dropdown with phases and counts", () => {
    mockUseConfig.mockReturnValue({
      phases: [
        { id: "phase-a", name: "Phase A", description: "" },
        { id: "phase-b", name: "Phase B", description: "" },
      ],
      readonly: false,
      version: "1.0",
      efforts: DEFAULT_EFFORTS,
    });
    render(<PhaseSelector />);

    expect(screen.getByLabelText("Phase:")).toBeInTheDocument();
    expect(screen.getByText("All (4)")).toBeInTheDocument();
    expect(screen.getByText("Phase A (2)")).toBeInTheDocument();
    expect(screen.getByText("Phase B (1)")).toBeInTheDocument();
  });

  it("calls setPhase when a phase is selected", async () => {
    mockUseConfig.mockReturnValue({
      phases: [
        { id: "phase-a", name: "Phase A", description: "" },
      ],
      readonly: false,
      version: "1.0",
      efforts: DEFAULT_EFFORTS,
    });
    render(<PhaseSelector />);

    await userEvent.selectOptions(screen.getByLabelText("Phase:"), "phase-a");
    expect(mockSetPhase).toHaveBeenCalledWith("phase-a");
  });

  it("calls setPhase with null when All is selected", async () => {
    mockUseConfig.mockReturnValue({
      phases: [
        { id: "phase-a", name: "Phase A", description: "" },
      ],
      readonly: false,
      version: "1.0",
      efforts: DEFAULT_EFFORTS,
    });
    render(<PhaseSelector />);

    await userEvent.selectOptions(screen.getByLabelText("Phase:"), "");
    expect(mockSetPhase).toHaveBeenCalledWith(null);
  });
});
