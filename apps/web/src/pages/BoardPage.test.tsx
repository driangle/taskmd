import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, fireEvent, waitFor } from "@testing-library/react";
import { BoardPage } from "./BoardPage.tsx";
import {
  createBoardTask,
  createBoardGroup,
  renderWithProviders,
} from "../test-utils/index.ts";

const mockGroups = [
  createBoardGroup({
    group: "pending",
    tasks: [
      createBoardTask({ id: "001", title: "Task A", status: "pending", priority: "high", effort: "small", type: "feature", tags: ["backend", "api"] }),
      createBoardTask({ id: "002", title: "Task B", status: "pending", priority: "low", effort: "large", type: "bug", tags: ["frontend"] }),
    ],
  }),
  createBoardGroup({
    group: "in-progress",
    tasks: [
      createBoardTask({ id: "003", title: "Task C", status: "in-progress", priority: "medium", effort: "medium", type: "chore", tags: ["api"] }),
    ],
  }),
];

let mockBoardData: typeof mockGroups | undefined = mockGroups;
let mockBoardError: Error | undefined;
let mockBoardLoading = false;
let mockPhases: { id: string; name: string; description: string }[] = [];
let mockReadonly = false;
const mockMutate = vi.fn();

vi.mock("../hooks/use-board.ts", () => ({
  useBoard: () => ({
    data: mockBoardData,
    error: mockBoardError,
    isLoading: mockBoardLoading,
    mutate: mockMutate,
  }),
}));

vi.mock("../hooks/use-phase.tsx", () => ({
  usePhase: () => ({ phase: null }),
}));

vi.mock("../hooks/use-project.ts", () => ({
  useProject: () => ({ project: null }),
}));

vi.mock("../hooks/use-config.ts", () => ({
  useConfig: () => ({ readonly: mockReadonly, phases: mockPhases, efforts: ["small", "medium", "large"] }),
}));

vi.mock("../api/client.ts", () => ({
  updateTask: vi.fn(),
}));

import { updateTask } from "../api/client.ts";
const mockUpdateTask = vi.mocked(updateTask);

function renderPage(initialEntries: string[] = ["/"]) {
  return renderWithProviders(<BoardPage />, { initialEntries });
}

describe("BoardPage", () => {
  beforeEach(() => {
    mockBoardData = mockGroups;
    mockBoardError = undefined;
    mockBoardLoading = false;
    mockPhases = [];
    mockReadonly = false;
    mockUpdateTask.mockReset();
  });

  describe("availableTags extraction", () => {
    it("extracts unique sorted tags from all groups", () => {
      renderPage();
      expect(screen.getByText("Task A")).toBeInTheDocument();
      expect(screen.getByText("Task B")).toBeInTheDocument();
      expect(screen.getByText("Task C")).toBeInTheDocument();
    });
  });

  describe("groupBy options", () => {
    it("shows base groupBy options without phases", () => {
      renderPage();
      const select = screen.getByRole("combobox");
      const options = Array.from(select.querySelectorAll("option")).map(o => o.textContent);
      expect(options).toEqual(["status", "priority", "effort", "type", "group", "tag"]);
    });

    it("includes phase option when phases exist", () => {
      mockPhases = [{ id: "mvp", name: "MVP", description: "" }, { id: "v2", name: "V2", description: "" }];
      renderPage();
      const select = screen.getByRole("combobox");
      const options = Array.from(select.querySelectorAll("option")).map(o => o.textContent);
      expect(options).toContain("phase");
    });
  });

  describe("groupBy from URL", () => {
    it("defaults to status when no groupBy param", () => {
      renderPage(["/"]);
      const select = screen.getByRole("combobox") as HTMLSelectElement;
      expect(select.value).toBe("status");
    });

    it("reads groupBy from URL", () => {
      renderPage(["/?groupBy=priority"]);
      const select = screen.getByRole("combobox") as HTMLSelectElement;
      expect(select.value).toBe("priority");
    });

    it("falls back to status for invalid groupBy", () => {
      renderPage(["/?groupBy=invalid"]);
      const select = screen.getByRole("combobox") as HTMLSelectElement;
      expect(select.value).toBe("status");
    });
  });

  describe("loading and error states", () => {
    it("shows loading state", () => {
      mockBoardData = undefined;
      mockBoardLoading = true;
      renderPage();
      expect(screen.queryByText("Task A")).not.toBeInTheDocument();
    });

    it("shows error state", () => {
      mockBoardData = undefined;
      mockBoardError = new Error("Network error");
      renderPage();
      expect(screen.getByText(/Network error/)).toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it("shows no tasks message when all groups are empty", () => {
      mockBoardData = [];
      renderPage();
      expect(screen.getByText("No tasks to display.")).toBeInTheDocument();
    });
  });

  describe("filtering", () => {
    it("renders all tasks when no filters are changed", () => {
      renderPage();
      expect(screen.getByText("Task A")).toBeInTheDocument();
      expect(screen.getByText("Task B")).toBeInTheDocument();
      expect(screen.getByText("Task C")).toBeInTheDocument();
    });
  });

  describe("groupBy change", () => {
    it("updates select value when groupBy is changed", () => {
      renderPage();
      const select = screen.getByRole("combobox") as HTMLSelectElement;
      fireEvent.change(select, { target: { value: "priority" } });
      expect(select.value).toBe("priority");
    });
  });

  describe("retry", () => {
    it("calls mutate when retry is clicked in error state", () => {
      mockBoardData = undefined;
      mockBoardError = new Error("Network error");
      renderPage();
      fireEvent.click(screen.getByText("Retry"));
      expect(mockMutate).toHaveBeenCalled();
    });
  });

  describe("loading skeleton", () => {
    it("shows loading skeleton when isLoading is true", () => {
      mockBoardData = undefined;
      mockBoardLoading = true;
      renderPage();
      expect(screen.getByRole("status", { name: /loading/i })).toBeInTheDocument();
    });
  });

  describe("groupBy change to status", () => {
    it("removes groupBy param when changed to status", () => {
      renderPage(["/?groupBy=priority"]);
      const select = screen.getByRole("combobox") as HTMLSelectElement;
      fireEvent.change(select, { target: { value: "status" } });
      expect(select.value).toBe("status");
    });
  });

  describe("sibling-only write guard", () => {
    function dropTask(taskId: string, sourceGroup: string, targetColumn: HTMLElement) {
      const dropEvent = new Event("drop", { bubbles: true }) as unknown as DragEvent;
      const store: Record<string, string> = {
        "text/plain": taskId,
        "application/x-source-group": sourceGroup,
      };
      Object.defineProperty(dropEvent, "dataTransfer", {
        value: { getData: (key: string) => store[key] ?? "" },
      });
      Object.defineProperty(dropEvent, "preventDefault", { value: vi.fn() });
      fireEvent(targetColumn, dropEvent);
    }

    it("surfaces the guard error naming the worktree when a move is blocked", async () => {
      const guardMessage =
        "task 001 exists only in worktree ../agent-b (branch dnc/001/parser); run taskmd there";
      mockUpdateTask.mockRejectedValueOnce(new Error(guardMessage));
      renderPage();

      const columns = screen.getAllByRole("group");
      dropTask("001", "pending", columns[1]);

      await waitFor(() => {
        expect(
          screen.getByText(`Failed to move task 001: ${guardMessage}`),
        ).toBeInTheDocument();
      });
    });
  });
});
