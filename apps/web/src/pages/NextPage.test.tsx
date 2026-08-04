import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { NextPage } from "./NextPage.tsx";
import type { Recommendation } from "../api/types.ts";

const mockRecommendations: Recommendation[] = [
  {
    rank: 1,
    id: "001",
    title: "Test task",
    file_path: "tasks/001-test.md",
    status: "pending",
    priority: "high",
    effort: "small",
    score: 85,
    reasons: ["high priority"],
    downstream_count: 0,
    on_critical_path: false,
  },
];

let mockNextData: Recommendation[] | undefined = mockRecommendations;
let mockNextError: Error | undefined;
let mockNextLoading = false;
let lastNextArgs: { limit: number; group?: string } = { limit: 5 };

const mockMutate = vi.fn();
vi.mock("../hooks/use-next.ts", () => ({
  useNext: (limit: number, group?: string) => {
    lastNextArgs = { limit, group };
    return {
      data: mockNextData,
      error: mockNextError,
      isLoading: mockNextLoading,
      mutate: mockMutate,
      isValidating: false,
    };
  },
}));

vi.mock("../hooks/use-project.ts", () => ({
  useProject: () => ({ project: null }),
}));

vi.mock("../hooks/use-tasks.ts", () => ({
  useTasks: vi.fn(() => ({
    data: [{ group: "cli" }, { group: "web" }],
  })),
}));

function renderPage(initialEntries: string[] = ["/"]) {
  lastNextArgs = { limit: 5 };
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <NextPage />
    </MemoryRouter>,
  );
}

describe("NextPage URL sync", () => {
  beforeEach(() => {
    mockNextData = mockRecommendations;
    mockNextError = undefined;
    mockNextLoading = false;
  });

  it("reads limit from URL query string", () => {
    renderPage(["/?limit=10"]);
    expect(lastNextArgs.limit).toBe(10);
  });

  it("defaults limit to 5 when not in URL", () => {
    renderPage(["/"]);
    expect(lastNextArgs.limit).toBe(5);
  });

  it("reads group from URL query string", () => {
    renderPage(["/?group=cli"]);
    expect(lastNextArgs.group).toBe("cli");
  });

  it("passes undefined group when not in URL", () => {
    renderPage(["/"]);
    expect(lastNextArgs.group).toBeUndefined();
  });

  it("passes both limit and group from URL", () => {
    renderPage(["/?limit=3&group=web"]);
    expect(lastNextArgs.limit).toBe(3);
    expect(lastNextArgs.group).toBe("web");
  });

  it("renders the folder filter input", () => {
    renderPage(["/"]);
    expect(screen.getByLabelText("Folder:")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("All folders")).toBeInTheDocument();
  });

  it("updates group when selecting a suggestion", async () => {
    renderPage(["/"]);
    const user = userEvent.setup();
    await user.click(screen.getByLabelText("Folder:"));
    await user.click(screen.getByText("cli"));
    expect(lastNextArgs.group).toBe("cli");
  });

  it("does not update group on every keystroke", async () => {
    renderPage(["/"]);
    const user = userEvent.setup();
    await user.type(screen.getByLabelText("Folder:"), "cl");
    // Group should still be undefined since nothing was selected
    expect(lastNextArgs.group).toBeUndefined();
  });

  it("shows empty state with filter controls when no results", () => {
    mockNextData = [];
    renderPage(["/"]);
    expect(
      screen.getByText(/No actionable tasks found/),
    ).toBeInTheDocument();
    expect(screen.getByLabelText("Folder:")).toBeInTheDocument();
  });

  it("calls mutate when retry is clicked in error state", () => {
    mockNextError = new Error("Server error");
    mockNextData = undefined;
    renderPage(["/"]);
    fireEvent.click(screen.getByText("Retry"));
    expect(mockMutate).toHaveBeenCalled();
  });

  it("shows loading state", () => {
    mockNextLoading = true;
    mockNextData = undefined;
    renderPage();
    expect(screen.getByRole("status", { name: /loading/i })).toBeInTheDocument();
  });

  it("returns null when data is undefined and not loading", () => {
    mockNextData = undefined;
    mockNextError = undefined;
    mockNextLoading = false;
    renderPage();
    expect(screen.queryByLabelText("Folder:")).not.toBeInTheDocument();
  });
});
