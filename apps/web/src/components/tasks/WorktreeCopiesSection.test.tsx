import { describe, it, expect } from "vitest";
import { render, screen, within } from "@testing-library/react";
import { WorktreeCopiesSection } from "./WorktreeCopiesSection.tsx";
import type { WorktreeCopy } from "../../api/types.ts";

const copies: WorktreeCopy[] = [
  { status: "pending", local: true },
  { worktree: "agent-b", branch: "dnc/001/parser", status: "in-progress", owner: "agent-b" },
];

describe("WorktreeCopiesSection", () => {
  it("renders one row per copy with status and owner", () => {
    render(<WorktreeCopiesSection copies={copies} />);
    expect(screen.getByText("Worktree Copies")).toBeInTheDocument();

    const rows = screen.getAllByRole("row");
    // Header row + one per copy.
    expect(rows).toHaveLength(3);
    expect(within(rows[1]).getByText("(local)")).toBeInTheDocument();
    expect(within(rows[1]).getByText("pending")).toBeInTheDocument();
    // "agent-b" appears as both the worktree name and the owner.
    expect(within(rows[2]).getAllByText("agent-b")).toHaveLength(2);
    expect(within(rows[2]).getByText("dnc/001/parser")).toBeInTheDocument();
    expect(within(rows[2]).getByText("in-progress")).toBeInTheDocument();
  });

  it("marks the local copy and dashes out its missing branch and owner", () => {
    render(<WorktreeCopiesSection copies={copies} />);
    const localRow = screen.getAllByRole("row")[1];
    expect(within(localRow).getByText("this worktree")).toBeInTheDocument();
    expect(within(localRow).getAllByText("—").length).toBeGreaterThanOrEqual(2);
  });
});
