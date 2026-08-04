import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { TracksPage } from "./TracksPage.tsx";
import { createTracksResult } from "../test-utils/index.ts";

vi.mock("../hooks/use-tracks.ts", () => ({
  useTracks: vi.fn(),
}));

vi.mock("../hooks/use-phase.tsx", () => ({
  usePhase: () => ({ phase: null, setPhase: vi.fn() }),
}));

vi.mock("../hooks/use-project.ts", () => ({
  useProject: () => ({ project: null, setProject: vi.fn() }),
}));

vi.mock("../components/tracks/TracksView.tsx", () => ({
  TracksView: () => <div data-testid="tracks-view">TracksView</div>,
}));

import { useTracks } from "../hooks/use-tracks.ts";
const mockUseTracks = vi.mocked(useTracks);

describe("TracksPage", () => {
  it("renders loading state", () => {
    mockUseTracks.mockReturnValue({
      data: undefined,
      error: undefined,
      isLoading: true,
      mutate: vi.fn(),
      isValidating: false,
    });
    render(<TracksPage />);
    expect(screen.getByRole("status", { name: /loading/i })).toBeInTheDocument();
  });

  it("renders error state", () => {
    mockUseTracks.mockReturnValue({
      data: undefined,
      error: new Error("Server error"),
      isLoading: false,
      mutate: vi.fn(),
      isValidating: false,
    });
    render(<TracksPage />);
    expect(screen.getByText("Something went wrong")).toBeInTheDocument();
  });

  it("renders TracksView when data is available", () => {
    mockUseTracks.mockReturnValue({
      data: createTracksResult(),
      error: undefined,
      isLoading: false,
      mutate: vi.fn(),
      isValidating: false,
    });
    render(<TracksPage />);
    expect(screen.getByTestId("tracks-view")).toBeInTheDocument();
  });

  it("calls mutate when retry is clicked in error state", () => {
    const mockMutate = vi.fn();
    mockUseTracks.mockReturnValue({
      data: undefined,
      error: new Error("Server error"),
      isLoading: false,
      mutate: mockMutate,
      isValidating: false,
    });
    render(<TracksPage />);
    fireEvent.click(screen.getByText("Retry"));
    expect(mockMutate).toHaveBeenCalled();
  });
});
