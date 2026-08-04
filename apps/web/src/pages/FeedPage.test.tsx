import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { FeedPage } from "./FeedPage.tsx";

vi.mock("../hooks/use-feed.ts", () => ({
  useFeed: vi.fn(),
}));

vi.mock("../hooks/use-project.ts", () => ({
  useProject: () => ({ project: null, setProject: vi.fn() }),
}));

vi.mock("../components/feed/FeedView.tsx", () => ({
  FeedView: ({ entries }: { entries: unknown[] }) => (
    <div data-testid="feed-view">{entries.length} entries</div>
  ),
}));

import { useFeed } from "../hooks/use-feed.ts";
const mockUseFeed = vi.mocked(useFeed);

describe("FeedPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders loading state", () => {
    mockUseFeed.mockReturnValue({
      data: undefined,
      error: undefined,
      isLoading: true,
      mutate: vi.fn(),
      isValidating: false,
    });
    render(<FeedPage />);
    expect(screen.getByRole("status", { name: /loading/i })).toBeInTheDocument();
  });

  it("renders error state", () => {
    mockUseFeed.mockReturnValue({
      data: undefined,
      error: new Error("Server error"),
      isLoading: false,
      mutate: vi.fn(),
      isValidating: false,
    });
    render(<FeedPage />);
    expect(screen.getByText("Something went wrong")).toBeInTheDocument();
  });

  it("renders empty state when no entries", () => {
    mockUseFeed.mockReturnValue({
      data: [],
      error: undefined,
      isLoading: false,
      mutate: vi.fn(),
      isValidating: false,
    });
    render(<FeedPage />);
    expect(screen.getByText("No recent activity")).toBeInTheDocument();
  });

  it("renders FeedView when data is available", () => {
    mockUseFeed.mockReturnValue({
      data: [
        { source: "git", timestamp: "2026-03-01T10:00:00Z", message: "test" },
      ],
      error: undefined,
      isLoading: false,
      mutate: vi.fn(),
      isValidating: false,
    });
    render(<FeedPage />);
    expect(screen.getByTestId("feed-view")).toBeInTheDocument();
    expect(screen.getByText("1 entries")).toBeInTheDocument();
  });

  it("calls mutate when retry is clicked", () => {
    const mockMutate = vi.fn();
    mockUseFeed.mockReturnValue({
      data: undefined,
      error: new Error("Server error"),
      isLoading: false,
      mutate: mockMutate,
      isValidating: false,
    });
    render(<FeedPage />);
    fireEvent.click(screen.getByText("Retry"));
    expect(mockMutate).toHaveBeenCalled();
  });

  it("renders filter controls", () => {
    mockUseFeed.mockReturnValue({
      data: [],
      error: undefined,
      isLoading: false,
      mutate: vi.fn(),
      isValidating: false,
    });
    render(<FeedPage />);
    expect(screen.getByText("Source")).toBeInTheDocument();
    expect(screen.getByText("Since")).toBeInTheDocument();
    expect(screen.getByText("Scope")).toBeInTheDocument();
  });
});
