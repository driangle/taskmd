import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { DEFAULT_EFFORTS } from "../components/tasks/TaskTable/constants.ts";
import { StatsPage } from "./StatsPage.tsx";
import { createStats, createTask } from "../test-utils/index.ts";

vi.mock("../hooks/use-stats.ts", () => ({
  useStats: vi.fn(),
}));

vi.mock("../hooks/use-phase.tsx", () => ({
  usePhase: () => ({ phase: null, setPhase: vi.fn() }),
}));

vi.mock("../hooks/use-project.ts", () => ({
  useProject: () => ({ project: null, setProject: vi.fn() }),
}));

vi.mock("../hooks/use-tasks.ts", () => ({
  useTasks: vi.fn(),
}));

vi.mock("../hooks/use-config.ts", () => ({
  useConfig: vi.fn(),
}));

vi.mock("../components/stats/StatsView.tsx", () => ({
  StatsView: ({ stats }: { stats: { total_tasks: number } }) => (
    <div data-testid="stats-view">Stats: {stats.total_tasks} tasks</div>
  ),
}));

import { useStats } from "../hooks/use-stats.ts";
const mockUseStats = vi.mocked(useStats);

import { useTasks } from "../hooks/use-tasks.ts";
const mockUseTasks = vi.mocked(useTasks);

import { useConfig } from "../hooks/use-config.ts";
const mockUseConfig = vi.mocked(useConfig);

describe("StatsPage", () => {
  beforeEach(() => {
    mockUseTasks.mockReturnValue({ data: undefined } as ReturnType<typeof useTasks>);
    mockUseConfig.mockReturnValue({ phases: [], readonly: false, version: "1.0.0", efforts: DEFAULT_EFFORTS } as ReturnType<typeof useConfig>);
  });

  it("renders loading state", () => {
    mockUseStats.mockReturnValue({
      data: undefined,
      error: undefined,
      isLoading: true,
      mutate: vi.fn(),
      isValidating: false,
    });
    render(<StatsPage />);
    expect(screen.getByRole("status", { name: /loading/i })).toBeInTheDocument();
  });

  it("renders error state", () => {
    mockUseStats.mockReturnValue({
      data: undefined,
      error: new Error("Server error"),
      isLoading: false,
      mutate: vi.fn(),
      isValidating: false,
    });
    render(<StatsPage />);
    expect(screen.getByText("Something went wrong")).toBeInTheDocument();
  });

  it("renders empty state when total_tasks is 0", () => {
    mockUseStats.mockReturnValue({
      data: createStats({ total_tasks: 0, tasks_by_status: {}, tasks_by_priority: {}, tasks_by_effort: {}, blocked_tasks_count: 0, critical_path_length: 0, max_dependency_depth: 0, avg_dependencies_per_task: 0, tags_by_count: [] }),
      error: undefined,
      isLoading: false,
      mutate: vi.fn(),
      isValidating: false,
    });
    render(<StatsPage />);
    expect(screen.getByText(/No tasks found/)).toBeInTheDocument();
  });

  it("renders StatsView when data is available", () => {
    mockUseStats.mockReturnValue({
      data: createStats({ total_tasks: 42 }),
      error: undefined,
      isLoading: false,
      mutate: vi.fn(),
      isValidating: false,
    });
    render(<StatsPage />);
    expect(screen.getByTestId("stats-view")).toBeInTheDocument();
    expect(screen.getByText("Stats: 42 tasks")).toBeInTheDocument();
  });

  it("calls mutate when retry is clicked in error state", () => {
    const mockMutate = vi.fn();
    mockUseStats.mockReturnValue({
      data: undefined,
      error: new Error("Server error"),
      isLoading: false,
      mutate: mockMutate,
      isValidating: false,
    });
    render(<StatsPage />);
    fireEvent.click(screen.getByText("Retry"));
    expect(mockMutate).toHaveBeenCalled();
  });

  it("returns null when data is undefined", () => {
    mockUseStats.mockReturnValue({
      data: undefined,
      error: undefined,
      isLoading: false,
      mutate: vi.fn(),
      isValidating: false,
    });
    const { container } = render(<StatsPage />);
    expect(container.innerHTML).toBe("");
  });

  it("computes phaseProgress when phases and tasks are available", () => {
    mockUseConfig.mockReturnValue({
      phases: [
        { id: "mvp", name: "MVP", description: "" },
        { id: "v2", name: "V2", description: "" },
      ],
      readonly: false,
      version: "1.0.0",
      efforts: DEFAULT_EFFORTS,
    } as ReturnType<typeof useConfig>);
    mockUseTasks.mockReturnValue({
      data: [
        createTask({ phase: "mvp", status: "completed" }),
        createTask({ phase: "mvp", status: "pending" }),
        createTask({ phase: "v2", status: "pending" }),
      ],
    } as ReturnType<typeof useTasks>);
    mockUseStats.mockReturnValue({
      data: createStats({ total_tasks: 3 }),
      error: undefined,
      isLoading: false,
      mutate: vi.fn(),
      isValidating: false,
    });
    render(<StatsPage />);
    expect(screen.getByTestId("stats-view")).toBeInTheDocument();
  });
});
