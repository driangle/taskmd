import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { TracksView } from "./TracksView.tsx";
import {
  createTrack,
  createTracksResult,
  resetFixtureCounter,
} from "../../test-utils/fixtures.ts";

vi.mock("./TrackColumn.tsx", () => ({
  TrackColumn: ({ track }: { track: { id: number } }) => (
    <div data-testid={`track-${track.id}`}>Track {track.id}</div>
  ),
}));

vi.mock("./FlexibleSection.tsx", () => ({
  FlexibleSection: ({ tasks }: { tasks: unknown[] }) =>
    tasks.length > 0 ? (
      <div data-testid="flexible">{tasks.length} flexible</div>
    ) : null,
}));

beforeEach(() => {
  resetFixtureCounter();
});

describe("TracksView", () => {
  it("renders 'Parallel Tracks' heading", () => {
    const data = createTracksResult();
    render(<TracksView data={data} limit={0} onLimitChange={() => {}} />);

    expect(screen.getByText("Parallel Tracks")).toBeInTheDocument();
  });

  it("renders limit buttons (All, 2, 3, 5)", () => {
    const data = createTracksResult();
    render(<TracksView data={data} limit={0} onLimitChange={() => {}} />);

    expect(screen.getByText("All")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
    expect(screen.getByText("5")).toBeInTheDocument();
  });

  it("marks the active limit button as pressed", () => {
    const data = createTracksResult();
    render(<TracksView data={data} limit={3} onLimitChange={() => {}} />);

    expect(screen.getByRole("button", { name: "3" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "All" })).toHaveAttribute("aria-pressed", "false");
  });

  it("calls onLimitChange when a limit button is clicked", () => {
    const onLimitChange = vi.fn();
    const data = createTracksResult();
    render(<TracksView data={data} limit={0} onLimitChange={onLimitChange} />);

    fireEvent.click(screen.getByText("3"));
    expect(onLimitChange).toHaveBeenCalledWith(3);

    fireEvent.click(screen.getByText("All"));
    expect(onLimitChange).toHaveBeenCalledWith(0);
  });

  it("shows empty state when no tracks and no flexible tasks", () => {
    const data = createTracksResult({ tracks: [], flexible: [] });
    render(<TracksView data={data} limit={0} onLimitChange={() => {}} />);

    expect(
      screen.getByText(/No actionable tasks found/),
    ).toBeInTheDocument();
  });

  it("renders track columns", () => {
    const data = createTracksResult({
      tracks: [
        createTrack({ id: 1 }),
        createTrack({ id: 2 }),
      ],
    });
    render(<TracksView data={data} limit={0} onLimitChange={() => {}} />);

    expect(screen.getByTestId("track-1")).toBeInTheDocument();
    expect(screen.getByTestId("track-2")).toBeInTheDocument();
  });

  it("shows warnings when present", () => {
    const data = createTracksResult({
      warnings: ["Circular dependency detected", "Missing scope definition"],
    });
    render(<TracksView data={data} limit={0} onLimitChange={() => {}} />);

    expect(screen.getByText("Warnings")).toBeInTheDocument();
    expect(screen.getByText("Circular dependency detected")).toBeInTheDocument();
    expect(screen.getByText("Missing scope definition")).toBeInTheDocument();
  });

  it("does not show warnings section when no warnings", () => {
    const data = createTracksResult({ warnings: [] });
    render(<TracksView data={data} limit={0} onLimitChange={() => {}} />);

    expect(screen.queryByText("Warnings")).not.toBeInTheDocument();
  });
});
