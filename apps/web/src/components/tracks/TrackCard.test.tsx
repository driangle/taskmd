import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { TrackCard } from "./TrackCard.tsx";
import { createTrackTask, resetFixtureCounter } from "../../test-utils/fixtures.ts";
import { beforeEach } from "vitest";

beforeEach(() => {
  resetFixtureCounter();
});

function renderCard(task: ReturnType<typeof createTrackTask>) {
  return render(
    <MemoryRouter>
      <TrackCard task={task} />
    </MemoryRouter>,
  );
}

describe("TrackCard", () => {
  it("renders task title and id", () => {
    const task = createTrackTask({ id: "042", title: "Implement login" });
    renderCard(task);

    expect(screen.getByText("Implement login")).toBeInTheDocument();
    expect(screen.getByText("042")).toBeInTheDocument();
  });

  it("renders title as a link to the task detail page", () => {
    const task = createTrackTask({ id: "042", title: "Implement login" });
    renderCard(task);

    const link = screen.getByRole("link", { name: "Implement login" });
    expect(link).toHaveAttribute("href", "/tasks/042");
  });

  it("renders score in points", () => {
    const task = createTrackTask({ score: 75 });
    renderCard(task);

    expect(screen.getByText("75 pts")).toBeInTheDocument();
  });

  it("renders priority badge when present", () => {
    const task = createTrackTask({ priority: "high" });
    renderCard(task);

    expect(screen.getByText("high")).toBeInTheDocument();
  });

  it("renders effort badge when present", () => {
    const task = createTrackTask({ effort: "small" });
    renderCard(task);

    expect(screen.getByText("small")).toBeInTheDocument();
  });

  it("renders touches when present", () => {
    const task = createTrackTask({ touches: ["api", "database"] });
    renderCard(task);

    expect(screen.getByText("api")).toBeInTheDocument();
    expect(screen.getByText("database")).toBeInTheDocument();
  });

  it("does not render touches section when empty", () => {
    const task = createTrackTask({ touches: [] });
    renderCard(task);

    // The touches wrapper div should not be present
    expect(screen.queryByText("api")).not.toBeInTheDocument();
  });

  it("does not render touches section when undefined", () => {
    const task = createTrackTask();
    renderCard(task);

    // Default fixture has no touches; the section should be absent
    expect(screen.queryByRole("list", { name: "Touched scopes" })).not.toBeInTheDocument();
  });
});
