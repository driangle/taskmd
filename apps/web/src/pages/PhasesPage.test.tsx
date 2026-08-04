import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";

vi.mock("../hooks/use-config.ts", () => ({ useConfig: vi.fn() }));
vi.mock("../hooks/use-project.ts", () => ({
  useProject: () => ({ project: null }),
}));
vi.mock("../hooks/use-tasks.ts", () => ({ useTasks: vi.fn() }));
vi.mock("../components/phases/PhasesView.tsx", () => ({
  PhasesView: () => <div data-testid="phases-view">PhasesView</div>,
}));

import { useConfig } from "../hooks/use-config.ts";
import { useTasks } from "../hooks/use-tasks.ts";
const mockUseConfig = vi.mocked(useConfig);
const mockUseTasks = vi.mocked(useTasks);

import { PhasesPage } from "./PhasesPage.tsx";

describe("PhasesPage", () => {
  beforeEach(() => {
    mockUseConfig.mockReturnValue({
      readonly: false,
      version: "1.0.0",
      phases: [],
    });
  });

  it("renders loading state", () => {
    mockUseTasks.mockReturnValue({
      data: undefined,
      error: undefined,
      isLoading: true,
      mutate: vi.fn(),
      isValidating: false,
    });
    render(<PhasesPage />);
    expect(screen.getByRole("status", { name: /loading/i })).toBeInTheDocument();
  });

  it("renders error state", () => {
    mockUseTasks.mockReturnValue({
      data: undefined,
      error: new Error("Failed to load"),
      isLoading: false,
      mutate: vi.fn(),
      isValidating: false,
    });
    render(<PhasesPage />);
    expect(screen.getByText("Something went wrong")).toBeInTheDocument();
  });

  it("renders PhasesView when data is available", () => {
    mockUseTasks.mockReturnValue({
      data: [],
      error: undefined,
      isLoading: false,
      mutate: vi.fn(),
      isValidating: false,
    });
    render(<PhasesPage />);
    expect(screen.getByTestId("phases-view")).toBeInTheDocument();
  });

  it("calls mutate when retry is clicked in error state", () => {
    const mockMutate = vi.fn();
    mockUseTasks.mockReturnValue({
      data: undefined,
      error: new Error("Failed to load"),
      isLoading: false,
      mutate: mockMutate,
      isValidating: false,
    });
    render(<PhasesPage />);
    fireEvent.click(screen.getByText("Retry"));
    expect(mockMutate).toHaveBeenCalled();
  });
});
