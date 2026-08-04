import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { ValidateView } from "./ValidateView.tsx";
import {
  createValidationResult,
  createValidationIssue,
} from "../../test-utils/fixtures.ts";

function renderView(result: Parameters<typeof ValidateView>[0]["result"]) {
  return render(
    <MemoryRouter>
      <ValidateView result={result} />
    </MemoryRouter>,
  );
}

describe("ValidateView", () => {
  it("shows 'All tasks are valid' when there are no issues", () => {
    renderView(createValidationResult());

    expect(screen.getByText("All tasks are valid")).toBeInTheDocument();
  });

  it("renders error and warning counts in summary cards", () => {
    renderView(createValidationResult({ errors: 3, warnings: 7 }));

    expect(screen.getByText("Errors")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
    expect(screen.getByText("Warnings")).toBeInTheDocument();
    expect(screen.getByText("7")).toBeInTheDocument();
  });

  it("groups issues by file path", () => {
    const issues = [
      createValidationIssue({
        file_path: "tasks/cli/001-task.md",
        message: "First issue",
      }),
      createValidationIssue({
        file_path: "tasks/cli/001-task.md",
        message: "Second issue",
      }),
      createValidationIssue({
        file_path: "tasks/web/002-task.md",
        message: "Third issue",
      }),
    ];

    renderView(createValidationResult({ issues, errors: 3, warnings: 0 }));

    expect(screen.getByText("tasks/cli/001-task.md")).toBeInTheDocument();
    expect(screen.getByText("tasks/web/002-task.md")).toBeInTheDocument();
    expect(screen.getByText("First issue")).toBeInTheDocument();
    expect(screen.getByText("Second issue")).toBeInTheDocument();
    expect(screen.getByText("Third issue")).toBeInTheDocument();
  });

  it("renders task_id as a link", () => {
    const issues = [
      createValidationIssue({ task_id: "042", message: "Bad field" }),
    ];

    renderView(createValidationResult({ issues, errors: 1, warnings: 0 }));

    const link = screen.getByRole("link", { name: "042" });
    expect(link).toBeInTheDocument();
    expect(link).toHaveAttribute("href", "/tasks/042");
  });

  it("labels each issue with its severity level", () => {
    const issues = [
      createValidationIssue({
        level: "error",
        task_id: "001",
        message: "Error message",
      }),
      createValidationIssue({
        level: "warning",
        task_id: "002",
        message: "Warning message",
      }),
    ];

    renderView(createValidationResult({ issues, errors: 1, warnings: 1 }));

    expect(screen.getByLabelText("error")).toBeInTheDocument();
    expect(screen.getByLabelText("warning")).toBeInTheDocument();
  });

  it("groups issues without file_path under (general)", () => {
    const issues = [
      createValidationIssue({
        file_path: undefined,
        message: "General issue",
      }),
    ];

    renderView(createValidationResult({ issues, errors: 1, warnings: 0 }));

    expect(screen.getByText("(general)")).toBeInTheDocument();
    expect(screen.getByText("General issue")).toBeInTheDocument();
  });

  it("shows multiple file headings when issues span different files", () => {
    const files = [
      "tasks/alpha/010-a.md",
      "tasks/beta/020-b.md",
      "tasks/gamma/030-c.md",
    ];

    const issues = files.map((file_path) =>
      createValidationIssue({ file_path, message: `Issue in ${file_path}` }),
    );

    renderView(createValidationResult({ issues, errors: 3, warnings: 0 }));

    for (const file of files) {
      expect(screen.getByText(file)).toBeInTheDocument();
    }
  });
});
