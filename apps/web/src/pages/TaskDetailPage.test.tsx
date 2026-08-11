import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { TaskDetailPage } from "./TaskDetailPage.tsx";
import type { Task } from "../api/types.ts";

function makeTask(overrides: Partial<Task> = {}): Task {
  return {
    id: "042",
    title: "Test task",
    status: "pending",
    priority: "high",
    effort: "medium",
    type: "feature",
    dependencies: null,
    tags: ["backend"],
    phase: "mvp",
    group: "cli",
    owner: "",
    parent: "",
    created: "2026-01-01",
    body: "Task body content",
    file_path: "tasks/042-test.md",
    ...overrides,
  };
}

let mockTask: Task | undefined = makeTask();
let mockError: Error | undefined;
let mockLoading = false;
const mockMutate = vi.fn();

vi.mock("../hooks/use-task-detail.ts", () => ({
  useTaskDetail: () => ({
    data: mockTask,
    error: mockError,
    isLoading: mockLoading,
    mutate: mockMutate,
  }),
}));

let mockWorklogData: import("../api/types.ts").WorklogEntry[] | undefined;

vi.mock("../hooks/use-worklog.ts", () => ({
  useWorklog: () => ({
    data: mockWorklogData,
  }),
}));

vi.mock("../hooks/use-config.ts", () => ({
  useConfig: () => ({ readonly: false, efforts: ["small", "medium", "large"] }),
}));

vi.mock("../hooks/use-project.ts", () => ({
  useProject: () => ({ project: null }),
}));


function renderPage() {
  return render(
    <MemoryRouter initialEntries={["/tasks/042"]}>
      <Routes>
        <Route path="/tasks/:id" element={<TaskDetailPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("TaskDetailPage", () => {
  beforeEach(() => {
    mockTask = makeTask();
    mockError = undefined;
    mockLoading = false;
    mockMutate.mockReset();
    mockWorklogData = undefined;
  });

  it("renders task details", () => {
    renderPage();
    expect(screen.getByText("Test task")).toBeInTheDocument();
    expect(screen.getByText("042")).toBeInTheDocument();
    expect(screen.getByText("high")).toBeInTheDocument();
  });

  it("shows task not found when task is undefined", () => {
    mockTask = undefined;
    renderPage();
    expect(screen.getByText(/Task not found/)).toBeInTheDocument();
  });

  it("shows loading state", () => {
    mockLoading = true;
    mockTask = undefined;
    renderPage();
    expect(screen.getByRole("status", { name: /loading/i })).toBeInTheDocument();
  });

  it("shows error state with retry", () => {
    mockError = new Error("Network error");
    mockTask = undefined;
    renderPage();
    expect(screen.getByText("Something went wrong")).toBeInTheDocument();
    fireEvent.click(screen.getByText("Retry"));
    expect(mockMutate).toHaveBeenCalled();
  });

  it("renders task dependencies as links", () => {
    mockTask = makeTask({ dependencies: ["010", "020"] });
    renderPage();
    expect(screen.getByText("Dependencies")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "010" })).toHaveAttribute("href", "/tasks/010");
    expect(screen.getByRole("link", { name: "020" })).toHaveAttribute("href", "/tasks/020");
  });

  it("does not render dependencies section when empty", () => {
    mockTask = makeTask({ dependencies: null });
    renderPage();
    expect(screen.queryByText("Dependencies")).not.toBeInTheDocument();
  });

  it("renders task tags as links", () => {
    mockTask = makeTask({ tags: ["backend", "api"] });
    renderPage();
    expect(screen.getByText("Tags")).toBeInTheDocument();
    expect(screen.getByText("backend")).toBeInTheDocument();
    expect(screen.getByText("api")).toBeInTheDocument();
  });

  it("does not render tags section when empty", () => {
    mockTask = makeTask({ tags: null });
    renderPage();
    expect(screen.queryByText("Tags")).not.toBeInTheDocument();
  });

  it("renders task body as markdown", () => {
    mockTask = makeTask({ body: "**Bold text** content" });
    renderPage();
    expect(screen.getByText("Bold text")).toBeInTheDocument();
  });

  it("does not render body section when empty", () => {
    mockTask = makeTask({ body: "" });
    renderPage();
    expect(screen.queryByRole("region", { name: "Task description" })).not.toBeInTheDocument();
  });

  it("renders file_path when present", () => {
    mockTask = makeTask({ file_path: "tasks/042-test.md" });
    renderPage();
    expect(screen.getByText("tasks/042-test.md")).toBeInTheDocument();
  });

  it("renders parent link when present", () => {
    mockTask = makeTask({ parent: "001" });
    renderPage();
    expect(screen.getByText("Parent")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "001" })).toHaveAttribute("href", "/tasks/001");
  });

  it("renders effort, phase, owner, group, created fields", () => {
    mockTask = makeTask({ effort: "large", phase: "mvp", owner: "alice", group: "web", created: "2026-03-01" });
    renderPage();
    expect(screen.getByText("large")).toBeInTheDocument();
    expect(screen.getByText("alice")).toBeInTheDocument();
    expect(screen.getByText("web")).toBeInTheDocument();
    expect(screen.getByText("2026-03-01")).toBeInTheDocument();
  });

  it("renders worklog section when entries exist", () => {
    mockWorklogData = [
      { timestamp: "2026-01-15T10:30:00Z", content: "Started work" },
    ];
    renderPage();
    expect(screen.getByText("Started work")).toBeInTheDocument();
  });
});
