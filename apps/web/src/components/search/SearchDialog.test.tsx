import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, fireEvent, within } from "@testing-library/react";
import { mockApi, resetMockApi } from "../../test-utils/mock-api.ts";
import { createSearchResult, resetFixtureCounter } from "../../test-utils/fixtures.ts";
import { renderWithProviders } from "../../test-utils/render.ts";
import { SearchDialog } from "./SearchDialog.tsx";

const mockNavigate = vi.fn();

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock("../../hooks/use-project.ts", () => ({
  useProject: () => mockApi.project,
}));

vi.mock("../../hooks/use-search.ts", () => ({
  useSearch: () => mockApi.search,
}));

beforeEach(() => {
  resetMockApi();
  resetFixtureCounter();
  mockNavigate.mockClear();
});

describe("SearchDialog", () => {
  it("returns null when open is false", () => {
    const { container } = renderWithProviders(
      <SearchDialog open={false} onClose={vi.fn()} />,
    );
    expect(container.innerHTML).toBe("");
  });

  it("renders dialog when open is true", () => {
    renderWithProviders(<SearchDialog open={true} onClose={vi.fn()} />);
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });

  it("shows placeholder text when no query is entered", () => {
    renderWithProviders(<SearchDialog open={true} onClose={vi.fn()} />);
    expect(screen.getByText("Start typing to search tasks...")).toBeInTheDocument();
  });

  it("shows no results message when query entered but no results", () => {
    mockApi.search = { data: [], error: undefined, isLoading: false };

    renderWithProviders(<SearchDialog open={true} onClose={vi.fn()} />);

    const input = screen.getByRole("combobox");
    fireEvent.change(input, { target: { value: "nonexistent" } });

    expect(screen.getByText(/No results found for/)).toBeInTheDocument();
  });

  it("renders search results with IDs and titles", () => {
    const results = [
      createSearchResult({ id: "001", title: "First Task", status: "pending" }),
      createSearchResult({ id: "002", title: "Second Task", status: "in-progress" }),
    ];
    mockApi.search = { data: results, error: undefined, isLoading: false };

    renderWithProviders(<SearchDialog open={true} onClose={vi.fn()} />);

    const input = screen.getByRole("combobox");
    fireEvent.change(input, { target: { value: "task" } });

    expect(screen.getByText("#001")).toBeInTheDocument();
    expect(screen.getByText("#002")).toBeInTheDocument();
    expect(screen.getByRole("listbox")).toBeInTheDocument();
    expect(screen.getAllByRole("option")).toHaveLength(2);
  });

  it("calls onClose when Escape is pressed", () => {
    const onClose = vi.fn();
    renderWithProviders(<SearchDialog open={true} onClose={onClose} />);

    const dialog = screen.getByRole("dialog");
    fireEvent.keyDown(dialog.parentElement!, { key: "Escape" });

    expect(onClose).toHaveBeenCalledOnce();
  });

  it("ArrowDown and ArrowUp navigate results with aria-selected", () => {
    const results = [
      createSearchResult({ id: "001", title: "Alpha" }),
      createSearchResult({ id: "002", title: "Beta" }),
      createSearchResult({ id: "003", title: "Gamma" }),
    ];
    mockApi.search = { data: results, error: undefined, isLoading: false };

    renderWithProviders(<SearchDialog open={true} onClose={vi.fn()} />);

    const input = screen.getByRole("combobox");
    fireEvent.change(input, { target: { value: "test" } });

    const container = screen.getByRole("dialog").parentElement!;

    // Initially no item is selected
    const options = screen.getAllByRole("option");
    expect(options[0]).toHaveAttribute("aria-selected", "false");
    expect(options[1]).toHaveAttribute("aria-selected", "false");
    expect(options[2]).toHaveAttribute("aria-selected", "false");

    // ArrowDown selects first item
    fireEvent.keyDown(container, { key: "ArrowDown" });
    expect(screen.getAllByRole("option")[0]).toHaveAttribute("aria-selected", "true");
    expect(screen.getAllByRole("option")[1]).toHaveAttribute("aria-selected", "false");

    // ArrowDown again selects second item
    fireEvent.keyDown(container, { key: "ArrowDown" });
    expect(screen.getAllByRole("option")[0]).toHaveAttribute("aria-selected", "false");
    expect(screen.getAllByRole("option")[1]).toHaveAttribute("aria-selected", "true");

    // ArrowUp goes back to first item
    fireEvent.keyDown(container, { key: "ArrowUp" });
    expect(screen.getAllByRole("option")[0]).toHaveAttribute("aria-selected", "true");
    expect(screen.getAllByRole("option")[1]).toHaveAttribute("aria-selected", "false");

    // ArrowUp from first wraps to last
    fireEvent.keyDown(container, { key: "ArrowUp" });
    expect(screen.getAllByRole("option")[2]).toHaveAttribute("aria-selected", "true");
  });

  it("Enter navigates to selected task and calls onClose", () => {
    const onClose = vi.fn();
    const results = [
      createSearchResult({ id: "042", title: "Target Task" }),
      createSearchResult({ id: "043", title: "Other Task" }),
    ];
    mockApi.search = { data: results, error: undefined, isLoading: false };

    renderWithProviders(<SearchDialog open={true} onClose={onClose} />);

    const input = screen.getByRole("combobox");
    fireEvent.change(input, { target: { value: "task" } });

    const container = screen.getByRole("dialog").parentElement!;

    // Select first item
    fireEvent.keyDown(container, { key: "ArrowDown" });
    // Press Enter
    fireEvent.keyDown(container, { key: "Enter" });

    expect(onClose).toHaveBeenCalledOnce();
    expect(mockNavigate).toHaveBeenCalledWith("/tasks/042");
  });

  it("clicking a result navigates and closes", () => {
    const onClose = vi.fn();
    const results = [
      createSearchResult({ id: "007", title: "Click Me" }),
    ];
    mockApi.search = { data: results, error: undefined, isLoading: false };

    renderWithProviders(<SearchDialog open={true} onClose={onClose} />);

    const input = screen.getByRole("combobox");
    fireEvent.change(input, { target: { value: "click" } });

    const button = within(screen.getByRole("option")).getByRole("button");
    fireEvent.click(button);

    expect(onClose).toHaveBeenCalledOnce();
    expect(mockNavigate).toHaveBeenCalledWith("/tasks/007");
  });

  it("clicking backdrop calls onClose", () => {
    const onClose = vi.fn();
    renderWithProviders(<SearchDialog open={true} onClose={onClose} />);

    // The backdrop is the outer div wrapping everything
    const backdrop = screen.getByRole("dialog").parentElement!;
    fireEvent.click(backdrop);

    expect(onClose).toHaveBeenCalledOnce();
  });

  it("clicking inside dialog does not close", () => {
    const onClose = vi.fn();
    renderWithProviders(<SearchDialog open={true} onClose={onClose} />);

    const dialog = screen.getByRole("dialog");
    fireEvent.click(dialog);

    expect(onClose).not.toHaveBeenCalled();
  });
});
