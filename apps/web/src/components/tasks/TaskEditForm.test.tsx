import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TaskEditForm } from "./TaskEditForm.tsx";
import type { Task } from "../../api/types.ts";

function makeTask(overrides: Partial<Task> = {}): Task {
  return {
    id: "001",
    title: "Test task",
    status: "pending",
    priority: "medium",
    effort: "small",
    type: "feature",
    dependencies: null,
    tags: ["backend", "api"],
    phase: "",
    group: "cli",
    owner: "alice",
    parent: "010",
    created: "2026-01-01",
    body: "Some body text",
    file_path: "tasks/cli/001-test.md",
    ...overrides,
  };
}

describe("TaskEditForm", () => {
  it("renders initial form state matching task props", () => {
    const task = makeTask();
    render(
      <TaskEditForm task={task} onSave={vi.fn()} onCancel={vi.fn()} error={null} />,
    );

    expect(screen.getByDisplayValue("Test task")).toBeInTheDocument();
    expect(screen.getByDisplayValue("pending")).toBeInTheDocument();
    expect(screen.getByDisplayValue("medium")).toBeInTheDocument();
    expect(screen.getByDisplayValue("small")).toBeInTheDocument();
    expect(screen.getByDisplayValue("feature")).toBeInTheDocument();
    expect(screen.getByDisplayValue("alice")).toBeInTheDocument();
    expect(screen.getByDisplayValue("010")).toBeInTheDocument();
    expect(screen.getByDisplayValue("backend, api")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Some body text")).toBeInTheDocument();
  });

  it("shows error message when error prop is provided", () => {
    render(
      <TaskEditForm task={makeTask()} onSave={vi.fn()} onCancel={vi.fn()} error="Something failed" />,
    );
    expect(screen.getByText("Something failed")).toBeInTheDocument();
  });

  it("does not show error message when error is null", () => {
    render(
      <TaskEditForm task={makeTask()} onSave={vi.fn()} onCancel={vi.fn()} error={null} />,
    );
    expect(screen.queryByText("Something failed")).not.toBeInTheDocument();
  });

  it("calls onCancel when no changes are made and form is submitted", async () => {
    const onCancel = vi.fn();
    const onSave = vi.fn();
    render(
      <TaskEditForm task={makeTask()} onSave={onSave} onCancel={onCancel} error={null} />,
    );

    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(onCancel).toHaveBeenCalledOnce();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("calls onSave with only changed fields", async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    const task = makeTask();
    render(
      <TaskEditForm task={task} onSave={onSave} onCancel={vi.fn()} error={null} />,
    );

    const titleInput = screen.getByDisplayValue("Test task");
    await userEvent.clear(titleInput);
    await userEvent.type(titleInput, "Updated title");

    await userEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(onSave).toHaveBeenCalledWith({ title: "Updated title" });
  });

  it("detects tag changes correctly", async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    render(
      <TaskEditForm task={makeTask()} onSave={onSave} onCancel={vi.fn()} error={null} />,
    );

    const tagsInput = screen.getByDisplayValue("backend, api");
    await userEvent.clear(tagsInput);
    await userEvent.type(tagsInput, "frontend, ui");

    await userEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(onSave).toHaveBeenCalledWith({ tags: ["frontend", "ui"] });
  });

  it("shows 'Saving...' and disables buttons during save", async () => {
    let resolvePromise: () => void;
    const savePromise = new Promise<void>((resolve) => {
      resolvePromise = resolve;
    });
    const onSave = vi.fn().mockReturnValue(savePromise);

    render(
      <TaskEditForm task={makeTask()} onSave={onSave} onCancel={vi.fn()} error={null} />,
    );

    // Make a change so onSave is called
    const titleInput = screen.getByDisplayValue("Test task");
    await userEvent.clear(titleInput);
    await userEvent.type(titleInput, "New");

    await userEvent.click(screen.getByRole("button", { name: "Save" }));

    // While saving
    expect(screen.getByText("Saving...")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Saving..." })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();

    // Resolve
    resolvePromise!();
    await savePromise;
  });

  it("calls onCancel when Cancel button is clicked", async () => {
    const onCancel = vi.fn();
    render(
      <TaskEditForm task={makeTask()} onSave={vi.fn()} onCancel={onCancel} error={null} />,
    );

    await userEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it("handles null tags in task", () => {
    render(
      <TaskEditForm task={makeTask({ tags: null })} onSave={vi.fn()} onCancel={vi.fn()} error={null} />,
    );
    // Tags input should be empty
    expect(screen.getByPlaceholderText("e.g. backend, api, feature")).toHaveValue("");
  });

  it("handles empty body in task", () => {
    render(
      <TaskEditForm task={makeTask({ body: "" })} onSave={vi.fn()} onCancel={vi.fn()} error={null} />,
    );
    expect(screen.getByLabelText("Body (Markdown)")).toHaveValue("");
  });
});
