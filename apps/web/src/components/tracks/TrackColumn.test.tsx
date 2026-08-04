import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { TrackColumn } from "./TrackColumn.tsx";
import { createTrack, createTrackTask, resetFixtureCounter } from "../../test-utils/fixtures.ts";
import { beforeEach } from "vitest";

beforeEach(() => {
  resetFixtureCounter();
});

function renderColumn(track: ReturnType<typeof createTrack>) {
  return render(
    <MemoryRouter>
      <TrackColumn track={track} />
    </MemoryRouter>,
  );
}

describe("TrackColumn", () => {
  it("renders track id and task count", () => {
    const tasks = [createTrackTask(), createTrackTask(), createTrackTask()];
    const track = createTrack({ id: 2, tasks });
    renderColumn(track);

    expect(screen.getByText(/Track 2/)).toBeInTheDocument();
    expect(screen.getByText("(3)")).toBeInTheDocument();
  });

  it("renders scope badges", () => {
    const track = createTrack({ scopes: ["api", "auth"] });
    renderColumn(track);

    expect(screen.getByText("api")).toBeInTheDocument();
    expect(screen.getByText("auth")).toBeInTheDocument();
  });

  it("renders task cards", () => {
    const tasks = [
      createTrackTask({ title: "First task" }),
      createTrackTask({ title: "Second task" }),
    ];
    const track = createTrack({ tasks });
    renderColumn(track);

    expect(screen.getByText("First task")).toBeInTheDocument();
    expect(screen.getByText("Second task")).toBeInTheDocument();
  });

  it("does not render scopes section when scopes is empty", () => {
    const track = createTrack({ scopes: [] });
    renderColumn(track);

    // The header should still render
    expect(screen.getByRole("heading", { level: 3 })).toHaveTextContent("Track 1");
    // With empty scopes the scopes list should not be rendered
    expect(screen.queryByRole("list", { name: "Scopes" })).not.toBeInTheDocument();
  });
});
