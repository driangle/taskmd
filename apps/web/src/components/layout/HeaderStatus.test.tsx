import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { HeaderStatus } from "./HeaderStatus.tsx";

describe("HeaderStatus", () => {
  it("shows the worktree overlay indicator with name and sibling count", () => {
    render(
      <HeaderStatus
        version="1.0.0"
        readonly={false}
        worktree={{ name: "agent-b", siblings: 3 }}
      />,
    );
    expect(screen.getByText(/worktree agent-b — 3 siblings/)).toBeInTheDocument();
  });

  it("uses the singular for one sibling and a generic label without a name", () => {
    render(
      <HeaderStatus version="" readonly={false} worktree={{ siblings: 1 }} />,
    );
    expect(screen.getByText(/worktrees — 1 sibling$/)).toBeInTheDocument();
  });

  it("shows no indicator in a single-worktree repo", () => {
    render(<HeaderStatus version="1.0.0" readonly={false} />);
    expect(screen.queryByText(/worktree/)).not.toBeInTheDocument();
  });

  it("still renders the read-only pill alongside the indicator", () => {
    render(
      <HeaderStatus
        version="1.0.0"
        readonly
        worktree={{ name: "agent-b", siblings: 2 }}
      />,
    );
    expect(screen.getByText("Read Only")).toBeInTheDocument();
    expect(screen.getByText(/worktree agent-b — 2 siblings/)).toBeInTheDocument();
  });
});
