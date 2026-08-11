import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { TaskDetailPage } from "./TaskDetailPage.tsx";
import { ApiRequestError } from "../api/client.ts";
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
const mockUpdateTask = vi.fn();

vi.mock("../hooks/use-task-detail.ts", () => ({
  useTaskDetail: () => ({
    data: mockTask,
    error: mockError,
    isLoading: mockLoading,
    mutate: mockMutate,
  }),
}));

vi.mock("../hooks/use-worklog.ts", () => ({
  useWorklog: () => ({ data: undefined }),
}));

vi.mock("../hooks/use-config.ts", () => ({
  useConfig: () => ({ readonly: false, efforts: ["small", "medium", "large"] }),
}));

vi.mock("../hooks/use-project.ts", () => ({
  useProject: () => ({ project: null }),
}));

vi.mock("../api/client.ts", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api/client.ts")>();
  return {
    ...actual,
    updateTask: (...args: unknown[]) => mockUpdateTask(...args),
  };
});

function renderPage() {
  return render(
    <MemoryRouter initialEntries={["/tasks/042"]}>
      <Routes>
        <Route path="/tasks/:id" element={<TaskDetailPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("TaskDetailPage edit flow", () => {
  beforeEach(() => {
    mockTask = makeTask();
    mockError = undefined;
    mockLoading = false;
    mockMutate.mockReset();
    mockUpdateTask.mockReset();
  });

  it("enters edit mode when Edit is clicked", async () => {
    renderPage();
    await userEvent.click(screen.getByText("Edit"));
    expect(screen.getByRole("button", { name: "Save" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();
  });

  it("save handler calls updateTask and mutate on success", async () => {
    const updatedTask = makeTask({ title: "Updated title" });
    mockUpdateTask.mockResolvedValueOnce(updatedTask);
    renderPage();
    await userEvent.click(screen.getByText("Edit"));
    const titleInput = screen.getByDisplayValue("Test task");
    await userEvent.clear(titleInput);
    await userEvent.type(titleInput, "Updated title");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => {
      expect(mockUpdateTask).toHaveBeenCalledWith("042", { title: "Updated title" });
    });
    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(updatedTask, false);
    });
  });

  it("displays ApiRequestError with details", async () => {
    mockUpdateTask.mockRejectedValueOnce(
      new ApiRequestError("Validation failed", ["title is required", "status is invalid"]),
    );
    renderPage();
    await userEvent.click(screen.getByText("Edit"));
    const titleInput = screen.getByDisplayValue("Test task");
    await userEvent.clear(titleInput);
    await userEvent.type(titleInput, "X");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => {
      expect(screen.getByText("Validation failed: title is required, status is invalid")).toBeInTheDocument();
    });
  });

  it("displays ApiRequestError without details", async () => {
    mockUpdateTask.mockRejectedValueOnce(new ApiRequestError("Server error"));
    renderPage();
    await userEvent.click(screen.getByText("Edit"));
    const titleInput = screen.getByDisplayValue("Test task");
    await userEvent.clear(titleInput);
    await userEvent.type(titleInput, "Y");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => {
      expect(screen.getByText("Server error")).toBeInTheDocument();
    });
  });

  it("displays generic error for non-ApiRequestError", async () => {
    mockUpdateTask.mockRejectedValueOnce(new Error("boom"));
    renderPage();
    await userEvent.click(screen.getByText("Edit"));
    const titleInput = screen.getByDisplayValue("Test task");
    await userEvent.clear(titleInput);
    await userEvent.type(titleInput, "Z");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => {
      expect(screen.getByText("An unexpected error occurred.")).toBeInTheDocument();
    });
  });

  it("Escape key cancels editing and clears error", async () => {
    mockUpdateTask.mockRejectedValueOnce(new ApiRequestError("Err"));
    renderPage();
    await userEvent.click(screen.getByText("Edit"));
    const titleInput = screen.getByDisplayValue("Test task");
    await userEvent.clear(titleInput);
    await userEvent.type(titleInput, "W");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => {
      expect(screen.getByText("Err")).toBeInTheDocument();
    });
    await userEvent.keyboard("{Escape}");
    await waitFor(() => {
      expect(screen.getByText("Edit")).toBeInTheDocument();
    });
    expect(screen.queryByText("Err")).not.toBeInTheDocument();
  });
});
