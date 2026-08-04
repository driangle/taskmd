import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { StatusBadge, PriorityBadge, TypeBadge, PhaseBadge, BlockedStatusBadge } from "./Badges.tsx";
import { STATUS_COLORS, PRIORITY_COLORS, TYPE_COLORS, getPhaseColor } from "./constants.ts";

describe("StatusBadge", () => {
  it.each(Object.keys(STATUS_COLORS))("renders the '%s' status label", (status) => {
    render(<StatusBadge status={status} />);
    expect(screen.getByText(status)).toBeInTheDocument();
  });

  it("still renders the label for an unknown status", () => {
    render(<StatusBadge status="unknown" />);
    expect(screen.getByText("unknown")).toBeInTheDocument();
  });
});

describe("PriorityBadge", () => {
  it.each(Object.keys(PRIORITY_COLORS))("renders the '%s' priority label", (priority) => {
    render(<PriorityBadge priority={priority} />);
    expect(screen.getByText(priority)).toBeInTheDocument();
  });

  it("still renders the label for an unknown priority", () => {
    render(<PriorityBadge priority="unknown" />);
    expect(screen.getByText("unknown")).toBeInTheDocument();
  });
});

describe("TypeBadge", () => {
  it.each(Object.keys(TYPE_COLORS))("renders the '%s' type label", (type) => {
    render(<TypeBadge type={type} />);
    expect(screen.getByText(type)).toBeInTheDocument();
  });

  it("still renders the label for an unknown type", () => {
    render(<TypeBadge type="unknown" />);
    expect(screen.getByText("unknown")).toBeInTheDocument();
  });
});

describe("PhaseBadge", () => {
  it("renders the phase name", () => {
    render(<PhaseBadge phase="alpha" />);
    expect(screen.getByText("alpha")).toBeInTheDocument();
  });

  // getPhaseColor is a pure hashing function; test its contract directly rather than
  // asserting the resulting Tailwind classes on the rendered badge.
  it("assigns different colors to different phases", () => {
    const unique = new Set([
      getPhaseColor("alpha"),
      getPhaseColor("beta"),
      getPhaseColor("gamma"),
    ]);
    // Hash collisions are theoretically possible but unlikely for three distinct inputs.
    expect(unique.size).toBeGreaterThanOrEqual(2);
  });

  it("assigns the same color to the same phase consistently", () => {
    expect(getPhaseColor("release-1")).toBe(getPhaseColor("release-1"));
  });
});

describe("BlockedStatusBadge", () => {
  it("renders Ready badge when dependencies is null", () => {
    render(<BlockedStatusBadge dependencies={null} />);
    expect(screen.getByText("Ready")).toBeInTheDocument();
    expect(screen.getByText("✓")).toBeInTheDocument();
    expect(screen.getByLabelText("Task is ready to work on")).toBeInTheDocument();
  });

  it("renders Ready badge when dependencies is empty array", () => {
    render(<BlockedStatusBadge dependencies={[]} />);
    expect(screen.getByText("Ready")).toBeInTheDocument();
  });

  it("renders Blocked badge with count for single dependency", () => {
    render(<BlockedStatusBadge dependencies={["005"]} />);
    expect(screen.getByText("(1)")).toBeInTheDocument();
    expect(screen.getByText("⚠")).toBeInTheDocument();
  });

  it("renders Blocked badge with count for multiple dependencies", () => {
    render(<BlockedStatusBadge dependencies={["005", "010", "015"]} />);
    expect(screen.getByText("(3)")).toBeInTheDocument();
  });

  it("shows tooltip with blocked-by IDs", () => {
    render(<BlockedStatusBadge dependencies={["005", "010"]} />);
    const badge = screen.getByLabelText("Blocked by: 005, 010");
    expect(badge).toHaveAttribute("title", "Blocked by: 005, 010");
  });

  it("shows Ready when all dependencies are completed", () => {
    const statusMap = new Map([["005", "completed"], ["010", "completed"]]);
    render(<BlockedStatusBadge dependencies={["005", "010"]} taskStatusMap={statusMap} />);
    expect(screen.getByText("Ready")).toBeInTheDocument();
  });

  it("shows Blocked only for unmet dependencies", () => {
    const statusMap = new Map([["005", "completed"], ["010", "pending"]]);
    render(<BlockedStatusBadge dependencies={["005", "010"]} taskStatusMap={statusMap} />);
    expect(screen.getByText("(1)")).toBeInTheDocument();
    expect(screen.getByLabelText("Blocked by: 010")).toBeInTheDocument();
  });

  it("treats missing tasks in statusMap as unmet", () => {
    const statusMap = new Map([["005", "completed"]]);
    render(<BlockedStatusBadge dependencies={["005", "999"]} taskStatusMap={statusMap} />);
    expect(screen.getByText("(1)")).toBeInTheDocument();
    expect(screen.getByLabelText("Blocked by: 999")).toBeInTheDocument();
  });
});
