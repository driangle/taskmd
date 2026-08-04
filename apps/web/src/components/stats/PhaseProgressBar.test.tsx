import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { PhaseProgressBar } from "./PhaseProgressBar.tsx";
import type { PhaseProgress } from "./PhaseProgressBar.tsx";

function makePhase(overrides: Partial<PhaseProgress> = {}): PhaseProgress {
  return {
    id: "phase-1",
    name: "Phase One",
    total: 10,
    completed: 5,
    ...overrides,
  };
}

describe("PhaseProgressBar", () => {
  it("renders phase name and task count label", () => {
    render(<PhaseProgressBar phase={makePhase()} />);
    expect(screen.getByText("Phase One")).toBeInTheDocument();
    expect(screen.getByText("5 / 10 tasks (50%)")).toBeInTheDocument();
  });

  it("renders 0% for a phase with zero tasks", () => {
    render(<PhaseProgressBar phase={makePhase({ total: 0, completed: 0 })} />);
    expect(screen.getByText("0 / 0 tasks (0%)")).toBeInTheDocument();
    expect(screen.getByRole("progressbar")).toHaveAttribute("aria-valuenow", "0");
  });

  it("renders 100% for fully completed phase", () => {
    render(<PhaseProgressBar phase={makePhase({ total: 8, completed: 8 })} />);
    expect(screen.getByText("8 / 8 tasks (100%)")).toBeInTheDocument();
    expect(screen.getByRole("progressbar")).toHaveAttribute("aria-valuenow", "100");
  });

  it("exposes completion as an accessible progressbar value", () => {
    render(<PhaseProgressBar phase={makePhase({ total: 4, completed: 3 })} />);
    const bar = screen.getByRole("progressbar");
    expect(bar).toHaveAttribute("aria-valuenow", "75");
    expect(bar).toHaveAttribute("aria-valuemin", "0");
    expect(bar).toHaveAttribute("aria-valuemax", "100");
  });

  it("sets progressbar value proportional to completion", () => {
    render(<PhaseProgressBar phase={makePhase({ total: 10, completed: 3 })} />);
    expect(screen.getByRole("progressbar")).toHaveAttribute("aria-valuenow", "30");
  });
});
